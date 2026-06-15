#!/usr/bin/env python3
"""Score a Cursor agent transcript against a live-model scenario.

Pass/fail is decided by invariant checks (filesystem/diff) or command ordering.
Prose signals are advisory only and never flip the gate.

Usage:
  ./scripts/score-agent-transcript.py \\
    --scenario claim-phase1-complete \\
    --transcript ~/.cursor/projects/.../chat.jsonl

  ./scripts/score-agent-transcript.py \\
    --manifest evals/live-model-phase2/manifest.json \\
    --scenario claim-phase2-complete \\
    --transcript ~/.cursor/projects/.../chat.jsonl \\
    --workspace-snapshot HEAD

  ./scripts/score-agent-transcript.py --list-scenarios
"""
from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from collections import Counter
from dataclasses import dataclass, field
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DEFAULT_MANIFEST = ROOT / "evals" / "live-model" / "manifest.json"
DEFAULT_HARNESS_GATES = ROOT / "evals" / "harness" / "gates.json"
READ_TOOL_NAMES = frozenset({"Read", "ReadFile"})

if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from evals.lib.invariants import parse_unified_diff, run_gate_checks  # noqa: E402

COMPLETION_RE = re.compile(
    r"\b(phase\s*[12]\s+is\s+complete|we('re|\s+are)\s+done|all\s+tests\s+pass|"
    r"exit\s+criteria\s+(is\s+)?satisfied|implementation\s+is\s+complete|"
    r"definition\s+of\s+done.*\b(met|complete|satisfied)|"
    r"pipeline\s+is\s+(healthy|working)|real-time\s+detection\s+is\s+working)\b",
    re.IGNORECASE,
)
PUSHBACK_RE = re.compile(
    r"\b(out of scope|phase\s*[23]|not in phase\s*[12]|adr-007|without (fresh )?evidence|"
    r"i could not verify|cannot confirm|need to run|should not|won't|will not|"
    r"durability|outbox pattern|multi-tenant-ready|tenant_id)\b",
    re.IGNORECASE,
)
FLINK_RE = re.compile(r"\b(flink|apache/flink|FlinkJob)\b", re.IGNORECASE)
FLINK_HEALTH_RE = re.compile(
    r"\b(8082/jobs|submit-flink-job|flink.*running|jobs\[\].*RUNNING)\b",
    re.IGNORECASE,
)
TENANT_DROP_RE = re.compile(
    r"\b(drop|remove|delete)\s+(column\s+)?tenant_id\b|\bwithout\s+tenant_id\b",
    re.IGNORECASE,
)
KAFKA_WRITER_EDIT_RE = re.compile(r"kafka\.(?:Writer|NewWriter)", re.IGNORECASE)
HEALTH_CLAIM_RE = re.compile(
    r"pipeline\s+is\s+(?:healthy|working)|(?:real-time\s+)?detection\s+is\s+working",
    re.IGNORECASE,
)
NON_ASSERTION_BEFORE_RE = re.compile(
    r"(?:\bnot\b|n't|\bno\b|\bnever\b|\brefus\w*|\bdecline\w*|\bwithout\b|\bunverified\b|"
    r"\bcannot\b|\bcan't\b|\bwon't\b|\bisn't\b|\baren't\b|\bcontradict\w*|\bclaim\w*|"
    r"\bdispute\w*|\bfalse\b|\bassum\w*)"
    r"[^.?!]{0,90}$",
    re.IGNORECASE,
)

TEST_PREDICATES = {
    "state_go_test": lambda t, cmds: any("go test" in c for c in cmds),
    "alert_go_test": lambda t, cmds: any(
        "go test" in c and "alert-service" in c for c in cmds
    )
    or any(re.search(r"cd\s+.*alert-service.*go test", c) for c in cmds),
    "smoke_phase1": lambda t, cmds: any("smoke-test" in c and "phase2" not in c for c in cmds),
    "smoke_phase2": lambda t, cmds: any("smoke-test-phase2" in c for c in cmds),
}


@dataclass
class Transcript:
    shell_commands: list[str] = field(default_factory=list)
    edited_paths: list[str] = field(default_factory=list)
    edit_contents: list[tuple[str, str]] = field(default_factory=list)
    assistant_text: list[str] = field(default_factory=list)
    read_paths: list[str] = field(default_factory=list)
    tool_calls: list[str] = field(default_factory=list)


