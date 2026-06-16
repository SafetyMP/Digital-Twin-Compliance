#!/usr/bin/env python3
"""Filesystem invariant checks shared by mechanical and behavioral eval lanes."""
from __future__ import annotations

import argparse
import fnmatch
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path

DEFAULT_TENANT = "00000000-0000-0000-0000-000000000001"
KAFKA_WRITER_RE = re.compile(r"kafka\.Writer")
PHASE3_TERMS_RE = re.compile(
    r"cedar-policy|immudb|neo4j|keycloak|gorules|gorules\.io",
    re.IGNORECASE,
)
PHASE1_SCOPE_TERMS_RE = re.compile(
    r"apache/flink|org\.apache\.flink|immudb|neo4j|keycloak|cedar-policy|gorules|next\.js",
    re.IGNORECASE,
)
TENANT_DROP_RE = re.compile(
    r"\b(drop|remove|delete)\s+(column\s+)?tenant_id\b",
    re.IGNORECASE,
)

ALLOWED_KAFKA_WRITER_SUFFIXES = (
    "internal/outbox/",
    "dlq.go",
)

CHECK_NAMES = frozenset(
    {
        "outbox-only-kafka-writer",
        "tenant-id-columns",
        "tenant-id-state-migrations",
        "tenant-id-alert-migrations",
        "phase3-scope-boundary",
        "scope-boundary-phase1",
    }
)


@dataclass
class InvariantResult:
    name: str
    passed: bool
    violations: list[str] = field(default_factory=list)

    def to_dict(self) -> dict[str, object]:
        return {"name": self.name, "passed": self.passed, "violations": self.violations}


def parse_unified_diff(diff_text: str) -> set[str]:
    """Return repo-relative paths touched by a unified diff."""
    paths: set[str] = set()
    for line in diff_text.splitlines():
        if line.startswith("+++ ") or line.startswith("--- "):
            raw = line[4:].strip()
            if raw in ("/dev/null", "dev/null"):
                continue
            if raw.startswith("a/") or raw.startswith("b/"):
                raw = raw[2:]
            elif raw.startswith("a\\") or raw.startswith("b\\"):
                raw = raw[2:]
            if "\t" in raw:
                raw = raw.split("\t", 1)[0]
            if raw:
                paths.add(raw.replace("\\", "/"))
    return paths


def _normalize(path: str) -> str:
    return path.replace("\\", "/").lstrip("./")


def _filter_paths(changed_paths: set[str] | None, candidates: list[Path], root: Path) -> list[Path]:
    if not changed_paths:
        return candidates
    normalized = {_normalize(p) for p in changed_paths}
    filtered: list[Path] = []
    for path in candidates:
        rel = _normalize(str(path.relative_to(root)))
        if rel in normalized or any(rel.endswith(n) or n.endswith(rel) for n in normalized):
            filtered.append(path)
    return filtered


def _glob_go_files(root: Path, rel_dir: str, changed_paths: set[str] | None) -> list[Path]:
    base = root / rel_dir
    if not base.is_dir():
        return []
    files = sorted(base.rglob("*.go"))
    return _filter_paths(changed_paths, files, root)


def check_outbox_only_kafka_writer(
    root: Path,
    changed_paths: set[str] | None = None,
) -> InvariantResult:
    name = "outbox-only-kafka-writer"
    violations: list[str] = []
    for path in _glob_go_files(root, "services/state-service", changed_paths):
        rel = _normalize(str(path.relative_to(root)))
        text = path.read_text(encoding="utf-8", errors="replace")
        if not KAFKA_WRITER_RE.search(text):
            continue
        norm = rel.lower()
        if any(token in norm for token in ALLOWED_KAFKA_WRITER_SUFFIXES):
            continue
        if path.name == "dlq.go":
            continue
        violations.append(f"kafka.Writer outside internal/outbox/: {rel}")
    return InvariantResult(name, not violations, violations)


def _migration_has_tenant_id(path: Path) -> bool:
    if not path.is_file():
        return False
    text = path.read_text(encoding="utf-8", errors="replace")
    return "tenant_id" in text and DEFAULT_TENANT in text


