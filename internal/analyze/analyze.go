// Package analyze compares two contract versions and classifies every
// difference into a change with a category (breaking, additive, behavioral),
// a severity and a stable location, then joins the changes against consumer
// manifests to compute a blast radius.
//
// Classification rules are documented inline and mirrored in docs/CONTRACT.md.
// The engine is deterministic: identical inputs always yield byte-identical
// reports because every collection is sorted before emission.
package analyze

import (
	"fmt"
	"sort"
	"strings"

	"github.com/seismic-labs/contractfault/internal/contract"
)

// Category is the top-level classification of a change.
type Category string

const (
	// Breaking changes can cause existing consumers to fail.
	Breaking Category = "breaking"
	// Additive changes extend the surface without breaking existing callers.
	Additive Category = "additive"
	// Behavioral changes keep the shape but alter runtime semantics.
	Behavioral Category = "behavioral"
)

// Severity ranks how disruptive a change is within its category.
type Severity int

const (
	// Info is a purely informational, low-risk change.
	Info Severity = iota + 1
	// Minor is a change consumers should be aware of.
	Minor
	// Major is a change likely to require consumer code updates.
	Major
	// Critical is a change that will break consumers unless they migrate.
	Critical
)

// String renders a severity as a stable lowercase label.
func (s Severity) String() string {
	switch s {
	case Critical:
		return "critical"
	case Major:
		return "major"
	case Minor:
		return "minor"
	default:
		return "info"
	}
}

// Change is a single classified difference between two contract versions.
type Change struct {
	// Code is a stable machine identifier ("endpoint.removed").
	Code string `json:"code"`
	// Category is breaking / additive / behavioral.
	Category Category `json:"category"`
	// Severity ranks impact within the category.
	Severity string `json:"severity"`
	// Location is a human/consumer-correlatable path ("getOrder",
	// "Order.total", "getOrder#query:status").
	Location string `json:"location"`
	// Endpoint is the operation ID this change is attached to, if any. Used to
	// join against consumer endpoint usage.
	Endpoint string `json:"endpoint,omitempty"`
	// FieldPath is the "Type.field" path this change touches, if any.
	FieldPath string `json:"fieldPath,omitempty"`
	// ParamKey is the parameter key this change touches, if any.
	ParamKey string `json:"paramKey,omitempty"`
	// Detail is a one-line human description of what changed.
