export type AppId = "alert" | "audit" | "graph" | "simulation" | "report";

export type JourneyMeta = {
  appId: AppId;
  label: string;
  journey: string;
  phase: number;
  short: string;
};

export const JOURNEYS: Record<AppId, JourneyMeta> = {
  alert: {
    appId: "alert",
    label: "Alert Console",
    journey: "Monitoring & alerts",
    phase: 2,
    short: "Monitoring",
  },
  audit: {
    appId: "audit",
    label: "Audit Explorer",
    journey: "Policy & audit",
    phase: 3,
    short: "Audit",
  },
  graph: {
    appId: "graph",
    label: "Graph Explorer",
    journey: "Graph & simulation",
    phase: 4,
    short: "Graph",
  },
  simulation: {
    appId: "simulation",
    label: "Simulation Console",
    journey: "Graph & simulation",
    phase: 4,
    short: "Simulation",
  },
  report: {
    appId: "report",
    label: "Report Console",
    journey: "Regulatory reporting",
    phase: 5,
    short: "Reporting",
  },
};

export const APP_SWITCHER_ORDER: AppId[] = [
  "alert",
  "audit",
  "graph",
  "simulation",
  "report",
];

export type ConsoleUrls = Record<AppId, string>;

export const DEFAULT_CONSOLE_URLS: ConsoleUrls = {
  alert: "http://localhost:3000",
  audit: "http://localhost:3002",
  graph: "http://localhost:3003",
  simulation: "http://localhost:3004",
  report: "http://localhost:3005",
};
