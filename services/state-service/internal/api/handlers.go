package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/digital-twin/platform/services/state-service/internal/store"
)

type Server struct {
	store *store.Store
}

func NewServer(s *store.Store) *Server {
	return &Server{store: s}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/personas/{personaId}", s.getPersona)
	mux.HandleFunc("GET /api/v1/personas", s.listPersonas)
	return mux
}

// health returns liveness status for the service.
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getPersona(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("personaId")
	persona, err := s.store.GetPersona(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "persona not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, persona)
}

func (s *Server) listPersonas(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 50)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	if limit > 200 {
		limit = 200
	}
	personaType := r.URL.Query().Get("personaType")

	personas, err := s.store.ListPersonas(r.Context(), personaType, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	if personas == nil {
		personas = []store.TwinPersona{}
	}
	writeJSON(w, http.StatusOK, personas)
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