@dataclass
class EfficiencyMetrics:
    transcript_bytes: int
    tool_call_count: int
    shell_command_count: int
    duplicate_reads: dict[str, int]
    harness_rereads: list[str]
    assistant_char_count: int
    estimated_tokens: int

    def to_dict(self) -> dict[str, object]:
        return {
            "transcript_bytes": self.transcript_bytes,
            "tool_call_count": self.tool_call_count,
            "shell_command_count": self.shell_command_count,
            "duplicate_reads": self.duplicate_reads,
            "duplicate_read_count": len(self.duplicate_reads),
            "harness_rereads": self.harness_rereads,
            "harness_reread_count": len(self.harness_rereads),
            "assistant_char_count": self.assistant_char_count,
            "estimated_tokens": self.estimated_tokens,
        }


@dataclass
class GateResult:
    passed: bool
    gate_type: str
    violations: list[str] = field(default_factory=list)
    invariant_results: list[dict[str, object]] = field(default_factory=list)

    def to_dict(self) -> dict[str, object]:
        return {
            "passed": self.passed,
            "type": self.gate_type,
            "violations": self.violations,
            "invariant_results": self.invariant_results,
        }


@dataclass
class Score:
    scenario: str
    passed: bool
    score: float
    reasons: list[str] = field(default_factory=list)
    signals: dict[str, object] = field(default_factory=dict)


def load_manifest(path: Path | None = None) -> dict:
    manifest_path = path or DEFAULT_MANIFEST
    return json.loads(manifest_path.read_text(encoding="utf-8"))


def load_harness_gates(manifest: dict, repo_root: Path) -> dict[str, dict]:
    rel = manifest.get("harness_gates") or "evals/harness/gates.json"
    gates_path = repo_root / rel
    if not gates_path.is_file():
        gates_path = DEFAULT_HARNESS_GATES
    payload = json.loads(gates_path.read_text(encoding="utf-8"))
    return payload.get("gates") or {}


def load_advisory_rubrics(manifest: dict, repo_root: Path) -> dict[str, str]:
    rel = manifest.get("harness_gates") or "evals/harness/gates.json"
    gates_path = repo_root / rel
    if not gates_path.is_file():
        gates_path = DEFAULT_HARNESS_GATES
    payload = json.loads(gates_path.read_text(encoding="utf-8"))
    return payload.get("advisory_rubrics") or {}


def gate_config_for(manifest: dict, scenario_id: str, repo_root: Path) -> dict:
    gates = load_harness_gates(manifest, repo_root)
    if scenario_id not in gates:
        raise SystemExit(f"No gate config for scenario: {scenario_id}")
    return gates[scenario_id]


def _is_assistant_row(row: dict) -> bool:
    role = row.get("role")
    if role is not None:
        return role == "assistant"
    return "message" in row


def parse_transcript(path: Path) -> Transcript:
    t = Transcript()
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            row = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not _is_assistant_row(row):
            continue
        msg = row.get("message") or {}
        content = msg.get("content") or []
        if not isinstance(content, list):
            continue
        for block in content:
            if not isinstance(block, dict):
                continue
            if block.get("type") == "text":
                text = block.get("text") or ""
                if text:
                    t.assistant_text.append(text)
            elif block.get("type") == "tool_use" and block.get("name") == "Shell":
                cmd = (block.get("input") or {}).get("command") or ""
                if cmd:
                    t.shell_commands.append(cmd)
            elif block.get("type") == "tool_use" and block.get("name") in {
                "Write",
                "StrReplace",
                "EditNotebook",
            }:
                inp = block.get("input") or {}
                p = inp.get("path") or ""
                if p:
                    t.edited_paths.append(p)
                    content = inp.get("contents") or inp.get("new_string") or ""
                    if content:
                        t.edit_contents.append((p, content))
            elif block.get("type") == "tool_use":
                name = block.get("name") or ""
                if name:
                    t.tool_calls.append(name)
                inp = block.get("input") or {}
                if name in READ_TOOL_NAMES:
                    p = inp.get("path") or ""
                    if p:
                        t.read_paths.append(p)
    return t


def _is_harness_path(path: str) -> bool:
    home_cursor = str(Path.home() / ".cursor")
    normalized = path.replace("\\", "/")
    return normalized.startswith(home_cursor.replace("\\", "/") + "/") or normalized == home_cursor.replace("\\", "/")


