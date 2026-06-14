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

const API = process.env.NEXT_PUBLIC_ALERT_SERVICE_URL || "http://localhost:8085";
const WS = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8085/ws/alerts";

function severityClass(severity: string) {
  if (severity === "Critical") return "bg-red-600";
  if (severity === "Warning") return "bg-amber-500 text-black";
  return "bg-slate-600";
}

export default function HomePage() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [connected, setConnected] = useState(false);

  const upsertAlert = useCallback((alert: Alert) => {
    setAlerts((prev) => {
      const rest = prev.filter((a) => a.alertId !== alert.alertId);
      return [alert, ...rest].sort(
        (a, b) => new Date(b.detectedAt).getTime() - new Date(a.detectedAt).getTime()
      );
    });
  }, []);

  useEffect(() => {
    fetch(`${API}/api/v1/alerts?status=Open&limit=50`)
      .then((r) => r.json())
      .then((data) => setAlerts(data))
      .catch(console.error);
  }, []);

  useEffect(() => {
    let ws: WebSocket | null = null;
    let backoff = 1000;
    let cancelled = false;

    const connect = () => {
      ws = new WebSocket(`${WS}?status=Open`);
      ws.onopen = () => {
        setConnected(true);
        backoff = 1000;
      };
      ws.onclose = () => {
        setConnected(false);
        if (!cancelled) setTimeout(connect, backoff);
        backoff = Math.min(backoff * 2, 10000);
      };
      ws.onmessage = (ev) => {
        const msg = JSON.parse(ev.data);
        if (msg.payload) upsertAlert(msg.payload);
      };
    };

    connect();
    return () => {
      cancelled = true;
      ws?.close();
    };
  }, [upsertAlert]);

  const acknowledge = async (alertId: string) => {
    await fetch(`${API}/api/v1/alerts/${alertId}/acknowledge`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ acknowledgedBy: "operator-dev" }),
    });
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
        {alerts.length === 0 && <p className="text-slate-400">No open alerts.</p>}
      </ul>
    </main>
  );
}
