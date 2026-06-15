package com.digitaltwin.jobs.cep;

import org.junit.jupiter.api.Test;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;

class JsonParsersTest {
    @Test
    void parsePaymentAcceptsDebeziumCreateEnvelope() {
        String raw = """
                {
                  "before": null,
                  "after": {
                    "payment_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
                    "source_account_id": "0abdf8bc-753c-4685-9807-5ea9aea25378",
                    "destination_account_id": "11111111-1111-1111-1111-111111111101",
                    "amount": "100.00",
                    "currency": "EUR",
                    "initiated_at": "2026-06-14T12:00:00Z"
                  },
                  "op": "c"
                }
                """;
        Optional<JsonParsers.PaymentEvent> payment = JsonParsers.parsePayment(raw);
        assertTrue(payment.isPresent());
        assertEquals("0abdf8bc-753c-4685-9807-5ea9aea25378", payment.get().sourceAccountId());
    }

    @Test
    void parsePaymentAcceptsConnectPayloadWrapper() {
        String raw = """
                {
                  "payload": {
                    "before": null,
                    "after": {
                      "payment_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
                      "source_account_id": "0abdf8bc-753c-4685-9807-5ea9aea25378",
                      "destination_account_id": "11111111-1111-1111-1111-111111111101",
                      "amount": "100.00",
                      "currency": "EUR",
                      "initiated_at": "2026-06-14T12:00:00Z"
                    },
                    "op": "c"
                  }
                }
                """;
        Optional<JsonParsers.PaymentEvent> payment = JsonParsers.parsePayment(raw);
        assertTrue(payment.isPresent());
        assertEquals("0abdf8bc-753c-4685-9807-5ea9aea25378", payment.get().sourceAccountId());
    }

    @Test
    void parsePaymentRejectsDeleteOperation() {
        String raw = """
                {"before":{},"after":null,"op":"d"}
                """;
        assertTrue(JsonParsers.parsePayment(raw).isEmpty());
    }
}