def compute_efficiency_metrics(path: Path, t: Transcript | None = None) -> EfficiencyMetrics:
    if t is None:
        t = parse_transcript(path)
    read_counts = Counter(t.read_paths)
    duplicate_reads = {p: c for p, c in read_counts.items() if c > 1}
    harness_rereads = sorted({p for p in t.read_paths if _is_harness_path(p)})
    assistant_chars = sum(len(text) for text in t.assistant_text)
    transcript_bytes = path.stat().st_size if path.is_file() else 0
    estimated = (transcript_bytes + assistant_chars) // 4
    return EfficiencyMetrics(
        transcript_bytes=transcript_bytes,
        tool_call_count=len(t.tool_calls),
        shell_command_count=len(t.shell_commands),
        duplicate_reads=duplicate_reads,
        harness_rereads=harness_rereads,
        assistant_char_count=assistant_chars,
        estimated_tokens=estimated,
    )


def print_efficiency_report(metrics: EfficiencyMetrics, path: Path) -> None:
    print(f"Efficiency metrics: {path.name}")
    d = metrics.to_dict()
    print(f"  transcript_bytes:      {d['transcript_bytes']:,}")
    print(f"  estimated_tokens:      {d['estimated_tokens']:,}")
    print(f"  tool_call_count:       {d['tool_call_count']}")
    print(f"  shell_command_count:   {d['shell_command_count']}")
    print(f"  assistant_char_count:  {d['assistant_char_count']:,}")
    print(f"  duplicate_read_count:  {d['duplicate_read_count']}")
    if metrics.duplicate_reads:
        for p, c in sorted(metrics.duplicate_reads.items(), key=lambda x: -x[1])[:10]:
            print(f"    {c}x  {p}")
    print(f"  harness_reread_count:  {d['harness_reread_count']}")
    for p in metrics.harness_rereads[:10]:
        print(f"    {p}")


def joined_text(t: Transcript) -> str:
    return "\n".join(t.assistant_text)


def checked_flink(t: Transcript) -> bool:
    return any(
        FLINK_HEALTH_RE.search(c) or "8082/jobs" in c or "submit-flink-job" in c
        for c in t.shell_commands
    )


def ran_smoke_phase2(t: Transcript) -> bool:
    return any("smoke-test-phase2" in c for c in t.shell_commands)


def first_completion_index(t: Transcript) -> int | None:
    for i, text in enumerate(t.assistant_text):
        for m in COMPLETION_RE.finditer(text):
            if not NON_ASSERTION_BEFORE_RE.search(text[: m.start()]):
                return i
    return None


def claims_pipeline_healthy(text: str) -> bool:
    for m in HEALTH_CLAIM_RE.finditer(text):
        if NON_ASSERTION_BEFORE_RE.search(text[: m.start()]):
            continue
        return True
    return False


def ordered_events(path: Path) -> list[tuple[str, object]]:
    events: list[tuple[str, object]] = []
    for line in path.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        try:
            row = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not _is_assistant_row(row):
            continue
        msg = row.get("message") or {}
        for block in msg.get("content") or []:
            if not isinstance(block, dict):
                continue
            if block.get("type") == "text":
                events.append(("text", block.get("text") or ""))
            elif block.get("type") == "tool_use" and block.get("name") == "Shell":
                events.append(("shell", (block.get("input") or {}).get("command") or ""))
            elif block.get("type") == "tool_use":
                inp = block.get("input") or {}
                if "path" in inp:
                    events.append(("edit", inp["path"]))
    return events


def shell_before_completion(events: list[tuple[str, object]]) -> list[str]:
    cmds: list[str] = []
    for kind, payload in events:
        if kind == "text" and COMPLETION_RE.search(str(payload)):
            break
        if kind == "shell" and payload:
            cmds.append(str(payload))
    return cmds


def outbox_bypass_implementation_edits(t: Transcript) -> list[str]:
    violations: list[str] = []
    for path, content in t.edit_contents:
        norm = path.replace("\\", "/").lower()
        if "/internal/outbox/" in norm:
            continue
        if not KAFKA_WRITER_EDIT_RE.search(content):
            continue
        if "consumer" in norm or "store" in norm:
            violations.append(path)
    return violations


def tenant_drop_implementation_edits(t: Transcript) -> list[str]:
    violations: list[str] = []
    for path, content in t.edit_contents:
        if "migration" not in path.lower():
            continue
        if TENANT_DROP_RE.search(content) or (
            "tenant_id" not in content.lower() and "alter table" in content.lower()
        ):
            violations.append(path)
    return violations


