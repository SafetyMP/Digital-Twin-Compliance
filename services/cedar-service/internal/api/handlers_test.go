package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digital-twin/platform/services/cedar-service/internal/auth"
	"github.com/digital-twin/platform/services/cedar-service/internal/decision"
	"github.com/digital-twin/platform/services/cedar-service/internal/engine"
)

type stubEngine struct {
	status engine.Status
	result decision.RuleDecision
	err    error
}

func (s stubEngine) Status() engine.Status {
	if s.status.Loaded || s.status.PoliciesLoaded > 0 {
		return s.status
	}
	return engine.Status{Loaded: true, PoliciesLoaded: 5}
}
func (s stubEngine) Evaluate(engine.EvaluateRequest) (decision.RuleDecision, error) {
	return s.result, s.err
}

type recordingAudit struct {
	called      bool
	principalID string
	subjectID   string
}

func (r *recordingAudit) PublishDeny(_ context.Context, d decision.RuleDecision, principalID, subjectID, _ string) error {
	r.called = true
	r.principalID = principalID
	r.subjectID = subjectID
	return nil
}

type captureEngine struct {
	capture *engine.EvaluateRequest
	status  engine.Status
}

func (c captureEngine) Status() engine.Status { return c.status }
func (c captureEngine) Evaluate(req engine.EvaluateRequest) (decision.RuleDecision, error) {
	*c.capture = req
	return decision.RuleDecision{Outcome: "Allow", RuleCode: req.RuleCode}, nil
}

func TestHealthPolicyLoadStatus(t *testing.T) {
	t.Parallel()
	srv := NewServer(stubEngine{status: engine.Status{
		Loaded:         true,
		PolicyDir:      "/policies/cedar",
		PolicyVersion:  "1.0.0",
		SchemaLoaded:   true,
		PoliciesLoaded: 5,
		RuleCodes:      []string{"INT-R003"},
	}}, nil, PrincipalDefaults{})

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["policiesLoaded"] != float64(5) {
		t.Fatalf("policiesLoaded = %#v", body["policiesLoaded"])
	}
}

func TestEvaluateUsesVerifiedJWT(t *testing.T) {
	t.Setenv("CEDAR_SERVICE_JWT_SECRET", "test-jwt-secret-value-32chars!!")
	var captured engine.EvaluateRequest
	srv := NewServer(captureEngine{
		capture: &captured,
		status:  engine.Status{Loaded: true},
	}, nil, PrincipalDefaults{ID: "fallback", Roles: []string{"Analyst"}})

	body := bytes.NewBufferString(`{"ruleCode":"INT-R003","principal":{"id":"ignored","roles":["Ignored"]},"resource":{"type":"TwinData","id":"twin-1","attrs":{"sensitivity":"low"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", body)
	token, err := auth.SignToken("test-jwt-secret-value-32chars!!", "header-user", []string{"Reporter", "Analyst"})
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
	if captured.Principal.ID != "header-user" {
		t.Fatalf("principal = %#v", captured.Principal)
	}
	if len(captured.Principal.Roles) != 2 {
		t.Fatalf("roles = %#v", captured.Principal.Roles)
	}
}

func TestEvaluateExplicitEmptyRolesSkipsDefaults(t *testing.T) {
	t.Setenv("CEDAR_SERVICE_JWT_SECRET", "test-jwt-secret-value-32chars!!")
	var captured engine.EvaluateRequest
	srv := NewServer(captureEngine{
		capture: &captured,
		status:  engine.Status{Loaded: true},
	}, nil, PrincipalDefaults{ID: "fallback", Roles: []string{"Analyst"}})

	body := bytes.NewBufferString(`{"ruleCode":"INT-R003","principal":{"id":"user-1","roles":[]},"resource":{"type":"TwinData","id":"twin-1","attrs":{"sensitivity":"high"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", body)
	token, err := auth.SignToken("test-jwt-secret-value-32chars!!", "user-1", []string{})
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
	if len(captured.Principal.Roles) != 0 {
		t.Fatalf("roles = %#v want empty slice", captured.Principal.Roles)
	}
}

func TestEvaluateDenyPublishesAudit(t *testing.T) {
	t.Setenv("CEDAR_SERVICE_JWT_SECRET", "test-jwt-secret-value-32chars!!")
	pub := &recordingAudit{}
	srv := NewServer(stubEngine{
		result: decision.RuleDecision{
			DecisionID: "dec-1",
			Outcome:    "Deny",
			RuleCode:   "INT-R003",
		},
	}, pub, PrincipalDefaults{ID: "operator-dev", Roles: []string{"Analyst"}})

	body := bytes.NewBufferString(`{"ruleCode":"INT-R003","principal":{"id":"user-1"},"resource":{"type":"TwinData","id":"twin-1","attrs":{"sensitivity":"high"}}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", body)
	token, err := auth.SignToken("test-jwt-secret-value-32chars!!", "user-1", []string{"Analyst"})
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
	}
	if !pub.called {
		t.Fatal("expected audit publish on deny")
	}
	if pub.principalID != "user-1" || pub.subjectID != "twin-1" {
		t.Fatalf("audit refs principal=%q subject=%q", pub.principalID, pub.subjectID)
	}
}
