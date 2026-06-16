package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestKafkaContract_AuditPendingPublisherAlert(t *testing.T) {
	t.Parallel()

	raw := readContractFile(t, "compliance.audit.pending/alert-audit.envelope.json")
	var env map[string]any
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env["eventType"] != "AuditPending" {
		t.Fatalf("eventType = %#v", env["eventType"])
	}
	if env["source"] != "alert-service" {
		t.Fatalf("source = %#v", env["source"])
	}
	payload, ok := env["payload"].(map[string]any)
	if !ok {
		t.Fatalf("payload type = %T", env["payload"])
	}
	if payload["entryType"] != "Alert" {
		t.Fatalf("entryType = %#v", payload["entryType"])
	}
	subject, ok := payload["subject"].(map[string]any)
	if !ok || subject["subjectType"] != "ComplianceAlert" {
		t.Fatalf("subject = %#v", payload["subject"])
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