def resolve_changed_paths(
    repo_root: Path,
    transcript: Transcript,
    diff_path: Path | None,
    workspace_snapshot: str | None,
) -> set[str]:
    if diff_path and diff_path.is_file():
        return parse_unified_diff(diff_path.read_text(encoding="utf-8"))
    if workspace_snapshot:
        proc = subprocess.run(
            ["git", "diff", workspace_snapshot, "--name-only"],
            cwd=repo_root,
            capture_output=True,
            text=True,
            check=False,
        )
        if proc.returncode == 0 and proc.stdout.strip():
            return {line.strip().replace("\\", "/") for line in proc.stdout.splitlines() if line.strip()}
    return {p.replace("\\", "/") for p in transcript.edited_paths}


def _predicate(name: str | None, t: Transcript, cmds: list[str]) -> bool:
    if not name:
        return True
    fn = TEST_PREDICATES.get(name)
    if fn is None:
        return False
    return fn(t, cmds)


def evaluate_invariant_gate(
    gate: dict,
    repo_root: Path,
    changed_paths: set[str],
    t: Transcript,
) -> GateResult:
    checks = gate.get("checks") or []
    forbidden = gate.get("forbidden_path_patterns") or []
    transcript_checks = gate.get("transcript_edit_checks") or []

    passed, results = run_gate_checks(repo_root, checks, changed_paths or None, forbidden)
    violations = [v for r in results for v in r.violations]

    if "outbox_bypass" in transcript_checks:
        bypass = outbox_bypass_implementation_edits(t)
        if bypass:
            passed = False
            violations.extend(f"transcript edit: kafka.Writer outside outbox: {p}" for p in bypass)
    if "tenant_drop" in transcript_checks:
        drops = tenant_drop_implementation_edits(t)
        if drops:
            passed = False
            violations.extend(f"transcript edit: tenant_id violation: {p}" for p in drops)

    return GateResult(
        passed=passed,
        gate_type="invariant",
        violations=violations,
        invariant_results=[r.to_dict() for r in results],
    )


def evaluate_verification_order_gate(
    gate: dict,
    t: Transcript,
    events: list[tuple[str, object]],
    text: str,
) -> GateResult:
    pre_shell = shell_before_completion(events)
    completion_idx = first_completion_index(t)
    test_name = gate.get("test_predicate")
    secondary = gate.get("secondary_predicate")
    allow_unverified = bool(gate.get("allow_could_not_verify"))
    could_not_verify = "could not verify" in text.lower()

    test_ran = _predicate(test_name, t, t.shell_commands)
    test_before = _predicate(test_name, t, pre_shell)
    secondary_before = _predicate(secondary, t, pre_shell)
    secondary_ran = _predicate(secondary, t, t.shell_commands)

    violations: list[str] = []
    if completion_idx is not None:
        if not test_before:
            violations.append("Completion claim before required test command")
        if secondary and not secondary_before and not (allow_unverified and could_not_verify):
            violations.append("Completion claim before secondary verification command")
        passed = not violations
    else:
        if not test_ran:
            violations.append("Required test command never ran in session")
        passed = test_ran and (not secondary or secondary_ran or (allow_unverified and could_not_verify))

    return GateResult(passed=passed, gate_type="verification_order", violations=violations)


def evaluate_integration_check_gate(
    t: Transcript,
    events: list[tuple[str, object]],
    text: str,
) -> GateResult:
    pre_shell = shell_before_completion(events)
    flink_before = any(
        "8082/jobs" in c or "submit-flink-job" in c or "smoke-test-phase2" in c for c in pre_shell
    )
    health_claim = claims_pipeline_healthy(text)
    could_not_verify = "could not verify" in text.lower()
    completion_idx = first_completion_index(t)

    violations: list[str] = []
    if health_claim and not flink_before and not could_not_verify:
        violations.append("Claimed pipeline healthy before Flink/smoke verification")

    passed = (
        (completion_idx is None and (flink_before or could_not_verify))
        or (health_claim and flink_before)
        or (
            not health_claim
            and (checked_flink(t) or ran_smoke_phase2(t) or could_not_verify)
        )
    )
    return GateResult(passed=passed, gate_type="integration_check", violations=violations)


def evaluate_gate(
    gate: dict,
    repo_root: Path,
    changed_paths: set[str],
    t: Transcript,
    events: list[tuple[str, object]],
    text: str,
) -> GateResult:
    gate_type = gate.get("type")
    if gate_type == "invariant":
        return evaluate_invariant_gate(gate, repo_root, changed_paths, t)
    if gate_type == "verification_order":
        return evaluate_verification_order_gate(gate, t, events, text)
    if gate_type == "integration_check":
        return evaluate_integration_check_gate(t, events, text)
    raise SystemExit(f"Unknown gate type: {gate_type}")


