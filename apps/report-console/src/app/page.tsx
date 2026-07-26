"use client";

import { useState } from "react";

type ReportMeta = {
  id: string;
  status: string;
  reportType?: string;
  taxonomyVersion?: string;
  objectKey?: string;
  evidenceRef?: string;
};

export default function ReportConsolePage() {
  const [reportType, setReportType] = useState("FINREP_F01");
  const [report, setReport] = useState<ReportMeta | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const createDraft = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch("/api/reports", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ reportType }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.detail || data.error || "create failed");
        return;
      }
      setReport(data);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  const validate = async () => {
    if (!report) return;
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/api/reports/${report.id}/validate`, { method: "POST" });
      const data = await res.json();
      if (!res.ok) {
        setError(data.detail || data.error || "validate failed");
        return;
      }
      setReport({ ...report, ...data });
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  const submit = async () => {
    if (!report) return;
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/api/reports/${report.id}/submit`, { method: "POST" });
      const data = await res.json();
      if (!res.ok) {
        setError(data.detail || data.error || "submit failed");
        return;
      }
      setReport({ ...report, ...data });
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="mx-auto max-w-3xl p-6">
      <h1 className="mb-2 text-2xl font-semibold">Regulatory reports</h1>
      <p className="mb-6 text-sm text-slate-400">
        Lifecycle enforced by Reporting Service: draft → validated → submitted
      </p>

      <div className="mb-6 flex flex-wrap gap-3">
        <select
          className="rounded border border-slate-700 bg-slate-900 px-3 py-2 text-sm"
          value={reportType}
          onChange={(e) => setReportType(e.target.value)}
        >
          <option value="FINREP_F01">FINREP F01 (XBRL)</option>
          <option value="ANACREDIT_T2">AnaCredit Table 2 (SDMX)</option>
          <option value="DORA_ICT">DORA ICT Register (XML)</option>
        </select>
        <button
          onClick={createDraft}
          disabled={loading}
          className="rounded bg-teal-700 px-4 py-2 text-sm hover:bg-teal-600 disabled:opacity-50"
        >
          Create draft
        </button>
        <button
          onClick={validate}
          disabled={loading || !report || report.status !== "draft"}
          className="rounded bg-slate-700 px-4 py-2 text-sm hover:bg-slate-600 disabled:opacity-50"
        >
          Validate
        </button>
        <button
          onClick={submit}
          disabled={loading || !report || report.status !== "validated"}
          className="rounded bg-slate-700 px-4 py-2 text-sm hover:bg-slate-600 disabled:opacity-50"
        >
          Submit
        </button>
      </div>

      {error && <p className="mb-4 text-sm text-red-300">{error}</p>}

      {report && (
        <dl className="space-y-2 rounded border border-slate-800 bg-slate-900/60 p-4 text-sm">
          <div className="flex justify-between gap-4">
            <dt className="text-slate-400">id</dt>
            <dd className="font-mono">{report.id}</dd>
          </div>
          <div className="flex justify-between gap-4">
            <dt className="text-slate-400">status</dt>
            <dd>{report.status}</dd>
          </div>
          {report.taxonomyVersion && (
            <div className="flex justify-between gap-4">
              <dt className="text-slate-400">taxonomyVersion</dt>
              <dd>{report.taxonomyVersion}</dd>
            </div>
          )}
          {report.objectKey && (
            <div className="flex justify-between gap-4">
              <dt className="text-slate-400">objectKey</dt>
              <dd className="font-mono">{report.objectKey}</dd>
            </div>
          )}
          {report.evidenceRef && (
            <div className="flex justify-between gap-4">
              <dt className="text-slate-400">evidenceRef</dt>
              <dd className="font-mono">{report.evidenceRef}</dd>
            </div>
          )}
        </dl>
      )}
    </main>
  );
}
