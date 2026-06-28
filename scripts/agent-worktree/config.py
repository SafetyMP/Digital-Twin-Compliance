#!/usr/bin/env python3
"""Read .cursor/worktrees.config.json for scope, compose, and test mapping."""
from __future__ import annotations

import fnmatch
import json
import os
import subprocess
import sys
from pathlib import Path


def load_config(path: Path) -> dict:
    if path.is_file():
        return json.loads(path.read_text(encoding="utf-8"))
    return {}


def infer_track(branch: str, prefix: str = "agent") -> str:
    if branch.startswith(f"{prefix}/backend/"):
        return "backend"
    if branch.startswith(f"{prefix}/frontend/"):
        return "frontend"
    if branch.startswith(f"{prefix}/docs/"):
        return "docs"
    return "experiment"


def changed_files(base: str, branch: str) -> list[str]:
    for spec in (f"{base}...{branch}",):
        r = subprocess.run(
            ["git", "diff", "--name-only", spec],
            capture_output=True,
            text=True,
            check=False,
        )
        if r.returncode == 0 and r.stdout.strip():
            return [ln for ln in r.stdout.splitlines() if ln]
    r = subprocess.run(
        ["git", "diff", "--name-only", base, branch],
        capture_output=True,
        text=True,
        check=False,
    )
    return [ln for ln in (r.stdout or "").splitlines() if ln]


def _matches_deny(path: str, deny_globs: list[str]) -> bool:
    for g in deny_globs:
        if fnmatch.fnmatch(path, g) or fnmatch.fnmatch(path, f"*{g}*"):
            return True
        if path.startswith(g.rstrip("*")):
            return True
    return False


def _matches_allow(path: str, allow_prefixes: list[str]) -> bool:
    if "*" in allow_prefixes:
        return True
    return any(path.startswith(p) for p in allow_prefixes)


def check_scope(
    branch: str,
    base: str,
    config: dict,
    track: str | None = None,
    strict: bool = False,
) -> int:
    prefix = config.get("branch_prefix", "agent")
    track = track or infer_track(branch, prefix)
    tracks = config.get("tracks") or {}
    rules = tracks.get(track) or tracks.get("experiment") or {}
    allow = rules.get("allow_prefixes", ["*"])
    deny = rules.get("deny_globs", [])
    warn_only = bool(rules.get("warn_only")) and track == "experiment" and not strict

    files = changed_files(base, branch)
    if not files:
        print(f"check-worktree-scope: no file changes on {branch} vs {base}")
        return 0

    violations: list[str] = []
    warnings: list[str] = []
    for f in files:
        if deny and _matches_deny(f, deny):
            msg = f"{f} (deny pattern)"
            if warn_only:
                warnings.append(msg)
            else:
                violations.append(msg)
        if allow and not _matches_allow(f, allow) and track != "experiment":
            violations.append(f)

    print(f"Scope check: branch={branch} track={track} base={base} files={len(files)}")
    for w in warnings:
        print(f"  WARN {w}")
    for v in violations:
        print(f"  VIOLATION {v}")

    if violations:
        print("check-worktree-scope: FAIL")
        return 1
    if warnings:
        print("check-worktree-scope: PASS with warnings")
        return 0
    print("check-worktree-scope: PASS")
    return 0


def compose_services(files: list[str], config: dict) -> list[str]:
    mapping = config.get("path_compose_services") or {}
    services: list[str] = []
    seen: set[str] = set()
    for f in files:
        for pattern, svcs in mapping.items():
            if fnmatch.fnmatch(f, pattern):
                for s in svcs:
                    if s not in seen:
                        seen.add(s)
                        services.append(s)
    return services


def package_tests(files: list[str], config: dict) -> list[str]:
    mapping = config.get("path_package_tests") or {}
    cmds: list[str] = []
    seen: set[str] = set()
    for f in files:
        for pattern, cmd in mapping.items():
            if fnmatch.fnmatch(f, pattern) and cmd not in seen:
                seen.add(cmd)
                cmds.append(cmd)
    return cmds


def contract_checks(files: list[str], config: dict) -> list[str]:
    mapping = config.get("path_contract_checks") or {}
    cmds: list[str] = []
    seen: set[str] = set()
    for f in files:
        for pattern, cmd in mapping.items():
            if fnmatch.fnmatch(f, pattern) and cmd not in seen:
                seen.add(cmd)
                cmds.append(cmd)
    return cmds


