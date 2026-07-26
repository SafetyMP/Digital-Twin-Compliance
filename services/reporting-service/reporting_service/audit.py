from __future__ import annotations

import json
import uuid
from datetime import datetime, timezone

from kafka import KafkaProducer

from reporting_service import config

_producer: KafkaProducer | None = None


def _get_producer() -> KafkaProducer:
    global _producer
    if _producer is None:
        _producer = KafkaProducer(
            bootstrap_servers=config.KAFKA_BROKERS.split(","),
            value_serializer=lambda v: json.dumps(v).encode("utf-8"),
        )
    return _producer


def publish_report_submitted(
    report_id: str,
    correlation_id: str,
    report_type: str,
    object_key: str,
    taxonomy_version: str,
) -> str:
    """Publish audit pending; returns evidenceRef (= report_id for smoke verify)."""
    evidence_ref = f"report:{report_id}"
    pending = {
        "entryType": "RegulatoryReport",
        "correlationId": correlation_id,
        "subject": {
            "subjectId": report_id,
            "subjectType": "RegulatoryReport",
        },
        "actor": {
            "actorId": config.SERVICE_SOURCE,
            "actorType": "Service",
        },
        "action": "ReportSubmitted",
        "payload": {
            "reportType": report_type,
            "objectKey": object_key,
            "taxonomyVersion": taxonomy_version,
            "evidenceRef": evidence_ref,
        },
    }
    envelope = {
        "eventId": str(uuid.uuid4()),
        "eventType": "AuditPending",
        "eventVersion": "1.0",
        "source": config.SERVICE_SOURCE,
        "correlationId": correlation_id,
        "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.%fZ"),
        "idempotencyKey": f"audit-report-{report_id}",
        "payload": pending,
    }
    producer = _get_producer()
    future = producer.send(
        config.KAFKA_AUDIT_PENDING_TOPIC,
        value=envelope,
        key=report_id.encode(),
    )
    future.get(timeout=10)
    producer.flush(timeout=10)
    return evidence_ref
