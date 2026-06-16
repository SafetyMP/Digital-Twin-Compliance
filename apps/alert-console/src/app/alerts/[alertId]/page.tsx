"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";

const AUDIT_EXPLORER =
  process.env.NEXT_PUBLIC_AUDIT_EXPLORER_URL || "http://localhost:3002";

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
  evidenceRef?: string | null;
};

export default function AlertDetailPage() {
  const params = useParams<{ alertId: string }>();
  const [alert, setAlert] = useState<Alert | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    fetch(`/api/alerts/${params.alertId}`, { cache: "no-store" })
      .then((r) => {
        if (!r.ok) throw new Error(`alert API ${r.status}`);
        return r.json();
      })
      .then((data) => {
        setAlert(data);
        setLoadError(null);
      })
      .catch((err) => {
        setLoadError(err instanceof Error ? err.message : "failed to load alert");
      });
  }, [params.alertId]);

  const acknowledge = async () => {
    const res = await fetch(`/api/alerts/${params.alertId}/acknowledge`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ acknowledgedBy: "operator-dev" }),
    });
    if (!res.ok) return;
    setAlert(await res.json());
  };

  if (loadError) {
    return (
      <main className="p-6">
        <Link href="/" className="text-sm text-slate-400 hover:underline">
          ← Back to feed
        </Link>
        <p className="mt-4 text-red-300">{loadError}</p>
      </main>
    );
  }

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
        <div>
          <dt className="text-slate-500">Audit evidence</dt>
          <dd>
            {alert.evidenceRef ? (
              <span className="flex flex-wrap items-center gap-2">
                <span className="font-mono text-xs">{alert.evidenceRef}</span>
                <a
                  href={`${AUDIT_EXPLORER}/entries/${alert.evidenceRef}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="rounded bg-emerald-800 px-2 py-1 text-xs hover:bg-emerald-700"
                >
                  View in Audit Explorer →
                </a>
              </span>
            ) : (
              <span className="text-slate-500">Pending (audit pipeline)</span>
            )}
          </dd>
        </div>
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
