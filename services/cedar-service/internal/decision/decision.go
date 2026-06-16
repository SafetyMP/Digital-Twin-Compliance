package decision

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type RuleDecision struct {
	DecisionID    string  `json:"decisionId"`
	RuleCode      string  `json:"ruleCode"`
	Outcome       string  `json:"outcome"`
	Score         float64 `json:"score"`
	Rationale     string  `json:"rationale"`
	PolicyVersion string  `json:"policyVersion"`
	EvaluatedAt   string  `json:"evaluatedAt"`
	InputHash     string  `json:"inputHash"`
}

func InputHash(v any) string {
	raw, _ := json.Marshal(v)
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func NewDecision(ruleCode, outcome, rationale, policyVersion string, req any) RuleDecision {
	score := 0.95
	if outcome == "Allow" {
		score = 1.0
	}
	return RuleDecision{
		DecisionID:    uuid.NewString(),
		RuleCode:      ruleCode,
		Outcome:       outcome,
		Score:         score,
		Rationale:     rationale,
		PolicyVersion: policyVersion,
		EvaluatedAt:   time.Now().UTC().Format(time.RFC3339Nano),
		InputHash:     InputHash(req),
	}
}
