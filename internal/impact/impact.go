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
