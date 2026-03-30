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
	default:
		return 1
	}
}

// Validate enforces manifest invariants.
func (m *Manifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("consumer: manifest has empty name")
	}
	switch strings.ToLower(m.Criticality) {
	case "low", "medium", "high", "":
	default:
		return fmt.Errorf("consumer %q: invalid criticality %q", m.Name, m.Criticality)
	}
	for i, u := range m.Uses {
		if strings.TrimSpace(u.Endpoint) == "" {
			return fmt.Errorf("consumer %q: usage %d has empty endpoint", m.Name, i)
		}
	}
	if strings.TrimSpace(m.Criticality) == "" {
		m.Criticality = "low"
	}
	return nil
}

// UsesEndpoint reports whether the consumer calls the given endpoint ID.
func (m Manifest) UsesEndpoint(id string) bool {
	for _, u := range m.Uses {
		if u.Endpoint == id {
			return true
		}
	}
	return false
}

// FieldPaths returns the union of read and write field paths across all usages.
func (m Manifest) FieldPaths() map[string]bool {
	out := map[string]bool{}
	for _, u := range m.Uses {
		for _, f := range u.ReadsFields {
			out[f] = true
		}
		for _, f := range u.WritesFields {
			out[f] = true
		}
	}
	return out
}

// ParamKeys returns the union of parameter keys used across all usages.
func (m Manifest) ParamKeys() map[string]bool {
	out := map[string]bool{}
	for _, u := range m.Uses {
		for _, p := range u.Params {
			out[p] = true
		}
	}
	return out
}

// Load parses a single consumer manifest file.
func Load(path string) (*Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("consumer: reading %s: %w", path, err)
	}
	var m Manifest
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("consumer: decoding %s: %w", path, err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// LoadAll parses every manifest at the given paths and returns them sorted by
// name for deterministic downstream ordering.
func LoadAll(paths []string) ([]Manifest, error) {
	out := make([]Manifest, 0, len(paths))
	seen := map[string]bool{}
	for _, p := range paths {
		m, err := Load(p)
		if err != nil {
			return nil, err
		}
		if seen[m.Name] {
			return nil, fmt.Errorf("consumer: duplicate consumer name %q", m.Name)
		}
		seen[m.Name] = true
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
