/**
 * Type definitions mirroring the JSON report emitted by the contractfault Go
 * CLI (schema tag "contractfault/v1"). Keeping these in one place lets the
 * viewer guard on the schema field and fail loudly on shape drift.
 */

export type Category = "breaking" | "additive" | "behavioral";
export type Severity = "critical" | "major" | "minor" | "info";
export type Verdict = "stable" | "tremor" | "shaken" | "rupture";

export interface Summary {
  breaking: number;
  additive: number;
  behavioral: number;
  affectedConsumers: number;
  magnitude: number;
  verdict: Verdict;
}

export interface ChangeImpact {
  code: string;
  category: Category;
  severity: Severity;
  location: string;
  endpoint?: string;
  fieldPath?: string;
  paramKey?: string;
  detail: string;
  migration: string;
  consumers: string[];
}

export interface ConsumerImpact {
  name: string;
  team: string;
  criticality: string;
  breaking: number;
  additive: number;
  behavioral: number;
  codes: string[];
  score: number;
}

export interface Report {
  schema: string;
  service: string;
  fromVersion: string;
  toVersion: string;
  summary: Summary;
  changes: ChangeImpact[];
  consumers: ConsumerImpact[];
}

export const SCHEMA_ID = "contractfault/v1";

/**
 * parseReport validates that an unknown JSON value conforms to the expected
 * report shape and schema, throwing a descriptive error otherwise. It performs
 * structural checks rather than trusting the input, so malformed reports are
 * rejected at the boundary instead of causing confusing failures later.
 */
export function parseReport(value: unknown): Report {
  if (typeof value !== "object" || value === null) {
    throw new Error("report: expected a JSON object");
  }
  const r = value as Record<string, unknown>;
  if (r.schema !== SCHEMA_ID) {
    throw new Error(`report: unsupported schema ${String(r.schema)} (want ${SCHEMA_ID})`);
  }
  if (typeof r.service !== "string") {
    throw new Error("report: missing service");
  }
  if (typeof r.summary !== "object" || r.summary === null) {
    throw new Error("report: missing summary");
  }
  if (!Array.isArray(r.changes)) {
    throw new Error("report: changes must be an array");
  }
  if (!Array.isArray(r.consumers)) {
    throw new Error("report: consumers must be an array");
  }
  return value as Report;
}