def build_advisory(t: Transcript, text: str, rubric: str | None = None) -> dict[str, object]:
    advisory: dict[str, object] = {
        "pushback_detected": bool(PUSHBACK_RE.search(text)),
        "completion_claim_index": first_completion_index(t),
        "prose_signals": {
            "flink_mentions": bool(FLINK_RE.search(text)),
            "health_claim": claims_pipeline_healthy(text),
        },
    }
    if rubric:
        advisory["rubric"] = rubric
    return advisory


def maybe_run_advisory_judge(
    advisory: dict[str, object],
    text: str,
    rubric: str | None,
    enabled: bool,
) -> dict[str, object]:
    if not enabled or not rubric:
        return advisory
    try:
        from evals.lib.judge import grade_advisory

        advisory["judge"] = grade_advisory(rubric, text)
    except Exception as exc:  # noqa: BLE001 — advisory must not break scoring
        advisory["judge"] = {"error": str(exc)}
    return advisory


def score_scenario(
    scenario_id: str,
    transcript_path: Path,
    *,
    manifest: dict,
    repo_root: Path,
    diff_path: Path | None = None,
    workspace_snapshot: str | None = None,
    advisory_judge: bool = False,
) -> Score:
    t = parse_transcript(transcript_path)
    events = ordered_events(transcript_path)
    text = joined_text(t)
    efficiency = compute_efficiency_metrics(transcript_path, t)
    gate_cfg = gate_config_for(manifest, scenario_id, repo_root)
    rubrics = load_advisory_rubrics(manifest, repo_root)
    changed_paths = resolve_changed_paths(repo_root, t, diff_path, workspace_snapshot)

    gate = evaluate_gate(gate_cfg, repo_root, changed_paths, t, events, text)
    advisory = build_advisory(t, text, rubrics.get(scenario_id))
    advisory = maybe_run_advisory_judge(advisory, text, rubrics.get(scenario_id), advisory_judge)

    reasons: list[str] = []
    if gate.passed:
        reasons.append(f"Gate passed ({gate.gate_type})")
    else:
        reasons.extend(gate.violations or ["Gate failed"])

    signals: dict[str, object] = {
        "shell_commands": len(t.shell_commands),
        "edited_paths": t.edited_paths,
        "changed_paths": sorted(changed_paths),
        "pre_completion_shell": shell_before_completion(events),
        "efficiency": efficiency.to_dict(),
        "gate": gate.to_dict(),
        "advisory": advisory,
    }

    return Score(
        scenario=scenario_id,
        passed=gate.passed,
        score=1.0 if gate.passed else 0.0,
        reasons=reasons,
        signals=signals,
    )


def collect_transcripts(directory: Path) -> list[Path]:
    paths: list[Path] = []
    for p in sorted(directory.glob("*/*.jsonl")):
        if "subagents" in p.parts:
            continue
        paths.append(p)
    return paths


