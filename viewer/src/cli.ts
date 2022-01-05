#!/usr/bin/env node
/**
 * contractfault-viewer reads a JSON report emitted by the Go CLI and renders
 * it as a terminal seismic map or writes an SVG impact seismograph.
 *
 * Usage:
 *   contractfault-viewer <report.json> [--svg out.svg] [--no-color]
 *
 * Exit codes mirror the analyzer: 2 when the report contains breaking changes,
 * 1 when it contains behavioral-only changes, 0 when stable. This lets the
 * viewer double as a CI gate when consuming a stored report.
 */

import { readFileSync, writeFileSync } from "node:fs";
import { parseReport } from "./types.js";
import { renderSvg, renderTerminal } from "./seismic.js";

interface Args {
  input: string;
  svgOut?: string;
  color: boolean;
}

export function parseArgs(argv: string[]): Args {
