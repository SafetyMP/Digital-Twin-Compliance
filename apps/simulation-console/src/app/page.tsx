"use client";

import Link from "next/link";
import { useState } from "react";

type RunResult = {
  runId: string;
  correlationId: string;
  metrics: {
    scenarioId: string;
    baselineCet1: number;
    stressedCet1: number;
    baselineTotalCapital: number;
    stressedTotalCapital: number;
    explainabilityRef: string;
  };
  decisions: Array<{ ruleCode: string; outcome: string; rationale: string }>;
};

export default function SimulationConsolePage() {
  const [scenarioId, setScenarioId] = useState("ecb-adverse-v1");
  const [result, setResult] = useState<RunResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const auditExplorer =
    process.env.NEXT_PUBLIC_AUDIT_EXPLORER_URL || "http://localhost:3002";

  const runSimulation = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("/api/simulations/run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ scenarioId, parameters: {} }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.detail || data.error || "run failed");
        return;
      }
      setResult(data);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="mx-auto max-w-3xl p-6">
      <h1 className="mb-2 text-2xl font-semibold">Stress Simulation</h1>
      <p className="mb-6 text-sm text-slate-400">Deterministic ECB Adverse-style scenario on seed graph</p>

      <div className="mb-6 flex flex-wrap gap-3">
        <select
          className="rounded border border-slate-700 bg-slate-900 px-3 py-2 text-sm"
          value={scenarioId}
          onChange={(e) => setScenarioId(e.target.value)}
        >
          <option value="ecb-adverse-v1">ECB Adverse v1</option>
        </select>
        <button
          onClick={runSimulation}
          disabled={loading}
          className="rounded bg-violet-700 px-4 py-2 text-sm hover:bg-violet-600 disabled:opacity-50"
        >
          {loading ? "Running…" : "Run simulation"}
        </button>
      </div>

      {error && <p className="mb-4 text-sm text-red-300">{error}</p>}

      {result && (
        <div className="space-y-4">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-slate-700 text-left text-slate-400">
                <th className="py-2">Metric</th>
                <th>Baseline</th>
                <th>Stressed</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-b border-slate-800">
                <td className="py-2">CET1 ratio</td>
                <td>{(result.metrics.baselineCet1 * 100).toFixed(2)}%</td>
                <td className="text-amber-300">{(result.metrics.stressedCet1 * 100).toFixed(2)}%</td>
              </tr>
              <tr className="border-b border-slate-800">
                <td className="py-2">Total capital</td>
                <td>{(result.metrics.baselineTotalCapital * 100).toFixed(2)}%</td>
                <td className="text-amber-300">{(result.metrics.stressedTotalCapital * 100).toFixed(2)}%</td>
              </tr>
            </tbody>
          </table>

          <div className="rounded border border-slate-800 bg-slate-900 p-4 text-sm">
            <p className="font-mono text-xs text-slate-500">runId: {result.runId}</p>
            <p className="mt-2">{result.metrics.explainabilityRef}</p>
            <ul className="mt-3 space-y-1">
              {result.decisions.map((d) => (
                <li key={d.ruleCode}>
                  <span className="font-mono text-violet-300">{d.ruleCode}</span>: {d.outcome} — {d.rationale}
                </li>
              ))}
            </ul>
          </div>

          <Link href={auditExplorer} className="text-sm text-sky-400 hover:underline">
            Open Audit Explorer →
          </Link>
        </div>
      )}
    </main>
  );
}
