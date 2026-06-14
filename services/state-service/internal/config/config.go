package config

import (
	"os"
	"strings"
)

type Config struct {
	StateDBURL         string
	KafkaBrokers       []string
	SchemaRegistryURL  string
	DefaultTenantID    string
	HTTPAddr           string
	ServiceSource      string
	DebeziumTopics     []string
	OutboxPollInterval string
}

func Load() Config {
	return Config{
		StateDBURL:         env("STATE_DB_URL", "postgres://state:state@localhost:5434/twin_state?sslmode=disable"),
		KafkaBrokers:       splitEnv("KAFKA_BROKERS", "localhost:9092"),
		SchemaRegistryURL:  env("SCHEMA_REGISTRY_URL", "http://localhost:8081"),
		DefaultTenantID:    env("DEFAULT_TENANT_ID", "00000000-0000-0000-0000-000000000001"),
		HTTPAddr:           env("STATE_SERVICE_HTTP_ADDR", ":8080"),
		ServiceSource:      env("STATE_SERVICE_SOURCE", "state-service"),
		DebeziumTopics: splitEnv("DEBEZIUM_TOPICS",
			"domain.events.public.legal_entities,domain.events.public.accounts,domain.events.public.instruments"),
		OutboxPollInterval: env("OUTBOX_POLL_INTERVAL", "1s"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitEnv(key, fallback string) []string {
	raw := env(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
