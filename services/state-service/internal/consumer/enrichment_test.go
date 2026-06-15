package consumer

import (
	"encoding/json"
	"testing"
)

func TestEnrichInstitutionStateDefaultLCR(t *testing.T) {
	row := map[string]any{"legal_name": "Test Bank"}
	out := enrichInstitutionState("11111111-1111-1111-1111-111111111101", row)
	liq, ok := out["liquidity"].(map[string]any)
	if !ok {
		t.Fatal("expected liquidity map")
	}
	if liq["lcr"].(float64) != 1.05 {
		t.Fatalf("expected lcr 1.05, got %v", liq["lcr"])
	}
}

func TestEnrichInstitutionStateLowLCR(t *testing.T) {
	row := map[string]any{"legal_name": "Delta Independent Bank"}
	out := enrichInstitutionState(lowLCRInstitution, row)
	liq := out["liquidity"].(map[string]any)
	if liq["lcr"].(float64) != 0.95 {
		t.Fatalf("expected lcr 0.95, got %v", liq["lcr"])
	}
}

func TestEnrichStateBytesLegalEntity(t *testing.T) {
	row := map[string]any{"legal_name": "Alpha"}
	b, err := enrichStateBytes("legal_entities", "11111111-1111-1111-1111-111111111101", row)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatal(err)
	}
	if _, ok := parsed["liquidity"]; !ok {
		t.Fatal("expected liquidity in enriched state")
	}
}
