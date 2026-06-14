package outbox

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/digital-twin/platform/services/state-service/internal/store"
)

func TestBuildTwinStateEnvelope(t *testing.T) {
	t.Parallel()

	payload, _ := json.Marshal(map[string]any{
		"personaId":        "11111111-1111-1111-1111-111111111101",
		"personaType":      "Institution",
		"sourceEntityId":   "11111111-1111-1111-1111-111111111101",
		"stateVersion":     2,
		"complianceStatus": "Unknown",
		"lastSyncedAt":     "2026-06-13T18:45:00Z",
	})

	row := store.OutboxRow{
		ID:           42,
		Topic:        "twin.state.updated",
		PartitionKey: "11111111-1111-1111-1111-111111111101",
		Payload:      payload,
	}

	envelope, body, err := buildTwinStateEnvelope("state-service", row)
	if err != nil {
		t.Fatalf("buildTwinStateEnvelope: %v", err)
	}
	if envelope.EventType != "TwinStateUpdated" {
		t.Fatalf("eventType = %q", envelope.EventType)
	}
	if envelope.IdempotencyKey != "outbox-42" {
		t.Fatalf("idempotencyKey = %q", envelope.IdempotencyKey)
	}
	if !strings.Contains(string(body), "TwinStateUpdated") {
		t.Fatalf("body missing event type: %s", body)
	}
}
