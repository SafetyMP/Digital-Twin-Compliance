import type { ReactNode } from "react";
import { AppSwitcher } from "./AppSwitcher";
import type { AppId, ConsoleUrls } from "./journeys";
import { JOURNEYS } from "./journeys";
import { PhaseBadge } from "./PhaseBadge";

function Mark() {
  return (
    <svg
      width="28"
      height="28"
      viewBox="0 0 32 32"
      aria-hidden="true"
      className="shrink-0"
    >
      <rect width="32" height="32" rx="6" fill="#0f172a" stroke="#334155" />
      <circle cx="11" cy="12" r="3.5" fill="none" stroke="#38bdf8" strokeWidth="2" />
      <circle cx="21" cy="20" r="3.5" fill="none" stroke="#e2e8f0" strokeWidth="2" />
      <path d="M14 14.5 18 17.5" stroke="#64748b" strokeWidth="1.5" />
    </svg>
  );
}

export function AppShell({
  activeApp,
  urls,
  envBadge = "Dev — no authentication",
  children,
}: {
  activeApp: AppId;
  urls: ConsoleUrls;
  envBadge?: string;
  children: ReactNode;
}) {
  const meta = JOURNEYS[activeApp];

  return (
    <div className="dt-shell min-h-screen bg-slate-950 text-slate-100">
      <header className="border-b border-slate-800 bg-slate-950/95">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-4 py-3">
          <div className="flex min-w-0 items-center gap-3">
            <Mark />
            <div className="min-w-0">
              <div className="truncate text-sm font-semibold tracking-tight text-slate-50">
                Digital Twin Compliance
              </div>
              <div className="mt-0.5 flex flex-wrap items-center gap-2">
                <span className="truncate text-xs text-slate-400">{meta.label}</span>
                <PhaseBadge appId={activeApp} />
              </div>
            </div>
          </div>
          <AppSwitcher activeApp={activeApp} urls={urls} />
        </div>
        {envBadge ? (
          <div className="border-t border-slate-800/80 bg-slate-900/50 px-4 py-1.5 text-center text-xs text-slate-400">
            {envBadge} · {meta.journey} · Phase {meta.phase}
          </div>
        ) : null}
      </header>
      {children}
    </div>
  );
}
