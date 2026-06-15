package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/digital-twin/platform/services/state-service/internal/store"
)

type fakePersonaStore struct {
	getPersona   func(ctx context.Context, personaID string) (store.TwinPersona, error)
	listPersonas func(ctx context.Context, personaType string, limit, offset int) ([]store.TwinPersona, error)
}

func (f *fakePersonaStore) GetPersona(ctx context.Context, personaID string) (store.TwinPersona, error) {
	if f.getPersona != nil {
		return f.getPersona(ctx, personaID)
	}
	return store.TwinPersona{}, nil
}

func (f *fakePersonaStore) ListPersonas(ctx context.Context, personaType string, limit, offset int) ([]store.TwinPersona, error) {
	if f.listPersonas != nil {
		return f.listPersonas(ctx, personaType, limit, offset)
	}
	return nil, nil
}

func TestHealth(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	NewServer(&fakePersonaStore{}).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status = %q", body["status"])
	}
}

func TestGetPersonaNotFound(t *testing.T) {
	t.Parallel()

	srv := NewServer(&fakePersonaStore{
		getPersona: func(ctx context.Context, personaID string) (store.TwinPersona, error) {
			return store.TwinPersona{}, store.ErrNotFound
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/personas/missing-id", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetPersonaOK(t *testing.T) {
	t.Parallel()

	want := store.TwinPersona{
		PersonaID:        "11111111-1111-1111-1111-111111111101",
		SourceEntityID:   "11111111-1111-1111-1111-111111111101",
		PersonaType:      "Institution",
		StateVersion:     2,
		ComplianceStatus: "Unknown",
		LastSyncedAt:     time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC),
	}
	srv := NewServer(&fakePersonaStore{
		getPersona: func(ctx context.Context, personaID string) (store.TwinPersona, error) {
			return want, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/personas/"+want.PersonaID, nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got store.TwinPersona
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.PersonaID != want.PersonaID || got.StateVersion != want.StateVersion {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestListPersonasCapsLimitAndDefaults(t *testing.T) {
	t.Parallel()

	var gotLimit, gotOffset int
	srv := NewServer(&fakePersonaStore{
		listPersonas: func(ctx context.Context, personaType string, limit, offset int) ([]store.TwinPersona, error) {
			gotLimit = limit
			gotOffset = offset
			if personaType != "Institution" {
				t.Fatalf("personaType = %q", personaType)
			}
			return []store.TwinPersona{{PersonaID: "a"}}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/personas?personaType=Institution&limit=999&offset=bad", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if gotLimit != 200 {
		t.Fatalf("limit = %d, want 200 cap", gotLimit)
	}
	if gotOffset != 0 {
		t.Fatalf("offset = %d, want default 0", gotOffset)
	}
}

func TestListPersonasEmptySlice(t *testing.T) {
	t.Parallel()

	srv := NewServer(&fakePersonaStore{
		listPersonas: func(ctx context.Context, personaType string, limit, offset int) ([]store.TwinPersona, error) {
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/personas", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	var got []store.TwinPersona
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("expected empty JSON array, got %v", got)
	}
}

func TestListPersonasStoreError(t *testing.T) {
	t.Parallel()

	srv := NewServer(&fakePersonaStore{
		listPersonas: func(ctx context.Context, personaType string, limit, offset int) ([]store.TwinPersona, error) {
			return nil, errors.New("db down")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/personas", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestParseIntDefault(t *testing.T) {
	t.Parallel()

	if got := parseIntDefault("", 50); got != 50 {
		t.Fatalf("empty = %d", got)
	}
	if got := parseIntDefault("10", 50); got != 10 {
		t.Fatalf("valid = %d", got)
	}
	if got := parseIntDefault("nope", 50); got != 50 {
		t.Fatalf("invalid = %d", got)
	}
}
