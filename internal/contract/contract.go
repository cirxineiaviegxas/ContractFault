// Package contract defines the documented JSON contract format that
// contractfault ingests and the loaders that parse it from disk.
//
// The format is intentionally OpenAPI-like but self-contained: it models
// endpoints (methods + paths + params + responses) and reusable named types
// (objects with fields). The goal is to be practical to author by hand while
// still expressing the properties that matter for consumer-impact analysis:
// required-ness, nullability, enum membership, deprecation and status codes.
package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Contract is the top-level document. A contract is a versioned snapshot of an
// API surface: a set of named types and a set of endpoints.
type Contract struct {
	// Service is the logical name of the API (e.g. "orders-api").
	Service string `json:"service"`
	// Version is a human-readable semantic version string ("1.4.0").
	Version string `json:"version"`
	// Types is the catalogue of reusable object types keyed by type name.
	Types map[string]Type `json:"types"`
	// Endpoints is the list of operations exposed by the service.
	Endpoints []Endpoint `json:"endpoints"`
}

// Type is a reusable object type: a named bag of fields plus optional enum
// value constraints for scalar types.
type Type struct {
	// Name is the type identifier, mirrored from the map key for convenience.
