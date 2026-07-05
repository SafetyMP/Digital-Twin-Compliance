package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digital-twin/platform/services/graph-service/internal/graph"
)

type stubStore struct {
	summary graph.Summary
	nodes   []graph.Node
	edges   []graph.Edge
	neoErr  error
}

func (s *stubStore) VerifyConnectivity(context.Context) error { return s.neoErr }
func (s *stubStore) Summary(context.Context) (graph.Summary, error) {
	return s.summary, nil
}
func (s *stubStore) ListNodes(context.Context, string, int) ([]graph.Node, error) {
	return s.nodes, nil
}
func (s *stubStore) ListEdges(context.Context, string, int) ([]graph.Edge, error) {
	return s.edges, nil
}

func TestHealthOK(t *testing.T) {
	srv := NewServer(&stubStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHealthDegraded(t *testing.T) {
	srv := NewServer(&stubStore{neoErr: errors.New("down")})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSummary(t *testing.T) {
	srv := NewServer(&stubStore{summary: graph.Summary{NodeCount: 12, EdgeCount: 50}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/summary", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}
