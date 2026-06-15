package consumer

import (
	"encoding/json"
	"testing"
)

func TestEnrichInstitutionStateDefaultLCR(t *testing.T) {
	row := map[string]any{
		"legal_name":            "Test Bank",
		"lcr":                   float64(1.05),
		"hqla":                  float64(500000000),
		"net_cash_outflows_30d": float64(476190476),
		"liquidity_currency":    "EUR",
	}
	out := enrichInstitutionState(row)
	liq, ok := out["liquidity"].(map[string]any)
	if !ok {
		t.Fatal("expected liquidity map")
	}
	if liq["lcr"].(float64) != 1.05 {
		t.Fatalf("expected lcr 1.05, got %v", liq["lcr"])
	}
	if _, exists := out["lcr"]; exists {
		t.Fatal("expected flat lcr column to be omitted from currentState")
	}
}

func TestEnrichInstitutionStateLowLCR(t *testing.T) {
	row := map[string]any{
		"legal_name":            "Delta Independent Bank",
		"entity_id":             "44444444-4444-4444-4444-444444444401",
		"lcr":                   float64(0.95),
		"hqla":                  float64(450000000),
		"net_cash_outflows_30d": float64(473684211),
		"liquidity_currency":    "EUR",
	}
	out := enrichInstitutionState(row)
	liq := out["liquidity"].(map[string]any)
	if liq["lcr"].(float64) != 0.95 {
		t.Fatalf("expected lcr 0.95, got %v", liq["lcr"])
	}
}

func TestEnrichInstitutionStateMissingLiquidity(t *testing.T) {
	row := map[string]any{"legal_name": "Legacy Bank"}
	out := enrichInstitutionState(row)
	if _, ok := out["liquidity"]; ok {
		t.Fatal("expected no liquidity block when source columns are absent")
	}
}

func TestEnrichStateBytesLegalEntity(t *testing.T) {
	row := map[string]any{
		"legal_name":            "Alpha",
		"lcr":                   float64(1.05),
		"hqla":                  float64(500000000),
		"net_cash_outflows_30d": float64(476190476),
		"liquidity_currency":    "EUR",
	}
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

func TestMapDebeziumToCDCInputLegalEntityLiquidity(t *testing.T) {
	t.Parallel()

	payload := DebeziumPayload{
		After: map[string]any{
			"entity_id":             "44444444-4444-4444-4444-444444444401",
			"legal_name":            "Delta Independent Bank",
			"entity_type":           "Bank",
			"jurisdiction":          "IE",
			"consolidation_scope":   "Solo",
			"lcr":                   "0.9500",
			"hqla":                  "450000000.00",
			"net_cash_outflows_30d": "473684211.00",
			"liquidity_currency":    "EUR",
			"updated_at":            "2026-06-14T01:00:00Z",
		},
		Op: "u",
	}
	payload.Source.Table = "legal_entities"

	input, _, err := MapDebeziumToCDCInput(payload)
	if err != nil {
		t.Fatalf("MapDebeziumToCDCInput: %v", err)
	}
	var state map[string]any
	if err := json.Unmarshal(input.CurrentState, &state); err != nil {
		t.Fatal(err)
	}
	liq, ok := state["liquidity"].(map[string]any)
	if !ok {
		t.Fatalf("expected liquidity in state: %s", string(input.CurrentState))
	}
	if liq["lcr"].(float64) != 0.95 {
		t.Fatalf("lcr = %v", liq["lcr"])
	}
}

func TestEnrichInstrumentStateStringNotional(t *testing.T) {
	row := map[string]any{
		"instrument_id":    "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"owner_entity_id":  "11111111-1111-1111-1111-111111111102",
		"counterparty_id":  "22222222-2222-2222-2222-222222222202",
		"notional_amount":  "6000000.00",
		"currency":         "EUR",
		"instrument_type":  "Loan",
		"regulatory_class": "F0610",
	}
	out := enrichInstrumentState(row)
	if out["notional_amount"].(float64) != 6000000.0 {
		t.Fatalf("notional_amount = %v", out["notional_amount"])
	}
}

func TestEnrichStateBytesInstrument(t *testing.T) {
	row := map[string]any{
		"instrument_id":   "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"notional_amount": "6000000.00",
		"currency":        "EUR",
	}
	b, err := enrichStateBytes("instruments", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", row)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["notional_amount"].(float64) != 6000000.0 {
		t.Fatalf("notional_amount = %v", parsed["notional_amount"])
	}
}

func TestMapDebeziumToCDCInputInstrumentEnrichedNumeric(t *testing.T) {
	t.Parallel()

	payload := DebeziumPayload{
		After: map[string]any{
			"instrument_id":    "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"owner_entity_id":  "11111111-1111-1111-1111-111111111102",
			"counterparty_id":  "22222222-2222-2222-2222-222222222202",
			"notional_amount":  "6000000.00",
			"currency":         "EUR",
			"instrument_type":  "Loan",
			"regulatory_class": "F0610",
			"updated_at":       "2026-06-14T01:00:00Z",
		},
		Op: "u",
	}
	payload.Source.Table = "instruments"

	input, _, err := MapDebeziumToCDCInput(payload)
	if err != nil {
		t.Fatalf("MapDebeziumToCDCInput: %v", err)
	}
	var state map[string]any
	if err := json.Unmarshal(input.CurrentState, &state); err != nil {
		t.Fatal(err)
	}
	if state["notional_amount"].(float64) != 6000000.0 {
		t.Fatalf("notional_amount = %v", state["notional_amount"])
	}
}
