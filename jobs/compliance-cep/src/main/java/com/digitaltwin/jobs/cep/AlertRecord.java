package com.digitaltwin.jobs.cep;

import java.util.Map;
import java.util.UUID;

public record AlertRecord(
        String alertId,
        String ruleCode,
        String regime,
        String severity,
        String personaId,
        String personaType,
        String summary,
        Map<String, String> details,
        String detectedAt,
        String idempotencyKey,
        String correlationId
) {
    public static AlertRecord create(
            String ruleCode,
            String regime,
            String severity,
            String personaId,
            String personaType,
            String summary,
            Map<String, String> details,
            String idempotencyKey
    ) {
        return new AlertRecord(
                UUID.randomUUID().toString(),
                ruleCode,
                regime,
                severity,
                personaId,
                personaType,
                summary,
                details,
                java.time.Instant.now().toString(),
                idempotencyKey,
                UUID.randomUUID().toString()
        );
    }
}
