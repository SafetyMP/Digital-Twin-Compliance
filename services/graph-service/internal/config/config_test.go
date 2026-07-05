package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.HTTPAddr != ":8093" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.ConsumerGroup != "graph-service" {
		t.Fatalf("ConsumerGroup = %q", cfg.ConsumerGroup)
	}
	if len(cfg.KafkaBrokers) == 0 {
		t.Fatal("expected kafka brokers")
	}
}
