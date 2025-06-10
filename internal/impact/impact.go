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
	case "major":
		return 2.5
	case "minor":
		return 1.0
	default:
		return 0.25
	}
}

// Build joins a diff against consumer manifests into a finished report.
func Build(diff *analyze.Diff, consumers []consumer.Manifest) *Report {
	r := &Report{
		Schema:      "contractfault/v1",
		Service:     diff.Service,
		FromVersion: diff.FromVersion,
		ToVersion:   diff.ToVersion,
	}

	// Per-consumer accumulators.
	type acc struct {
		m          consumer.Manifest
		breaking   int
		additive   int
		behavioral int
		codes      map[string]bool
		score      float64
	}
	accs := make(map[string]*acc, len(consumers))
	order := make([]string, 0, len(consumers))
	for _, m := range consumers {
		accs[m.Name] = &acc{m: m, codes: map[string]bool{}}
		order = append(order, m.Name)
	}

	var magnitude float64
	for _, ch := range diff.Changes {
		affected := affectedConsumers(ch, consumers)
		ci := ChangeImpact{Change: ch, Consumers: affected}
		r.Changes = append(r.Changes, ci)

		// The base tremor magnitude of a change is its severity weight; it is
		// amplified by each affected consumer's criticality.
		base := severityWeight(ch.Severity)
		for _, name := range affected {
			a := accs[name]
			switch ch.Category {
			case analyze.Breaking:
				a.breaking++
			case analyze.Additive:
				a.additive++
			case analyze.Behavioral:
				a.behavioral++
			}
			a.codes[ch.Code] = true
			contribution := base * float64(a.m.CriticalityWeight())
			a.score += contribution
			magnitude += contribution
		}
	}

	// Materialize consumer rollups in stable order.
	affectedCount := 0
	for _, name := range order {
		a := accs[name]
		total := a.breaking + a.additive + a.behavioral
		if total > 0 {
			affectedCount++
		}
		codes := make([]string, 0, len(a.codes))
		for c := range a.codes {
			codes = append(codes, c)
		}
		sort.Strings(codes)
		r.Consumers = append(r.Consumers, ConsumerImpact{
			Name:        a.m.Name,
			Team:        a.m.Team,
			Criticality: a.m.Criticality,
			Breaking:    a.breaking,
			Additive:    a.additive,
			Behavioral:  a.behavioral,
			Codes:       codes,
			Score:       round1(a.score),
		})
	}
	sort.SliceStable(r.Consumers, func(i, j int) bool {
		if r.Consumers[i].Score != r.Consumers[j].Score {
			return r.Consumers[i].Score > r.Consumers[j].Score
		}
		return r.Consumers[i].Name < r.Consumers[j].Name
	})

	counts := diff.Counts()
	r.Summary = Summary{
		Breaking:          counts[analyze.Breaking],
		Additive:          counts[analyze.Additive],
		Behavioral:        counts[analyze.Behavioral],
		AffectedConsumers: affectedCount,
		Magnitude:         magnitudeScale(magnitude),
	}
	r.Summary.Verdict = verdict(r.Summary)
	return r
}

// affectedConsumers returns the sorted names of consumers impacted by a change.
func affectedConsumers(ch analyze.Change, consumers []consumer.Manifest) []string {
	// Always non-nil so the JSON report renders "consumers": [] rather than
	// null, which keeps the TypeScript viewer's array handling simple.
	out := []string{}
	for _, m := range consumers {
		if consumerAffected(ch, m) {
			out = append(out, m.Name)
		}
	}
	sort.Strings(out)
