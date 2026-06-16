package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/digital-twin/platform/services/audit-service/internal/chain"
	"github.com/digital-twin/platform/services/audit-service/internal/events"
	"github.com/digital-twin/platform/services/audit-service/internal/store"
)

type EntryIndexStore interface {
	ListEntries(ctx context.Context, filter store.ListFilter) ([]store.EntryIndex, error)
	GetIndex(ctx context.Context, entryID string) (store.EntryIndex, error)
	ListForVerify(ctx context.Context, fromSeq, toSeq int64) ([]store.EntryIndex, error)
}

type LedgerReader interface {
	Ping(ctx context.Context) error
	GetEntry(ctx context.Context, entryID string) (events.AuditEntry, error)
}

type Server struct {
	index  EntryIndexStore
	ledger LedgerReader
}

func NewServer(index EntryIndexStore, ledger LedgerReader) *Server {
	return &Server{index: index, ledger: ledger}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/audit/entries", s.listEntries)
	mux.HandleFunc("GET /api/v1/audit/entries/{entryId}", s.getEntry)
	mux.HandleFunc("GET /api/v1/audit/verify", s.verify)
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	code := http.StatusOK
	if s.ledger != nil {
		if err := s.ledger.Ping(r.Context()); err != nil {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
	}
	writeJSON(w, code, map[string]string{"status": status})
}

func (s *Server) listEntries(w http.ResponseWriter, r *http.Request) {
	filter := store.ListFilter{
		RuleCode:  r.URL.Query().Get("ruleCode"),
		SubjectID: r.URL.Query().Get("subjectId"),
		Limit:     parseIntDefault(r.URL.Query().Get("limit"), 50),
		Offset:    parseIntDefault(r.URL.Query().Get("offset"), 0),
	}
	if from := r.URL.Query().Get("from"); from != "" {
		t, err := parseTime(from)
		if err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid from timestamp")
			return
		}
		filter.From = &t
	}
	if to := r.URL.Query().Get("to"); to != "" {
		t, err := parseTime(to)
		if err != nil {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid to timestamp")
			return
		}
		filter.To = &t
	}

	entries, err := s.index.ListEntries(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	if entries == nil {
		entries = []store.EntryIndex{}
	}
	writeJSON(w, http.StatusOK, entries)
}

func (s *Server) getEntry(w http.ResponseWriter, r *http.Request) {
	entryID := r.PathValue("entryId")
	entry, err := s.ledger.GetEntry(r.Context(), entryID)
	if err == nil {
		writeJSON(w, http.StatusOK, entry)
		return
	}

	idx, idxErr := s.index.GetIndex(r.Context(), entryID)
	if idxErr != nil {
		if errors.Is(idxErr, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "audit entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", idxErr.Error())
		return
	}
	writeJSON(w, http.StatusOK, idx)
}

func (s *Server) verify(w http.ResponseWriter, r *http.Request) {
	fromSeq := int64(parseIntDefault(r.URL.Query().Get("fromSequence"), 0))
	toSeq := int64(parseIntDefault(r.URL.Query().Get("toSequence"), 0))

	indexRows, err := s.index.ListForVerify(r.Context(), fromSeq, toSeq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	entries := make([]events.AuditEntry, 0, len(indexRows))
	for _, idx := range indexRows {
		entry, err := s.ledger.GetEntry(r.Context(), idx.EntryID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
			return
		}
		entries = append(entries, entry)
	}

	result := chain.VerifyEntries(entries)
	status := http.StatusOK
	if !result.Valid {
		status = http.StatusConflict
	}
	writeJSON(w, status, result)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": message, "code": code})
}

func parseIntDefault(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func parseTime(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, raw)
}
