"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

type Node = { entityId: string; name: string; lcr?: number; cet1Ratio?: number };
type Edge = {
  fromEntityId: string;
  toEntityId: string;
  exposureType: string;
  notionalEur: number;
  layer: string;
};
type Summary = { nodeCount: number; edgeCount: number };

const LAYERS = ["", "ShortTerm", "LongTerm", "Contingent"] as const;

export default function GraphExplorerPage() {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [edges, setEdges] = useState<Edge[]>([]);
  const [search, setSearch] = useState("");
  const [layer, setLayer] = useState<(typeof LAYERS)[number]>("");
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [sRes, nRes, eRes] = await Promise.all([
        fetch("/api/graph/summary"),
        fetch(`/api/graph/nodes?name=${encodeURIComponent(search)}&limit=100`),
        fetch(`/api/graph/edges?layer=${encodeURIComponent(layer)}&limit=500`),
      ]);
      setSummary(await sRes.json());
      setNodes(await nRes.json());
      setEdges(await eRes.json());
    } finally {
      setLoading(false);
    }
  }, [search, layer]);

  useEffect(() => {
    load();
  }, [load]);

  const positions = useMemo(() => {
    const pos = new Map<string, { x: number; y: number }>();
    nodes.forEach((n, i) => {
      const angle = (i / Math.max(nodes.length, 1)) * Math.PI * 2;
      pos.set(n.entityId, { x: 200 + Math.cos(angle) * 140, y: 180 + Math.sin(angle) * 120 });
    });
    return pos;
  }, [nodes]);

  return (
    <main className="mx-auto max-w-6xl p-6">
      <header className="mb-6">
        <h1 className="text-2xl font-semibold">Exposure Graph</h1>
        <p className="text-sm text-slate-400">
          {summary ? `${summary.nodeCount} institutions · ${summary.edgeCount} exposures` : "Loading summary…"}
        </p>
      </header>

      <div className="mb-4 flex flex-wrap gap-3">
        <input
          className="rounded border border-slate-700 bg-slate-900 px-3 py-2 text-sm"
          placeholder="Search institution"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select
          className="rounded border border-slate-700 bg-slate-900 px-3 py-2 text-sm"
          value={layer}
          onChange={(e) => setLayer(e.target.value as (typeof LAYERS)[number])}
        >
          {LAYERS.map((l) => (
            <option key={l || "all"} value={l}>
              {l || "All layers"}
            </option>
          ))}
        </select>
        <button onClick={load} className="rounded bg-sky-700 px-4 py-2 text-sm hover:bg-sky-600">
          Refresh
        </button>
      </div>

      {loading && <p className="text-slate-400">Loading graph…</p>}

      <div className="grid gap-6 lg:grid-cols-2">
        <svg viewBox="0 0 400 360" className="h-80 w-full rounded-lg border border-slate-800 bg-slate-900">
          {edges.map((e, idx) => {
            const from = positions.get(e.fromEntityId);
            const to = positions.get(e.toEntityId);
            if (!from || !to) return null;
            return (
              <line
                key={`${e.fromEntityId}-${e.toEntityId}-${idx}`}
                x1={from.x}
                y1={from.y}
                x2={to.x}
                y2={to.y}
                stroke="#38bdf8"
                strokeOpacity={0.5}
                strokeWidth={1}
              />
            );
          })}
          {nodes.map((n) => {
            const p = positions.get(n.entityId);
            if (!p) return null;
            return (
              <g key={n.entityId}>
                <circle cx={p.x} cy={p.y} r={10} fill="#0ea5e9" />
                <text x={p.x + 14} y={p.y + 4} className="fill-slate-200 text-[8px]">
                  {n.name.slice(0, 18)}
                </text>
              </g>
            );
          })}
        </svg>

        <ul className="max-h-80 space-y-2 overflow-y-auto text-sm">
          {edges.slice(0, 40).map((e, idx) => (
            <li key={idx} className="rounded border border-slate-800 bg-slate-900 p-2 font-mono text-xs">
              {e.fromEntityId.slice(0, 8)} → {e.toEntityId.slice(0, 8)} · {e.layer} · €
              {e.notionalEur.toLocaleString()}
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
