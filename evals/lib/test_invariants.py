from __future__ import annotations

import textwrap
from pathlib import Path

import pytest

from evals.lib.invariants import (
    check_forbidden_path_patterns,
    check_outbox_only_kafka_writer,
    check_phase3_scope_boundary,
    check_tenant_id_migrations,
    parse_unified_diff,
)

FIXTURES = Path(__file__).resolve().parents[1] / "fixtures" / "workspaces"


def test_parse_unified_diff_extracts_paths() -> None:
    diff = textwrap.dedent(
        """\
        diff --git a/services/state-service/internal/consumer/consumer.go b/services/state-service/internal/consumer/consumer.go
        --- a/services/state-service/internal/consumer/consumer.go
        +++ b/services/state-service/internal/consumer/consumer.go
        @@ -1 +1 @@
        +writer *kafka.Writer
        """
    )
    paths = parse_unified_diff(diff)
    assert "services/state-service/internal/consumer/consumer.go" in paths


def test_outbox_violation_detected_in_fixture_workspace() -> None:
    root = FIXTURES / "outbox-violation"
    result = check_outbox_only_kafka_writer(root)
    assert not result.passed
    assert any("consumer.go" in v for v in result.violations)


def test_outbox_clean_fixture_passes() -> None:
    root = FIXTURES / "outbox-clean"
    result = check_outbox_only_kafka_writer(root)
    assert result.passed


def test_tenant_drop_detected_in_fixture_workspace() -> None:
    root = FIXTURES / "tenant-violation"
    migration = root / "services/state-service/migrations/001_init.sql"
    changed = {str(migration.relative_to(root)).replace("\\", "/")}
    result = check_tenant_id_migrations(root, changed)
    assert not result.passed


def test_phase3_cedar_detected_in_fixture_workspace() -> None:
    root = FIXTURES / "cedar-violation"
    changed = {"services/policy-service/main.go"}
    result = check_phase3_scope_boundary(root, changed)
    assert not result.passed


def test_forbidden_path_patterns() -> None:
    root = Path("/tmp/unused")
    result = check_forbidden_path_patterns(
        root,
        ["services/state-service/internal/consumer/**"],
        {"services/state-service/internal/consumer/consumer.go"},
    )
    assert not result.passed
