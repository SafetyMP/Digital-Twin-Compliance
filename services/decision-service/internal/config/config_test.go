package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("DECISION_SERVICE_HTTP_ADDR", "")
	t.Setenv("ZEN_POLICY_DIR", "")
	t.Setenv("KAFKA_AUDIT_PENDING_TOPIC", "")

	cfg := Load()
	if cfg.HTTPAddr != ":8092" {
		t.Fatalf("HTTPAddr = %q, want :8092", cfg.HTTPAddr)
	}
	if cfg.PolicyDir != "policies/zen" {
		t.Fatalf("PolicyDir = %q, want policies/zen", cfg.PolicyDir)
	}
	if cfg.AuditTopic != "compliance.audit.pending" {
		t.Fatalf("AuditTopic = %q", cfg.AuditTopic)
	}
	if len(cfg.KafkaBrokers) != 1 || cfg.KafkaBrokers[0] != "localhost:9092" {
		t.Fatalf("KafkaBrokers = %v", cfg.KafkaBrokers)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("DECISION_SERVICE_HTTP_ADDR", ":9099")
	t.Setenv("ZEN_POLICY_DIR", "/tmp/zen")
	t.Setenv("KAFKA_BROKERS", "a:9092,b:9092")

	cfg := Load()
	if cfg.HTTPAddr != ":9099" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.PolicyDir != "/tmp/zen" {
		t.Fatalf("PolicyDir = %q", cfg.PolicyDir)
	}
	if len(cfg.KafkaBrokers) != 2 {
		t.Fatalf("KafkaBrokers = %v", cfg.KafkaBrokers)
	}
}
