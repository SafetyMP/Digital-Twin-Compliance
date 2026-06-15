package consumer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestKafkaContract_TwinStateUpdated_InstrumentPublisher(t *testing.T) {
	t.Parallel()

	after := loadContractJSON(t, "twin.state.updated/cdc/instrument.after.json")
	payload := DebeziumPayload{After: after, Op: "u"}
	payload.Source.Table = "instruments"

	input, _, err := MapDebeziumToCDCInput(payload)
	if err != nil {
		t.Fatalf("MapDebeziumToCDCInput: %v", err)
	}

	got, err := buildTwinStateContractPayload(input.PersonaID, input.PersonaType, input.SourceEntityID, 2, input.CurrentState)
	if err != nil {
		t.Fatal(err)
	}

	want := readContractFile(t, "twin.state.updated/instrument.payload.json")
	assertJSONEqual(t, got, want)
}

func TestKafkaContract_TwinStateUpdated_InstitutionPublisher(t *testing.T) {
	t.Parallel()

	after := loadContractJSON(t, "twin.state.updated/cdc/institution.after.json")
	payload := DebeziumPayload{After: after, Op: "u"}
	payload.Source.Table = "legal_entities"

	input, _, err := MapDebeziumToCDCInput(payload)
	if err != nil {
		t.Fatalf("MapDebeziumToCDCInput: %v", err)
	}

	got, err := buildTwinStateContractPayload(input.PersonaID, input.PersonaType, input.SourceEntityID, 2, input.CurrentState)
	if err != nil {
		t.Fatal(err)
	}

	want := readContractFile(t, "twin.state.updated/institution.payload.json")
	assertJSONEqual(t, got, want)
}

func buildTwinStateContractPayload(personaID, personaType, sourceEntityID string, stateVersion int, currentState []byte) ([]byte, error) {
	return json.Marshal(map[string]any{
		"personaId":      personaID,
		"personaType":    personaType,
		"sourceEntityId": sourceEntityID,
		"stateVersion":   stateVersion,
		"currentState":   json.RawMessage(currentState),
	})
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
	path := filepath.Join(repoRoot(t), "contracts", "kafka", filepath.FromSlash(rel))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read contract %s: %v", rel, err)
	}
	return raw
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "contracts", "kafka", "README.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found (contracts/kafka/README.md)")
		}
		dir = parent
	}
}

func assertJSONEqual(t *testing.T, got, want []byte) {
	t.Helper()
	var g any
	var w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("unmarshal got: %v", err)
	}
	if err := json.Unmarshal(want, &w); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if !reflect.DeepEqual(g, w) {
		gotPretty, _ := json.MarshalIndent(g, "", "  ")
		wantPretty, _ := json.MarshalIndent(w, "", "  ")
		t.Fatalf("contract mismatch:\n got:  %s\n want: %s", gotPretty, wantPretty)
	}
}
