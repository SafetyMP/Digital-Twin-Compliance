package com.digitaltwin.jobs.cep;

import com.digitaltwin.jobs.cep.patterns.ExposureChecker;
import com.digitaltwin.jobs.cep.patterns.LcrChecker;
import com.digitaltwin.jobs.cep.patterns.VelocityChecker;

import java.util.Optional;

public final class PatternEngine {
    private final VelocityChecker velocity;
    private final ExposureChecker exposure;
    private final LcrChecker lcr;

    public PatternEngine(JobConfig config, RedisFeatureStore redis) {
        this(config, redis, null);
    }

    public PatternEngine(JobConfig config, RedisFeatureStore redis, DecisionServiceClient decisionClient) {
        this.velocity = new VelocityChecker(config, redis);
        this.exposure = new ExposureChecker(config, redis);
        this.lcr = new LcrChecker(config, redis, decisionClient);
    }

    public Optional<AlertRecord> onPayment(JsonParsers.PaymentEvent payment) {
        return velocity.check(payment);
    }

    public Optional<AlertRecord> onTwinState(JsonParsers.TwinStateEvent twin) {
        if ("Institution".equals(twin.personaType())) {
            return lcr.check(twin);
        }
        if ("Instrument".equals(twin.personaType())) {
            return exposure.check(twin);
        }
        return Optional.empty();
    }
}
