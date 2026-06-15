package com.digitaltwin.jobs.cep.patterns;

import com.digitaltwin.jobs.cep.AlertRecord;
import com.digitaltwin.jobs.cep.JobConfig;
import com.digitaltwin.jobs.cep.JsonParsers;
import com.digitaltwin.jobs.cep.RedisFeatureStore;

import java.util.Map;
import java.util.Optional;

public final class VelocityChecker {
    private final JobConfig config;
    private final RedisFeatureStore redis;

    public VelocityChecker(JobConfig config, RedisFeatureStore redis) {
        this.config = config;
        this.redis = redis;
    }

    public Optional<AlertRecord> check(JsonParsers.PaymentEvent payment) {
        long count = redis.incrementVelocity(payment.sourceAccountId());
        if (count <= config.velocityMaxPerHour) {
            return Optional.empty();
        }
        String windowEnd = java.time.Instant.now().truncatedTo(java.time.temporal.ChronoUnit.HOURS).toString();
        String idempotencyKey = "INT-M001-" + payment.sourceAccountId() + "-" + windowEnd;
        return Optional.of(AlertRecord.create(
                "INT-M001",
                "Internal",
                "Warning",
                payment.sourceAccountId(),
                "Account",
                "Transaction velocity exceeded threshold",
                Map.of(
                        "count", Long.toString(count),
                        "threshold", Integer.toString(config.velocityMaxPerHour),
                        "window", "1h"
                ),
                idempotencyKey
        ));
    }

    public boolean shouldAlert(long count) {
        return count > config.velocityMaxPerHour;
    }
}
