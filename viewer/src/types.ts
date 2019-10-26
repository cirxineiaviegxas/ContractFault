/**
 * Type definitions mirroring the JSON report emitted by the contractfault Go
 * CLI (schema tag "contractfault/v1"). Keeping these in one place lets the
 * viewer guard on the schema field and fail loudly on shape drift.
 */

export type Category = "breaking" | "additive" | "behavioral";
export type Severity = "critical" | "major" | "minor" | "info";
export type Verdict = "stable" | "tremor" | "shaken" | "rupture";

export interface Summary {
