from __future__ import annotations

from contextlib import contextmanager
from pathlib import Path

import psycopg
from psycopg.rows import dict_row

from reporting_service import config


def connect() -> psycopg.Connection:
    return psycopg.connect(config.DB_URL, row_factory=dict_row)


@contextmanager
def cursor():
    with connect() as conn:
        with conn.cursor() as cur:
            yield cur
        conn.commit()


def migrate() -> None:
    sql_path = Path(__file__).resolve().parent.parent / "migrations" / "001_init.sql"
    with connect() as conn:
        with conn.cursor() as cur:
            cur.execute(sql_path.read_text(encoding="utf-8"))
        conn.commit()


def seed_taxonomies(tenant_id: str, version: str) -> None:
    rows = [
        ("FINREP_F01", "assets.cash", "mi_cash"),
        ("FINREP_F01", "assets.loans", "mi_loans"),
        ("FINREP_F01", "liab.deposits", "mi_deposits"),
        ("ANACREDIT_T2", "loan", "CRDT"),
        ("ANACREDIT_T2", "bond", "DBT"),
        ("DORA_ICT", "core-banking", "ICT-CORE"),
        ("DORA_ICT", "kafka", "ICT-MSG"),
        ("DORA_ICT", "immudb", "ICT-LEDGER"),
    ]
    with cursor() as cur:
        for report_type, source, target in rows:
            cur.execute(
                """
                INSERT INTO taxonomy_maps
                  (tenant_id, report_type, source_code, target_code, version, effective_from)
                VALUES (%s, %s, %s, %s, %s, CURRENT_DATE)
                ON CONFLICT (tenant_id, report_type, source_code, version) DO UPDATE
                  SET target_code = EXCLUDED.target_code
                """,
                (tenant_id, report_type, source, target, version),
            )
