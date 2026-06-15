package com.digitaltwin.jobs.cep;

import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;

class TwinStateContractTest {
    @Test
    void parseInstrumentTwinStateMatchesStateServiceEnvelope() throws Exception {
        String raw = readResource("contract/instrument-twin-state.json");
        Optional<JsonParsers.TwinStateEvent> twin = JsonParsers.parseTwinState(raw);
        assertTrue(twin.isPresent());
        assertEquals("Instrument", twin.get().personaType());
        assertEquals(2, twin.get().stateVersion());
        assertEquals(
                6000000.0,
                JsonParsers.parseDouble(twin.get().currentState(), "notional_amount"),
                0.01
        );
        assertEquals(
                "11111111-1111-1111-1111-111111111102",
                JsonParsers.text(twin.get().currentState(), "owner_entity_id")
        );
    }

    @Test
    void parseTwinStateAcceptsStringStateVersion() {
        String raw = """
                {
                  "personaId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
                  "personaType": "Instrument",
                  "sourceEntityId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
                  "stateVersion": "3",
                  "currentState": {
                    "notional_amount": "6000000.00",
                    "currency": "EUR"
                  }
                }
                """;
        Optional<JsonParsers.TwinStateEvent> twin = JsonParsers.parseTwinState(raw);
        assertTrue(twin.isPresent());
        assertEquals(3, twin.get().stateVersion());
        assertEquals(6000000.0, JsonParsers.parseDouble(twin.get().currentState(), "notional_amount"), 0.01);
    }

    private static String readResource(String path) throws IOException {
        try (InputStream in = TwinStateContractTest.class.getClassLoader().getResourceAsStream(path)) {
            assertNotNull(in, "missing resource: " + path);
            return new String(in.readAllBytes(), StandardCharsets.UTF_8);
        }
    }
}
