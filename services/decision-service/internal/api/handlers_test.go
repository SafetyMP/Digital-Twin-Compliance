package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digital-twin/platform/services/decision-service/internal/audit"
	"github.com/digital-twin/platform/services/decision-service/internal/decision"
	"github.com/digital-twin/platform/services/decision-service/internal/engine"
)

type stubEvaluator struct {
	decision decision.RuleDecision
	err      error
	rules    []engine.RuleInfo
}

func (s stubEvaluator) Evaluate(_ string, _ map[string]any) (decision.RuleDecision, error) {
	return s.decision, s.err
}

func (s stubEvaluator) ListRules() []engine.RuleInfo {
	return s.rules
}

type recordingPublisher struct {
	called bool
	input  audit.PublishInput
}

func (r *recordingPublisher) PublishRuleDecision(_ context.Context, in audit.PublishInput) error {
	r.called = true
	r.input = in
	return nil
}

func (r *recordingPublisher) Close() error { return nil }

func TestHealth(t *testing.T) {
	srv := NewServer(stubEvaluator{}, audit.NoopPublisher{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListRules(t *testing.T) {
	srv := NewServer(stubEvaluator{rules: []engine.RuleInfo{
		{RuleCode: "INT-R001", File: "int-r001.json", Version: "1.0.0"},
	}}, audit.NoopPublisher{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rules", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string][]engine.RuleInfo
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body["rules"]) != 1 {
		t.Fatalf("rules = %v", body["rules"])
	}
}

func TestEvaluateAllowNoAudit(t *testing.T) {
	score := 0.0
	pub := &recordingPublisher{}
	srv := NewServer(stubEvaluator{decision: decision.RuleDecision{
		DecisionID: "d1",
		RuleCode:   "INT-R001",
		Outcome:    decision.OutcomeAllow,
		Score:      &score,
	}}, pub)

	body := bytes.NewBufferString(`{"ruleCode":"INT-R001","input":{"velocity":10}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", body)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if pub.called {
		t.Error("audit should not publish for Allow")
	}
}

func TestEvaluateDenyPublishesAudit(t *testing.T) {
	score := 0.95
	pub := &recordingPublisher{}
	srv := NewServer(stubEvaluator{decision: decision.RuleDecision{
		DecisionID:    "d2",
		RuleCode:      "BASEL-R001",
		Outcome:       decision.OutcomeDeny,
		Score:         &score,
		PolicyVersion: "1.0.0",
	}}, pub)

	body := bytes.NewBufferString(`{"ruleCode":"BASEL-R001","input":{"lcr":0.9,"personaId":"p1"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", body)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !pub.called {
		t.Fatal("expected audit publish for Deny")
	}
	if pub.input.Decision.DecisionID != "d2" {
		t.Fatalf("published decision = %+v", pub.input.Decision)
	}
}

func TestEvaluateUnknownRule(t *testing.T) {
	srv := NewServer(stubEvaluator{err: fmt.Errorf("unknown rule code %q", "NOPE-R001")}, audit.NoopPublisher{})
	body := bytes.NewBufferString(`{"ruleCode":"NOPE-R001","input":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", body)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}
