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
