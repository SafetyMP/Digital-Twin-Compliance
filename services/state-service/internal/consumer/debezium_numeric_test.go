package consumer

import (
	"strings"
	"testing"
)

func TestDecodeDebeziumDecimalBase64(t *testing.T) {
	t.Parallel()

	got := decodeDebeziumDecimal("I8NGAA==", 2)
	if got != "6000000.00" {
		t.Fatalf("decode base64 = %v, want 6000000.00", got)
	}
}

func TestDecodeDebeziumDecimalPlainString(t *testing.T) {
	t.Parallel()

	got := decodeDebeziumDecimal("1234.56", 2)
	if got != "1234.56" {
		t.Fatalf("decode plain = %v", got)
	}
}

func TestDecodeDebeziumDecimalFloat(t *testing.T) {
	t.Parallel()

	got := decodeDebeziumDecimal(float64(99.5), 2)
	if got != float64(99.5) {
		t.Fatalf("decode float = %v", got)
	}
}

func TestNormalizeDebeziumRowInstruments(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"instrument_id":   "abc",
		"notional_amount": "I8NGAA==",
		"maturity_date":   float64(21531),
		"currency":        "EUR",
	}
	out := normalizeDebeziumRow("instruments", row)
	if out["notional_amount"] != "6000000.00" {
		t.Fatalf("notional_amount = %v", out["notional_amount"])
	}
	if out["maturity_date"] != "2028-12-13" {
		t.Fatalf("maturity_date = %v", out["maturity_date"])
	}
}

func TestMapDebeziumToCDCInputInstrumentNumeric(t *testing.T) {
	t.Parallel()

	payload := DebeziumPayload{
		After: map[string]any{
			"instrument_id":    "f2e55fa0-f533-44b1-b23b-1d9c100e827a",
			"owner_entity_id":  "11111111-1111-1111-1111-111111111102",
			"counterparty_id":  "22222222-2222-2222-2222-222222222202",
			"notional_amount":  "I8NGAA==",
			"currency":         "EUR",
			"instrument_type":  "Bond",
			"isin":             "DE000TEST001",
			"maturity_date":    "2027-12-31",
			"regulatory_class": "Standard",
			"updated_at":       "2026-06-13T18:45:00Z",
		},
		Op: "u",
	}
	payload.Source.Table = "instruments"

	input, _, err := MapDebeziumToCDCInput(payload)
	if err != nil {
		t.Fatalf("MapDebeziumToCDCInput: %v", err)
	}
	if input.PersonaType != "Instrument" {
		t.Fatalf("personaType = %q", input.PersonaType)
	}
	if !strings.Contains(string(input.CurrentState), "6000000.00") {
		t.Fatalf("currentState missing decoded notional: %s", string(input.CurrentState))
	}
}
