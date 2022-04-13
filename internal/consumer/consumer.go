// Package consumer models the usage manifests that named downstream services
// publish to declare which parts of a contract they actually depend on.
//
// The impact analyzer joins the set of detected contract changes against these
// manifests so that a report can say "change X breaks consumer Y" rather than
// merely "change X exists". A manifest is deliberately coarse: it lists the
// endpoints a consumer calls and the fields it reads or writes. That is enough
// to compute a precise blast radius without forcing consumers to publish their
// entire source tree.
package consumer

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Manifest describes one named consumer's dependence on a contract.
type Manifest struct {
	// Name is the consumer's service name ("checkout-web").
	Name string `json:"name"`
	// Team is the owning team, surfaced in reports for routing.
	Team string `json:"team"`
	// Criticality is one of "low", "medium", "high"; it weights the severity
	// of impacts in the aggregate risk score.
	Criticality string `json:"criticality"`
	// Uses lists the specific contract elements the consumer relies on.
	Uses []Usage `json:"uses"`
}

// Usage is a single dependency edge from a consumer to a contract element.
type Usage struct {
	// Endpoint is the operation ID the consumer calls.
	Endpoint string `json:"endpoint"`
	// ReadsFields lists "Type.field" paths the consumer reads from responses.
	ReadsFields []string `json:"readsFields,omitempty"`
	// WritesFields lists "Type.field" paths the consumer sends in requests.
	WritesFields []string `json:"writesFields,omitempty"`
	// Params lists request parameter keys ("query:status") the consumer sets.
	Params []string `json:"params,omitempty"`
}

// CriticalityWeight maps the criticality label to a numeric multiplier.
func (m Manifest) CriticalityWeight() int {
	switch strings.ToLower(m.Criticality) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
