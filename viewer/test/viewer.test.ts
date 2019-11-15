import { test } from "node:test";
import assert from "node:assert/strict";
import { parseReport, SCHEMA_ID, type Report } from "../src/types.ts";
import { magnitudeBar, renderSvg, renderTerminal } from "../src/seismic.ts";
import { exitCodeFor, parseArgs } from "../src/cli.ts";

function sample(overrides: Partial<Report> = {}): Report {
  return {
    schema: SCHEMA_ID,
    service: "orders-api",
    fromVersion: "1.4.0",
    toVersion: "2.0.0",
