import httpx

from simulation_service import config


class GraphClient:
    def __init__(self, base_url: str | None = None):
        self.base_url = (base_url or config.GRAPH_SERVICE_URL).rstrip("/")

    async def health(self) -> dict:
        async with httpx.AsyncClient(timeout=10.0) as client:
            res = await client.get(f"{self.base_url}/api/v1/health")
            res.raise_for_status()
            return res.json()

    async def summary(self) -> dict:
        async with httpx.AsyncClient(timeout=10.0) as client:
            res = await client.get(f"{self.base_url}/api/v1/graph/summary")
            res.raise_for_status()
            return res.json()

    async def nodes(self, limit: int = 500) -> list[dict]:
        async with httpx.AsyncClient(timeout=30.0) as client:
            res = await client.get(
                f"{self.base_url}/api/v1/graph/nodes",
                params={"limit": limit},
            )
            res.raise_for_status()
            return res.json()

    async def edges(self, limit: int = 1000) -> list[dict]:
        async with httpx.AsyncClient(timeout=30.0) as client:
            res = await client.get(
                f"{self.base_url}/api/v1/graph/edges",
                params={"limit": limit},
            )
            res.raise_for_status()
            return res.json()
