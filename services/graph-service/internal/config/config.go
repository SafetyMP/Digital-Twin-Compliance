package config

import (
	"os"
	"strings"
)

type Config struct {
	Neo4jURI         string
	Neo4jUser        string
	Neo4jPassword    string
	KafkaBrokers     []string
	HTTPAddr         string
	DefaultTenantID  string
	TwinTopic        string
	InstrumentsTopic string
	ConsumerGroup    string
	ServiceSource    string
}

func Load() Config {
	return Config{
		Neo4jURI:         env("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:        env("NEO4J_USER", "neo4j"),
		Neo4jPassword:    env("NEO4J_PASSWORD", "changeme"),
		KafkaBrokers:     splitCSV(env("KAFKA_BROKERS", "localhost:9092")),
		HTTPAddr:         env("GRAPH_SERVICE_HTTP_ADDR", ":8093"),
		DefaultTenantID:  env("DEFAULT_TENANT_ID", "00000000-0000-0000-0000-000000000001"),
		TwinTopic:        env("TWIN_STATE_TOPIC", "twin.state.updated"),
		InstrumentsTopic: env("DOMAIN_INSTRUMENTS_TOPIC", "domain.events.public.instruments"),
		ConsumerGroup:    env("GRAPH_CONSUMER_GROUP", "graph-service"),
		ServiceSource:    env("GRAPH_SERVICE_SOURCE", "graph-service"),
	}
}

func (c Config) InstrumentsConsumerGroup() string {
	if g := env("GRAPH_INSTRUMENTS_CONSUMER_GROUP", ""); g != "" {
		return g
	}
	return c.ConsumerGroup + "-instruments"
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
