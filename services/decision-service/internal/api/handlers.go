package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/digital-twin/platform/services/decision-service/internal/audit"
	"github.com/digital-twin/platform/services/decision-service/internal/decision"
	"github.com/digital-twin/platform/services/decision-service/internal/engine"
)

type RuleEvaluator interface {
	Evaluate(ruleCode string, input map[string]any) (decision.RuleDecision, error)
	ListRules() []engine.RuleInfo
}

type AuditPublisher interface {
	PublishRuleDecision(ctx context.Context, in audit.PublishInput) error
}

type Server struct {
	evaluator RuleEvaluator
	audit     AuditPublisher
}

func NewServer(eval RuleEvaluator, pub AuditPublisher) *Server {
	if pub == nil {
		pub = audit.NoopPublisher{}
	}
	return &Server{evaluator: eval, audit: pub}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/rules", s.listRules)
	mux.HandleFunc("POST /api/v1/evaluate", s.evaluate)
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) listRules(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"rules": s.evaluator.ListRules()})
}

type evaluateRequest struct {
	RuleCode      string          `json:"ruleCode"`
	Input         json.RawMessage `json:"input"`
	CorrelationID string          `json:"correlationId"`
}

func (s *Server) evaluate(w http.ResponseWriter, r *http.Request) {
	var req evaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}
	if req.RuleCode == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "ruleCode is required")
		return
	}

	input, err := decision.ParseInput(req.Input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	result, err := s.evaluator.Evaluate(req.RuleCode, input)
	if err != nil {
		if strings.Contains(err.Error(), "unknown rule code") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if decision.RequiresAudit(result.Outcome) {
		if err := s.audit.PublishRuleDecision(r.Context(), audit.PublishInput{
			Decision:      result,
			Input:         input,
			CorrelationID: req.CorrelationID,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to publish audit event")
			return
		}
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, errCode, message string) {
	writeJSON(w, code, map[string]string{"error": message, "code": errCode})
}
