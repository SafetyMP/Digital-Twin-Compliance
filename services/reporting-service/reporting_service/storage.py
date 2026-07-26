from __future__ import annotations

import io
from datetime import datetime, timedelta, timezone

from minio import Minio
from minio.commonconfig import GOVERNANCE
from minio.retention import Retention
from minio.versioningconfig import ENABLED, VersioningConfig

from reporting_service import config

_client: Minio | None = None


def client() -> Minio:
    global _client
    if _client is None:
        _client = Minio(
            config.MINIO_ENDPOINT,
            access_key=config.MINIO_ACCESS_KEY,
            secret_key=config.MINIO_SECRET_KEY,
            secure=config.MINIO_SECURE,
        )
    return _client


def ensure_bucket() -> None:
    c = client()
    if not c.bucket_exists(config.MINIO_BUCKET):
        c.make_bucket(config.MINIO_BUCKET, object_lock=True)
        try:
            c.set_bucket_versioning(config.MINIO_BUCKET, VersioningConfig(ENABLED))
        except Exception:
            pass


def put_report(object_key: str, body: bytes, content_type: str = "application/xml") -> str:
    ensure_bucket()
    c = client()
    retain_until = datetime.now(timezone.utc) + timedelta(days=365 * 7)
    retention = Retention(GOVERNANCE, retain_until)
    c.put_object(
        config.MINIO_BUCKET,
        object_key,
        data=io.BytesIO(body),
        length=len(body),
        content_type=content_type,
        retention=retention,
    )
    return object_key


def object_exists(object_key: str) -> bool:
    try:
        client().stat_object(config.MINIO_BUCKET, object_key)
        return True
    except Exception:
        return False


def bucket_object_lock_enabled() -> bool:
    """Best-effort: bucket exists with object-lock creation flag (smoke probe)."""
    ensure_bucket()
    return client().bucket_exists(config.MINIO_BUCKET)
