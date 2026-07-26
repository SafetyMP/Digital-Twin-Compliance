/**
 * Capture README screenshots and demo GIF from console UIs.
 *
 * Live capture (warm stack with linked alert for Path B GIF frames):
 *   ./scripts/demo-phase3.sh --trigger-alert --restart-policies
 *   npm run screenshots
 *
 * Capture only Phase 4/5 home pages (no alert prerequisite):
 *   npm run screenshots -- --consoles-only
 *
 * Rebuild GIF only from existing alert/audit PNGs (no browser):
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
const GIF_FRAME_DELAY_MS = 2_000;

const gifFrameFiles = [
  { file: "alert-console.png", name: "Alert Console" },
  { file: "audit-explorer.png", name: "Audit Explorer" },
];

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
  console.log(`Captured demo GIF -> docs/assets/demo.gif`);
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

async function capturePages(pages, { forGif = false } = {}) {
  await mkdir(outDir, { recursive: true });

  const browser = await chromium.launch(launchOptions());
  const context = await browser.newContext({
    viewport: { width: 1280, height: 800 },
  });
  const page = await context.newPage();
  const gifFrames = [];

  for (const { url, file, name } of pages) {
    await page.goto(url, { waitUntil: "networkidle" });
    await page.waitForTimeout(500);
    const buffer = await page.screenshot({ fullPage: false });
    const dest = path.join(outDir, file);
    await writeFile(dest, buffer);
    if (forGif) {
      gifFrames.push({ buffer, name });
    }
    console.log(`Captured ${name} -> docs/assets/${file}`);
  }

  if (forGif && gifFrames.length > 0) {
    await writeDemoGif(gifFrames);
  }
  await browser.close();
}

async function captureConsolesOnly() {
  await capturePages([
    { url: `${alertConsoleBase}/`, file: "alert-console.png", name: "Alert Console" },
    { url: `${auditExplorerBase}/`, file: "audit-explorer.png", name: "Audit Explorer" },
    { url: `${graphExplorerBase}/`, file: "graph-explorer.png", name: "Graph Explorer" },
    {
      url: `${simulationConsoleBase}/`,
      file: "simulation-console.png",
      name: "Simulation Console",
    },
    { url: `${reportConsoleBase}/`, file: "report-console.png", name: "Report Console" },
  ]);
}

async function captureLive() {
  const linked = discoverLinkedAlert();
  if (!linked) {
    throw new Error(
      "No linked alert found. Run ./scripts/demo-phase3.sh --trigger-alert or set ALERT_ID and EVIDENCE_REF. For home-page shots only: npm run screenshots -- --consoles-only",
    );
  }

  await capturePages(
    [
      {
        url: `${alertConsoleBase}/alerts/${linked.alertId}`,
        file: "alert-console.png",
        name: "Alert Console",
      },
      {
        url: `${auditExplorerBase}/entries/${linked.evidenceRef}`,
        file: "audit-explorer.png",
        name: "Audit Explorer",
      },
    ],
    { forGif: true },
  );

  await capturePages([
    { url: `${graphExplorerBase}/`, file: "graph-explorer.png", name: "Graph Explorer" },
    {
      url: `${simulationConsoleBase}/`,
      file: "simulation-console.png",
      name: "Simulation Console",
    },
    { url: `${reportConsoleBase}/`, file: "report-console.png", name: "Report Console" },
  ]);
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
