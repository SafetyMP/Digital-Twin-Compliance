package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/digital-twin/platform/services/alert-service/internal/store"
)

type AlertStore interface {
	ListAlerts(ctx context.Context, status, severity string, limit, offset int) ([]store.Alert, error)
	GetAlert(ctx context.Context, alertID string) (store.Alert, error)
	Acknowledge(ctx context.Context, alertID, acknowledgedBy string) (store.Alert, error)
}

type AlertBroadcaster interface {
	Broadcast(msgType string, alert store.Alert)
	ServeHTTP(w http.ResponseWriter, r *http.Request, initial []store.Alert)
}

type Server struct {
	store  AlertStore
	hub    AlertBroadcaster
	wsPath string
}

func NewServer(st AlertStore, h AlertBroadcaster, wsPath string) *Server {
	return &Server{store: st, hub: h, wsPath: wsPath}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/alerts", s.listAlerts)
	mux.HandleFunc("GET /api/v1/alerts/{alertId}", s.getAlert)
	mux.HandleFunc("POST /api/v1/alerts/{alertId}/acknowledge", s.acknowledge)
	mux.HandleFunc("GET "+s.wsPath, s.wsAlerts)
	return mux
}

// health is a liveness probe: it returns 200 OK to signal the process is up.
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listAlerts(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 50)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	if limit < 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	if limit > 200 {
		limit = 200
	}
	alerts, err := s.store.ListAlerts(r.Context(), r.URL.Query().Get("status"), r.URL.Query().Get("severity"), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	if alerts == nil {
		alerts = []store.Alert{}
	}
	writeJSON(w, http.StatusOK, alerts)
}

func (s *Server) getAlert(w http.ResponseWriter, r *http.Request) {
	alert, err := s.store.GetAlert(r.Context(), r.PathValue("alertId"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "alert not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

type acknowledgeRequest struct {
	AcknowledgedBy string `json:"acknowledgedBy"`
}

func (s *Server) acknowledge(w http.ResponseWriter, r *http.Request) {
	var req acknowledgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AcknowledgedBy == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "acknowledgedBy required")
		return
	}
	alert, err := s.store.Acknowledge(r.Context(), r.PathValue("alertId"), req.AcknowledgedBy)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "alert not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	s.hub.Broadcast("alert.acknowledged", alert)
	writeJSON(w, http.StatusOK, alert)
}

func (s *Server) wsAlerts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := 50
	alerts, err := s.store.ListAlerts(r.Context(), status, "", limit, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	s.hub.ServeHTTP(w, r, alerts)
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
