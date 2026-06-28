package com.digitaltwin.jobs.cep.patterns;

import com.digitaltwin.jobs.cep.AlertRecord;
import com.digitaltwin.jobs.cep.DecisionServiceClient;
import com.digitaltwin.jobs.cep.JobConfig;
import com.digitaltwin.jobs.cep.JsonParsers;
import com.digitaltwin.jobs.cep.RedisFeatureStore;
import com.fasterxml.jackson.databind.JsonNode;

import java.util.Map;
import java.util.Optional;
import java.util.logging.Logger;

public final class LcrChecker {
    private static final Logger LOG = Logger.getLogger(LcrChecker.class.getName());

    private final JobConfig config;
    private final RedisFeatureStore redis;
    private final DecisionServiceClient decisionClient;

    public LcrChecker(JobConfig config, RedisFeatureStore redis) {
        this(config, redis, null);
    }

    public LcrChecker(JobConfig config, RedisFeatureStore redis, DecisionServiceClient decisionClient) {
        this.config = config;
        this.redis = redis;
        this.decisionClient = decisionClient;
    }

    public Optional<AlertRecord> check(JsonParsers.TwinStateEvent twin) {
        JsonNode state = twin.currentState();
        if (state == null || state.isNull()) {
            return Optional.empty();
        }
        JsonNode liquidity = state.get("liquidity");
        if (liquidity == null || liquidity.isNull()) {
            return Optional.empty();
        }
        JsonNode lcrNode = liquidity.get("lcr");
        if (lcrNode == null || lcrNode.isNull()) {
            return Optional.empty();
        }
        double lcr = JsonParsers.parseDouble(liquidity, "lcr");
        if (redis != null) {
            redis.setLcr(twin.personaId(), lcr);
        }

        if (decisionClient != null) {
            return checkWithZen(twin.personaId(), lcr);
        }
        if (!shouldAlert(lcr, config.lcrMinimum)) {
            return Optional.empty();
        }
        return Optional.of(buildAlert(twin.personaId(), lcr, config.lcrMinimum, Map.of()));
    }

    private Optional<AlertRecord> checkWithZen(String personaId, double lcr) {
        try {
            String outcome = decisionClient.evaluateBaselLcr(lcr, personaId, config.tenantId);
            if (!requiresAlert(outcome)) {
                return Optional.empty();
            }
            return Optional.of(buildAlert(
                    personaId,
                    lcr,
                    config.lcrMinimum,
                    Map.of("zenRuleCode", "BASEL-R001", "zenOutcome", outcome)
            ));
        } catch (Exception e) {
            LOG.warning("Zen BASEL-R001 evaluation failed; falling back to inline LCR threshold: " + e.getMessage());
            if (!shouldAlert(lcr, config.lcrMinimum)) {
                return Optional.empty();
            }
            return Optional.of(buildAlert(personaId, lcr, config.lcrMinimum, Map.of("zenFallback", "true")));
        }
    }

    private AlertRecord buildAlert(String personaId, double lcr, double threshold, Map<String, String> extra) {
        int lcrFloor = (int) Math.floor(lcr * 100);
        String idempotencyKey = "BASEL-M001-" + personaId + "-" + lcrFloor;
        var details = new java.util.LinkedHashMap<String, String>();
        details.put("lcr", Double.toString(lcr));
        details.put("threshold", Double.toString(threshold));
        details.put("metric", "lcr");
        details.putAll(extra);
        return AlertRecord.create(
                "BASEL-M001",
                "Basel",
                "Critical",
                personaId,
                "Institution",
                "LCR below minimum threshold",
                details,
                idempotencyKey
        );
    }

    public boolean shouldAlert(double lcr, double minimum) {
        return lcr < minimum;
    }

    static boolean requiresAlert(String outcome) {
        return "Deny".equals(outcome) || "Flag".equals(outcome) || "Escalate".equals(outcome);
    }
}
