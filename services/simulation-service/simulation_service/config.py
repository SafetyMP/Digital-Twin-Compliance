import os


def env(key: str, default: str) -> str:
    return os.environ.get(key, default)


GRAPH_SERVICE_URL = env("GRAPH_SERVICE_URL", "http://localhost:8093")
DECISION_SERVICE_URL = env("DECISION_SERVICE_URL", "http://localhost:8092")
KAFKA_BROKERS = env("KAFKA_BROKERS", "localhost:9092")
KAFKA_AUDIT_PENDING_TOPIC = env("KAFKA_AUDIT_PENDING_TOPIC", "compliance.audit.pending")
DEFAULT_TENANT_ID = env("DEFAULT_TENANT_ID", "00000000-0000-0000-0000-000000000001")
SERVICE_SOURCE = env("SIMULATION_SERVICE_SOURCE", "simulation-service")
HTTP_ADDR = env("SIMULATION_SERVICE_HTTP_ADDR", ":8094")
