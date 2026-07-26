import type { AppId } from "./journeys";
import { JOURNEYS } from "./journeys";

export function PhaseBadge({ appId }: { appId: AppId }) {
  const meta = JOURNEYS[appId];
  return (
    <span className="inline-flex items-center gap-1.5 rounded-md border border-slate-700 bg-slate-900/80 px-2 py-0.5 text-xs text-slate-300">
      <span className="font-medium text-slate-100">{meta.short}</span>
      <span className="text-slate-500">·</span>
      <span>Phase {meta.phase}</span>
    </span>
  );
}
