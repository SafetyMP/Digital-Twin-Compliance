package chain

import (
	"encoding/json"
	"testing"

	"github.com/digital-twin/platform/services/audit-service/internal/events"
)

func TestPayloadHashDeterministic(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"ruleCode":"INT-M001","severity":"Warning"}`)
	metadata := json.RawMessage(`{"regime":"Internal","policyVersion":"phase2-inline"}`)

	h1, err := PayloadHash(payload, metadata)
	if err != nil {
		t.Fatalf("PayloadHash: %v", err)
	}
	h2, err := PayloadHash(payload, metadata)
	if err != nil {
		t.Fatalf("PayloadHash: %v", err)
	}
	if h1 != h2 {
		t.Fatalf("hashes differ: %q vs %q", h1, h2)
	}
	if len(h1) < 10 || h1[:7] != "sha256:" {
		t.Fatalf("hash format = %q", h1)
	}
}

func TestPayloadHashKeyOrderIndependent(t *testing.T) {
	t.Parallel()

	a := json.RawMessage(`{"b":2,"a":1}`)
	b := json.RawMessage(`{"a":1,"b":2}`)
	h1, err := PayloadHash(a, json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	h2, err := PayloadHash(b, json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("expected same hash for reordered keys: %q vs %q", h1, h2)
	}
}

func TestVerifyEntriesValidChain(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"alertId":"a1","ruleCode":"INT-M001"}`)
	metadata := json.RawMessage(`{"regime":"Internal"}`)
	h1, err := PayloadHash(payload, metadata)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := PayloadHash(json.RawMessage(`{"x":1}`), json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}

	entries := []events.AuditEntry{
		{
			SequenceNumber: 1,
			Payload:        payload,
			Metadata:       metadata,
			PayloadHash:    h1,
			PreviousHash:   "",
		},
		{
			SequenceNumber: 2,
			Payload:        json.RawMessage(`{"x":1}`),
			Metadata:       json.RawMessage(`{}`),
			PayloadHash:    h2,
			PreviousHash:   h1,
		},
	}

	result := VerifyEntries(entries)
	if !result.Valid {
		t.Fatalf("expected valid chain: %+v", result)
	}
	if result.CheckedCount != 2 {
		t.Fatalf("checked = %d", result.CheckedCount)
	}
}

func TestVerifyEntriesBrokenPreviousHash(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"a":1}`)
	metadata := json.RawMessage(`{}`)
	h1, err := PayloadHash(payload, metadata)
	if err != nil {
		t.Fatal(err)
	}

	entries := []events.AuditEntry{
		{SequenceNumber: 1, Payload: payload, Metadata: metadata, PayloadHash: h1, PreviousHash: ""},
		{SequenceNumber: 2, Payload: payload, Metadata: metadata, PayloadHash: h1, PreviousHash: "sha256:deadbeef"},
	}

	result := VerifyEntries(entries)
	if result.Valid {
		t.Fatal("expected broken chain")
	}
	if result.BrokenAt != 2 {
		t.Fatalf("brokenAt = %d", result.BrokenAt)
	}
}

func TestVerifyEntriesGenesisBySequenceNumber(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"a":1}`)
	metadata := json.RawMessage(`{}`)
	h1, err := PayloadHash(payload, metadata)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := PayloadHash(json.RawMessage(`{"b":2}`), metadata)
	if err != nil {
		t.Fatal(err)
	}

	entries := []events.AuditEntry{
		{
			SequenceNumber: 3,
			Payload:        json.RawMessage(`{"b":2}`),
			Metadata:       metadata,
			PayloadHash:    h2,
			PreviousHash:   h1,
		},
	}

	result := VerifyEntries(entries)
	if result.Valid {
		t.Fatal("expected previousHash mismatch when chain starts after sequence 1")
	}
	if result.BrokenAt != 3 {
		t.Fatalf("brokenAt = %d", result.BrokenAt)
	}
}

func TestVerifyEntriesTamperedPayloadHash(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"a":1}`)
	metadata := json.RawMessage(`{}`)

	entries := []events.AuditEntry{
		{
			SequenceNumber: 1,
			Payload:        payload,
			Metadata:       metadata,
			PayloadHash:    "sha256:invalid",
			PreviousHash:   "",
		},
	}

	result := VerifyEntries(entries)
	if result.Valid {
		t.Fatal("expected invalid hash")
	}
}
