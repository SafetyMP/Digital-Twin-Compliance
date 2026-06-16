package config

import (
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers     []string
	HTTPAddr         string
	PolicyDir        string
	AuditTopic       string
	DefaultPrincipal string
	DefaultRoles     []string
}

func Load() Config {
	return Config{
		KafkaBrokers:     splitCSV(env("KAFKA_BROKERS", "localhost:9092")),
		HTTPAddr:         env("CEDAR_SERVICE_HTTP_ADDR", ":8091"),
		PolicyDir:        env("CEDAR_POLICY_DIR", "policies/cedar"),
		AuditTopic:       env("KAFKA_AUDIT_PENDING_TOPIC", "compliance.audit.pending"),
		DefaultPrincipal: env("DEFAULT_PRINCIPAL", "operator-dev"),
		DefaultRoles:     splitCSV(env("DEFAULT_ROLES", "Analyst")),
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
