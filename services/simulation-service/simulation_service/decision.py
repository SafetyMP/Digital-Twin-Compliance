import httpx

from simulation_service import config


async def evaluate_corep(stressed_cet1: float, stressed_total_capital: float, persona_id: str, correlation_id: str) -> list[dict]:
    decisions = []
    async with httpx.AsyncClient(timeout=15.0) as client:
        specs = [
            ("COREP-R001", {"cet1Ratio": stressed_cet1}),
            ("COREP-R002", {"totalCapitalRatio": stressed_total_capital}),
        ]
        for rule_code, fields in specs:
            payload = {
                "ruleCode": rule_code,
                "input": {
                    **fields,
                    "personaId": persona_id,
                    "tenantId": config.DEFAULT_TENANT_ID,
                },
                "correlationId": correlation_id,
            }
            res = await client.post(
                f"{config.DECISION_SERVICE_URL.rstrip('/')}/api/v1/evaluate",
                json=payload,
            )
            res.raise_for_status()
            decisions.append(res.json())
    return decisions
