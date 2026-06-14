package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

const maxInstitutionDepth = 3

// ValidateInstitutionDepth ensures adding an entity with parentEntityID
// does not exceed the 3-level hierarchy (parent → subsidiary → sub-subsidiary).
func (s *Store) ValidateInstitutionDepth(ctx context.Context, entityID, parentEntityID string) error {
	if parentEntityID == "" {
		return nil
	}
	if parentEntityID == entityID {
		return ErrHierarchyDepth
	}

	parentDepth, err := s.institutionDepth(ctx, parentEntityID)
	if errors.Is(err, ErrNotFound) {
		// Parent not yet synced (common during Debezium snapshot); allow one level.
		parentDepth = 1
	} else if err != nil {
		return err
	}

	if parentDepth+1 > maxInstitutionDepth {
		return ErrHierarchyDepth
	}
	return nil
}

func (s *Store) institutionDepth(ctx context.Context, entityID string) (int, error) {
	depth := 1
	current := entityID
	visited := map[string]struct{}{current: {}}

	for {
		parentID, err := s.parentEntityID(ctx, current)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return depth, nil
			}
			return 0, err
		}
		if parentID == "" {
			return depth, nil
		}
		if _, seen := visited[parentID]; seen {
			return 0, ErrHierarchyDepth
		}
		visited[parentID] = struct{}{}
		depth++
		if depth > maxInstitutionDepth {
			return depth, ErrHierarchyDepth
		}
		current = parentID
	}
}

func (s *Store) parentEntityID(ctx context.Context, entityID string) (string, error) {
	var state json.RawMessage
	err := s.pool.QueryRow(ctx, `
		SELECT current_state
		FROM twin_personas
		WHERE tenant_id = $1 AND source_entity_id = $2 AND persona_type = 'Institution'
	`, s.tenantID, entityID).Scan(&state)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}

	var row map[string]any
	if err := json.Unmarshal(state, &row); err != nil {
		return "", err
	}
	parent := stringField(row, "parent_entity_id")
	if parent == "" || parent == "<nil>" {
		return "", nil
	}
	return parent, nil
}

// InstitutionDepthFromChain computes depth from a root-to-leaf chain of entity IDs.
// Used by unit tests and validation helpers.
func InstitutionDepthFromChain(parentChain []string) (int, error) {
	depth := 1
	for _, parent := range parentChain {
		if parent == "" {
			break
		}
		depth++
		if depth > maxInstitutionDepth {
			return depth, ErrHierarchyDepth
		}
	}
	return depth, nil
}
