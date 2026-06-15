package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("ALERT_DB_URL", "")
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("DEFAULT_TENANT_ID", "")

	cfg := Load()
	if cfg.AlertDBURL == "" {
		t.Fatal("expected default AlertDBURL")
	}
	if len(cfg.KafkaBrokers) != 1 || cfg.KafkaBrokers[0] != "localhost:9092" {
		t.Fatalf("KafkaBrokers = %v", cfg.KafkaBrokers)
	}
	if cfg.DefaultTenantID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("DefaultTenantID = %q", cfg.DefaultTenantID)
	}
	if cfg.WSSPath != "/ws/alerts" {
		t.Fatalf("WSSPath = %q", cfg.WSSPath)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("ALERT_DB_URL", "postgres://custom/alerts")
	t.Setenv("KAFKA_BROKERS", "a:9092,b:9092")
	t.Setenv("ALERT_SERVICE_HTTP_ADDR", ":9091")
	t.Setenv("COMPLIANCE_ALERTS_TOPIC", "alerts.custom")
	t.Setenv("COMPLIANCE_ALERTS_DLQ_TOPIC", "alerts.custom.dlq")
	t.Setenv("ALERT_CONSUMER_GROUP", "alerts-group")
	t.Setenv("ALERT_SERVICE_WS_PATH", "/ws/custom")

	cfg := Load()
	if cfg.AlertDBURL != "postgres://custom/alerts" {
		t.Fatalf("AlertDBURL = %q", cfg.AlertDBURL)
	}
	if len(cfg.KafkaBrokers) != 2 {
		t.Fatalf("KafkaBrokers = %v", cfg.KafkaBrokers)
	}
	if cfg.HTTPAddr != ":9091" || cfg.AlertsTopic != "alerts.custom" {
		t.Fatalf("cfg = %+v", cfg)
	}
	if cfg.AlertsDLQTopic != "alerts.custom.dlq" {
		t.Fatalf("AlertsDLQTopic = %q", cfg.AlertsDLQTopic)
	}
	if cfg.ConsumerGroup != "alerts-group" || cfg.WSSPath != "/ws/custom" {
		t.Fatalf("cfg = %+v", cfg)
	}
}
