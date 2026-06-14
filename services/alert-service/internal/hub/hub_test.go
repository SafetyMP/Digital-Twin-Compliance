package hub

import (
	"testing"
	"time"

	"github.com/digital-twin/platform/services/alert-service/internal/store"
)

func TestBroadcast_NoClients(t *testing.T) {
	t.Parallel()

	h := New()
	h.Broadcast("alert.raised", store.Alert{
		AlertID:    "550e8400-e29b-41d4-a716-446655440000",
		RuleCode:   "INT-M001",
		DetectedAt: time.Now().UTC(),
	})
}