def check_tenant_id_migrations(
    root: Path,
    changed_paths: set[str] | None = None,
    *,
    service: str = "state-service",
) -> InvariantResult:
    name = "tenant-id-state-migrations" if service == "state-service" else "tenant-id-alert-migrations"
    migration_dir = root / "services" / service / "migrations"
    violations: list[str] = []
    if not migration_dir.is_dir():
        return InvariantResult(name, False, [f"missing migrations dir: {migration_dir}"])

    migrations = sorted(migration_dir.glob("*.sql"))
    scoped = _filter_paths(changed_paths, migrations, root) if changed_paths else migrations

    if changed_paths:
        for path in scoped:
            rel = _normalize(str(path.relative_to(root)))
            text = path.read_text(encoding="utf-8", errors="replace")
            if TENANT_DROP_RE.search(text):
                violations.append(f"tenant_id drop language in {rel}")
            elif "tenant_id" not in text:
                violations.append(f"tenant_id missing from edited migration {rel}")
            elif DEFAULT_TENANT not in text and "tenant_id" in text:
                violations.append(f"default tenant UUID missing from {rel}")
    else:
        init = migration_dir / ("001_init.sql" if service == "state-service" else "001_alerts.sql")
        if not _migration_has_tenant_id(init):
            violations.append(f"tenant_id/default tenant missing from {init.relative_to(root)}")

    return InvariantResult(name, not violations, violations)


def check_tenant_id_columns(root: Path, changed_paths: set[str] | None = None) -> InvariantResult:
    state = check_tenant_id_migrations(root, changed_paths, service="state-service")
    if not state.passed:
        return InvariantResult("tenant-id-columns", False, state.violations)
    return InvariantResult("tenant-id-columns", True, [])


def _scan_files_for_terms(
    root: Path,
    terms_re: re.Pattern[str],
    search_roots: list[str],
    globs: list[str],
    changed_paths: set[str] | None,
    exclude_substrings: tuple[str, ...] = (),
) -> list[str]:
    hits: list[str] = []
    normalized_changes = {_normalize(p) for p in changed_paths} if changed_paths else None

    for rel_root in search_roots:
        base = root / rel_root
        if not base.exists():
            continue
        for glob in globs:
            for path in base.rglob(glob.split("/")[-1] if "/" not in glob else glob):
                if not path.is_file():
                    continue
                rel = _normalize(str(path.relative_to(root)))
                if exclude_substrings and any(x in rel for x in exclude_substrings):
                    continue
                if normalized_changes is not None and rel not in normalized_changes:
                    continue
                text = path.read_text(encoding="utf-8", errors="replace")
                if terms_re.search(text):
                    hits.append(rel)
    return hits


def check_phase3_scope_boundary(
    root: Path,
    changed_paths: set[str] | None = None,
) -> InvariantResult:
    name = "phase3-scope-boundary"
    # Phase 3 stack landed — mechanical full-repo check is N/A (mirrors scope-boundary-phase1 + compliance-cep).
    if (root / "scripts" / "smoke-test-phase3.sh").is_file():
        return InvariantResult(name, True, [])
    search_roots = ["services", "jobs", "apps", "mocks", "schemas"]
    globs = ["*.go", "*.sql", "*.yml", "*.yaml", "*.avsc", "*.sh", "*.java", "*.tsx", "*.ts"]
    exclude = ("evals/", "scripts/run-live-evals", "scripts/score-agent-transcript")
    hits = _scan_files_for_terms(
        root,
        PHASE3_TERMS_RE,
        search_roots,
        globs,
        changed_paths,
        exclude_substrings=exclude,
    )
    compose = root / "docker-compose.dev.yml"
    if compose.is_file():
        rel = _normalize(str(compose.relative_to(root)))
        if changed_paths is None or rel in {_normalize(p) for p in changed_paths}:
            text = compose.read_text(encoding="utf-8", errors="replace")
            if PHASE3_TERMS_RE.search(text):
                hits.append(rel)
    violations = [f"Phase 3 term in {hit}" for hit in sorted(set(hits))]
    return InvariantResult(name, not violations, violations)


