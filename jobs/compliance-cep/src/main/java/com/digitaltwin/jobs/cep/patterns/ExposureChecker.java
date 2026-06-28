package com.digitaltwin.jobs.cep.patterns;

import com.digitaltwin.jobs.cep.AlertRecord;
import com.digitaltwin.jobs.cep.DecisionServiceClient;
import com.digitaltwin.jobs.cep.JobConfig;
import com.digitaltwin.jobs.cep.JsonParsers;
import com.digitaltwin.jobs.cep.RedisFeatureStore;
import com.fasterxml.jackson.databind.JsonNode;

import java.time.LocalDate;
import java.util.Map;
import java.util.Optional;
import java.util.logging.Logger;

public final class ExposureChecker {
    private static final Logger LOG = Logger.getLogger(ExposureChecker.class.getName());

    private final JobConfig config;
    private final RedisFeatureStore redis;
    private final DecisionServiceClient decisionClient;

    public ExposureChecker(JobConfig config, RedisFeatureStore redis) {
        this(config, redis, null);
    }

    public ExposureChecker(JobConfig config, RedisFeatureStore redis, DecisionServiceClient decisionClient) {
        this.config = config;
        this.redis = redis;
        this.decisionClient = decisionClient;
    }

    public Optional<AlertRecord> check(JsonParsers.TwinStateEvent twin) {
        JsonNode state = twin.currentState();
        if (state == null || state.isNull()) {
            return Optional.empty();
        }
        String ownerId = text(state, "owner_entity_id");
        String counterpartyId = text(state, "counterparty_id");
        if (ownerId.isEmpty() || counterpartyId.isEmpty()) {
            return Optional.empty();
        }
        double notional = JsonParsers.parseDouble(state, "notional_amount");
        String currency = text(state, "currency");
        double notionalEur = toEur(notional, currency);
        double total = redis.applyExposureDelta(
                twin.personaId(),
                ownerId,
                counterpartyId,
                notionalEur,
                twin.stateVersion()
        );
        return evaluateExposure(ownerId, counterpartyId, total);
    }

    Optional<AlertRecord> evaluateExposure(String ownerId, String counterpartyId, double total) {
        if (decisionClient != null) {
            return checkWithZen(ownerId, counterpartyId, total);
        }
        if (total <= config.exposureLimitEur) {
            return Optional.empty();
        }
        return Optional.of(buildAlert(ownerId, counterpartyId, total, config.exposureLimitEur, Map.of()));
    }

    private Optional<AlertRecord> checkWithZen(String ownerId, String counterpartyId, double total) {
        try {
            String outcome = decisionClient.evaluateIntExposure(
                    total, ownerId, counterpartyId, config.tenantId);
            if (!DecisionServiceClient.requiresAlert(outcome)) {
                return Optional.empty();
            }
            return Optional.of(buildAlert(
                    ownerId,
                    counterpartyId,
                    total,
                    config.exposureLimitEur,
                    Map.of("zenRuleCode", "INT-R002", "zenOutcome", outcome)
            ));
        } catch (Exception e) {
            LOG.warning("Zen INT-R002 evaluation failed; falling back to inline exposure threshold: " + e.getMessage());
            if (!shouldAlert(total, config.exposureLimitEur)) {
                return Optional.empty();
            }
            return Optional.of(buildAlert(
                    ownerId, counterpartyId, total, config.exposureLimitEur, Map.of("zenFallback", "true")));
        }
    }

    private AlertRecord buildAlert(
            String ownerId,
            String counterpartyId,
            double total,
            double threshold,
            Map<String, String> extra
    ) {
        String day = LocalDate.now().toString();
        String idempotencyKey = "INT-M002-" + ownerId + "-" + counterpartyId + "-" + day;
        var details = new java.util.LinkedHashMap<String, String>();
        details.put("institutionId", ownerId);
        details.put("counterpartyId", counterpartyId);
        details.put("totalExposureEur", Double.toString(total));
        details.put("thresholdEur", Double.toString(threshold));
        details.putAll(extra);
        return AlertRecord.create(
                "INT-M002",
                "Internal",
                "Critical",
                ownerId,
                "Institution",
                "Counterparty exposure limit exceeded",
                details,
                idempotencyKey
        );
    }

    public boolean shouldAlert(double total, double limit) {
        return total > limit;
    }

    static double toEur(double amount, String currency) {
        return switch (currency) {
            case "USD" -> amount * 0.92;
            case "GBP" -> amount * 1.17;
            default -> amount;
        };
    }

    private static String text(JsonNode node, String field) {
        JsonNode v = node.get(field);
        return v == null || v.isNull() ? "" : v.asText();
    }
}
