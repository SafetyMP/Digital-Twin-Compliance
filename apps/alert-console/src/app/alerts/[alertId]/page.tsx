"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";

const API = process.env.NEXT_PUBLIC_ALERT_SERVICE_URL || "http://localhost:8085";

type Alert = {
  alertId: string;
  ruleCode: string;
  regime: string;
  severity: string;
  status: string;
  personaId: string;
  personaType: string;
  summary: string;
  details: Record<string, unknown>;
  detectedAt: string;
};

export default function AlertDetailPage() {
  const params = useParams<{ alertId: string }>();
  const [alert, setAlert] = useState<Alert | null>(null);

  useEffect(() => {
    fetch(`${API}/api/v1/alerts/${params.alertId}`)
      .then((r) => r.json())
      .then(setAlert)
      .catch(console.error);
  }, [params.alertId]);

  const acknowledge = async () => {
    const res = await fetch(`${API}/api/v1/alerts/${params.alertId}/acknowledge`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ acknowledgedBy: "operator-dev" }),
    });
    setAlert(await res.json());
  };

  if (!alert) return <main className="p-6">Loading…</main>;

  return (
    <main className="mx-auto max-w-2xl p-6">
      <Link href="/" className="text-sm text-slate-400 hover:underline">
        ← Back to feed
      </Link>
      <h1 className="mt-4 text-2xl font-semibold">{alert.summary}</h1>
      <dl className="mt-6 space-y-2 text-sm">
        <div><dt className="text-slate-500">Rule</dt><dd>{alert.ruleCode} ({alert.regime})</dd></div>
        <div><dt className="text-slate-500">Severity</dt><dd>{alert.severity}</dd></div>
        <div><dt className="text-slate-500">Status</dt><dd>{alert.status}</dd></div>
        <div><dt className="text-slate-500">Persona</dt><dd>{alert.personaType} · {alert.personaId}</dd></div>
        <div><dt className="text-slate-500">Detected</dt><dd>{new Date(alert.detectedAt).toLocaleString()}</dd></div>
      </dl>
      <pre className="mt-6 overflow-auto rounded bg-slate-900 p-4 text-xs">
        {JSON.stringify(alert.details, null, 2)}
      </pre>
      {alert.status === "Open" && (
        <button onClick={acknowledge} className="mt-6 rounded bg-slate-700 px-4 py-2 hover:bg-slate-600">
          Acknowledge
        </button>
      )}
    </main>
  );
}
