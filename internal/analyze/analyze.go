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
