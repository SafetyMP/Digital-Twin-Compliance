package com.digitaltwin.jobs.cep.patterns;

import com.digitaltwin.jobs.cep.AlertRecord;
import com.digitaltwin.jobs.cep.JobConfig;
import com.digitaltwin.jobs.cep.JsonParsers;
import com.digitaltwin.jobs.cep.RedisFeatureStore;
import com.fasterxml.jackson.databind.JsonNode;

import java.time.LocalDate;
import java.util.Map;
import java.util.Optional;

public final class ExposureChecker {
    private final JobConfig config;
    private final RedisFeatureStore redis;

    public ExposureChecker(JobConfig config, RedisFeatureStore redis) {
        this.config = config;
        this.redis = redis;
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
        double notional = parseDouble(state, "notional_amount");
        String currency = text(state, "currency");
        double notionalEur = toEur(notional, currency);
        double total = redis.applyExposureDelta(
                twin.personaId(),
                ownerId,
                counterpartyId,
                notionalEur,
                twin.stateVersion()
        );
        if (total <= config.exposureLimitEur) {
            return Optional.empty();
        }
        String day = LocalDate.now().toString();
        String idempotencyKey = "INT-M002-" + ownerId + "-" + counterpartyId + "-" + day;
        return Optional.of(AlertRecord.create(
                "INT-M002",
                "Internal",
                "Critical",
                ownerId,
                "Institution",
                "Counterparty exposure limit exceeded",
                Map.of(
                        "institutionId", ownerId,
                        "counterpartyId", counterpartyId,
                        "totalExposureEur", Double.toString(total),
                        "thresholdEur", Double.toString(config.exposureLimitEur)
                ),
                idempotencyKey
        ));
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

    static double parseDouble(JsonNode node, String field) {
        JsonNode v = node.get(field);
        if (v == null || v.isNull()) {
            return 0.0;
        }
        if (v.isNumber()) {
            return v.asDouble();
        }
        String text = v.asText("");
        if (text.isEmpty()) {
            return 0.0;
        }
        try {
            return Double.parseDouble(text);
        } catch (NumberFormatException e) {
            return 0.0;
        }
    }
}
