package com.digitaltwin.jobs.cep.patterns;

import com.digitaltwin.jobs.cep.AlertRecord;
import com.digitaltwin.jobs.cep.DecisionServiceClient;
import com.digitaltwin.jobs.cep.JobConfig;
import com.digitaltwin.jobs.cep.JsonParsers;
import com.digitaltwin.jobs.cep.RedisFeatureStore;

import java.util.Map;
import java.util.Optional;
import java.util.logging.Logger;

public final class VelocityChecker {
    private static final Logger LOG = Logger.getLogger(VelocityChecker.class.getName());

    private final JobConfig config;
    private final RedisFeatureStore redis;
    private final DecisionServiceClient decisionClient;

    public VelocityChecker(JobConfig config, RedisFeatureStore redis) {
        this(config, redis, null);
    }

    public VelocityChecker(JobConfig config, RedisFeatureStore redis, DecisionServiceClient decisionClient) {
        this.config = config;
        this.redis = redis;
        this.decisionClient = decisionClient;
    }

    public Optional<AlertRecord> check(JsonParsers.PaymentEvent payment) {
        long count = redis.incrementVelocity(payment.sourceAccountId());
        return evaluateCount(payment.sourceAccountId(), count);
    }

    Optional<AlertRecord> evaluateCount(String accountId, long count) {
        if (decisionClient != null) {
            return checkWithZen(accountId, count);
        }
        if (count <= config.velocityMaxPerHour) {
            return Optional.empty();
        }
        return Optional.of(buildAlert(accountId, count, config.velocityMaxPerHour, Map.of()));
    }

    private Optional<AlertRecord> checkWithZen(String accountId, long count) {
        try {
            String outcome = decisionClient.evaluateIntVelocity(count, accountId, config.tenantId);
            if (!DecisionServiceClient.requiresAlert(outcome)) {
                return Optional.empty();
            }
            return Optional.of(buildAlert(
                    accountId,
                    count,
                    config.velocityMaxPerHour,
                    Map.of("zenRuleCode", "INT-R001", "zenOutcome", outcome)
            ));
        } catch (Exception e) {
            LOG.warning("Zen INT-R001 evaluation failed; falling back to inline velocity threshold: " + e.getMessage());
            if (!shouldAlert(count)) {
                return Optional.empty();
            }
            return Optional.of(buildAlert(accountId, count, config.velocityMaxPerHour, Map.of("zenFallback", "true")));
        }
    }

    private AlertRecord buildAlert(String accountId, long count, int threshold, Map<String, String> extra) {
        String windowEnd = java.time.Instant.now().truncatedTo(java.time.temporal.ChronoUnit.HOURS).toString();
        String idempotencyKey = "INT-M001-" + accountId + "-" + windowEnd;
        var details = new java.util.LinkedHashMap<String, String>();
        details.put("count", Long.toString(count));
        details.put("threshold", Integer.toString(threshold));
        details.put("window", "1h");
        details.putAll(extra);
        return AlertRecord.create(
                "INT-M001",
                "Internal",
                "Warning",
                accountId,
                "Account",
                "Transaction velocity exceeded threshold",
                details,
                idempotencyKey
        );
    }

    public boolean shouldAlert(long count) {
        return count > config.velocityMaxPerHour;
    }
}
