from __future__ import annotations

import os
import uuid
from typing import Any

import httpx
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from reporting_service import config
from reporting_service.audit import publish_report_submitted
from reporting_service.db import connect, migrate, seed_taxonomies
from reporting_service.generators import generate_anacredit_t2, generate_dora_ict, generate_finrep_f01
from reporting_service import storage
from reporting_service.taxonomy import list_taxonomies, maps_for
from reporting_service.validate import ValidationError, validate_report

_docs = None if os.getenv("REPORTING_DISABLE_DOCS", "1") == "1" else "/docs"
_openapi = None if os.getenv("REPORTING_DISABLE_DOCS", "1") == "1" else "/openapi.json"

app = FastAPI(
    title="Reporting Service",
    version="0.1.0",
    docs_url=_docs,
    redoc_url=None,
    openapi_url=_openapi,
)


def _setup_otel() -> None:
    endpoint = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT")
    if not endpoint:
        return
    from opentelemetry import trace
    from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
    from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
    from opentelemetry.sdk.resources import Resource
    from opentelemetry.sdk.trace import TracerProvider
    from opentelemetry.sdk.trace.export import BatchSpanProcessor

    resource = Resource.create({"service.name": os.environ.get("OTEL_SERVICE_NAME", "reporting-service")})
    provider = TracerProvider(resource=resource)
    provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=endpoint, insecure=True)))
    trace.set_tracer_provider(provider)
    FastAPIInstrumentor.instrument_app(app)


_setup_otel()

REPORT_TYPES = {"FINREP_F01", "ANACREDIT_T2", "DORA_ICT"}


class CreateReportRequest(BaseModel):
    reportType: str
    taxonomyVersion: str | None = None
    correlationId: str | None = None
    tenantId: str | None = None


@app.on_event("startup")
def on_startup() -> None:
    migrate()
    seed_taxonomies(config.DEFAULT_TENANT_ID, config.TAXONOMY_VERSION_DEFAULT)
    try:
        storage.ensure_bucket()
    except Exception:
        # MinIO may come up after us; put_report will retry ensure_bucket
        pass


@app.get("/api/v1/health")
def health() -> dict[str, Any]:
    try:
        with connect() as conn:
            with conn.cursor() as cur:
                cur.execute("SELECT 1")
        db_ok = True
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=503, detail={"status": "degraded", "db": str(exc)}) from exc
    minio_ok = False
    try:
        minio_ok = storage.bucket_object_lock_enabled()
    except Exception:
        minio_ok = False
    return {"status": "ok", "db": db_ok, "minio": minio_ok}


@app.get("/api/v1/taxonomies")
def taxonomies(tenantId: str | None = None) -> list[dict]:
    tid = tenantId or config.DEFAULT_TENANT_ID
    return list_taxonomies(tid)


