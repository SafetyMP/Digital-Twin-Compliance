package com.digitaltwin.jobs.cep.patterns;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class PatternLogicTest {
    @Test
    void velocityShouldAlertAboveThreshold() {
        VelocityChecker checker = new VelocityChecker(
                new com.digitaltwin.jobs.cep.JobConfig(java.util.Map.of("velocityMax", "50")),
                null
        );
        assertTrue(checker.shouldAlert(51));
        assertFalse(checker.shouldAlert(50));
    }

    @Test
    void exposureShouldAlertAboveLimit() {
        ExposureChecker checker = new ExposureChecker(
                new com.digitaltwin.jobs.cep.JobConfig(java.util.Map.of("exposureLimit", "10000000")),
                null
        );
        assertTrue(checker.shouldAlert(10_000_001, 10_000_000));
        assertFalse(checker.shouldAlert(9_000_000, 10_000_000));
    }

    @Test
    void lcrShouldAlertBelowMinimum() {
        LcrChecker checker = new LcrChecker(
                new com.digitaltwin.jobs.cep.JobConfig(java.util.Map.of("lcrMinimum", "1.0")),
                null
        );
        assertTrue(checker.shouldAlert(0.95, 1.0));
        assertFalse(checker.shouldAlert(1.05, 1.0));
    }

    @Test
    void exposureConvertsUsdToEur() {
        assertEquals(9200.0, ExposureChecker.toEur(10000, "USD"), 0.01);
    }

    @Test
    void exposureDeltaSkipsStaleStateVersion() {
        assertEquals(0.0, com.digitaltwin.jobs.cep.RedisFeatureStore.exposureDeltaAmount(
                5_000_000, 6_000_000, 2, 2));
    }

    @Test
    void exposureDeltaComputesNotionalChange() {
        assertEquals(1_000_000.0, com.digitaltwin.jobs.cep.RedisFeatureStore.exposureDeltaAmount(
                5_000_000, 6_000_000, 1, 2));
        assertEquals(6_000_000.0, com.digitaltwin.jobs.cep.RedisFeatureStore.exposureDeltaAmount(
                0, 6_000_000, 0, 1));
    }
}
