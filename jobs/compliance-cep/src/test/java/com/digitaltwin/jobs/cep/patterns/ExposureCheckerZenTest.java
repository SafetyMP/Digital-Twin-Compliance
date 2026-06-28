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

class ExposureCheckerZenTest {
    private HttpServer server;

    @AfterEach
    void tearDown() {
        if (server != null) {
            server.stop(0);
        }
    }

    @Test
    void zenFlagProducesIntM002AlertWithMetadata() throws Exception {
        int port = startDecisionServer("""
                {"ruleCode":"INT-R002","outcome":"Flag"}
                """);
        JobConfig config = new JobConfig(Map.of(
                "exposureLimit", "10000000",
                "decisionServiceUrl", "http://127.0.0.1:" + port
        ));
        ExposureChecker checker = new ExposureChecker(
                config, null, new DecisionServiceClient(config.decisionServiceUrl));

        Optional<AlertRecord> alert = checker.evaluateExposure(
                "11111111-1111-1111-1111-111111111102",
                "22222222-2222-2222-2222-222222222202",
                12_500_000);

        assertTrue(alert.isPresent());
        assertEquals("INT-M002", alert.get().ruleCode());
        assertEquals("Flag", alert.get().details().get("zenOutcome"));
        assertEquals("INT-R002", alert.get().details().get("zenRuleCode"));
    }

    @Test
    void zenAllowProducesNoAlert() throws Exception {
        int port = startDecisionServer("""
                {"ruleCode":"INT-R002","outcome":"Allow"}
                """);
        JobConfig config = new JobConfig(Map.of(
                "exposureLimit", "10000000",
                "decisionServiceUrl", "http://127.0.0.1:" + port
        ));
        ExposureChecker checker = new ExposureChecker(
                config, null, new DecisionServiceClient(config.decisionServiceUrl));

        assertTrue(checker.evaluateExposure("owner", "cp", 8_500_000).isEmpty());
    }

    @Test
    void zenFailureFallsBackToInlineThreshold() throws Exception {
        int port = startDecisionServer(503, "unavailable");
        JobConfig config = new JobConfig(Map.of(
                "exposureLimit", "10000000",
                "decisionServiceUrl", "http://127.0.0.1:" + port
        ));
        ExposureChecker checker = new ExposureChecker(
                config, null, new DecisionServiceClient(config.decisionServiceUrl));

        Optional<AlertRecord> alert = checker.evaluateExposure("owner", "cp", 11_000_000);

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
