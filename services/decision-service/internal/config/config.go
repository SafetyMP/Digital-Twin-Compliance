package config

import (
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers    []string
	HTTPAddr        string
	PolicyDir       string
	AuditTopic      string
	DefaultTenantID string
}

func Load() Config {
	return Config{
		KafkaBrokers:    splitCSV(env("KAFKA_BROKERS", "localhost:9092")),
		HTTPAddr:        env("DECISION_SERVICE_HTTP_ADDR", ":8092"),
		PolicyDir:       env("ZEN_POLICY_DIR", "policies/zen"),
		AuditTopic:      env("KAFKA_AUDIT_PENDING_TOPIC", "compliance.audit.pending"),
		DefaultTenantID: env("DEFAULT_TENANT_ID", "00000000-0000-0000-0000-000000000001"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
