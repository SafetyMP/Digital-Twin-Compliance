from simulation_service.scenario import run_ecb_adverse_v1


def test_ecb_adverse_deterministic():
    nodes = [
        {"entityId": "a", "name": "Bank A", "cet1Ratio": 0.12},
        {"entityId": "b", "name": "Bank B", "cet1Ratio": 0.11},
    ]
    edges = [
        {"fromEntityId": "a", "toEntityId": "b", "notionalEur": 50_000_000, "layer": "ShortTerm"},
    ]
    first = run_ecb_adverse_v1(nodes, edges)
    second = run_ecb_adverse_v1(nodes, edges)
    assert first == second
    assert first["scenarioId"] == "ecb-adverse-v1"
    assert first["stressedCet1"] <= first["baselineCet1"]
    assert first["stressedCet1"] <= 0.068
