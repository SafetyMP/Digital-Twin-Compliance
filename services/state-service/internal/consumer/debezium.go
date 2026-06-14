package consumer

import (
	"encoding/json"
	"fmt"
)

type DebeziumPayload struct {
	Before map[string]any `json:"before"`
	After  map[string]any `json:"after"`
	Source struct {
		Table  string `json:"table"`
		DB     string `json:"db"`
		Schema string `json:"schema"`
	} `json:"source"`
	Op string `json:"op"`
}

// ParseDebeziumMessage accepts Debezium JSON with or without Connect envelope.
func ParseDebeziumMessage(raw []byte) (DebeziumPayload, error) {
	var wrapped struct {
		Payload DebeziumPayload `json:"payload"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Payload.Op != "" {
		return wrapped.Payload, nil
	}

	var direct DebeziumPayload
	if err := json.Unmarshal(raw, &direct); err != nil {
		return DebeziumPayload{}, fmt.Errorf("parse debezium message: %w", err)
	}
	if direct.Op == "" {
		return DebeziumPayload{}, fmt.Errorf("parse debezium message: missing op field")
	}
	return direct, nil
}