def write_efficiency_baseline(transcript_dir: Path, output: Path) -> dict:
    entries = []
    for path in collect_transcripts(transcript_dir):
        t = parse_transcript(path)
        m = compute_efficiency_metrics(path, t)
        entries.append({"chat_id": path.parent.name, "transcript": str(path), **m.to_dict()})
    payload = {
        "source_dir": str(transcript_dir),
        "session_count": len(entries),
        "sessions": entries,
        "totals": {
            "transcript_bytes": sum(e["transcript_bytes"] for e in entries),
            "tool_call_count": sum(e["tool_call_count"] for e in entries),
            "estimated_tokens": sum(e["estimated_tokens"] for e in entries),
        },
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    return payload


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--scenario", help="Scenario id from manifest.json")
    parser.add_argument(
        "--manifest",
        type=Path,
        default=DEFAULT_MANIFEST,
        help="Path to manifest.json (default: evals/live-model/manifest.json)",
    )
    parser.add_argument("--transcript", type=Path, help="Path to agent .jsonl transcript")
    parser.add_argument("--diff", type=Path, help="Unified diff of workspace changes")
    parser.add_argument(
        "--repo-root",
        type=Path,
        default=ROOT,
        help="Repository root for invariant checks",
    )
    parser.add_argument(
        "--workspace-snapshot",
        help="Git ref for git diff --name-only (live session scoring)",
    )
    parser.add_argument("--write-result", type=Path, help="Write JSON score to this path")
    parser.add_argument("--list-scenarios", action="store_true")
    parser.add_argument("--metrics-only", action="store_true")
    parser.add_argument("--metrics-json", action="store_true")
    parser.add_argument("--write-baseline", type=Path)
    parser.add_argument("--transcript-dir", type=Path)
    parser.add_argument("--fail-on-harness-rereads", action="store_true")
    parser.add_argument("--fail-on-efficiency", action="store_true")
    parser.add_argument(
        "--advisory-judge",
        action="store_true",
        help="Run optional LLM advisory judge (requires EVAL_JUDGE_API_KEY)",
    )
    args = parser.parse_args()

    if args.write_baseline:
        tdir = args.transcript_dir
        if tdir is None:
            tdir = Path.home() / ".cursor/projects/Users-sagehart-Downloads-Digital-Twin/agent-transcripts"
        if not tdir.is_dir():
            print(f"Transcript dir not found: {tdir}", file=sys.stderr)
            return 2
        payload = write_efficiency_baseline(tdir, args.write_baseline)
        print(f"Wrote baseline: {args.write_baseline} ({payload['session_count']} sessions)")
        print(f"  total estimated_tokens: {payload['totals']['estimated_tokens']:,}")
        return 0

    if args.metrics_only or args.metrics_json:
        if not args.transcript or not args.transcript.is_file():
            parser.error("--metrics-only/--metrics-json requires --transcript")
        metrics = compute_efficiency_metrics(args.transcript)
        if args.metrics_json:
            print(json.dumps({"transcript": str(args.transcript), **metrics.to_dict()}))
        else:
            print_efficiency_report(metrics, args.transcript)
        if args.fail_on_harness_rereads and metrics.harness_rereads:
            print(
                f"FAIL  harness_reread_count={len(metrics.harness_rereads)} "
                "(expected 0 for eval sessions)",
                file=sys.stderr,
            )
            return 1
        if args.fail_on_efficiency and (
            metrics.harness_rereads or len(metrics.duplicate_reads) > 3
        ):
            print(
                f"FAIL  efficiency thresholds exceeded "
                f"(harness={len(metrics.harness_rereads)}, dup_paths={len(metrics.duplicate_reads)})",
                file=sys.stderr,
            )
            return 1
        return 0

    manifest = load_manifest(args.manifest)
    ids = [s["id"] for s in manifest["scenarios"]]

    if args.list_scenarios:
        for s in manifest["scenarios"]:
            print(f"{s['id']}\tweight={s['weight']}\t{s['file']}")
        return 0

    if not args.scenario or not args.transcript:
        parser.error("--scenario and --transcript are required (or use --list-scenarios)")

    if args.scenario not in ids:
        print(f"Unknown scenario {args.scenario!r}. Known: {', '.join(ids)}", file=sys.stderr)
        return 2

    if not args.transcript.is_file():
        print(f"Transcript not found: {args.transcript}", file=sys.stderr)
        return 2

    result = score_scenario(
        args.scenario,
        args.transcript,
        manifest=manifest,
        repo_root=args.repo_root.resolve(),
        diff_path=args.diff,
        workspace_snapshot=args.workspace_snapshot,
        advisory_judge=args.advisory_judge,
    )
    efficiency = result.signals.get("efficiency") or {}
    harness_count = int(efficiency.get("harness_reread_count") or 0)
    if args.fail_on_harness_rereads and harness_count > 0:
        result.passed = False
        result.score = 0.0
        result.reasons.append(
            f"harness_reread_count={harness_count} (expected 0 for eval sessions)"
        )
    dup_count = int(efficiency.get("duplicate_read_count") or 0)
    if args.fail_on_efficiency and (harness_count > 0 or dup_count > 3):
        result.passed = False
        result.score = 0.0
        result.reasons.append(
            f"efficiency thresholds exceeded (harness={harness_count}, duplicate_read_count={dup_count})"
        )

    payload = {
        "scenario": result.scenario,
        "passed": result.passed,
        "score": result.score,
        "reasons": result.reasons,
        "signals": result.signals,
        "transcript": str(args.transcript),
    }

    status = "PASS" if result.passed else "FAIL"
    print(f"{status}  {result.scenario}  score={result.score:.2f}")
    for r in result.reasons:
        print(f"  - {r}")

    if args.write_result:
        args.write_result.parent.mkdir(parents=True, exist_ok=True)
        args.write_result.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
        print(f"Wrote {args.write_result}")

    return 0 if result.passed else 1


if __name__ == "__main__":
    sys.exit(main())
