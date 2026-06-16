package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/digital-twin/platform/services/audit-service/internal/events"
)

func PayloadHash(payload, metadata json.RawMessage) (string, error) {
	canonical, err := canonicalConcat(payload, metadata)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func canonicalConcat(payload, metadata json.RawMessage) ([]byte, error) {
	payloadCanon, err := canonicalJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("canonical payload: %w", err)
	}
	metadataCanon, err := canonicalJSON(metadata)
	if err != nil {
		return nil, fmt.Errorf("canonical metadata: %w", err)
	}
	out := make([]byte, 0, len(payloadCanon)+len(metadataCanon))
	out = append(out, payloadCanon...)
	out = append(out, metadataCanon...)
	return out, nil
}

func canonicalJSON(raw json.RawMessage) ([]byte, error) {
	if len(raw) == 0 {
		return []byte("{}"), nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	normalized, err := normalizeValue(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}

func normalizeValue(v any) (any, error) {
	switch typed := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for k := range typed {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]any, len(keys))
		for _, k := range keys {
			normalized, err := normalizeValue(typed[k])
			if err != nil {
				return nil, err
			}
			out[k] = normalized
		}
		return out, nil
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			normalized, err := normalizeValue(item)
			if err != nil {
				return nil, err
			}
			out[i] = normalized
		}
		return out, nil
	default:
		return typed, nil
	}
}

type VerifyResult struct {
	Valid        bool   `json:"valid"`
	CheckedCount int    `json:"checkedCount"`
	BrokenAt     int64  `json:"brokenAt,omitempty"`
	Message      string `json:"message,omitempty"`
}

func VerifyEntries(entries []events.AuditEntry) VerifyResult {
	if len(entries) == 0 {
		return VerifyResult{Valid: true, CheckedCount: 0, Message: "empty chain"}
	}

	var expectedPrevious string
	for i, entry := range entries {
		if entry.SequenceNumber == 1 {
			if entry.PreviousHash != "" {
				return VerifyResult{
					Valid:        false,
					CheckedCount: i,
					BrokenAt:     entry.SequenceNumber,
					Message:      "genesis entry must have empty previousHash",
				}
			}
		} else if entry.PreviousHash != expectedPrevious {
			return VerifyResult{
				Valid:        false,
				CheckedCount: i,
				BrokenAt:     entry.SequenceNumber,
				Message:      fmt.Sprintf("previousHash mismatch at sequence %d", entry.SequenceNumber),
			}
		}

		computed, err := PayloadHash(entry.Payload, entry.Metadata)
		if err != nil {
			return VerifyResult{
				Valid:        false,
				CheckedCount: i,
				BrokenAt:     entry.SequenceNumber,
				Message:      fmt.Sprintf("hash compute failed: %v", err),
			}
		}
		if computed != entry.PayloadHash {
			return VerifyResult{
				Valid:        false,
				CheckedCount: i,
				BrokenAt:     entry.SequenceNumber,
				Message:      fmt.Sprintf("payloadHash mismatch at sequence %d", entry.SequenceNumber),
			}
		}
		expectedPrevious = entry.PayloadHash
	}

	return VerifyResult{Valid: true, CheckedCount: len(entries), Message: "chain intact"}
}
