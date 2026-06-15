package outbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/digital-twin/platform/services/state-service/internal/consumer"
	"github.com/digital-twin/platform/services/state-service/internal/store"
)

func TestKafkaContract_TwinStateUpdated_InstrumentEnvelopePublisher(t *testing.T) {
	t.Parallel()

	after := loadContractJSON(t, "twin.state.updated/cdc/instrument.after.json")
	payload := consumer.DebeziumPayload{After: after, Op: "u"}
	payload.Source.Table = "instruments"

	input, _, err := consumer.MapDebeziumToCDCInput(payload)
	if err != nil {
		t.Fatalf("MapDebeziumToCDCInput: %v", err)
	}

	inner, err := buildTwinStateContractPayload(
		input.PersonaID, input.PersonaType, input.SourceEntityID, 2, input.CurrentState,
	)
	if err != nil {
		t.Fatal(err)
	}

	row := store.OutboxRow{
		ID:           42,
		Topic:        "twin.state.updated",
		PartitionKey: input.PersonaID,
		Payload:      inner,
	}
	_, body, err := buildTwinStateEnvelope("state-service", row)
	if err != nil {
		t.Fatal(err)
	}

	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		t.Fatal(err)
	}
	if root["eventType"] != "TwinStateUpdated" {
		t.Fatalf("eventType = %v", root["eventType"])
	}
	payloadStr, ok := root["payload"].(string)
	if !ok || payloadStr == "" {
		t.Fatalf("payload must be a JSON string, got %T", root["payload"])
	}

	var innerParsed map[string]any
	if err := json.Unmarshal([]byte(payloadStr), &innerParsed); err != nil {
		t.Fatal(err)
	}
	state, ok := innerParsed["currentState"].(map[string]any)
	if !ok {
		t.Fatalf("currentState = %T", innerParsed["currentState"])
	}
	if state["owner_entity_id"] != "11111111-1111-1111-1111-111111111102" {
		t.Fatalf("owner_entity_id = %v", state["owner_entity_id"])
	}
	if state["notional_amount"].(float64) != 6000000.0 {
		t.Fatalf("notional_amount = %v", state["notional_amount"])
	}
}

func loadContractJSON(t *testing.T, rel string) map[string]any {
	t.Helper()
	raw := readContractFile(t, rel)
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal %s: %v", rel, err)
	}
	return out
}

func readContractFile(t *testing.T, rel string) []byte {
	t.Helper()
	path := contractPath(t, rel)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read contract %s: %v", rel, err)
	}
	return raw
}

func contractPath(t *testing.T, rel string) string {
	t.Helper()
	return repoRoot(t) + "/contracts/kafka/" + rel
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(dir + "/contracts/kafka/README.md"); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

func buildTwinStateContractPayload(personaID, personaType, sourceEntityID string, stateVersion int, currentState []byte) ([]byte, error) {
	return json.Marshal(map[string]any{
		"personaId":        personaID,
		"personaType":      personaType,
		"sourceEntityId":   sourceEntityID,
		"stateVersion":     stateVersion,
		"complianceStatus": "Unknown",
		"lastSyncedAt":     "2026-06-14T01:00:00Z",
		"currentState":     json.RawMessage(currentState),
	})
}
