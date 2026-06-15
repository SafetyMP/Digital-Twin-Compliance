package com.digitaltwin.jobs.cep;

import java.io.Serializable;
import java.util.Map;

import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer;

public final class JobConfig implements Serializable {
    private static final long serialVersionUID = 1L;
    public final String kafkaBrokers;
    public final String redisHost;
    public final int redisPort;
    public final String tenantId;
    public final int velocityMaxPerHour;
    public final double exposureLimitEur;
    public final double lcrMinimum;
    public final int parallelism;
    public final String paymentsOffset;
    public final String twinOffset;

    public JobConfig(Map<String, String> params) {
        this.kafkaBrokers = params.getOrDefault("kafka", "kafka:9092");
        this.redisHost = params.getOrDefault("redisHost", "redis");
        this.redisPort = Integer.parseInt(params.getOrDefault("redisPort", "6379"));
        this.tenantId = params.getOrDefault("tenantId", "00000000-0000-0000-0000-000000000001");
        this.velocityMaxPerHour = Integer.parseInt(params.getOrDefault("velocityMax", "50"));
        this.exposureLimitEur = Double.parseDouble(params.getOrDefault("exposureLimit", "10000000"));
        this.lcrMinimum = Double.parseDouble(params.getOrDefault("lcrMinimum", "1.0"));
        this.parallelism = Integer.parseInt(params.getOrDefault("parallelism", "1"));
        this.paymentsOffset = params.getOrDefault("paymentsOffset", "earliest");
        this.twinOffset = params.getOrDefault("twinOffset", "earliest");
    }

    public OffsetsInitializer paymentsOffsets() {
        return offsetsInitializer(paymentsOffset);
    }

    public OffsetsInitializer twinOffsets() {
        return offsetsInitializer(twinOffset);
    }

    private static OffsetsInitializer offsetsInitializer(String mode) {
        if ("latest".equalsIgnoreCase(mode)) {
            return OffsetsInitializer.latest();
        }
        return OffsetsInitializer.earliest();
    }
}
