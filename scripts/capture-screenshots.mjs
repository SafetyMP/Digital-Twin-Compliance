/**
 * Capture README screenshots and demo GIF from the Phase 3 alert + audit UIs.
 *
 * Live capture (warm stack with linked alert):
 *   ./scripts/demo-phase3.sh --trigger-alert --restart-policies
 *   npm run screenshots
 *
 * Rebuild GIF only from existing PNGs (no browser):
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
const alertDbUrl =
  process.env.ALERT_DB_URL ?? "postgres://alert:alert@localhost:5435/alerts?sslmode=disable";

/** Frame duration in centiseconds (200 = 2 seconds per screen). */
const GIF_FRAME_DELAY_CS = 200;

const frameFiles = [
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
    encoder.writeFrame(index, width, height, { palette, delay: GIF_FRAME_DELAY_CS });
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
  for (const { file, name } of frameFiles) {
    const buffer = await readFile(path.join(outDir, file));
    frames.push({ buffer, name });
    console.log(`Loaded ${name} -> docs/assets/${file}`);
  }
  await writeDemoGif(frames);
}

async function captureLive() {
  const linked = discoverLinkedAlert();
  if (!linked) {
    throw new Error(
      "No linked alert found. Run ./scripts/demo-phase3.sh --trigger-alert or set ALERT_ID and EVIDENCE_REF.",
    );
  }

  const pages = [
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
  ];

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
    gifFrames.push({ buffer, name });
    console.log(`Captured ${name} -> docs/assets/${file}`);
  }

  await writeDemoGif(gifFrames);
  await browser.close();
}

async function main() {
  if (process.argv.includes("--from-existing")) {
    await rebuildGifFromExisting();
    return;
  }
  await captureLive();
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
