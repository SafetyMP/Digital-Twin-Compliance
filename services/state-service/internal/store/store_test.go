package store

import (
	"encoding/json"
	"testing"
)

func TestStringField(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"name":   "Alpha Bank",
		"empty":  "",
		"nilVal": nil,
		"num":    float64(42),
	}
	if got := stringField(row, "name"); got != "Alpha Bank" {
		t.Fatalf("name = %q", got)
	}
	if got := stringField(row, "missing"); got != "" {
		t.Fatalf("missing = %q", got)
	}
	if got := stringField(row, "nilVal"); got != "" {
		t.Fatalf("nilVal = %q", got)
	}
	if got := stringField(row, "num"); got != "42" {
		t.Fatalf("num = %q", got)
	}
}

func TestDateField(t *testing.T) {
	t.Parallel()

	row := map[string]any{"maturity_date": "2027-12-31"}
	got := dateField(row, "maturity_date")
	if got == nil || *got != "2027-12-31" {
		t.Fatalf("date = %v", got)
	}
	if dateField(row, "missing") != nil {
		t.Fatal("expected nil for missing date")
	}
}

func TestOutboxPayloadShape(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(map[string]any{
		"personaId":        "11111111-1111-1111-1111-111111111101",
		"personaType":      "Institution",
		"sourceEntityId":   "11111111-1111-1111-1111-111111111101",
		"stateVersion":     2,
		"complianceStatus": "Unknown",
		"lastSyncedAt":     "2026-06-13T18:45:00Z",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"personaId", "personaType", "sourceEntityId", "stateVersion"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing key %q in outbox payload", key)
		}
	}
}
