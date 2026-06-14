package com.digitaltwin.jobs.cep.patterns;

import com.digitaltwin.jobs.cep.AlertRecord;
import com.digitaltwin.jobs.cep.JobConfig;
import com.digitaltwin.jobs.cep.JsonParsers;
import com.digitaltwin.jobs.cep.RedisFeatureStore;
import com.fasterxml.jackson.databind.JsonNode;

import java.util.Map;
import java.util.Optional;

public final class LcrChecker {
    private final JobConfig config;
    private final RedisFeatureStore redis;

    public LcrChecker(JobConfig config, RedisFeatureStore redis) {
        this.config = config;
        this.redis = redis;
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
        double lcr = liquidity.get("lcr").asDouble();
        redis.setLcr(twin.personaId(), lcr);
        if (lcr >= config.lcrMinimum) {
            return Optional.empty();
        }
        int lcrFloor = (int) Math.floor(lcr * 100);
        String idempotencyKey = "BASEL-M001-" + twin.personaId() + "-" + lcrFloor;
        return Optional.of(AlertRecord.create(
                "BASEL-M001",
                "Basel",
                "Critical",
                twin.personaId(),
                "Institution",
                "LCR below minimum threshold",
                Map.of(
                        "lcr", Double.toString(lcr),
                        "threshold", Double.toString(config.lcrMinimum),
                        "metric", "lcr"
                ),
                idempotencyKey
        ));
    }

    public boolean shouldAlert(double lcr, double minimum) {
        return lcr < minimum;
    }
}
