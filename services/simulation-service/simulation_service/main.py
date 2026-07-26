import os
import uuid
from pathlib import Path

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from simulation_service import config
from simulation_service.audit import publish_audit, publish_simulation_run
from simulation_service.contagion import run_contagion
from simulation_service.decision import evaluate_corep
from simulation_service.effectiveness import load_breaches, replay
from simulation_service.graph_client import GraphClient
from simulation_service.scenario import SCENARIO_ECB_ADVERSE_V1, run_ecb_adverse_v1, stable_run_id

_docs = None if os.getenv("SIMULATION_DISABLE_DOCS", "1") == "1" else "/docs"
_openapi = None if os.getenv("SIMULATION_DISABLE_DOCS", "1") == "1" else "/openapi.json"

app = FastAPI(
    title="Simulation Service",
    version="0.1.0",
    docs_url=_docs,
    redoc_url=None,
    openapi_url=_openapi,
)
graph_client = GraphClient()


class RunRequest(BaseModel):
    scenarioId: str = Field(default=SCENARIO_ECB_ADVERSE_V1)
    parameters: dict = Field(default_factory=dict)
    correlationId: str | None = None
    personaId: str = Field(default="44444444-4444-4444-4444-444444444401")


class EffectivenessRequest(BaseModel):
    detections: list[str] = Field(default_factory=lambda: ["INT-M001", "INT-M002", "BASEL-M001"])
    correlationId: str | None = None
    breachesPath: str | None = None


class ContagionRequest(BaseModel):
    seedEntityId: str | None = None
    maxHops: int = 3
    correlationId: str | None = None


@app.get("/api/v1/health")
async def health():
    try:
        graph_health = await graph_client.health()
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=503, detail={"status": "degraded", "graph": "unreachable"}) from exc
    return {"status": "ok", "graph": graph_health.get("status", "unknown")}


@app.post("/api/v1/simulations/run")
async def run_simulation(body: RunRequest):
    if body.scenarioId != SCENARIO_ECB_ADVERSE_V1:
        raise HTTPException(status_code=400, detail="unsupported scenarioId")

    nodes = await graph_client.nodes()
    edges = await graph_client.edges()
    if not nodes:
        raise HTTPException(status_code=503, detail="graph not seeded")

    metrics = run_ecb_adverse_v1(nodes, edges)
    run_id = stable_run_id(body.scenarioId, body.parameters)
    correlation_id = body.correlationId or run_id
    metrics["runId"] = run_id

    decisions = await evaluate_corep(
        metrics["stressedCet1"],
        metrics["stressedTotalCapital"],
        body.personaId,
        correlation_id,
    )

    try:
        publish_simulation_run(run_id, correlation_id, metrics)
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=503, detail="audit publish failed") from exc

    return {
        "runId": run_id,
        "correlationId": correlation_id,
        "metrics": metrics,
        "decisions": decisions,
    }


@app.post("/api/v1/effectiveness/replay")
async def effectiveness_replay(body: EffectivenessRequest):
    path = body.breachesPath or os.environ.get(
        "EFFECTIVENESS_BREACHES_PATH",
        "/fixtures/phase7/breaches.json",
    )
    if not Path(path).is_file():
        # Host/dev fallback relative to repo
        alt = Path(__file__).resolve().parents[3] / "fixtures" / "phase7" / "breaches.json"
        path = str(alt)
    breaches = load_breaches(path)
    metrics = replay(breaches, set(body.detections))
    run_id = str(uuid.uuid4())
    correlation_id = body.correlationId or run_id
    try:
        evidence_ref = publish_audit(
            entry_type="ControlEffectivenessRun",
            subject_type="ControlEffectivenessRun",
            subject_id=run_id,
            action="EffectivenessReplayCompleted",
            correlation_id=correlation_id,
            payload={"metrics": metrics},
            idempotency_key=f"audit-effectiveness-{run_id}",
        )
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=503, detail="audit publish failed") from exc
    return {
        "runId": run_id,
        "correlationId": correlation_id,
        "metrics": metrics,
        "evidenceRef": evidence_ref,
    }


@app.post("/api/v1/contagion/run")
async def contagion_run(body: ContagionRequest):
    nodes = await graph_client.nodes()
    edges = await graph_client.edges()
    if not nodes:
        raise HTTPException(status_code=503, detail="graph not seeded")
    seed = body.seedEntityId or str(nodes[0].get("entityId"))
    result = run_contagion(nodes, edges, seed, max_hops=body.maxHops)
    run_id = str(uuid.uuid4())
    correlation_id = body.correlationId or run_id
    try:
        evidence_ref = publish_audit(
            entry_type="ContagionRun",
            subject_type="ContagionRun",
            subject_id=run_id,
            action="ContagionRunCompleted",
            correlation_id=correlation_id,
            payload={
                "seedEntityId": result["seedEntityId"],
                "infectedCount": result["infectedCount"],
                "pathExplainability": result["pathExplainability"][:20],
                "mode": result["mode"],
            },
            idempotency_key=f"audit-contagion-{run_id}",
        )
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=503, detail="audit publish failed") from exc
    return {
        "runId": run_id,
        "correlationId": correlation_id,
        "result": result,
        "evidenceRef": evidence_ref,
    }


def main():
    import uvicorn

    host, port = "0.0.0.0", 8094
    addr = config.HTTP_ADDR
    if addr.startswith(":"):
        port = int(addr[1:])
    elif ":" in addr:
        host, port_str = addr.rsplit(":", 1)
        port = int(port_str)
    uvicorn.run("simulation_service.main:app", host=host, port=port, reload=False)


if __name__ == "__main__":
    main()
