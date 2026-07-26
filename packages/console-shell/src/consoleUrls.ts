import { DEFAULT_CONSOLE_URLS, type ConsoleUrls } from "./journeys";

/** Build console URLs from Next.js NEXT_PUBLIC_* env with localhost defaults. */
export function consoleUrlsFromEnv(
  env: Record<string, string | undefined> = typeof process !== "undefined" ? process.env : {}
): ConsoleUrls {
  return {
    alert: env.NEXT_PUBLIC_ALERT_CONSOLE_URL || DEFAULT_CONSOLE_URLS.alert,
    audit: env.NEXT_PUBLIC_AUDIT_EXPLORER_URL || DEFAULT_CONSOLE_URLS.audit,
    graph: env.NEXT_PUBLIC_GRAPH_EXPLORER_URL || DEFAULT_CONSOLE_URLS.graph,
    simulation:
      env.NEXT_PUBLIC_SIMULATION_CONSOLE_URL || DEFAULT_CONSOLE_URLS.simulation,
    report: env.NEXT_PUBLIC_REPORT_CONSOLE_URL || DEFAULT_CONSOLE_URLS.report,
  };
}
