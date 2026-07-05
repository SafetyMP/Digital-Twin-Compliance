package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseTwinInstitutionFixture(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "contracts", "kafka", "twin.state.updated")
	raw, err := os.ReadFile(filepath.Join(root, "institution.payload.json"))
	if err != nil {
		t.Skip("fixture not found:", err)
	}
	p, err := ParseTwinPayload(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	if p.PersonaType != "Institution" {
		t.Fatalf("personaType = %q", p.PersonaType)
	}
	var inst InstitutionState
	if err := json.Unmarshal(p.CurrentState, &inst); err != nil {
		t.Fatal(err)
	}
	if inst.LegalName != "Delta Independent Bank" {
		t.Fatalf("legal_name = %q", inst.LegalName)
	}
}

func TestParseTwinInstrumentFixture(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..", "contracts", "kafka", "twin.state.updated")
	raw, err := os.ReadFile(filepath.Join(root, "instrument.payload.json"))
	if err != nil {
		t.Skip("fixture not found:", err)
	}
	p, err := ParseTwinPayload(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	if p.PersonaType != "Instrument" {
		t.Fatalf("personaType = %q", p.PersonaType)
	}
}
