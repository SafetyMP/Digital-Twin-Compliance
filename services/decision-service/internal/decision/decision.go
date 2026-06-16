package decision

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	OutcomeAllow    = "Allow"
	OutcomeDeny     = "Deny"
	OutcomeFlag     = "Flag"
	OutcomeEscalate = "Escalate"
)

type RuleDecision struct {
	DecisionID    string   `json:"decisionId"`
	RuleCode      string   `json:"ruleCode"`
	Outcome       string   `json:"outcome"`
	Score         *float64 `json:"score,omitempty"`
	Rationale     string   `json:"rationale"`
	PolicyVersion string   `json:"policyVersion"`
	EvaluatedAt   string   `json:"evaluatedAt"`
	InputHash     string   `json:"inputHash"`
}

type ZenOutput struct {
	RuleCode      string  `json:"ruleCode"`
	Outcome       string  `json:"outcome"`
	Score         float64 `json:"score"`
	Rationale     string  `json:"rationale"`
	PolicyVersion string  `json:"policyVersion"`
}

func RequiresAudit(outcome string) bool {
	switch outcome {
	case OutcomeDeny, OutcomeFlag, OutcomeEscalate:
		return true
	default:
		return false
	}
}

func RegimeForRuleCode(ruleCode string) string {
	prefix := strings.ToUpper(strings.Split(ruleCode, "-")[0])
	switch prefix {
	case "INT":
		return "Internal"
	case "BASEL":
		return "Basel"
	case "COREP":
		return "COREP"
	default:
		return "Unknown"
	}
}

func SubjectFromInput(input map[string]any) (subjectID, subjectType string) {
	for _, key := range []string{"personaId", "accountId", "ownerEntityId", "tenantId"} {
		if v, ok := input[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				subjectType = "TwinPersona"
				if key == "accountId" {
					subjectType = "Account"
				} else if key == "ownerEntityId" {
					subjectType = "LegalEntity"
				}
				return s, subjectType
			}
		}
	}
	return "unknown", "Unknown"
}

func InputHash(input map[string]any) (string, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func BuildDecision(ruleCode string, zen ZenOutput, input map[string]any, evaluatedAt time.Time) (RuleDecision, error) {
	hash, err := InputHash(input)
	if err != nil {
		return RuleDecision{}, err
	}
	score := zen.Score
	return RuleDecision{
		DecisionID:    uuid.NewString(),
		RuleCode:      ruleCode,
		Outcome:       zen.Outcome,
		Score:         &score,
		Rationale:     zen.Rationale,
		PolicyVersion: zen.PolicyVersion,
		EvaluatedAt:   evaluatedAt.UTC().Format(time.RFC3339Nano),
		InputHash:     hash,
	}, nil
}

func ParseInput(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var input map[string]any
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if input == nil {
		input = map[string]any{}
	}
	return input, nil
}
