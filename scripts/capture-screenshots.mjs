/**
 * Capture README screenshots and demo GIF from all five consoles.
 *
 * Default (warm stack; prefers linked alert detail when available):
 *   npm run screenshots
 *
 * Home pages only:
 *   npm run screenshots -- --consoles-only
 *
 * Rebuild GIF from existing PNGs (no browser):
 *   npm run screenshots:rebuild-gif
 *
 * CI: set CI=1 to use bundled Chromium instead of system Chrome.
 */
import { execFileSync } from "node:child_process";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { chromium } from "playwright";
import gifenc from "gifenc";
import { PNG } from "pngjs";

const { GIFEncoder, quantize, applyPalette } = gifenc;

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.join(__dirname, "..");
const outDir = path.join(repoRoot, "docs", "assets");

const alertConsoleBase = process.env.ALERT_CONSOLE_URL ?? "http://localhost:3000";
const auditExplorerBase = process.env.AUDIT_EXPLORER_URL ?? "http://localhost:3002";
const graphExplorerBase = process.env.GRAPH_EXPLORER_URL ?? "http://localhost:3003";
const simulationConsoleBase =
  process.env.SIMULATION_CONSOLE_URL ?? "http://localhost:3004";
const reportConsoleBase = process.env.REPORT_CONSOLE_URL ?? "http://localhost:3005";
const alertDbUrl =
  process.env.ALERT_DB_URL ?? "postgres://alert:alert@localhost:5435/alerts?sslmode=disable";

/** Frame duration in milliseconds (gifenc stores delay/10 as GIF centiseconds). */
const GIF_FRAME_DELAY_MS = 1_800;

const gifFrameFiles = [
  { file: "alert-console.png", name: "Alert Console" },
  { file: "audit-explorer.png", name: "Audit Explorer" },
  { file: "graph-explorer.png", name: "Graph Explorer" },
  { file: "simulation-console.png", name: "Simulation Console" },
  { file: "report-console.png", name: "Report Console" },
];

const ERROR_TEXT_RE =
  /(API \d{3}|request failed|failed to load|Reconnecting…|\bRetry\b)/i;

function launchOptions() {
  if (process.env.CI) {
    return { headless: true };
  }
  return { channel: "chrome", headless: true };
}

function discoverLinkedAlert() {
  if (process.env.ALERT_ID?.trim() && process.env.EVIDENCE_REF?.trim()) {
    return {
      alertId: process.env.ALERT_ID.trim(),
      evidenceRef: process.env.EVIDENCE_REF.trim(),
    };
  }

  try {
    const row = execFileSync(
      "psql",
      [
        alertDbUrl,
        "-Atqc",
        "SELECT alert_id, evidence_ref FROM compliance_alerts WHERE evidence_ref IS NOT NULL ORDER BY detected_at DESC LIMIT 1;",
      ],
      { encoding: "utf8" },
    ).trim();
    if (!row) {
      return null;
    }
    const [alertId, evidenceRef] = row.split("|");
    if (!alertId || !evidenceRef) {
      return null;
    }
    return { alertId, evidenceRef };
  } catch {
    return null;
  }
}

async function writeDemoGif(frames) {
  if (frames.length !== gifFrameFiles.length) {
    throw new Error(
      `demo.gif requires ${gifFrameFiles.length} frames, got ${frames.length}`,
    );
  }
  const encoder = GIFEncoder();
  for (const { buffer, name } of frames) {
    const { data, width, height } = PNG.sync.read(buffer);
    const palette = quantize(data, 256);
    const index = applyPalette(data, palette);
    encoder.writeFrame(index, width, height, { palette, delay: GIF_FRAME_DELAY_MS });
    console.log(`GIF frame: ${name}`);
  }
  encoder.finish();
  const gifPath = path.join(outDir, "demo.gif");
  await writeFile(gifPath, Buffer.from(encoder.bytes()));
  console.log(`Captured demo GIF -> docs/assets/demo.gif (${frames.length} frames)`);
}

async function rebuildGifFromExisting() {
  await mkdir(outDir, { recursive: true });
  const frames = [];
  for (const { file, name } of gifFrameFiles) {
    const buffer = await readFile(path.join(outDir, file));
    frames.push({ buffer, name });
    console.log(`Loaded ${name} -> docs/assets/${file}`);
  }
  await writeDemoGif(frames);
}

async function assertHealthyPage(page, name) {
  const bodyText = await page.locator("body").innerText();
  if (ERROR_TEXT_RE.test(bodyText)) {
    throw new Error(
      `${name} shows an error state in the screenshot surface. Restore backends (audit/graph/APIs) and retry.\nMatched text excerpt: ${bodyText.slice(0, 240)}`,
    );
  }
}

async function capturePages(pages) {
  await mkdir(outDir, { recursive: true });

  const browser = await chromium.launch(launchOptions());
  const context = await browser.newContext({
    viewport: { width: 1280, height: 800 },
  });
  const page = await context.newPage();
  const gifFrames = [];

  for (const { url, file, name, readySelector } of pages) {
    await page.goto(url, { waitUntil: "networkidle" });
    if (readySelector) {
      await page.waitForSelector(readySelector, { timeout: 30_000 });
    }
    await page.waitForTimeout(700);
    await assertHealthyPage(page, name);
    const buffer = await page.screenshot({ fullPage: false });
    const dest = path.join(outDir, file);
    await writeFile(dest, buffer);
    gifFrames.push({ buffer, name });
    console.log(`Captured ${name} -> docs/assets/${file}`);
  }

  await writeDemoGif(gifFrames);
  await browser.close();
}

function consoleHomePages({ alertUrl, auditUrl } = {}) {
  return [
    {
      url: alertUrl ?? `${alertConsoleBase}/`,
      file: "alert-console.png",
      name: "Alert Console",
      readySelector: "main h1",
    },
    {
      url: auditUrl ?? `${auditExplorerBase}/`,
      file: "audit-explorer.png",
      name: "Audit Explorer",
      readySelector: "main h1",
    },
    {
      url: `${graphExplorerBase}/`,
      file: "graph-explorer.png",
      name: "Graph Explorer",
      readySelector: "main h1",
    },
    {
      url: `${simulationConsoleBase}/`,
      file: "simulation-console.png",
      name: "Simulation Console",
      readySelector: "main h1",
    },
    {
      url: `${reportConsoleBase}/`,
      file: "report-console.png",
      name: "Report Console",
      readySelector: "main h1",
    },
  ];
}

async function captureConsolesOnly() {
  await capturePages(consoleHomePages());
}

async function captureLive() {
  const linked = discoverLinkedAlert();
  const alertUrl = linked
    ? `${alertConsoleBase}/alerts/${linked.alertId}`
    : `${alertConsoleBase}/`;
  const auditUrl = linked
    ? `${auditExplorerBase}/entries/${linked.evidenceRef}`
    : `${auditExplorerBase}/`;

  if (!linked) {
    console.warn(
      "No linked alert found; capturing console home pages. Run ./scripts/demo-phase3.sh --trigger-alert for detail views.",
    );
  }

  await capturePages(consoleHomePages({ alertUrl, auditUrl }));
}

async function main() {
  if (process.argv.includes("--from-existing")) {
    await rebuildGifFromExisting();
    return;
  }
  if (process.argv.includes("--consoles-only")) {
    await captureConsolesOnly();
    return;
  }
  await captureLive();
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
