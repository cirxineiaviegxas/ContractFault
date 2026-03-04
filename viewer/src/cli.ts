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
  const args: Args = { input: "", color: true };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === "--svg") {
      args.svgOut = argv[++i];
    } else if (a === "--no-color") {
      args.color = false;
    } else if (!a.startsWith("-") && args.input === "") {
      args.input = a;
    } else {
      throw new Error(`unknown argument: ${a}`);
    }
  }
  if (args.input === "") {
    throw new Error("usage: contractfault-viewer <report.json> [--svg out.svg] [--no-color]");
  }
  return args;
}

export function exitCodeFor(report: ReturnType<typeof parseReport>): number {
  if (report.summary.breaking > 0) return 2;
  if (report.summary.behavioral > 0) return 1;
  return 0;
}

function main(argv: string[]): number {
  let args: Args;
  try {
    args = parseArgs(argv);
  } catch (err) {
    process.stderr.write(`contractfault-viewer: ${(err as Error).message}\n`);
    return 3;
  }

  let report;
  try {
    const raw = readFileSync(args.input, "utf8");
    report = parseReport(JSON.parse(raw));
  } catch (err) {
    process.stderr.write(`contractfault-viewer: ${(err as Error).message}\n`);
    return 3;
  }