def get_waves(config: dict) -> list[dict]:
    waves = config.get("waves") or []
    return sorted(waves, key=lambda w: (w.get("order", 0), w.get("id", "")))


def validate_waves_config(config: dict) -> list[str]:
    waves = get_waves(config)
    if not waves:
        return []
    errors: list[str] = []
    ids = {w.get("id") for w in waves}
    for w in waves:
        wid = w.get("id")
        if not wid:
            errors.append("wave missing id")
            continue
        for dep in w.get("depends_on") or []:
            if dep not in ids:
                errors.append(f"wave {wid}: unknown dependency {dep}")
    graph = {w["id"]: w.get("depends_on") or [] for w in waves if w.get("id")}

    def has_cycle(node: str, visiting: set[str], visited: set[str]) -> bool:
        if node in visiting:
            return True
        if node in visited:
            return False
        visiting.add(node)
        for nxt in graph.get(node, []):
            if has_cycle(nxt, visiting, visited):
                return True
        visiting.remove(node)
        visited.add(node)
        return False

    visited: set[str] = set()
    for wid in graph:
        if has_cycle(wid, set(), visited):
            errors.append(f"dependency cycle involving {wid}")
            break
    return errors


def load_wave_state(path: Path) -> dict:
    if path.is_file():
        return json.loads(path.read_text(encoding="utf-8"))
    return {}


def wave_state_path(root: Path, task: str) -> Path:
    return root / ".worktrees" / "waves" / task / "state.json"


def wave_completed(state: dict, wave_id: str) -> bool:
    entry = (state.get("completed") or {}).get(wave_id) or {}
    return bool(entry.get("merged") or entry.get("skipped"))


def wave_ready(wave_id: str, state: dict, config: dict) -> tuple[bool, list[str]]:
    waves = {w["id"]: w for w in get_waves(config)}
    if wave_id not in waves:
        return False, [f"unknown wave: {wave_id}"]
    missing: list[str] = []
    for dep in waves[wave_id].get("depends_on") or []:
        if wave_completed(state, dep):
            continue
        dep_wave = waves.get(dep) or {}
        if dep_wave.get("optional"):
            continue
        missing.append(dep)
    return (len(missing) == 0, missing)


def cmd_waves_validate(config: dict) -> int:
    errors = validate_waves_config(config)
    if errors:
        for e in errors:
            print(f"  ERROR {e}")
        print("waves validate: FAIL")
        return 1
    n = len(get_waves(config))
    print(f"waves validate: PASS ({n} waves)")
    return 0


def cmd_waves_plan(config: dict) -> int:
    waves = get_waves(config)
    if not waves:
        print("waves plan: no waves defined in config")
        return 0
    print("Dependency waves (in order):")
    for w in waves:
        deps = ", ".join(w.get("depends_on") or []) or "(none)"
        runner = w.get("runner", "child")
        opt = " optional" if w.get("optional") else ""
        par = " parallel" if w.get("parallel") else ""
        print(f"  [{w.get('order', '?')}] {w['id']}{opt}{par}  runner={runner}  depends=[{deps}]")
        if w.get("description"):
            print(f"      {w['description']}")
        if w.get("tracks"):
            print(f"      tracks: {', '.join(w['tracks'])}")
        if w.get("path_hints"):
            print(f"      paths: {', '.join(w['path_hints'])}")
    return 0


def cmd_waves_status(state: dict, config: dict, task: str) -> int:
    waves = get_waves(config)
    completed = state.get("completed") or {}
    print(f"Wave task: {task}")
    print(f"Base: {state.get('base', 'HEAD')}")
    for w in waves:
        wid = w["id"]
        entry = completed.get(wid)
        if entry and entry.get("skipped"):
            status = "skipped"
        elif entry and entry.get("merged"):
            status = f"done ({entry.get('branch', 'merged')})"
        elif entry:
            status = "in progress"
        else:
            ready, missing = wave_ready(wid, state, config)
            status = "ready" if ready else f"blocked ({', '.join(missing)})"
        print(f"  {wid}: {status}")
    return 0


def cmd_waves_ready(wave_id: str, state: dict, config: dict) -> int:
    ready, missing = wave_ready(wave_id, state, config)
    if ready:
        print(f"wave {wave_id}: ready")
        return 0
    print(f"wave {wave_id}: blocked — waiting on {', '.join(missing)}")
    return 1


