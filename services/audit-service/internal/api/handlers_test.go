package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digital-twin/platform/services/audit-service/internal/chain"
	"github.com/digital-twin/platform/services/audit-service/internal/events"
	"github.com/digital-twin/platform/services/audit-service/internal/store"
)

type fakeIndex struct {
	verifyRows []store.EntryIndex
}

func (f *fakeIndex) ListEntries(ctx context.Context, filter store.ListFilter) ([]store.EntryIndex, error) {
	return nil, nil
}

func (f *fakeIndex) GetIndex(ctx context.Context, entryID string) (store.EntryIndex, error) {
	return store.EntryIndex{}, store.ErrNotFound
}

func (f *fakeIndex) ListForVerify(ctx context.Context, fromSeq, toSeq int64) ([]store.EntryIndex, error) {
	return f.verifyRows, nil
}

type fakeLedger struct {
	entries map[string]events.AuditEntry
	pingErr error
}

func (f *fakeLedger) Ping(ctx context.Context) error { return f.pingErr }

func (f *fakeLedger) GetEntry(ctx context.Context, entryID string) (events.AuditEntry, error) {
	e, ok := f.entries[entryID]
	if !ok {
		return events.AuditEntry{}, store.ErrNotFound
	}
	return e, nil
}

func TestHealthOK(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	NewServer(&fakeIndex{}, &fakeLedger{}).Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestVerifyValidChain(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"ruleCode":"INT-M001"}`)
	metadata := json.RawMessage(`{"regime":"Internal"}`)
	h1, err := chain.PayloadHash(payload, metadata)
	if err != nil {
		t.Fatal(err)
	}

	entry := events.AuditEntry{
		EntryID:        "e1",
		SequenceNumber: 1,
		Payload:        payload,
		Metadata:       metadata,
		PayloadHash:    h1,
		PreviousHash:   "",
	}

	srv := NewServer(&fakeIndex{
		verifyRows: []store.EntryIndex{{EntryID: "e1", SequenceNumber: 1}},
	}, &fakeLedger{entries: map[string]events.AuditEntry{"e1": entry}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/verify", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var result chain.VerifyResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Valid {
		t.Fatalf("result = %+v", result)
	}
}

func TestVerifyBrokenChain(t *testing.T) {
	t.Parallel()

	entry := events.AuditEntry{
		EntryID:        "e1",
		SequenceNumber: 1,
		Payload:        json.RawMessage(`{"a":1}`),
		Metadata:       json.RawMessage(`{}`),
		PayloadHash:    "sha256:bad",
		PreviousHash:   "",
	}

	srv := NewServer(&fakeIndex{
		verifyRows: []store.EntryIndex{{EntryID: "e1", SequenceNumber: 1}},
	}, &fakeLedger{entries: map[string]events.AuditEntry{"e1": entry}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/verify", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d", rec.Code)
	}
}
