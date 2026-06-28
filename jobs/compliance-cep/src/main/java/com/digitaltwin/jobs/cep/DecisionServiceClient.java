package com.digitaltwin.jobs.cep;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.logging.Logger;

/**
 * HTTP client for Decision Service Zen evaluation (Phase 3b hot path).
 */
public final class DecisionServiceClient {
    private static final ObjectMapper MAPPER = new ObjectMapper();
    private static final Logger LOG = Logger.getLogger(DecisionServiceClient.class.getName());
    private static final Duration TIMEOUT = Duration.ofSeconds(3);

    private final String baseUrl;
    private final HttpClient http;

    public DecisionServiceClient(String baseUrl) {
        this.baseUrl = trimTrailingSlash(baseUrl);
        this.http = HttpClient.newBuilder()
                .connectTimeout(TIMEOUT)
                .build();
    }

    public String evaluateBaselLcr(double lcr, String personaId, String tenantId) throws IOException, InterruptedException {
        ObjectNode input = MAPPER.createObjectNode();
        input.put("lcr", lcr);
        input.put("personaId", personaId);
        input.put("tenantId", tenantId);
        return evaluate("BASEL-R001", input);
    }

    public String evaluateIntVelocity(long velocity, String accountId, String tenantId) throws IOException, InterruptedException {
        ObjectNode input = MAPPER.createObjectNode();
        input.put("velocity", velocity);
        input.put("accountId", accountId);
        input.put("tenantId", tenantId);
        return evaluate("INT-R001", input);
    }

    public String evaluateIntExposure(
            double exposureEur,
            String ownerEntityId,
            String counterpartyId,
            String tenantId
    ) throws IOException, InterruptedException {
        ObjectNode input = MAPPER.createObjectNode();
        input.put("exposureEur", exposureEur);
        input.put("ownerEntityId", ownerEntityId);
        input.put("counterpartyId", counterpartyId);
        input.put("tenantId", tenantId);
        return evaluate("INT-R002", input);
    }

    /**
     * @return Zen outcome: Allow, Deny, Flag, or Escalate
     */
    public String evaluate(String ruleCode, ObjectNode input) throws IOException, InterruptedException {
        ObjectNode body = MAPPER.createObjectNode();
        body.put("ruleCode", ruleCode);
        body.set("input", input);

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/api/v1/evaluate"))
                .timeout(TIMEOUT)
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(MAPPER.writeValueAsString(body)))
                .build();

        HttpResponse<String> response = http.send(request, HttpResponse.BodyHandlers.ofString());
        if (response.statusCode() != 200) {
            throw new IOException("decision service returned " + response.statusCode() + ": " + response.body());
        }

        JsonNode root = MAPPER.readTree(response.body());
        JsonNode outcome = root.get("outcome");
        if (outcome == null || outcome.isNull() || outcome.asText().isBlank()) {
            throw new IOException("decision service response missing outcome");
        }
        String value = outcome.asText();
        LOG.fine(() -> "Zen " + ruleCode + " outcome=" + value + " input=" + input);
        return value;
    }

    public static boolean requiresAlert(String outcome) {
        return "Deny".equals(outcome) || "Flag".equals(outcome) || "Escalate".equals(outcome);
    }

    private static String trimTrailingSlash(String url) {
        if (url == null) {
            return "";
        }
        return url.endsWith("/") ? url.substring(0, url.length() - 1) : url;
    }
}
