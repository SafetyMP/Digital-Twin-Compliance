package com.digitaltwin.jobs.cep.patterns;

import com.digitaltwin.jobs.cep.AlertRecord;
import com.digitaltwin.jobs.cep.DecisionServiceClient;
import com.digitaltwin.jobs.cep.JobConfig;
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

class VelocityCheckerZenTest {
    private HttpServer server;

    @AfterEach
    void tearDown() {
        if (server != null) {
            server.stop(0);
        }
    }

    @Test
    void zenFlagProducesIntM001AlertWithMetadata() throws Exception {
        int port = startDecisionServer("""
                {"ruleCode":"INT-R001","outcome":"Flag"}
                """);
        JobConfig config = new JobConfig(Map.of(
                "velocityMax", "50",
                "decisionServiceUrl", "http://127.0.0.1:" + port
        ));
        VelocityChecker checker = new VelocityChecker(
                config, null, new DecisionServiceClient(config.decisionServiceUrl));

        Optional<AlertRecord> alert = checker.evaluateCount("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", 55);

        assertTrue(alert.isPresent());
        assertEquals("INT-M001", alert.get().ruleCode());
        assertEquals("Flag", alert.get().details().get("zenOutcome"));
        assertEquals("INT-R001", alert.get().details().get("zenRuleCode"));
    }

    @Test
    void zenAllowProducesNoAlert() throws Exception {
        int port = startDecisionServer("""
                {"ruleCode":"INT-R001","outcome":"Allow"}
                """);
        JobConfig config = new JobConfig(Map.of(
                "velocityMax", "50",
                "decisionServiceUrl", "http://127.0.0.1:" + port
        ));
        VelocityChecker checker = new VelocityChecker(
                config, null, new DecisionServiceClient(config.decisionServiceUrl));

        assertTrue(checker.evaluateCount("acct-1", 40).isEmpty());
    }

    @Test
    void zenFailureFallsBackToInlineThreshold() throws Exception {
        int port = startDecisionServer(503, "unavailable");
        JobConfig config = new JobConfig(Map.of(
                "velocityMax", "50",
                "decisionServiceUrl", "http://127.0.0.1:" + port
        ));
        VelocityChecker checker = new VelocityChecker(
                config, null, new DecisionServiceClient(config.decisionServiceUrl));

        Optional<AlertRecord> alert = checker.evaluateCount("acct-1", 51);

        assertTrue(alert.isPresent());
        assertEquals("true", alert.get().details().get("zenFallback"));
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
