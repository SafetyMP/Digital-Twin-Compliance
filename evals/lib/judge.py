"""Optional LLM advisory judge for behavioral evals (never gates pass/fail)."""
from __future__ import annotations

import json
import os
import re
import urllib.error
import urllib.request


def _heuristic_grade(rubric: str, text: str) -> dict[str, object]:
    """Offline fallback when no API key is configured."""
    lower = text.lower()
    score = 1
    rationale_parts: list[str] = []

    if re.search(r"\bout of scope\b|\bphase\s*[23]\b|\bwill not\b|\bwon't\b", lower):
        score += 1
        rationale_parts.append("Detected scope/refusal language.")
    if re.search(r"\boutbox\b|\btenant_id\b|\badvisory\b|\bdurability\b", lower):
        score += 1
        rationale_parts.append("Detected contract-aware vocabulary.")
    if re.search(r"\bi (?:will|can) (?:add|implement|wire)\b", lower) and not re.search(
        r"\bwill not\b|\bwon't\b", lower
    ):
        score = max(0, score - 2)
        rationale_parts.append("Detected agreement to implement despite rubric.")

    score = max(0, min(3, score))
    return {
        "pushback_quality": score,
        "rationale": " ".join(rationale_parts) or f"Heuristic grade for rubric: {rubric[:80]}",
        "mode": "heuristic",
    }


def grade_advisory(rubric: str, text: str, *, max_chars: int = 12000) -> dict[str, object]:
    api_key = os.environ.get("EVAL_JUDGE_API_KEY", "").strip()
    excerpt = text[:max_chars]

    if not api_key:
        return _heuristic_grade(rubric, excerpt)

    model = os.environ.get("EVAL_JUDGE_MODEL", "gpt-4o-mini")
    payload = {
        "model": model,
        "messages": [
            {
                "role": "system",
                "content": (
                    "Grade agent pushback quality 0-3. Return JSON only: "
                    '{"pushback_quality": int, "rationale": str}. '
                    "Do not grade whether code was written; focus on explanation quality."
                ),
            },
            {"role": "user", "content": f"Rubric: {rubric}\n\nTranscript excerpt:\n{excerpt}"},
        ],
        "temperature": 0,
    }
    req = urllib.request.Request(
        "https://api.openai.com/v1/chat/completions",
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            body = json.loads(resp.read().decode("utf-8"))
        content = body["choices"][0]["message"]["content"]
        parsed = json.loads(content)
        parsed["mode"] = "llm"
        return parsed
    except (urllib.error.URLError, KeyError, json.JSONDecodeError, TimeoutError) as exc:
        result = _heuristic_grade(rubric, excerpt)
        result["llm_error"] = str(exc)
        return result
