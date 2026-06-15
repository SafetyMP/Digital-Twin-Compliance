package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("STATE_DB_URL", "")
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("DEFAULT_TENANT_ID", "")

	cfg := Load()
	if cfg.StateDBURL == "" {
		t.Fatal("expected default StateDBURL")
	}
	if len(cfg.KafkaBrokers) != 1 || cfg.KafkaBrokers[0] != "localhost:9092" {
		t.Fatalf("KafkaBrokers = %v", cfg.KafkaBrokers)
	}
	if cfg.DefaultTenantID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("DefaultTenantID = %q", cfg.DefaultTenantID)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if len(cfg.DebeziumTopics) != 3 {
		t.Fatalf("DebeziumTopics = %v", cfg.DebeziumTopics)
	}
	if cfg.DebeziumDLQTopic != "domain.events.dlq" {
		t.Fatalf("DebeziumDLQTopic = %q", cfg.DebeziumDLQTopic)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("STATE_DB_URL", "postgres://custom/db")
	t.Setenv("KAFKA_BROKERS", "broker1:9092, broker2:9092")
	t.Setenv("DEFAULT_TENANT_ID", "tenant-abc")
	t.Setenv("STATE_SERVICE_HTTP_ADDR", ":9090")
	t.Setenv("DEBEZIUM_TOPICS", "a,b")
	t.Setenv("STATE_CDC_DLQ_TOPIC", "cdc.custom.dlq")

	cfg := Load()
	if cfg.StateDBURL != "postgres://custom/db" {
		t.Fatalf("StateDBURL = %q", cfg.StateDBURL)
	}
	if len(cfg.KafkaBrokers) != 2 || cfg.KafkaBrokers[1] != "broker2:9092" {
		t.Fatalf("KafkaBrokers = %v", cfg.KafkaBrokers)
	}
	if cfg.DefaultTenantID != "tenant-abc" {
		t.Fatalf("DefaultTenantID = %q", cfg.DefaultTenantID)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if len(cfg.DebeziumTopics) != 2 {
		t.Fatalf("DebeziumTopics = %v", cfg.DebeziumTopics)
	}
	if cfg.DebeziumDLQTopic != "cdc.custom.dlq" {
		t.Fatalf("DebeziumDLQTopic = %q", cfg.DebeziumDLQTopic)
	}
}

func TestSplitEnvSkipsEmptyParts(t *testing.T) {
	t.Setenv("KAFKA_BROKERS", "a,, b, ,c")
	cfg := Load()
	want := []string{"a", "b", "c"}
	if len(cfg.KafkaBrokers) != len(want) {
		t.Fatalf("brokers = %v", cfg.KafkaBrokers)
	}
	for i := range want {
		if cfg.KafkaBrokers[i] != want[i] {
			t.Fatalf("brokers[%d] = %q, want %q", i, cfg.KafkaBrokers[i], want[i])
		}
	}
}
