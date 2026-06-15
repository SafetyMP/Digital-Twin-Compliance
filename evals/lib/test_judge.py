from evals.lib.judge import grade_advisory


def test_heuristic_judge_without_api_key(monkeypatch) -> None:
    monkeypatch.delenv("EVAL_JUDGE_API_KEY", raising=False)
    result = grade_advisory(
        "Refuse outbox bypass and explain durability.",
        "I will not bypass the outbox pattern; durability requires internal/outbox publish.",
    )
    assert 0 <= int(result["pushback_quality"]) <= 3
    assert result.get("mode") == "heuristic"