def cmd_waves_handoff(wave_id: str, config: dict, task: str) -> int:
    waves = {w["id"]: w for w in get_waves(config)}
    w = waves.get(wave_id)
    if not w:
        print(f"unknown wave: {wave_id}", file=sys.stderr)
        return 1
    runner = w.get("runner", "child")
    if runner == "parent":
        print(f"Wave {wave_id} is parent-only — run from main root, do not spawn children.")
        return 0
    tracks = ", ".join(w.get("tracks") or ["experiment"])
    print(f"=== Wave handoff: {wave_id} (task: {task}) ===")
    print(w.get("description") or "")
    print()
    print(f"Tracks: {tracks}")
    if w.get("path_hints"):
        print(f"Path hints: {', '.join(w['path_hints'])}")
    if w.get("child_verify"):
        print("Child verify before return:")
        for c in w["child_verify"]:
            print(f"  ./{c}")
    print()
    print("Spawn only after parent confirms prior waves complete:")
    for dep in w.get("depends_on") or []:
        print(f"  - {dep}")
    print()
    print(f"  ./scripts/agent-worktree.sh create --track <track> --name {task}-{wave_id}")
    return 0


def handle_waves_cmd(args: list[str]) -> int:
    sub = args[0] if args else "plan"
    if sub in ("validate", "plan"):
        cfg_path = Path(args[1]) if len(args) > 1 else Path(".cursor/worktrees.config.json")
        config = load_config(cfg_path)
        if sub == "validate":
            return cmd_waves_validate(config)
        return cmd_waves_plan(config)

    if sub == "handoff":
        if len(args) < 3:
            print("usage: config.py waves handoff WAVE CONFIG [TASK]", file=sys.stderr)
            return 2
        wave_id = args[1]
        cfg_path = Path(args[2])
        task = args[3] if len(args) > 3 else "task"
        config = load_config(cfg_path)
        return cmd_waves_handoff(wave_id, config, task)

    if sub in ("status", "ready"):
        task = os.environ.get("AWT_WAVE_TASK", "")
        wave_id = ""
        cfg_path: Path | None = None
        i = 1
        while i < len(args):
            if args[i] == "--task" and i + 1 < len(args):
                task = args[i + 1]
                i += 2
                continue
            if sub == "ready" and not wave_id:
                wave_id = args[i]
                i += 1
                continue
            if cfg_path is None and not args[i].startswith("--"):
                cfg_path = Path(args[i])
                i += 1
                continue
            i += 1
        if cfg_path is None:
            cfg_path = Path(".cursor/worktrees.config.json")
        if not task:
            print("usage: config.py waves status CONFIG --task TASK", file=sys.stderr)
            print("       config.py waves ready WAVE CONFIG --task TASK", file=sys.stderr)
            return 2
        config = load_config(cfg_path)
        root = cfg_path.parent.parent
        state = load_wave_state(wave_state_path(root, task))
        if sub == "status":
            return cmd_waves_status(state, config, task)
        if not wave_id:
            print("usage: config.py waves ready WAVE CONFIG --task TASK", file=sys.stderr)
            return 2
        return cmd_waves_ready(wave_id, state, config)

    print(f"unknown waves subcommand: {sub}", file=sys.stderr)
    return 2


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: config.py scope|services|tests|contracts|waves ...", file=sys.stderr)
        return 2
    cmd = sys.argv[1]

    if cmd == "waves":
        return handle_waves_cmd(sys.argv[2:])

    cfg_path = Path(sys.argv[2]) if len(sys.argv) > 2 else Path(".cursor/worktrees.config.json")
    config = load_config(cfg_path)

    if cmd == "scope":
        branch = sys.argv[3]
        base = sys.argv[4] if len(sys.argv) > 4 else "HEAD"
        track = sys.argv[5] if len(sys.argv) > 5 and sys.argv[5] != "--strict" else None
        strict = "--strict" in sys.argv
        return check_scope(branch, base, config, track, strict)

    if cmd == "files":
        branch = sys.argv[3]
        base = sys.argv[4] if len(sys.argv) > 4 else "HEAD"
        for f in changed_files(base, branch):
            print(f)
        return 0

    if cmd == "services":
        files = sys.argv[3:]
        for s in compose_services(files, config):
            print(s)
        return 0

    if cmd == "tests":
        files = sys.argv[3:]
        for t in package_tests(files, config):
            print(t)
        return 0

    if cmd == "contracts":
        files = sys.argv[3:]
        for c in contract_checks(files, config):
            print(c)
        return 0

    print(f"unknown command: {cmd}", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main())
