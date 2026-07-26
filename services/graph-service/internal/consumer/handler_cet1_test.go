package consumer

import (
	"testing"

	"github.com/digital-twin/platform/services/graph-service/internal/events"
)

func TestInstitutionCET1FromCapital(t *testing.T) {
	t.Parallel()
	cet1 := 0.13
	got := institutionCET1(events.InstitutionState{
		Capital: map[string]any{"cet1_ratio": cet1},
	})
	if got != cet1 {
		t.Fatalf("got %v want %v", got, cet1)
	}
}

func TestInstitutionCET1MissingIsZero(t *testing.T) {
	t.Parallel()
	if got := institutionCET1(events.InstitutionState{}); got != 0 {
		t.Fatalf("got %v want 0", got)
	}
}
