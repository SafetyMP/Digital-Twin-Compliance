package events

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKafkaContract_ComplianceAlertsConsumer(t *testing.T) {
	t.Parallel()

	raw := readContractFile(t, "compliance.alerts/basel-alert-raised.envelope.json")
	env, err := ParseEnvelope(raw)
	if err != nil {
		t.Fatalf("ParseEnvelope: %v", err)
	}
	if env.EventType != "ComplianceAlertRaised" {
		t.Fatalf("eventType = %q", env.EventType)
	}
	if env.Source != "flink-compliance-cep" {
		t.Fatalf("source = %q", env.Source)
	}

	alert, err := ParseAlertPayload(env.Payload)
	if err != nil {
		t.Fatalf("ParseAlertPayload: %v", err)
	}
	if alert.RuleCode != "BASEL-M001" {
		t.Fatalf("ruleCode = %q", alert.RuleCode)
	}
	if alert.PersonaID != "44444444-4444-4444-4444-444444444401" {
		t.Fatalf("personaId = %q", alert.PersonaID)
	}
	if _, err := ParseDetectedAt(alert.DetectedAt); err != nil {
		t.Fatalf("ParseDetectedAt: %v", err)
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
