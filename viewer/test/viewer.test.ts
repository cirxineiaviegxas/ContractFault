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
    summary: {
      breaking: 2,
      additive: 3,
      behavioral: 1,
      affectedConsumers: 2,
      magnitude: 7.2,
      verdict: "rupture",
    },
    changes: [
      {
        code: "endpoint.removed",
        category: "breaking",
        severity: "critical",
        location: "getReceipt",
        endpoint: "getReceipt",
        detail: "endpoint removed",
        migration: "stop calling it",
        consumers: ["analytics-etl", "receipts-mailer"],
      },
      {
        code: "field.added",
        category: "additive",
        severity: "info",
        location: "Error.traceId",
        fieldPath: "Error.traceId",
        detail: "field added",
        migration: "adopt it",
        consumers: [],
      },
    ],
    consumers: [
      {
        name: "analytics-etl",
        team: "data",
        criticality: "medium",
        breaking: 1,
        additive: 0,
        behavioral: 0,
        codes: ["endpoint.removed"],
        score: 5,
      },
      {
        name: "quiet-svc",
        team: "misc",
        criticality: "low",
        breaking: 0,
        additive: 0,
        behavioral: 0,
        codes: [],
        score: 0,
      },
    ],
    ...overrides,
  };
}

test("parseReport accepts a valid report", () => {
  const r = parseReport(sample());
  assert.equal(r.service, "orders-api");
});

test("parseReport rejects wrong schema", () => {
  assert.throws(() => parseReport({ ...sample(), schema: "other" }), /unsupported schema/);
});

test("parseReport rejects non-object", () => {
  assert.throws(() => parseReport(42), /expected a JSON object/);
});

test("magnitudeBar clamps and fills proportionally", () => {
  assert.equal(magnitudeBar(0, 10), "[----------]");
  assert.equal(magnitudeBar(10, 10), "[##########]");
  assert.equal(magnitudeBar(5, 10), "[#####-----]");
  // Out-of-range values are clamped, never overflow the bar width.
  assert.equal(magnitudeBar(99, 10), "[##########]");
});

test("renderTerminal without color contains key facts and no escapes", () => {
  const out = renderTerminal(sample(), { color: false });
  assert.match(out, /orders-api/);
  assert.match(out, /magnitude 7\.2/);
  assert.match(out, /RUPTURE/);
  assert.match(out, /analytics-etl/);
  assert.ok(!out.includes("\x1b["), "expected no ANSI escapes when color disabled");
});

test("renderTerminal with color emits ANSI escapes", () => {
  const out = renderTerminal(sample(), { color: true });
  assert.ok(out.includes("\x1b["), "expected ANSI escapes when color enabled");
});

test("renderSvg produces valid standalone SVG with a station per consumer", () => {
  const svg = renderSvg(sample());
  assert.match(svg, /^<\?xml/);
  assert.match(svg, /<svg[^>]+width="720"/);
  assert.match(svg, /analytics-etl/);
  assert.match(svg, /quiet-svc/);
  assert.match(svg, /<animate/); // animated
  assert.ok(!svg.includes("http://") || svg.includes("http://www.w3.org/2000/svg"),
    "no remote references other than the SVG namespace");
});

test("renderSvg escapes XML special characters", () => {
  const r = sample({ service: "a<b>&\"c" });
  const svg = renderSvg(r);
  assert.ok(!svg.includes("a<b>&\"c"));
  assert.match(svg, /a&lt;b&gt;&amp;&quot;c/);
});

test("exitCodeFor maps summary to CI codes", () => {
  assert.equal(exitCodeFor(sample()), 2);
  assert.equal(exitCodeFor(sample({ summary: { ...sample().summary, breaking: 0 } })), 1);
  assert.equal(
    exitCodeFor(sample({ summary: { ...sample().summary, breaking: 0, behavioral: 0 } })),
    0,
  );
});

test("parseArgs parses input, svg and no-color", () => {
  const a = parseArgs(["report.json", "--svg", "out.svg", "--no-color"]);
  assert.equal(a.input, "report.json");
  assert.equal(a.svgOut, "out.svg");
  assert.equal(a.color, false);
});

test("parseArgs requires an input file", () => {
  assert.throws(() => parseArgs([]), /usage:/);
});

test("parseArgs rejects unknown flags", () => {
  assert.throws(() => parseArgs(["r.json", "--bogus"]), /unknown argument/);
});
