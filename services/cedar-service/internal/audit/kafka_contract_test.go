package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestKafkaContract_RuleDecisionPayloadGolden(t *testing.T) {
	t.Parallel()

	raw := readContractFile(t, "rule-decision/basel-r001-deny.payload.json")
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	if body["ruleCode"] != "BASEL-R001" || body["outcome"] != "Deny" {
		t.Fatalf("payload = %#v", body)
	}
}

func readContractFile(t *testing.T, rel string) []byte {
	t.Helper()
	path := filepath.Join(repoRoot(t), "contracts", "kafka", filepath.FromSlash(rel))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read contract %s: %v", rel, err)
	}
	return raw
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "contracts", "kafka", "README.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}
