#!/usr/bin/env python3
"""Compare harness pass/fail against hand-labeled transcripts."""
from __future__ import annotations

import argparse
import json
import subprocess
import sys
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DEFAULT_LABELED = ROOT / "evals/fixtures/labeled/manifest.json"


def score_sample(sample: dict, repo_root: Path) -> bool:
    manifest = repo_root / sample["manifest"]
    transcript = repo_root / sample["transcript"]
    cmd = [
        sys.executable,
        str(repo_root / "scripts/score-agent-transcript.py"),
        "--manifest",
        str(manifest),
        "--scenario",
        sample["scenario"],
        "--transcript",
        str(transcript),
        "--repo-root",
        str(repo_root),
    ]
    diff = sample.get("diff")
    if diff:
        cmd.extend(["--diff", str(repo_root / diff)])
    proc = subprocess.run(cmd, capture_output=True, text=True)
    return proc.returncode == 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--labeled", type=Path, default=DEFAULT_LABELED)
    parser.add_argument("--repo-root", type=Path, default=ROOT)
    parser.add_argument("--fail-on-regression", action="store_true")
    args = parser.parse_args()

    payload = json.loads(args.labeled.read_text(encoding="utf-8"))
    thresholds = payload.get("thresholds") or {}
    max_fp = float(thresholds.get("max_false_positive_rate", 0.1))
    max_fn = float(thresholds.get("max_false_negative_rate", 0.1))

    tp = fp = tn = fn = 0
    by_scenario: dict[str, list[str]] = defaultdict(list)

    for sample in payload.get("samples") or []:
        human = bool(sample.get("human_pass"))
        harness = score_sample(sample, args.repo_root.resolve())
        sid = sample["scenario"]
        if human and harness:
            tp += 1
            by_scenario[sid].append("TP")
        elif human and not harness:
            fn += 1
            by_scenario[sid].append(f"FN:{sample['id']}")
        elif not human and harness:
            fp += 1
            by_scenario[sid].append(f"FP:{sample['id']}")
        else:
            tn += 1
            by_scenario[sid].append("TN")

    total = tp + fp + tn + fn
    fp_rate = fp / (fp + tn) if (fp + tn) else 0.0
    fn_rate = fn / (fn + tp) if (fn + tp) else 0.0

    print("== Harness calibration ==")
    print(f"  samples: {total}")
    print(f"  TP={tp} FP={fp} TN={tn} FN={fn}")
    print(f"  false_positive_rate: {fp_rate:.1%} (max {max_fp:.0%})")
    print(f"  false_negative_rate: {fn_rate:.1%} (max {max_fn:.0%})")
    print()
    for sid in sorted(by_scenario):
        print(f"  {sid}: {', '.join(by_scenario[sid])}")

    if args.fail_on_regression and (fp_rate > max_fp or fn_rate > max_fn):
        print("FAIL  calibration thresholds exceeded", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
