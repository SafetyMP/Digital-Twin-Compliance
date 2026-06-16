package consumer

import (
	"fmt"

	"github.com/digital-twin/platform/services/state-service/internal/store"
)

func MapDebeziumToCDCInput(p DebeziumPayload) (store.CDCInput, string, error) {
	if p.Op == "d" {
		return store.CDCInput{}, "", nil
	}
	row := normalizeDebeziumRow(p.Source.Table, p.After)
	if row == nil {
		return store.CDCInput{}, "", nil
	}

	mapping, ok := tableMapping[p.Source.Table]
	if !ok {
		return store.CDCInput{}, "", nil
	}

	pk := stringField(row, mapping.pkColumn)
	if pk == "" {
		return store.CDCInput{}, "", nil
	}

	updatedAt, err := parseTimestamp(row["updated_at"])
	if err != nil {
		return store.CDCInput{}, "", fmt.Errorf("%s/%s: %w", p.Source.Table, pk, err)
	}
	idempotencyKey := fmt.Sprintf("%s-%s-%d", p.Source.Table, pk, updatedAt.UnixNano())

	stateBytes, err := enrichStateBytes(p.Source.Table, pk, row)
	if err != nil {
		return store.CDCInput{}, "", err
	}

	input := store.CDCInput{
		IdempotencyKey:  idempotencyKey,
		SourceTable:     p.Source.Table,
		PersonaID:       pk,
		SourceEntityID:  pk,
		PersonaType:     mapping.personaType,
		CurrentState:    stateBytes,
		SourceTimestamp: updatedAt,
	}

	parentID := ""
	if p.Source.Table == "legal_entities" {
		parentID = stringField(row, "parent_entity_id")
	}

	return input, parentID, nil
}
