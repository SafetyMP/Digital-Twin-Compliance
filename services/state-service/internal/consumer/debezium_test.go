package consumer

import "testing"

func TestParseDebeziumMessageWrapped(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"schema": {},
		"payload": {
			"before": null,
			"after": {
				"entity_id": "11111111-1111-1111-1111-111111111101",
				"legal_name": "Group Alpha",
				"updated_at": "2026-06-13T18:45:00Z"
			},
			"source": {"table": "legal_entities", "db": "core_banking", "schema": "public"},
			"op": "c"
		}
	}`)

	payload, err := ParseDebeziumMessage(raw)
	if err != nil {
		t.Fatalf("ParseDebeziumMessage: %v", err)
	}
	if payload.Op != "c" {
		t.Fatalf("op = %q", payload.Op)
	}
	if payload.Source.Table != "legal_entities" {
		t.Fatalf("table = %q", payload.Source.Table)
	}
	if payload.After["entity_id"] != "11111111-1111-1111-1111-111111111101" {
		t.Fatalf("unexpected after payload: %v", payload.After)
	}
}

func TestParseDebeziumMessageDirect(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"before": null,
		"after": {"account_id": "aaaa", "updated_at": 1718294700000000},
		"source": {"table": "accounts"},
		"op": "u"
	}`)

	payload, err := ParseDebeziumMessage(raw)
	if err != nil {
		t.Fatalf("ParseDebeziumMessage: %v", err)
	}
	if payload.Op != "u" {
		t.Fatalf("op = %q", payload.Op)
	}
}

func TestMapDebeziumToCDCInput(t *testing.T) {
	t.Parallel()

	payload := DebeziumPayload{
		After: map[string]any{
			"entity_id":  "11111111-1111-1111-1111-111111111101",
			"legal_name": "Group Alpha",
			"updated_at": "2026-06-13T18:45:00Z",
		},
		Op: "c",
	}
	payload.Source.Table = "legal_entities"

	input, parentID, err := MapDebeziumToCDCInput(payload)
	if err != nil {
		t.Fatalf("MapDebeziumToCDCInput: %v", err)
	}
	if input.PersonaType != "Institution" {
		t.Fatalf("personaType = %q", input.PersonaType)
	}
	if input.PersonaID != "11111111-1111-1111-1111-111111111101" {
		t.Fatalf("personaID = %q", input.PersonaID)
	}
	if parentID != "" {
		t.Fatalf("parentID = %q", parentID)
	}
	if input.IdempotencyKey == "" {
		t.Fatal("expected idempotency key")
	}
}

func TestMapDebeziumToCDCInputWithParent(t *testing.T) {
	t.Parallel()

	parent := "22222222-2222-2222-2222-222222222202"
	payload := DebeziumPayload{
		After: map[string]any{
			"entity_id":        "33333333-3333-3333-3333-333333333301",
			"legal_name":       "Branch Alpha 1a",
			"parent_entity_id": parent,
			"updated_at":       "2026-06-13T18:45:00Z",
		},
		Op: "c",
	}
	payload.Source.Table = "legal_entities"

	input, parentID, err := MapDebeziumToCDCInput(payload)
	if err != nil {
		t.Fatalf("MapDebeziumToCDCInput: %v", err)
	}
	if parentID != parent {
		t.Fatalf("parentID = %q, want %q", parentID, parent)
	}
	if input.PersonaType != "Institution" {
		t.Fatalf("personaType = %q", input.PersonaType)
	}
}

func TestMapDebeziumToCDCInputDeleteSkipped(t *testing.T) {
	t.Parallel()

	payload := DebeziumPayload{Op: "d"}
	payload.Source.Table = "legal_entities"

	input, _, err := MapDebeziumToCDCInput(payload)
	if err != nil {
		t.Fatalf("MapDebeziumToCDCInput: %v", err)
	}
	if input.PersonaID != "" {
		t.Fatal("expected empty input for delete")
	}
}
