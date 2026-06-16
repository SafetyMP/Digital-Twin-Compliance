package engine_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/digital-twin/platform/services/decision-service/internal/engine"
)

func policyDir(t *testing.T) string {
	t.Helper()
	dir := engine.PolicyDirFromRepoRoot()
	if _, err := os.Stat(filepath.Join(dir, "int-r001.json")); err != nil {
		t.Skipf("policies not found at %s: %v", dir, err)
	}
	return dir
}

func TestZenFixtures(t *testing.T) {
	dir := policyDir(t)
	eval, err := engine.NewEvaluator(dir)
	if err != nil {
		t.Fatalf("NewEvaluator: %v", err)
	}
	defer eval.Close()

	fixtureDir := filepath.Join(dir, "fixtures")
	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	for _, ent := range entries {
		if ent.IsDir() || filepath.Ext(ent.Name()) != ".json" {
			continue
		}
		t.Run(ent.Name(), func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(fixtureDir, ent.Name()))
			if err != nil {
				t.Fatal(err)
			}
			var spec struct {
				RuleCode string `json:"ruleCode"`
				Cases    []struct {
					Name     string         `json:"name"`
					Input    map[string]any `json:"input"`
					Expected struct {
						RuleCode      string  `json:"ruleCode"`
						Outcome       string  `json:"outcome"`
						Score         float64 `json:"score"`
						Rationale     string  `json:"rationale"`
						PolicyVersion string  `json:"policyVersion"`
					} `json:"expected"`
				} `json:"cases"`
			}
			if err := json.Unmarshal(raw, &spec); err != nil {
				t.Fatal(err)
			}
			for _, c := range spec.Cases {
				t.Run(c.Name, func(t *testing.T) {
					got, err := eval.Evaluate(spec.RuleCode, c.Input)
					if err != nil {
						t.Fatalf("evaluate: %v", err)
					}
					if got.Outcome != c.Expected.Outcome {
						t.Fatalf("outcome = %q want %q", got.Outcome, c.Expected.Outcome)
					}
					if got.Rationale != c.Expected.Rationale {
						t.Fatalf("rationale = %q want %q", got.Rationale, c.Expected.Rationale)
					}
					if got.PolicyVersion != c.Expected.PolicyVersion {
						t.Fatalf("policyVersion = %q want %q", got.PolicyVersion, c.Expected.PolicyVersion)
					}
					if got.Score == nil || *got.Score != c.Expected.Score {
						t.Fatalf("score = %v want %v", got.Score, c.Expected.Score)
					}
				})
			}
		})
	}
}
