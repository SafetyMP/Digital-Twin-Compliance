"""Continuous/agentic contagion over exposure graph (ADR-013 D37)."""

from __future__ import annotations

from collections import defaultdict, deque


def run_contagion(
    nodes: list[dict],
    edges: list[dict],
    seed_entity_id: str,
    max_hops: int = 3,
) -> dict:
    adj: dict[str, list[str]] = defaultdict(list)
    for e in edges:
        a, b = e.get("fromEntityId"), e.get("toEntityId")
        if a and b:
            adj[str(a)].append(str(b))
            adj[str(b)].append(str(a))

    names = {str(n.get("entityId")): n.get("name", "") for n in nodes}
    if seed_entity_id not in names and nodes:
        # fall back to first node
        seed_entity_id = str(nodes[0].get("entityId"))

    infected: dict[str, int] = {seed_entity_id: 0}
    q: deque[str] = deque([seed_entity_id])
    while q:
        cur = q.popleft()
        hop = infected[cur]
        if hop >= max_hops:
            continue
        for nxt in adj.get(cur, []):
            if nxt not in infected:
                infected[nxt] = hop + 1
                q.append(nxt)

    path_explain = sorted(
        [{"entityId": eid, "name": names.get(eid, ""), "hop": hop} for eid, hop in infected.items()],
        key=lambda r: (r["hop"], r["entityId"]),
    )
    return {
        "seedEntityId": seed_entity_id,
        "maxHops": max_hops,
        "infectedCount": len(infected),
        "pathExplainability": path_explain,
        "mode": "agentic-bfs-contagion-v1",
    }
