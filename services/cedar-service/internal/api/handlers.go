package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/digital-twin/platform/services/cedar-service/internal/audit"
	"github.com/digital-twin/platform/services/cedar-service/internal/auth"
	"github.com/digital-twin/platform/services/cedar-service/internal/decision"
	"github.com/digital-twin/platform/services/cedar-service/internal/engine"
)

type Evaluator interface {
	Status() engine.Status
	Evaluate(req engine.EvaluateRequest) (decision.RuleDecision, error)
}

type AuditPublisher interface {
	PublishDeny(ctx context.Context, d decision.RuleDecision, principalID, subjectID, correlationID string) error
}

type Server struct {
	engine Evaluator
	audit  AuditPublisher
	cfg    PrincipalDefaults
}

type PrincipalDefaults struct {
	ID    string
	Roles []string
}

func NewServer(eng Evaluator, pub AuditPublisher, defaults PrincipalDefaults) *Server {
	if pub == nil {
		pub = audit.Noop{}
	}
	return &Server{engine: eng, audit: pub, cfg: defaults}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("POST /api/v1/evaluate", s.evaluate)
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.engine.Status()
	status := "ok"
	code := http.StatusOK
	if !st.Loaded {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}
	writeJSON(w, code, map[string]any{
		"status":         status,
		"policyDir":      st.PolicyDir,
		"policyVersion":  st.PolicyVersion,
		"schemaLoaded":   st.SchemaLoaded,
		"policiesLoaded": st.PoliciesLoaded,
		"ruleCodes":      st.RuleCodes,
	})
}

type evaluateBody struct {
	RuleCode  string               `json:"ruleCode"`
	Principal principalBody        `json:"principal"`
	Action    string               `json:"action"`
	Resource  engine.ResourceInput `json:"resource"`
	Context   map[string]any       `json:"context"`
}

// principalBody tracks whether roles were present in JSON so empty [] is not replaced by dev defaults.
type principalBody struct {
	ID            string
	Roles         []string
	rolesExplicit bool
}

func (p *principalBody) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	type alias struct {
		ID    string   `json:"id"`
		Roles []string `json:"roles"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	p.ID = a.ID
	p.Roles = a.Roles
	_, p.rolesExplicit = raw["roles"]
	return nil
}

func (s *Server) evaluate(w http.ResponseWriter, r *http.Request) {
	var body evaluateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}

	principalID, roles, err := auth.PrincipalFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}
	body.Principal.ID = principalID
	body.Principal.Roles = roles
	body.Principal.rolesExplicit = true

	if body.Resource.Attrs == nil {
		body.Resource.Attrs = map[string]any{}
	}

	req := engine.EvaluateRequest{
		RuleCode:  body.RuleCode,
		Principal: engine.PrincipalInput{ID: body.Principal.ID, Roles: body.Principal.Roles},
		Action:    body.Action,
		Resource:  body.Resource,
		Context:   body.Context,
	}
	result, err := s.engine.Evaluate(req)
	if err != nil {
		if strings.Contains(err.Error(), "unknown rule") {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if result.Outcome == "Deny" {
		correlationID := r.Header.Get("X-Correlation-Id")
		if err := s.audit.PublishDeny(r.Context(), result, body.Principal.ID, body.Resource.ID, correlationID); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func headerOrDefault(v, def string) string {
	if v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": message, "code": code})
}
