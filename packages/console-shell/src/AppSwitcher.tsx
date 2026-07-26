import type { AppId, ConsoleUrls } from "./journeys";
import { APP_SWITCHER_ORDER, JOURNEYS } from "./journeys";

export function AppSwitcher({
  activeApp,
  urls,
}: {
  activeApp: AppId;
  urls: ConsoleUrls;
}) {
  return (
    <nav aria-label="Console switcher" className="flex flex-wrap items-center gap-1">
      {APP_SWITCHER_ORDER.map((id) => {
        const meta = JOURNEYS[id];
        const active = id === activeApp;
        return (
          <a
            key={id}
            href={urls[id]}
            aria-current={active ? "page" : undefined}
            className={
              active
                ? "rounded-md bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-900"
                : "rounded-md px-2.5 py-1 text-xs text-slate-400 hover:bg-slate-800 hover:text-slate-100"
            }
          >
            {meta.label}
          </a>
        );
      })}
    </nav>
  );
}
