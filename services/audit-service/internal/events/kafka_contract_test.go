package events

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKafkaContract_AuditPendingAlertEnvelope(t *testing.T) {
	t.Parallel()
	testAuditPendingGolden(t, "compliance.audit.pending/alert-audit.envelope.json", "alert-service", "Alert")
}

func TestKafkaContract_AuditPendingRuleDecisionEnvelope(t *testing.T) {
	t.Parallel()
	testAuditPendingGolden(t, "compliance.audit.pending/rule-decision-audit.envelope.json", "decision-service", "RuleDecision")
}

func testAuditPendingGolden(t *testing.T, rel, wantSource, wantEntryType string) {
	t.Helper()

	raw := readContractFile(t, rel)
	env, err := ParseEnvelope(raw)
	if err != nil {
		t.Fatalf("ParseEnvelope: %v", err)
	}
	if env.EventType != EventTypeAuditPending {
		t.Fatalf("eventType = %q", env.EventType)
	}
	if env.Source != wantSource {
		t.Fatalf("source = %q want %q", env.Source, wantSource)
	}
	if env.IdempotencyKey == "" {
		t.Fatal("missing idempotencyKey")
	}

	pending, err := ParseAuditPending(env.Payload)
	if err != nil {
		t.Fatalf("ParseAuditPending: %v", err)
	}
	if pending.EntryType != wantEntryType {
		t.Fatalf("entryType = %q want %q", pending.EntryType, wantEntryType)
	}
	if pending.Subject.SubjectID == "" || pending.Subject.SubjectType == "" {
		t.Fatalf("subject = %#v", pending.Subject)
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
			t.Fatal("repo root not found (contracts/kafka/README.md)")
		}
		dir = parent
	}
}
