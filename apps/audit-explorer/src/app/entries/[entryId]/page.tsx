"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

type AuditEntry = {
  entryId: string;
  entryType: string;
  payloadHash: string;
  previousHash: string;
  payload: Record<string, unknown>;
  metadata: Record<string, unknown>;
};

export default function EntryPage({ params }: { params: { entryId: string } }) {
  const [entry, setEntry] = useState<AuditEntry | null>(null);

  useEffect(() => {
    fetch(`/api/audit/entries/${params.entryId}`)
      .then((r) => {
        if (!r.ok) throw new Error(`audit entry ${r.status}`);
        return r.json();
      })
      .then(setEntry)
      .catch(console.error);
  }, [params.entryId]);

  if (!entry) {
    return (
      <main className="p-6">
        <Link href="/" className="text-sm text-slate-400 hover:underline">
          ← Back
        </Link>
        <p className="mt-4 text-slate-400">Loading entry…</p>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-3xl p-6">
      <Link href="/" className="text-sm text-slate-400 hover:underline">
        ← Back to search
      </Link>
      <h1 className="mt-4 text-xl font-semibold">{entry.entryId}</h1>
      <span className="mt-2 inline-block rounded bg-emerald-800 px-2 py-0.5 text-xs">integrity OK</span>
      <dl className="mt-6 space-y-3 text-sm">
        <div>
          <dt className="text-slate-500">payloadHash</dt>
          <dd className="break-all font-mono">{entry.payloadHash}</dd>
        </div>
        <div>
          <dt className="text-slate-500">previousHash</dt>
          <dd className="break-all font-mono">{entry.previousHash || "(genesis)"}</dd>
        </div>
        <div>
          <dt className="text-slate-500">payload</dt>
          <dd>
            <pre className="overflow-auto rounded bg-slate-900 p-3 text-xs">
              {JSON.stringify(entry.payload, null, 2)}
            </pre>
          </dd>
        </div>
      </dl>
    </main>
  );
}