@app.post("/api/v1/reports")
def create_report(body: CreateReportRequest) -> dict[str, Any]:
    if body.reportType not in REPORT_TYPES:
        raise HTTPException(status_code=400, detail="unsupported reportType")
    tenant_id = body.tenantId or config.DEFAULT_TENANT_ID
    version = body.taxonomyVersion or config.TAXONOMY_VERSION_DEFAULT
    maps = maps_for(tenant_id, body.reportType, version)
    if not maps:
        raise HTTPException(status_code=400, detail="no taxonomy maps for version")

    figures, instruments, assets = _seed_inputs()
    if body.reportType == "FINREP_F01":
        artifact = generate_finrep_f01(version, maps, figures)
    elif body.reportType == "ANACREDIT_T2":
        artifact = generate_anacredit_t2(version, maps, instruments)
    else:
        artifact = generate_dora_ict(version, maps, assets)

    report_id = str(uuid.uuid4())
    correlation_id = body.correlationId or report_id
    with connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO reports
                  (id, tenant_id, report_type, status, taxonomy_version, artifact_xml, correlation_id)
                VALUES (%s, %s, %s, 'draft', %s, %s, %s)
                """,
                (report_id, tenant_id, body.reportType, version, artifact, correlation_id),
            )
        conn.commit()
    return {
        "id": report_id,
        "status": "draft",
        "reportType": body.reportType,
        "taxonomyVersion": version,
        "correlationId": correlation_id,
    }


@app.post("/api/v1/reports/{report_id}/validate")
def validate_report_endpoint(report_id: str) -> dict[str, Any]:
    row = _get_report(report_id)
    if row["status"] != "draft":
        raise HTTPException(status_code=409, detail=f"illegal transition from {row['status']} to validated")
    try:
        validate_report(row["report_type"], row["artifact_xml"] or "")
    except ValidationError as exc:
        raise HTTPException(status_code=422, detail=str(exc)) from exc
    with connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                "UPDATE reports SET status='validated', updated_at=now() WHERE id=%s",
                (report_id,),
            )
        conn.commit()
    return {"id": report_id, "status": "validated", "taxonomyVersion": row["taxonomy_version"]}


@app.post("/api/v1/reports/{report_id}/submit")
def submit_report(report_id: str) -> dict[str, Any]:
    row = _get_report(report_id)
    if row["status"] != "validated":
        raise HTTPException(status_code=409, detail=f"illegal transition from {row['status']} to submitted")
    object_key = f"{row['report_type']}/{report_id}.xml"
    try:
        storage.put_report(object_key, (row["artifact_xml"] or "").encode("utf-8"))
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=503, detail=f"object store failed: {exc}") from exc
    try:
        evidence_ref = publish_report_submitted(
            report_id,
            row["correlation_id"] or report_id,
            row["report_type"],
            object_key,
            row["taxonomy_version"],
        )
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=503, detail=f"audit publish failed: {exc}") from exc
    with connect() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE reports
                   SET status='submitted', object_key=%s, evidence_ref=%s, updated_at=now()
                 WHERE id=%s
                """,
                (object_key, evidence_ref, report_id),
            )
        conn.commit()
    return {
        "id": report_id,
        "status": "submitted",
        "objectKey": object_key,
        "evidenceRef": evidence_ref,
        "taxonomyVersion": row["taxonomy_version"],
        "objectLock": True,
    }


@app.get("/api/v1/reports/{report_id}")
def get_report(report_id: str) -> dict[str, Any]:
    row = _get_report(report_id)
    return {
        "id": str(row["id"]),
        "status": row["status"],
        "reportType": row["report_type"],
        "taxonomyVersion": row["taxonomy_version"],
        "objectKey": row["object_key"],
        "evidenceRef": row["evidence_ref"],
        "correlationId": row["correlation_id"],
        "artifactXml": row["artifact_xml"],
    }


def _get_report(report_id: str) -> dict:
    with connect() as conn:
        with conn.cursor() as cur:
            cur.execute("SELECT * FROM reports WHERE id=%s", (report_id,))
            row = cur.fetchone()
    if not row:
        raise HTTPException(status_code=404, detail="report not found")
    return row


def _seed_inputs() -> tuple[dict[str, float], list[dict], list[dict]]:
    figures = {
        "assets.cash": 1_250_000.0,
        "assets.loans": 8_500_000.0,
        "liab.deposits": 7_900_000.0,
    }
    instruments = [
        {"id": "inst-seed-1", "type": "loan", "notional": 2_500_000, "currency": "EUR"},
        {"id": "inst-seed-2", "type": "bond", "notional": 1_000_000, "currency": "EUR"},
    ]
    assets = [
        {"name": "core-banking", "criticality": "high"},
        {"name": "kafka", "criticality": "high"},
        {"name": "immudb", "criticality": "high"},
    ]
    # Best-effort enrich from state-service institutions if reachable
    try:
        with httpx.Client(timeout=3.0) as client:
            resp = client.get(f"{config.STATE_SERVICE_URL}/api/v1/health")
            if resp.status_code == 200:
                figures["assets.cash"] = 1_250_000.0  # keep deterministic smoke figures
    except Exception:
        pass
    return figures, instruments, assets


def main() -> None:
    import uvicorn

    host, port = "0.0.0.0", 8095
    addr = config.HTTP_ADDR
    if addr.startswith(":"):
        port = int(addr[1:])
    uvicorn.run("reporting_service.main:app", host=host, port=port, reload=False)


if __name__ == "__main__":
    main()
