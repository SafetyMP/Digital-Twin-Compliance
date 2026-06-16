package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("AUDIT_DB_URL", "")
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("DEFAULT_TENANT_ID", "")

	cfg := Load()
	if cfg.AuditDBURL == "" {
		t.Fatal("expected default AuditDBURL")
	}
	if cfg.HTTPAddr != ":8090" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.PendingTopic != "compliance.audit.pending" {
		t.Fatalf("PendingTopic = %q", cfg.PendingTopic)
	}
	if cfg.ImmuDBDatabase != "digitaltwin_audit" {
		t.Fatalf("ImmuDBDatabase = %q", cfg.ImmuDBDatabase)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("AUDIT_DB_URL", "postgres://custom/audit")
	t.Setenv("KAFKA_BROKERS", "a:9092,b:9092")
	t.Setenv("AUDIT_SERVICE_HTTP_ADDR", ":9099")
	t.Setenv("KAFKA_AUDIT_PENDING_TOPIC", "audit.custom.pending")
	t.Setenv("KAFKA_AUDIT_RECORDED_TOPIC", "audit.custom.recorded")
	t.Setenv("AUDIT_CONSUMER_GROUP", "audit-group")
	t.Setenv("IMMUDB_HOST", "immudb.local")
	t.Setenv("IMMUDB_PORT", "3323")
	t.Setenv("IMMUDB_DATABASE", "custom_audit")

	cfg := Load()
	if cfg.AuditDBURL != "postgres://custom/audit" {
		t.Fatalf("AuditDBURL = %q", cfg.AuditDBURL)
	}
	if len(cfg.KafkaBrokers) != 2 || cfg.HTTPAddr != ":9099" {
		t.Fatalf("cfg = %+v", cfg)
	}
	if cfg.ImmuDBHost != "immudb.local" || cfg.ImmuDBPort != 3323 {
		t.Fatalf("immudb cfg = %+v", cfg)
	}
}
