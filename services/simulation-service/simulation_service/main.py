from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from simulation_service import config
from simulation_service.audit import publish_simulation_run
from simulation_service.decision import evaluate_corep
from simulation_service.graph_client import GraphClient
from simulation_service.scenario import SCENARIO_ECB_ADVERSE_V1, run_ecb_adverse_v1, stable_run_id

app = FastAPI(title="Simulation Service", version="0.1.0")
graph_client = GraphClient()


class RunRequest(BaseModel):
    scenarioId: str = Field(default=SCENARIO_ECB_ADVERSE_V1)
    parameters: dict = Field(default_factory=dict)
    correlationId: str | None = None
    personaId: str = Field(default="44444444-4444-4444-4444-444444444401")


@app.get("/api/v1/health")
async def health():
    try:
        graph_health = await graph_client.health()
    except Exception as exc:  # noqa: BLE001
        raise HTTPException(status_code=503, detail={"status": "degraded", "graph": str(exc)}) from exc
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

    publish_simulation_run(run_id, correlation_id, metrics)

    return {
        "runId": run_id,
        "correlationId": correlation_id,
        "metrics": metrics,
        "decisions": decisions,
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
