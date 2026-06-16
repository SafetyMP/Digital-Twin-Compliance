package config

import (
	"os"
	"strings"
)

type Config struct {
	AlertDBURL         string
	KafkaBrokers       []string
	HTTPAddr           string
	DefaultTenantID    string
	AlertsTopic        string
	AlertsDLQTopic     string
	ConsumerGroup      string
	WSSPath            string
	AuditPendingTopic  string
	AuditRecordedTopic string
	AuditConsumerGroup string
	ServiceSource      string
}

func Load() Config {
	return Config{
		AlertDBURL:         env("ALERT_DB_URL", "postgres://alert:alert@localhost:5435/alerts?sslmode=disable"),
		KafkaBrokers:       splitCSV(env("KAFKA_BROKERS", "localhost:9092")),
		HTTPAddr:           env("ALERT_SERVICE_HTTP_ADDR", ":8085"),
		DefaultTenantID:    env("DEFAULT_TENANT_ID", "00000000-0000-0000-0000-000000000001"),
		AlertsTopic:        env("COMPLIANCE_ALERTS_TOPIC", "compliance.alerts"),
		AlertsDLQTopic:     env("COMPLIANCE_ALERTS_DLQ_TOPIC", "compliance.alerts.dlq"),
		ConsumerGroup:      env("ALERT_CONSUMER_GROUP", "alert-service"),
		WSSPath:            env("ALERT_SERVICE_WS_PATH", "/ws/alerts"),
		AuditPendingTopic:  env("KAFKA_AUDIT_PENDING_TOPIC", "compliance.audit.pending"),
		AuditRecordedTopic: env("KAFKA_AUDIT_RECORDED_TOPIC", "compliance.audit.recorded"),
		AuditConsumerGroup: env("ALERT_AUDIT_CONSUMER_GROUP", "alert-service-audit"),
		ServiceSource:      env("ALERT_SERVICE_SOURCE", "alert-service"),
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