def check_scope_boundary_phase1(root: Path, changed_paths: set[str] | None = None) -> InvariantResult:
    name = "scope-boundary-phase1"
    if (root / "jobs" / "compliance-cep" / "pom.xml").is_file():
        return InvariantResult(name, True, [])
    search_roots = ["services", "mocks", "schemas"]
    globs = ["*.go", "*.sql", "*.yml", "*.yaml", "*.avsc", "*.sh"]
    hits = _scan_files_for_terms(
        root,
        PHASE1_SCOPE_TERMS_RE,
        search_roots,
        globs,
        changed_paths,
        exclude_substrings=("evals/", "scripts/run-live-evals", "scripts/score-agent-transcript"),
    )
    compose = root / "docker-compose.dev.yml"
    if compose.is_file():
        rel = _normalize(str(compose.relative_to(root)))
        if changed_paths is None or rel in {_normalize(p) for p in changed_paths}:
            text = compose.read_text(encoding="utf-8", errors="replace")
            if PHASE1_SCOPE_TERMS_RE.search(text):
                hits.append(rel)
    violations = [f"Phase 2+ term in Phase 1 scope check: {hit}" for hit in sorted(set(hits))]
    return InvariantResult(name, not violations, violations)


def check_forbidden_path_patterns(
    root: Path,
    patterns: list[str],
    changed_paths: set[str] | None,
) -> InvariantResult:
    name = "forbidden-path-patterns"
    if not changed_paths:
        return InvariantResult(name, True, [])
    violations: list[str] = []
    for raw in sorted(changed_paths):
        rel = _normalize(raw)
        for pattern in patterns:
            if fnmatch.fnmatch(rel, pattern) or fnmatch.fnmatch(rel.lower(), pattern.lower()):
                violations.append(f"forbidden path touched: {rel} (pattern {pattern})")
                break
    return InvariantResult(name, not violations, violations)


def run_named_check(name: str, root: Path, changed_paths: set[str] | None = None) -> InvariantResult:
    if name == "outbox-only-kafka-writer":
        return check_outbox_only_kafka_writer(root, changed_paths)
    if name == "tenant-id-columns":
        return check_tenant_id_columns(root, changed_paths)
    if name == "tenant-id-state-migrations":
        return check_tenant_id_migrations(root, changed_paths, service="state-service")
    if name == "tenant-id-alert-migrations":
        return check_tenant_id_migrations(root, changed_paths, service="alert-service")
    if name == "phase3-scope-boundary":
        return check_phase3_scope_boundary(root, changed_paths)
    if name == "scope-boundary-phase1":
        return check_scope_boundary_phase1(root, changed_paths)
    raise ValueError(f"unknown check: {name}")


def run_gate_checks(
    root: Path,
    check_names: list[str],
    changed_paths: set[str] | None,
    forbidden_patterns: list[str] | None = None,
) -> tuple[bool, list[InvariantResult]]:
    results: list[InvariantResult] = []
    for check_name in check_names:
        results.append(run_named_check(check_name, root, changed_paths))
    if forbidden_patterns:
        results.append(check_forbidden_path_patterns(root, forbidden_patterns, changed_paths))
    passed = all(r.passed for r in results)
    return passed, results


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", required=True, choices=sorted(CHECK_NAMES))
    parser.add_argument("--repo-root", type=Path, default=Path.cwd())
    parser.add_argument("--diff", type=Path, help="Unified diff; limit checks to changed paths")
    args = parser.parse_args(argv)

    changed: set[str] | None = None
    if args.diff and args.diff.is_file():
        changed = parse_unified_diff(args.diff.read_text(encoding="utf-8"))

    result = run_named_check(args.check, args.repo_root.resolve(), changed)
    if result.violations:
        for v in result.violations:
            print(v, file=sys.stderr)
    print(f"{'ok' if result.passed else 'FAIL'}  - {result.name}")
    return 0 if result.passed else 1


if __name__ == "__main__":
    sys.exit(main())
