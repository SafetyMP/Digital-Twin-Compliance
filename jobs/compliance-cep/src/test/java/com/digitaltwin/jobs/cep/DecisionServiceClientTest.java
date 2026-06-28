package com.digitaltwin.jobs.cep;

import com.sun.net.httpserver.HttpServer;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;

import static org.junit.jupiter.api.Assertions.*;

class DecisionServiceClientTest {
    private HttpServer server;

    @AfterEach
    void tearDown() {
        if (server != null) {
            server.stop(0);
        }
    }

    @Test
    void evaluateBaselLcrReturnsOutcome() throws Exception {
        int port = startServer("""
                {"ruleCode":"BASEL-R001","outcome":"Deny","rationale":"LCR below 100% minimum threshold"}
                """);
        DecisionServiceClient client = new DecisionServiceClient("http://127.0.0.1:" + port);

        String outcome = client.evaluateBaselLcr(0.9, "44444444-4444-4444-4444-444444444401", "00000000-0000-0000-0000-000000000001");

        assertEquals("Deny", outcome);
    }

    @Test
    void evaluateBaselLcrFailsOnNon200() throws Exception {
        int port = startServer(503, "service unavailable");
        DecisionServiceClient client = new DecisionServiceClient("http://127.0.0.1:" + port);

        assertThrows(IOException.class, () ->
                client.evaluateBaselLcr(0.9, "p1", "00000000-0000-0000-0000-000000000001"));
    }

    @Test
    void evaluateIntVelocityReturnsOutcome() throws Exception {
        int port = startServer("""
                {"ruleCode":"INT-R001","outcome":"Flag"}
                """);
        DecisionServiceClient client = new DecisionServiceClient("http://127.0.0.1:" + port);

        assertEquals("Flag", client.evaluateIntVelocity(
                55, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "00000000-0000-0000-0000-000000000001"));
    }

    @Test
    void evaluateIntExposureReturnsOutcome() throws Exception {
        int port = startServer("""
                {"ruleCode":"INT-R002","outcome":"Flag"}
                """);
        DecisionServiceClient client = new DecisionServiceClient("http://127.0.0.1:" + port);

        assertEquals("Flag", client.evaluateIntExposure(
                12_500_000,
                "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
                "cccccccc-cccc-cccc-cccc-cccccccccccc",
                "00000000-0000-0000-0000-000000000001"));
    }

    @Test
    void requiresAlertMatchesZenOutcomes() {
        assertTrue(DecisionServiceClient.requiresAlert("Flag"));
        assertFalse(DecisionServiceClient.requiresAlert("Allow"));
    }

    private int startServer(String body) throws IOException {
        return startServer(200, body);
    }

    private int startServer(int status, String body) throws IOException {
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
