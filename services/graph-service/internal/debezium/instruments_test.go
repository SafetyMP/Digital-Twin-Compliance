package debezium

import "testing"

func TestStringFieldConnectUUID(t *testing.T) {
	row := map[string]any{
		"owner_entity_id": map[string]any{"string": "11111111-1111-1111-1111-111111111101"},
	}
	if got := StringField(row, "owner_entity_id"); got != "11111111-1111-1111-1111-111111111101" {
		t.Fatalf("owner = %q", got)
	}
}

func TestDecodeBase64Notional(t *testing.T) {
	row := normalizeInstrumentRow(map[string]any{
		"notional_amount": "I8NGAA==",
	})
	v := FloatField(row, "notional_amount")
	if v <= 0 {
		t.Fatalf("notional = %v", v)
	}
}
