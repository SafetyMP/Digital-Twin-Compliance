"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

type Alert = {
  alertId: string;
  ruleCode: string;
  severity: string;
  status: string;
  personaId: string;
  personaType: string;
  summary: string;
  detectedAt: string;
};

const POLL_MS = 5000;

function severityClass(severity: string) {
  if (severity === "Critical") return "bg-red-600";
  if (severity === "Warning") return "bg-amber-500 text-black";
  return "bg-slate-600";
}

export default function HomePage() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [connected, setConnected] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);

  const loadAlerts = useCallback(async () => {
    const res = await fetch("/api/alerts?status=Open&limit=50", { cache: "no-store" });
    if (!res.ok) {
      throw new Error(`alerts API ${res.status}`);
    }
    const data = (await res.json()) as Alert[];
    setAlerts(Array.isArray(data) ? data : []);
    setConnected(true);
    setLoadError(null);
  }, []);

  useEffect(() => {
    let cancelled = false;

    const refresh = async () => {
      try {
        await loadAlerts();
      } catch (err) {
        if (!cancelled) {
          setConnected(false);
          setLoadError(err instanceof Error ? err.message : "failed to load alerts");
        }
      }
    };

    refresh();
    const timer = window.setInterval(refresh, POLL_MS);
    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [loadAlerts]);

  const acknowledge = async (alertId: string) => {
    const res = await fetch(`/api/alerts/${alertId}/acknowledge`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ acknowledgedBy: "operator-dev" }),
    });
    if (!res.ok) return;
    setAlerts((prev) =>
      prev.map((a) => (a.alertId === alertId ? { ...a, status: "Acknowledged" } : a))
    );
  };

  return (
    <main className="mx-auto max-w-4xl p-6">
      <header className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Compliance Alerts</h1>
        <span className={`text-sm ${connected ? "text-emerald-400" : "text-slate-400"}`}>
          {connected ? "Live" : "Reconnecting…"}
        </span>
      </header>
      {loadError && (
        <p className="mb-4 rounded border border-red-800 bg-red-950/40 p-3 text-sm text-red-200">
          {loadError}
        </p>
      )}
      <ul className="space-y-3">
        {alerts.map((alert) => (
          <li key={alert.alertId} className="rounded-lg border border-slate-800 bg-slate-900 p-4">
            <div className="flex items-start justify-between gap-4">
              <div>
                <div className="mb-2 flex items-center gap-2">
                  <span className={`rounded px-2 py-0.5 text-xs font-medium ${severityClass(alert.severity)}`}>
                    {alert.severity}
                  </span>
                  <span className="font-mono text-xs text-slate-400">{alert.ruleCode}</span>
                  <span className="text-xs text-slate-500">{alert.status}</span>
                </div>
                <Link href={`/alerts/${alert.alertId}`} className="font-medium hover:underline">
                  {alert.summary}
                </Link>
                <p className="mt-1 text-sm text-slate-400">
                  {alert.personaType} · {alert.personaId}
                </p>
                <p className="text-xs text-slate-500">{new Date(alert.detectedAt).toLocaleString()}</p>
              </div>
              {alert.status === "Open" && (
                <button
                  onClick={() => acknowledge(alert.alertId)}
                  className="rounded bg-slate-700 px-3 py-1 text-sm hover:bg-slate-600"
                >
                  Acknowledge
                </button>
              )}
            </div>
          </li>
        ))}
        {!loadError && alerts.length === 0 && <p className="text-slate-400">No open alerts.</p>}
      </ul>
    </main>
  );
}
