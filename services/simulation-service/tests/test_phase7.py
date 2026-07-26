from pathlib import Path

from simulation_service.contagion import run_contagion
from simulation_service.effectiveness import load_breaches, replay


def test_effectiveness_metrics():
    path = Path(__file__).resolve().parents[3] / "fixtures" / "phase7" / "breaches.json"
    breaches = load_breaches(path)
    metrics = replay(breaches, {"INT-M001", "INT-M002", "BASEL-M001"})
    assert metrics["detected"] == 4
    assert metrics["missed"] == 1
    assert metrics["coverage"] >= 0.8


def test_contagion_bfs():
    nodes = [{"entityId": "a", "name": "A"}, {"entityId": "b", "name": "B"}]
    edges = [{"fromEntityId": "a", "toEntityId": "b"}]
    result = run_contagion(nodes, edges, "a", max_hops=2)
    assert result["infectedCount"] == 2
    assert result["pathExplainability"][0]["hop"] == 0
