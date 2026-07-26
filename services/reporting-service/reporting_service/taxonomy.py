from __future__ import annotations

from reporting_service.db import cursor


def list_taxonomies(tenant_id: str) -> list[dict]:
    with cursor() as cur:
        cur.execute(
            """
            SELECT report_type, source_code, target_code, version,
                   effective_from::text, effective_to::text
            FROM taxonomy_maps
            WHERE tenant_id = %s
            ORDER BY report_type, source_code, version
            """,
            (tenant_id,),
        )
        return list(cur.fetchall())


def maps_for(tenant_id: str, report_type: str, version: str) -> dict[str, str]:
    with cursor() as cur:
        cur.execute(
            """
            SELECT source_code, target_code
            FROM taxonomy_maps
            WHERE tenant_id = %s AND report_type = %s AND version = %s
            """,
            (tenant_id, report_type, version),
        )
        rows = cur.fetchall()
    return {r["source_code"]: r["target_code"] for r in rows}
