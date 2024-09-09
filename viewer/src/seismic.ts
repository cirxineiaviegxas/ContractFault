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
