package outbox

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/digital-twin/platform/services/state-service/internal/store"
)

func TestBuildTwinStateEnvelopeUsesOutboxRowID(t *testing.T) {
	t.Parallel()

	for _, id := range []int64{1, 99, 1234567890} {
		payload, _ := json.Marshal(map[string]any{"personaId": "abc"})
		row := store.OutboxRow{ID: id, Topic: "twin.state.updated", PartitionKey: "abc", Payload: payload}
		envelope, _, err := buildTwinStateEnvelope("state-service", row)
		if err != nil {
			t.Fatalf("id=%d: %v", id, err)
		}
		want := fmt.Sprintf("outbox-%d", id)
		if envelope.IdempotencyKey != want {
			t.Fatalf("id=%d: idempotencyKey = %q, want %q", id, envelope.IdempotencyKey, want)
		}
	}
}
