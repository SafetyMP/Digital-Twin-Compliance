import json
import uuid
from datetime import datetime, timezone

from kafka import KafkaProducer

from simulation_service import config

_producer: KafkaProducer | None = None


def _get_producer() -> KafkaProducer:
    global _producer
    if _producer is None:
        _producer = KafkaProducer(
            bootstrap_servers=config.KAFKA_BROKERS.split(","),
            value_serializer=lambda v: json.dumps(v).encode("utf-8"),
        )
    return _producer


def publish_audit(
    *,
    entry_type: str,
    subject_type: str,
    subject_id: str,
    action: str,
    correlation_id: str,
    payload: dict,
    idempotency_key: str,
) -> str:
    evidence_ref = f"{entry_type.lower()}:{subject_id}"
    pending = {
        "entryType": entry_type,
        "correlationId": correlation_id,
        "subject": {
            "subjectId": subject_id,
            "subjectType": subject_type,
        },
        "actor": {
            "actorId": config.SERVICE_SOURCE,
            "actorType": "Service",
        },
        "action": action,
        "payload": {**payload, "evidenceRef": evidence_ref},
    }
    envelope = {
        "eventId": str(uuid.uuid4()),
        "eventType": "AuditPending",
        "eventVersion": "1.0",
        "source": config.SERVICE_SOURCE,
        "correlationId": correlation_id,
        "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.%fZ"),
        "idempotencyKey": idempotency_key,
        "payload": pending,
    }
    producer = _get_producer()
    future = producer.send(
        config.KAFKA_AUDIT_PENDING_TOPIC,
        value=envelope,
        key=subject_id.encode(),
    )
    future.get(timeout=10)
    producer.flush(timeout=10)
    return evidence_ref


def publish_simulation_run(
    run_id: str,
    correlation_id: str,
    metrics: dict,
) -> None:
    publish_audit(
        entry_type="SimulationRun",
        subject_type="SimulationRun",
        subject_id=run_id,
        action="SimulationRunCompleted",
        correlation_id=correlation_id,
        payload={
            "scenarioId": metrics["scenarioId"],
            "baselineCet1": metrics["baselineCet1"],
            "stressedCet1": metrics["stressedCet1"],
            "explainabilityRef": metrics["explainabilityRef"],
        },
        idempotency_key=f"audit-simulation-{run_id}",
    )
