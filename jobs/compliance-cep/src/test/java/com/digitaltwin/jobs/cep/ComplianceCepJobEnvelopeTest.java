package com.digitaltwin.jobs.cep;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

class ComplianceCepJobEnvelopeTest {
    private static final ObjectMapper MAPPER = new ObjectMapper();

    @Test
    void toEnvelopeJsonEmbedsPayloadObject() throws Exception {
        AlertRecord alert = AlertRecord.create(
                "INT-M001",
                "Internal",
                "Warning",
                "23cced6e-1907-4f3f-b70b-1680418a9dd7",
                "Account",
                "Transaction velocity exceeded threshold",
                Map.of("count", "51", "threshold", "50", "window", "1h"),
                "INT-M001-test-key"
        );

        JsonNode root = MAPPER.readTree(ComplianceCepJob.toEnvelopeJson(alert));
        assertEquals("ComplianceAlertRaised", root.get("eventType").asText());
        assertTrue(root.get("payload").isObject(), "payload must be a JSON object");
        assertEquals("INT-M001", root.get("payload").get("ruleCode").asText());
        assertEquals("23cced6e-1907-4f3f-b70b-1680418a9dd7", root.get("payload").get("personaId").asText());
    }
}
