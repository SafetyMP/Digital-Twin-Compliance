package com.digitaltwin.jobs.cep.patterns;

import com.digitaltwin.jobs.cep.AlertRecord;
import com.digitaltwin.jobs.cep.DecisionServiceClient;
import com.digitaltwin.jobs.cep.JobConfig;
import com.digitaltwin.jobs.cep.JsonParsers;
import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.Map;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;

class LcrCheckerZenTest {
    private HttpServer server;

    @AfterEach
    void tearDown() {
        if (server != null) {
            server.stop(0);
        }
    }

    @Test
    void zenDenyProducesBaselAlertWithZenMetadata() throws Exception {
        int port = startDecisionServer("""
                {"ruleCode":"BASEL-R001","outcome":"Deny"}
                """);
        JobConfig config = new JobConfig(Map.of(
                "lcrMinimum", "1.0",
                "decisionServiceUrl", "http://127.0.0.1:" + port
        ));
        LcrChecker checker = new LcrChecker(config, null, new DecisionServiceClient(config.decisionServiceUrl));
        JsonParsers.TwinStateEvent twin = institutionTwin(0.9);

        Optional<AlertRecord> alert = checker.check(twin);

        assertTrue(alert.isPresent());
        assertEquals("BASEL-M001", alert.get().ruleCode());
        assertEquals("Deny", alert.get().details().get("zenOutcome"));
        assertEquals("BASEL-R001", alert.get().details().get("zenRuleCode"));
    }

    @Test
    void zenAllowProducesNoAlert() throws Exception {
        int port = startDecisionServer("""
                {"ruleCode":"BASEL-R001","outcome":"Allow"}
                """);
        JobConfig config = new JobConfig(Map.of(
                "lcrMinimum", "1.0",
                "decisionServiceUrl", "http://127.0.0.1:" + port
        ));
        LcrChecker checker = new LcrChecker(config, null, new DecisionServiceClient(config.decisionServiceUrl));

        assertTrue(checker.check(institutionTwin(1.05)).isEmpty());
    }

    @Test
    void zenFailureFallsBackToInlineThreshold() throws Exception {
        int port = startDecisionServer(503, "unavailable");
        JobConfig config = new JobConfig(Map.of(
                "lcrMinimum", "1.0",
                "decisionServiceUrl", "http://127.0.0.1:" + port
        ));
        LcrChecker checker = new LcrChecker(config, null, new DecisionServiceClient(config.decisionServiceUrl));

        Optional<AlertRecord> alert = checker.check(institutionTwin(0.9));

        assertTrue(alert.isPresent());
        assertEquals("true", alert.get().details().get("zenFallback"));
    }

    @Test
    void requiresAlertMatchesDecisionServiceOutcomes() {
        assertTrue(LcrChecker.requiresAlert("Deny"));
        assertTrue(LcrChecker.requiresAlert("Flag"));
        assertTrue(LcrChecker.requiresAlert("Escalate"));
        assertFalse(LcrChecker.requiresAlert("Allow"));
    }

    private static JsonParsers.TwinStateEvent institutionTwin(double lcr) throws Exception {
        String raw = """
                {
                  "personaId": "44444444-4444-4444-4444-444444444401",
                  "personaType": "Institution",
                  "stateVersion": 2,
                  "currentState": {
                    "liquidity": {
                      "lcr": %s
                    }
                  }
                }
                """.formatted(lcr);
        return JsonParsers.parseTwinState(raw).orElseThrow();
    }

    private int startDecisionServer(String body) throws IOException {
        return startDecisionServer(200, body);
    }

    private int startDecisionServer(int status, String body) throws IOException {
        server = HttpServer.create(new InetSocketAddress("127.0.0.1", 0), 0);
        server.createContext("/api/v1/evaluate", exchange -> {
            byte[] bytes = body.getBytes(StandardCharsets.UTF_8);
            exchange.sendResponseHeaders(status, bytes.length);
            try (OutputStream out = exchange.getResponseBody()) {
                out.write(bytes);
            }
        });
        server.start();
        return server.getAddress().getPort();
    }
}
