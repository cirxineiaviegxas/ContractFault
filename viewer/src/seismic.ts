/**
 * seismic renders a contractfault Report into two human-facing forms:
 *
 *  - renderTerminal: a colorized, aligned seismic map for the console, showing
 *    the magnitude meter, per-change tremors and the consumer blast radius.
 *  - renderSvg: a self-contained SVG "impact seismograph" that plots each
 *    affected consumer as a station on a fault line, sized by its risk score.
 *
 * Both are pure functions of the Report so the viewer is deterministic.
 */

import type { Category, ConsumerImpact, Report, Severity } from "./types.js";

const RESET = "\x1b[0m";
const BOLD = "\x1b[1m";
const RED = "\x1b[31m";
const YELLOW = "\x1b[33m";
const GREEN = "\x1b[32m";
const CYAN = "\x1b[36m";
const DIM = "\x1b[2m";

interface RenderOptions {
  /** color disables ANSI escapes when false (e.g. piping to a file). */
  color: boolean;
}

function severityColor(sev: Severity): string {
  switch (sev) {
    case "critical":
      return RED;
    case "major":
      return YELLOW;
    case "minor":
      return CYAN;
    default:
      return DIM;
  }
}

function categoryGlyph(cat: Category): string {
  switch (cat) {
    case "breaking":
      return "!!";
    case "behavioral":
      return "~~";
    default:
      return "++";
  }
}

function paint(s: string, color: string, on: boolean): string {
  return on ? `${color}${s}${RESET}` : s;
}

/** magnitudeBar draws a 40-cell Richter-style meter for a 0-10 magnitude. */
export function magnitudeBar(magnitude: number, width = 40): string {
  const clamped = Math.max(0, Math.min(10, magnitude));
  const filled = Math.round((clamped / 10) * width);
  return "[" + "#".repeat(filled) + "-".repeat(width - filled) + "]";
}

/** renderTerminal produces the console seismic map. */
export function renderTerminal(report: Report, opts: RenderOptions = { color: true }): string {
  const c = opts.color;
  const lines: string[] = [];
  const s = report.summary;

  lines.push(paint("═══ contractfault seismic map ═══", BOLD + CYAN, c));
  lines.push(`service : ${report.service}`);
  lines.push(`shift   : ${report.fromVersion} → ${report.toVersion}`);

  const verdictColor = s.verdict === "rupture" ? RED : s.verdict === "shaken" ? YELLOW : GREEN;
  lines.push(
    `${paint("magnitude", BOLD, c)} ${s.magnitude.toFixed(1)}  ` +
      paint(s.verdict.toUpperCase(), BOLD + verdictColor, c),
  );
  lines.push(paint(magnitudeBar(s.magnitude), verdictColor, c));
  lines.push(
    `${paint("breaking " + s.breaking, RED, c)}  ` +
      `${paint("additive " + s.additive, GREEN, c)}  ` +
      `${paint("behavioral " + s.behavioral, YELLOW, c)}  ` +
      `affected ${s.affectedConsumers}`,
  );
  lines.push("");

  lines.push(paint("TREMORS", BOLD, c));
  for (const ch of report.changes) {
    const glyph = paint(categoryGlyph(ch.category), severityColor(ch.severity), c);
    const loc = ch.location.padEnd(28).slice(0, 28);
    lines.push(`  ${glyph} ${paint(ch.severity.padEnd(8), severityColor(ch.severity), c)} ${loc} ${ch.detail}`);
    if (ch.consumers && ch.consumers.length > 0) {
      lines.push(`      ${paint("hits", DIM, c)} ${ch.consumers.join(", ")}`);
    }
  }
  lines.push("");

  lines.push(paint("BLAST RADIUS", BOLD, c));
  for (const cons of report.consumers) {
    lines.push(renderConsumerLine(cons, c));
  }
  return lines.join("\n") + "\n";
}

function renderConsumerLine(cons: ConsumerImpact, color: boolean): string {
  const total = cons.breaking + cons.additive + cons.behavioral;
  const name = cons.name.padEnd(20);
  if (total === 0) {
    return `  ${name} ${paint("quiet", DIM, color)}`;
  }
  const bar = "▓".repeat(Math.min(20, Math.max(1, Math.round(cons.score / 2))));
  const barColor = cons.breaking > 0 ? RED : cons.behavioral > 0 ? YELLOW : GREEN;
  return (
    `  ${name} ${paint(bar, barColor, color)} ` +
    `${paint("score " + cons.score.toFixed(1), BOLD, color)} ` +
    `(b:${cons.breaking} a:${cons.additive} v:${cons.behavioral}) ` +
    `${paint(cons.criticality, DIM, color)}`
  );
}

