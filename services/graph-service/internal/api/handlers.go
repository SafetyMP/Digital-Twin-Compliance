package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/digital-twin/platform/services/graph-service/internal/graph"
)

type GraphReader interface {
	VerifyConnectivity(ctx context.Context) error
	Summary(ctx context.Context) (graph.Summary, error)
	ListNodes(ctx context.Context, nameQuery string, limit int) ([]graph.Node, error)
	ListEdges(ctx context.Context, layer string, limit int) ([]graph.Edge, error)
	ShortestPath(ctx context.Context, fromID, toID string) (graph.PathResult, error)
	DegreeCentrality(ctx context.Context, limit int) ([]graph.CentralityRow, error)
}

type Server struct {
	store GraphReader
}

func NewServer(st GraphReader) *Server {
	return &Server{store: st}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/graph/summary", s.summary)
	mux.HandleFunc("GET /api/v1/graph/nodes", s.nodes)
	mux.HandleFunc("GET /api/v1/graph/edges", s.edges)
	mux.HandleFunc("GET /api/v1/graph/paths", s.paths)
	mux.HandleFunc("GET /api/v1/graph/centrality", s.centrality)
	return withOptionalInternalToken(mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if err := s.store.VerifyConnectivity(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "degraded",
			"neo4j":  "unreachable",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": status, "neo4j": "connected"})
}

func (s *Server) summary(w http.ResponseWriter, r *http.Request) {
	sum, err := s.store.Summary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

func (s *Server) nodes(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 100)
	nodes, err := s.store.ListNodes(r.Context(), r.URL.Query().Get("name"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}
	if nodes == nil {
		nodes = []graph.Node{}
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (s *Server) edges(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 500)
	edges, err := s.store.ListEdges(r.Context(), r.URL.Query().Get("layer"), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}
	if edges == nil {
		edges = []graph.Edge{}
	}
	writeJSON(w, http.StatusOK, edges)
}

func (s *Server) paths(w http.ResponseWriter, r *http.Request) {
	fromID := r.URL.Query().Get("from")
	toID := r.URL.Query().Get("to")
	if fromID == "" || toID == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "from and to query params required")
		return
	}
	path, err := s.store.ShortestPath(r.Context(), fromID, toID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no path between institutions")
		return
	}
	writeJSON(w, http.StatusOK, path)
}

func (s *Server) centrality(w http.ResponseWriter, r *http.Request) {
	limit := parseIntDefault(r.URL.Query().Get("limit"), 20)
	rows, err := s.store.DegreeCentrality(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}
	if rows == nil {
		rows = []graph.CentralityRow{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"metric": "degree",
		"rows":   rows,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"error": msg, "code": code})
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
