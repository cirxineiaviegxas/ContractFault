// Package impact joins classified contract changes against consumer manifests
// to compute a blast radius: which named consumers are affected by which
// changes, at what severity, and an aggregate seismic risk score.
//
// The metaphor throughout contractfault is seismic: a contract change is a
// tremor, the affected consumers are the towns along the fault line, and the
// risk score is the magnitude. This package produces the Report struct that is
// serialized to JSON for the TypeScript viewer and rendered to text for CI.
package impact

import (
	"sort"

	"github.com/seismic-labs/contractfault/internal/analyze"
	"github.com/seismic-labs/contractfault/internal/consumer"
)

// Report is the top-level, serializable output of the analyzer.
type Report struct {
	// Schema is a version tag so the TypeScript viewer can guard on shape.
	Schema string `json:"schema"`
	// Service is the analyzed service name.
	Service string `json:"service"`
	// FromVersion / ToVersion bracket the comparison.
	FromVersion string `json:"fromVersion"`
	ToVersion   string `json:"toVersion"`
	// Summary aggregates counts and the overall magnitude.
	Summary Summary `json:"summary"`
	// Changes is the full classified change set, each annotated with the
	// consumers it impacts.
	Changes []ChangeImpact `json:"changes"`
	// Consumers is the per-consumer rollup of impacts.
	Consumers []ConsumerImpact `json:"consumers"`
}

// Summary holds aggregate metrics for the whole comparison.
type Summary struct {
	Breaking   int `json:"breaking"`
	Additive   int `json:"additive"`
	Behavioral int `json:"behavioral"`
	// AffectedConsumers is the number of consumers with at least one impact.
	AffectedConsumers int `json:"affectedConsumers"`
	// Magnitude is the seismic risk score on a 0.0-10.0 scale.
	Magnitude float64 `json:"magnitude"`
	// Verdict is a plain-language summary ("stable", "shaken", "rupture").
	Verdict string `json:"verdict"`
}

// ChangeImpact is a classified change plus the consumers it affects.
type ChangeImpact struct {
	analyze.Change
	// Consumers lists the names of consumers impacted by this change.
	Consumers []string `json:"consumers"`
}

// ConsumerImpact is one consumer's rollup across all changes.
type ConsumerImpact struct {
	Name        string `json:"name"`
	Team        string `json:"team"`
	Criticality string `json:"criticality"`
	// Breaking / Additive / Behavioral count impacts by category.
	Breaking   int `json:"breaking"`
	Additive   int `json:"additive"`
	Behavioral int `json:"behavioral"`
	// Codes lists the change codes affecting this consumer, sorted.
	Codes []string `json:"codes"`
	// Score is this consumer's weighted contribution to the magnitude.
	Score float64 `json:"score"`
}

// severityWeight maps a severity label to a magnitude weight.
func severityWeight(sev string) float64 {
	switch sev {
	case "critical":
		return 4.0
