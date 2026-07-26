import os


def env(key: str, default: str) -> str:
    return os.environ.get(key, default)


DB_URL = env(
    "REPORTING_DB_URL",
    "postgres://reporting:reporting@localhost:5437/reporting?sslmode=disable",
)
STATE_SERVICE_URL = env("STATE_SERVICE_URL", "http://localhost:8080")
MINIO_ENDPOINT = env("MINIO_ENDPOINT", "localhost:9000")
MINIO_ACCESS_KEY = env("MINIO_ACCESS_KEY", "minio")
MINIO_SECRET_KEY = env("MINIO_SECRET_KEY", "minioadmin")
MINIO_BUCKET = env("MINIO_BUCKET", "regulatory-reports")
MINIO_SECURE = env("MINIO_SECURE", "0") == "1"
KAFKA_BROKERS = env("KAFKA_BROKERS", "localhost:9092")
KAFKA_AUDIT_PENDING_TOPIC = env("KAFKA_AUDIT_PENDING_TOPIC", "compliance.audit.pending")
DEFAULT_TENANT_ID = env("DEFAULT_TENANT_ID", "00000000-0000-0000-0000-000000000001")
SERVICE_SOURCE = env("REPORTING_SERVICE_SOURCE", "reporting-service")
HTTP_ADDR = env("REPORTING_SERVICE_HTTP_ADDR", ":8095")
TAXONOMY_VERSION_DEFAULT = env("TAXONOMY_VERSION_DEFAULT", "2024.1")
FIXTURES_DIR = env("REPORTING_FIXTURES_DIR", "/fixtures/taxonomies")
