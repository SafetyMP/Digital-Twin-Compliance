package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AuditDBURL      string
	KafkaBrokers    []string
	HTTPAddr        string
	DefaultTenantID string
	PendingTopic    string
	RecordedTopic   string
	PendingDLQTopic string
	ConsumerGroup   string
	ImmuDBHost      string
	ImmuDBPort      int
	ImmuDBDatabase  string
	ImmuDBUser      string
	ImmuDBPassword  string
}

func Load() Config {
	port, _ := strconv.Atoi(env("IMMUDB_PORT", "3322"))
	return Config{
		AuditDBURL:      env("AUDIT_DB_URL", "postgres://audit:audit@localhost:5436/audit?sslmode=disable"),
		KafkaBrokers:    splitCSV(env("KAFKA_BROKERS", "localhost:9092")),
		HTTPAddr:        env("AUDIT_SERVICE_HTTP_ADDR", ":8090"),
		DefaultTenantID: env("DEFAULT_TENANT_ID", "00000000-0000-0000-0000-000000000001"),
		PendingTopic:    env("KAFKA_AUDIT_PENDING_TOPIC", "compliance.audit.pending"),
		RecordedTopic:   env("KAFKA_AUDIT_RECORDED_TOPIC", "compliance.audit.recorded"),
		PendingDLQTopic: env("KAFKA_AUDIT_PENDING_DLQ_TOPIC", "compliance.audit.pending.dlq"),
		ConsumerGroup:   env("AUDIT_CONSUMER_GROUP", "audit-service"),
		ImmuDBHost:      env("IMMUDB_HOST", "localhost"),
		ImmuDBPort:      port,
		ImmuDBDatabase:  env("IMMUDB_DATABASE", "digitaltwin_audit"),
		ImmuDBUser:      env("IMMUDB_USER", "immudb"),
		ImmuDBPassword:  env("IMMUDB_PASSWORD", "immudb"),
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
