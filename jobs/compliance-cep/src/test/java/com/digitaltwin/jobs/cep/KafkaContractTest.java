package com.digitaltwin.jobs.cep;

import com.digitaltwin.jobs.cep.patterns.ExposureChecker;
import com.digitaltwin.jobs.cep.patterns.LcrChecker;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;

class KafkaContractTest {
    @Test
    void twinStateUpdatedInstrumentConsumer() throws Exception {
        String raw = readContract("twin.state.updated/instrument.payload.json");
        Optional<JsonParsers.TwinStateEvent> twin = JsonParsers.parseTwinState(raw);
        assertTrue(twin.isPresent());
        assertEquals("Instrument", twin.get().personaType());
        assertEquals(2, twin.get().stateVersion());
        assertEquals(
                6000000.0,
                JsonParsers.parseDouble(twin.get().currentState(), "notional_amount"),
                0.01
        );

        ExposureChecker exposure = new ExposureChecker(
                new JobConfig(java.util.Map.of("exposureLimit", "10000000")),
                null
        );
        assertFalse(exposure.check(twin.get()).isPresent(),
                "single instrument below exposure limit should not alert");
    }

    @Test
    void twinStateUpdatedInstitutionConsumer() throws Exception {
        String raw = readContract("twin.state.updated/institution.payload.json");
        Optional<JsonParsers.TwinStateEvent> twin = JsonParsers.parseTwinState(raw);
        assertTrue(twin.isPresent());
        assertEquals("Institution", twin.get().personaType());
        assertEquals(2, twin.get().stateVersion());

        LcrChecker lcr = new LcrChecker(
                new JobConfig(java.util.Map.of("lcrMinimum", "1.0")),
                null
        );
        var alert = lcr.check(twin.get());
        assertTrue(alert.isPresent(), "LCR 0.95 should raise BASEL-M001");
        assertEquals("BASEL-M001", alert.get().ruleCode());
    }

    @Test
    void paymentCreateCdcConsumer() throws Exception {
        String raw = readContract("domain.events.public.payments/payment-create.cdc.json");
        Optional<JsonParsers.PaymentEvent> payment = JsonParsers.parsePayment(raw);
        assertTrue(payment.isPresent());
        assertEquals("0abdf8bc-753c-4685-9807-5ea9aea25378", payment.get().sourceAccountId());
        assertEquals(100.0, payment.get().amount(), 0.01);
    }

    @Test
    void complianceAlertsPublisherShape() throws Exception {
        String raw = readContract("compliance.alerts/basel-alert-raised.envelope.json");
        var root = new com.fasterxml.jackson.databind.ObjectMapper().readTree(raw);
        assertEquals("ComplianceAlertRaised", root.get("eventType").asText());
        assertEquals("flink-compliance-cep", root.get("source").asText());
        assertTrue(root.get("payload").isObject());
        assertEquals("BASEL-M001", root.get("payload").get("ruleCode").asText());
        assertEquals("Institution", root.get("payload").get("personaType").asText());
    }

    private static String readContract(String rel) throws IOException {
        String path = "kafka-contracts/" + rel;
        try (InputStream in = KafkaContractTest.class.getClassLoader().getResourceAsStream(path)) {
            assertNotNull(in, "missing contract resource: " + path);
            return new String(in.readAllBytes(), StandardCharsets.UTF_8);
        }
    }
}
