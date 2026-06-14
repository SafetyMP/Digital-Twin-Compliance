package outbox

import (
	"encoding/json"
	"fmt"

	"github.com/digital-twin/platform/services/state-service/internal/events"
	"github.com/digital-twin/platform/services/state-service/internal/store"
)

func buildTwinStateEnvelope(source string, row store.OutboxRow) (events.Envelope, []byte, error) {
	var twin events.TwinStateUpdated
	if err := json.Unmarshal(row.Payload, &twin); err != nil {
		return events.Envelope{}, nil, err
	}

	idempotencyKey := fmt.Sprintf("outbox-%d", row.ID)
	envelope, err := events.NewEnvelope("TwinStateUpdated", source, idempotencyKey, twin)
	if err != nil {
		return events.Envelope{}, nil, err
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return events.Envelope{}, nil, err
	}
	return envelope, body, nil
}
