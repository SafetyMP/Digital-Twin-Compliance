"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";

type EntryIndex = {
  entryId: string;
  entryType: string;
  ruleCode?: string;
  subjectId?: string;
  recordedAt: string;
  payloadHash: string;
  previousHash: string;
};

type VerifyResult = {
  valid: boolean;
  checkedCount: number;
  message?: string;
};

export default function HomePage() {
  const [entries, setEntries] = useState<EntryIndex[]>([]);
  const [ruleCode, setRuleCode] = useState("");
  const [verify, setVerify] = useState<VerifyResult | null>(null);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const q = new URLSearchParams();
      if (ruleCode) q.set("ruleCode", ruleCode);
      q.set("limit", "50");
      const res = await fetch(`/api/audit/entries?${q}`);
      const data = await res.json();
      setEntries(Array.isArray(data) ? data : []);
    } finally {
      setLoading(false);
    }
  }, [ruleCode]);

  useEffect(() => {
    load();
  }, [load]);

  const runVerify = async () => {
    const res = await fetch("/api/audit/verify");
    setVerify(await res.json());
  };

  return (
    <main className="mx-auto max-w-5xl p-6">
      <header className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Audit Ledger</h1>
          <p className="text-sm text-slate-400">Search immudb-backed compliance audit entries</p>
        </div>
        <button
          onClick={runVerify}
          className="rounded bg-emerald-700 px-4 py-2 text-sm hover:bg-emerald-600"
        >
          Verify chain
        </button>
      </header>

      {verify && (
        <div
          className={`mb-4 rounded-lg border p-3 text-sm ${
            verify.valid
              ? "border-emerald-700 bg-emerald-950/50 text-emerald-200"
              : "border-red-700 bg-red-950/50 text-red-200"
          }`}
        >
          Chain {verify.valid ? "valid" : "broken"} — {verify.checkedCount} entries
          {verify.message ? ` (${verify.message})` : ""}
        </div>
      )}

      <form
        className="mb-6 flex flex-wrap gap-3"
        onSubmit={(e) => {
          e.preventDefault();
          load();
        }}
      >
        <input
          className="rounded border border-slate-700 bg-slate-900 px-3 py-2 text-sm"
          placeholder="ruleCode e.g. BASEL-M001"
          value={ruleCode}
          onChange={(e) => setRuleCode(e.target.value)}
        />
        <button type="submit" className="rounded bg-slate-700 px-4 py-2 text-sm hover:bg-slate-600">
          Search
        </button>
      </form>

      {loading && <p className="text-slate-400">Loading…</p>}
      <ul className="space-y-3">
        {entries.map((e) => (
          <li key={e.entryId} className="rounded-lg border border-slate-800 bg-slate-900 p-4">
            <div className="mb-2 flex flex-wrap items-center gap-2">
              <span className="rounded bg-emerald-800 px-2 py-0.5 text-xs">integrity OK</span>
              <span className="font-mono text-xs text-slate-400">{e.entryType}</span>
              {e.ruleCode && <span className="font-mono text-xs text-amber-400">{e.ruleCode}</span>}
            </div>
            <Link href={`/entries/${e.entryId}`} className="font-medium hover:underline">
              {e.entryId}
            </Link>
            <p className="mt-1 text-xs text-slate-500">{new Date(e.recordedAt).toLocaleString()}</p>
            <p className="mt-2 truncate font-mono text-xs text-slate-500">{e.payloadHash}</p>
          </li>
        ))}
        {!loading && entries.length === 0 && <p className="text-slate-400">No audit entries found.</p>}
      </ul>
    </main>
  );
}
