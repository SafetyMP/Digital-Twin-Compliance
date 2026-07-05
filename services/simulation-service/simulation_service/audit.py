import json
import uuid
from datetime import datetime, timezone

from kafka import KafkaProducer

from simulation_service import config


def publish_simulation_run(
    run_id: str,
    correlation_id: str,
    metrics: dict,
) -> None:
    pending = {
        "entryType": "SimulationRun",
        "correlationId": correlation_id,
        "subject": {
            "subjectId": run_id,
            "subjectType": "SimulationRun",
        },
        "actor": {
            "actorId": config.SERVICE_SOURCE,
            "actorType": "Service",
        },
        "action": "SimulationRunCompleted",
        "payload": {
            "scenarioId": metrics["scenarioId"],
            "baselineCet1": metrics["baselineCet1"],
            "stressedCet1": metrics["stressedCet1"],
            "explainabilityRef": metrics["explainabilityRef"],
        },
    }
    envelope = {
        "eventId": str(uuid.uuid4()),
        "eventType": "AuditPending",
        "eventVersion": "1.0",
        "source": config.SERVICE_SOURCE,
        "correlationId": correlation_id,
        "timestamp": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.%fZ"),
        "idempotencyKey": f"audit-simulation-{run_id}",
        "payload": pending,
    }
    producer = KafkaProducer(
        bootstrap_servers=config.KAFKA_BROKERS.split(","),
        value_serializer=lambda v: json.dumps(v).encode("utf-8"),
    )
    try:
        producer.send(config.KAFKA_AUDIT_PENDING_TOPIC, value=envelope, key=run_id.encode())
        producer.flush(timeout=10)
    finally:
        producer.close()
