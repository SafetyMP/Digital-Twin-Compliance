"""Control-effectiveness twin replay (ADR-013 D36)."""

from __future__ import annotations

import json
from pathlib import Path


def load_breaches(path: str | Path) -> list[dict]:
    data = json.loads(Path(path).read_text(encoding="utf-8"))
    return list(data.get("breaches", data))


def replay(breaches: list[dict], detections: set[str]) -> dict:
    """Compare labeled breach IDs to detected control IDs; emit coverage/miss."""
    total = len(breaches)
    hit = 0
    misses: list[str] = []
    for b in breaches:
        bid = str(b.get("id"))
        control = str(b.get("expectedControl", ""))
        if control in detections or bid in detections:
            hit += 1
        else:
            misses.append(bid)
    coverage = (hit / total) if total else 0.0
    miss_rate = 1.0 - coverage
    return {
        "totalBreaches": total,
        "detected": hit,
        "missed": len(misses),
        "coverage": round(coverage, 4),
        "missRate": round(miss_rate, 4),
        "missIds": misses,
        "detections": sorted(detections),
    }
