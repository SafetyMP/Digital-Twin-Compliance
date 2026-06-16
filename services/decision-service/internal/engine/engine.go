package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/digital-twin/platform/services/decision-service/internal/decision"
	zen "github.com/gorules/zen-go"
)

var ruleCodeToFile = map[string]string{
	"INT-R001":   "int-r001.json",
	"INT-R002":   "int-r002.json",
	"BASEL-R001": "basel-r001.json",
	"COREP-R001": "corep-r001.json",
	"COREP-R002": "corep-r002.json",
}

type RuleInfo struct {
	RuleCode string `json:"ruleCode"`
	File     string `json:"file"`
	Version  string `json:"version"`
}

type Evaluator struct {
	engine    zen.Engine
	policyDir string
	rules     map[string]RuleInfo
	mu        sync.RWMutex
}

func NewEvaluator(policyDir string) (*Evaluator, error) {
	absDir, err := filepath.Abs(policyDir)
	if err != nil {
		return nil, fmt.Errorf("resolve policy dir: %w", err)
	}

	e := &Evaluator{
		policyDir: absDir,
		rules:     make(map[string]RuleInfo),
	}

	e.engine = zen.NewEngine(zen.EngineConfig{Loader: e.loadDecision})
	if err := e.indexRules(); err != nil {
		e.engine.Dispose()
		return nil, err
	}
	return e, nil
}

func (e *Evaluator) loadDecision(key string) ([]byte, error) {
	path := filepath.Join(e.policyDir, key)
	return os.ReadFile(path)
}

func (e *Evaluator) indexRules() error {
	for ruleCode, file := range ruleCodeToFile {
		path := filepath.Join(e.policyDir, file)
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("load policy %s: %w", file, err)
		}
		version, err := parsePolicyVersion(raw)
		if err != nil {
			return fmt.Errorf("parse version for %s: %w", file, err)
		}
		decision, err := e.engine.CreateDecision(raw)
		if err != nil {
			return fmt.Errorf("compile policy %s: %w", file, err)
		}
		decision.Dispose()
		e.rules[strings.ToUpper(ruleCode)] = RuleInfo{
			RuleCode: strings.ToUpper(ruleCode),
			File:     file,
			Version:  version,
		}
	}
	return nil
}

func parsePolicyVersion(raw []byte) (string, error) {
	var doc struct {
		Metadata struct {
			Version string `json:"version"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", err
	}
	if doc.Metadata.Version == "" {
		return "unknown", nil
	}
	return doc.Metadata.Version, nil
}

func (e *Evaluator) ListRules() []RuleInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]RuleInfo, 0, len(e.rules))
	for _, info := range e.rules {
		out = append(out, info)
	}
	sortRules(out)
	return out
}

func sortRules(rules []RuleInfo) {
	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[j].RuleCode < rules[i].RuleCode {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}

func (e *Evaluator) Evaluate(ruleCode string, input map[string]any) (decision.RuleDecision, error) {
	code := strings.ToUpper(strings.TrimSpace(ruleCode))
	info, ok := e.rules[code]
	if !ok {
		return decision.RuleDecision{}, fmt.Errorf("unknown rule code %q", ruleCode)
	}

	resp, err := e.engine.Evaluate(info.File, input)
	if err != nil {
		return decision.RuleDecision{}, fmt.Errorf("evaluate %s: %w", code, err)
	}

	var zenOut decision.ZenOutput
	if err := json.Unmarshal(resp.Result, &zenOut); err != nil {
		return decision.RuleDecision{}, fmt.Errorf("parse zen output: %w", err)
	}
	if zenOut.RuleCode == "" {
		zenOut.RuleCode = code
	}
	if zenOut.PolicyVersion == "" {
		zenOut.PolicyVersion = info.Version
	}

	return decision.BuildDecision(code, zenOut, input, time.Now())
}

func (e *Evaluator) Close() {
	if e.engine != nil {
		e.engine.Dispose()
	}
}

func PolicyDirFromRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "policies/zen"
	}
	dir := wd
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, "policies", "zen")
		if _, err := os.Stat(filepath.Join(candidate, "int-r001.json")); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "policies/zen"
}
