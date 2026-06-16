package com.digitaltwin.jobs.cep;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.util.Optional;

public final class JsonParsers {
    private static final ObjectMapper MAPPER = new ObjectMapper();

    private JsonParsers() {}

    public static JsonNode parseRoot(String raw) throws Exception {
        JsonNode root = MAPPER.readTree(raw);
        if (root.has("payload") && root.get("payload").has("op")) {
            return root.get("payload");
        }
        return root;
    }

    public static Optional<PaymentEvent> parsePayment(String raw) {
        try {
            JsonNode node = parseRoot(raw);
            if (!"c".equals(text(node, "op")) && !"u".equals(text(node, "op"))) {
                return Optional.empty();
            }
            JsonNode after = node.get("after");
            if (after == null || after.isNull()) {
                return Optional.empty();
            }
            String sourceAccountId = text(after, "source_account_id");
            if (sourceAccountId.isEmpty()) {
                return Optional.empty();
            }
            return Optional.of(new PaymentEvent(
                    text(after, "payment_id"),
                    sourceAccountId,
                    text(after, "destination_account_id"),
                    parseDouble(after, "amount"),
                    text(after, "currency"),
                    text(after, "initiated_at")
            ));
        } catch (Exception e) {
            return Optional.empty();
        }
    }

    public static Optional<TwinStateEvent> parseTwinState(String raw) {
        try {
            JsonNode root = MAPPER.readTree(raw);
            JsonNode payload;
            if (root.has("payload")) {
                JsonNode p = root.get("payload");
                payload = p.isTextual() ? MAPPER.readTree(p.asText()) : p;
            } else {
                payload = root;
            }
            if (payload == null || payload.isNull()) {
                return Optional.empty();
            }
            String personaType = text(payload, "personaType");
            if (personaType.isEmpty()) {
                return Optional.empty();
            }
            return Optional.of(new TwinStateEvent(
                    text(payload, "personaId"),
                    personaType,
                    text(payload, "sourceEntityId"),
                    parseInt(payload, "stateVersion"),
                    payload.get("currentState")
            ));
        } catch (Exception e) {
            return Optional.empty();
        }
    }

    public static String text(JsonNode node, String field) {
        JsonNode v = node.get(field);
        return v == null || v.isNull() ? "" : v.asText();
    }

    public static double parseDouble(JsonNode node, String field) {
        JsonNode v = node.get(field);
        if (v == null || v.isNull()) {
            return 0.0;
        }
        if (v.isNumber()) {
            return v.asDouble();
        }
        String raw = v.asText("");
        if (raw.isEmpty()) {
            return 0.0;
        }
        try {
            return Double.parseDouble(raw);
        } catch (NumberFormatException e) {
            // Debezium JSON converter may emit NUMERIC as base64; velocity rules ignore amount.
            return 0.0;
        }
    }

    public static int parseInt(JsonNode node, String field) {
        JsonNode v = node.get(field);
        if (v == null || v.isNull()) {
            return 0;
        }
        if (v.isNumber()) {
            return v.asInt();
        }
        String raw = v.asText("").trim();
        if (raw.isEmpty()) {
            return 0;
        }
        try {
            return Integer.parseInt(raw);
        } catch (NumberFormatException e) {
            return 0;
        }
    }

    public record PaymentEvent(
            String paymentId,
            String sourceAccountId,
            String destinationAccountId,
            double amount,
            String currency,
            String initiatedAt
    ) {}

    public record TwinStateEvent(
            String personaId,
            String personaType,
            String sourceEntityId,
            int stateVersion,
            JsonNode currentState
    ) {}
}
