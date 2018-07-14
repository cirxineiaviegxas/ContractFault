// Command contractfault is a deterministic API contract consumer-impact
// analyzer. It ingests two versions of a documented JSON contract plus a set
// of consumer usage manifests, classifies every difference as breaking,
// additive or behavioral, joins those changes against the consumers that
// depend on the affected elements, and emits a JSON or text seismic report
// with a CI-friendly exit code.
//
// Usage:
//
//	contractfault -old before.json -new after.json -consumers 'consumers/*.json' [flags]
