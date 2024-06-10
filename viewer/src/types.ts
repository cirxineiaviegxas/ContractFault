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