const SVG_COLORS: Record<string, string> = {
  breaking: "#e5484d",
  behavioral: "#f5a623",
  additive: "#30a46c",
  quiet: "#6b7280",
};

/**
 * renderSvg produces a standalone impact seismograph as an SVG string. Each
 * consumer is a station plotted left-to-right by descending score; its dot
 * radius encodes the score and its color encodes the worst category it saw.
 * A dashed fault line runs through the stations and the magnitude is printed
 * as a header. The output embeds no remote references.
 */
export function renderSvg(report: Report): string {
  const width = 720;
  const height = 360;
  const consumers = report.consumers;
  const maxScore = Math.max(1, ...consumers.map((x) => x.score));
  const marginX = 60;
  const usable = width - marginX * 2;
  const step = consumers.length > 1 ? usable / (consumers.length - 1) : 0;
  const faultY = 210;

  const stations = consumers
    .map((cons, i) => {
      const x = marginX + (consumers.length > 1 ? step * i : usable / 2);
      const total = cons.breaking + cons.additive + cons.behavioral;
      const worst =
        cons.breaking > 0 ? "breaking" : cons.behavioral > 0 ? "behavioral" : total > 0 ? "additive" : "quiet";
      const color = SVG_COLORS[worst];
      const r = 6 + (cons.score / maxScore) * 26;
      const amp = 8 + (cons.score / maxScore) * 60;
      return { cons, x, color, r, amp };
    });

  // Build a seismograph trace: a jagged polyline whose spike heights follow
  // each station's amplitude, giving the "impact seismograph" its shape.
  const tracePoints: string[] = [];
  tracePoints.push(`${marginX - 40},${faultY}`);
  for (const st of stations) {
    tracePoints.push(`${(st.x - 14).toFixed(1)},${faultY}`);
    tracePoints.push(`${st.x.toFixed(1)},${(faultY - st.amp).toFixed(1)}`);
    tracePoints.push(`${(st.x + 14).toFixed(1)},${(faultY + st.amp * 0.4).toFixed(1)}`);
    tracePoints.push(`${(st.x + 18).toFixed(1)},${faultY}`);
  }
  tracePoints.push(`${width - marginX + 40},${faultY}`);

  const verdictColor =
    report.summary.verdict === "rupture"
      ? SVG_COLORS.breaking
      : report.summary.verdict === "shaken"
        ? SVG_COLORS.behavioral
        : SVG_COLORS.additive;

  const stationSvg = stations
    .map((st) => {
      const label = st.cons.name;
      return `
    <g>
      <circle cx="${st.x.toFixed(1)}" cy="${faultY}" r="${st.r.toFixed(1)}" fill="${st.color}" fill-opacity="0.85">
        <animate attributeName="r" values="${st.r.toFixed(1)};${(st.r + 4).toFixed(1)};${st.r.toFixed(1)}" dur="${(2 + st.amp / 40).toFixed(2)}s" repeatCount="indefinite" />
      </circle>
      <text x="${st.x.toFixed(1)}" y="${faultY + 34}" text-anchor="middle" font-size="12" fill="#c9d1d9">${escapeXml(label)}</text>
      <text x="${st.x.toFixed(1)}" y="${faultY + 50}" text-anchor="middle" font-size="10" fill="#8b949e">${st.cons.score.toFixed(1)}</text>
    </g>`;
    })
    .join("");

  return `<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" role="img" aria-label="contractfault impact seismograph for ${escapeXml(report.service)}">
  <rect width="${width}" height="${height}" fill="#0d1117" />
  <text x="${marginX - 40}" y="48" font-size="22" font-family="monospace" fill="#c9d1d9">${escapeXml(report.service)} ${escapeXml(report.fromVersion)} → ${escapeXml(report.toVersion)}</text>
  <text x="${marginX - 40}" y="78" font-size="16" font-family="monospace" fill="${verdictColor}">magnitude ${report.summary.magnitude.toFixed(1)} · ${report.summary.verdict.toUpperCase()}</text>
  <line x1="${marginX - 40}" y1="${faultY}" x2="${width - marginX + 40}" y2="${faultY}" stroke="#30363d" stroke-width="1" stroke-dasharray="6 5" />
  <polyline points="${tracePoints.join(" ")}" fill="none" stroke="${verdictColor}" stroke-width="2" stroke-opacity="0.9">
    <animate attributeName="stroke-opacity" values="0.4;1;0.4" dur="3s" repeatCount="indefinite" />
  </polyline>
  ${stationSvg}
  <text x="${marginX - 40}" y="${height - 20}" font-size="11" font-family="monospace" fill="#8b949e">breaking ${report.summary.breaking} · additive ${report.summary.additive} · behavioral ${report.summary.behavioral} · affected ${report.summary.affectedConsumers}</text>
</svg>
`;
}

function escapeXml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
