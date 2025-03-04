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
	Name string `json:"name,omitempty"`
	// Kind is one of "object", "string", "integer", "number", "boolean".
	// Object types carry Fields; scalar types may carry Enum.
	Kind string `json:"kind"`
	// Fields are the members of an object type keyed by field name.
	Fields map[string]Field `json:"fields,omitempty"`
	// Enum lists the permitted values for a scalar type. An empty slice means
	// the scalar is unconstrained.
	Enum []string `json:"enum,omitempty"`
	// Deprecated marks the whole type as scheduled for removal.
	Deprecated bool `json:"deprecated,omitempty"`
}

// Field is a member of an object type.
type Field struct {
	// TypeRef names the type of the field. It is either a builtin scalar
	// ("string", "integer", "number", "boolean") or the name of a Type.
	TypeRef string `json:"type"`
	// Required indicates the field must always be present in a payload.
	Required bool `json:"required,omitempty"`
	// Nullable indicates the field may carry an explicit null value.
	Nullable bool `json:"nullable,omitempty"`
	// Array indicates the field is a homogeneous list of TypeRef values.
	Array bool `json:"array,omitempty"`
	// Deprecated marks the field as scheduled for removal.
	Deprecated bool `json:"deprecated,omitempty"`
	// Doc is a short human description used in migration hints.
	Doc string `json:"doc,omitempty"`
}

// Endpoint is a single API operation.
type Endpoint struct {
	// ID is a stable operation identifier ("getOrder"). It is the primary key
	// used to correlate an endpoint across versions.
	ID string `json:"id"`
	// Method is the HTTP verb, upper-cased on load.
	Method string `json:"method"`
	// Path is the templated URL path ("/orders/{id}").
	Path string `json:"path"`
	// Params are the request parameters (path, query, header).
	Params []Param `json:"params,omitempty"`
	// RequestType names the request body Type, empty for bodiless operations.
	RequestType string `json:"requestType,omitempty"`
	// Responses maps status code (as string) to the response body Type name.
	Responses map[string]string `json:"responses,omitempty"`
	// Deprecated marks the operation as scheduled for removal.
	Deprecated bool `json:"deprecated,omitempty"`
	// Idempotent documents whether repeated calls are safe; a change here is a
	// behavioral change even though the shape is identical.
	Idempotent bool `json:"idempotent,omitempty"`
}

// Param is a request parameter.
type Param struct {
	// Name is the parameter name.
	Name string `json:"name"`
	// In is the location: "path", "query" or "header".
	In string `json:"in"`
	// TypeRef is the scalar type of the parameter value.
	TypeRef string `json:"type"`
	// Required indicates the parameter must be supplied.
	Required bool `json:"required,omitempty"`
	// Enum constrains the permitted values.
	Enum []string `json:"enum,omitempty"`
}

// Key returns the correlation key for a parameter within an endpoint.
func (p Param) Key() string { return p.In + ":" + p.Name }

// EndpointByID indexes the contract's endpoints by their stable ID.
func (c *Contract) EndpointByID() map[string]Endpoint {
	out := make(map[string]Endpoint, len(c.Endpoints))
	for _, e := range c.Endpoints {
		out[e.ID] = e
	}
	return out
}

// Validate checks structural invariants and returns a descriptive error when
// the document is internally inconsistent. It is deliberately strict so that a
// malformed contract fails loudly rather than producing a misleading report.
func (c *Contract) Validate() error {
	if strings.TrimSpace(c.Service) == "" {
		return fmt.Errorf("contract: service name is empty")
	}
	if strings.TrimSpace(c.Version) == "" {
		return fmt.Errorf("contract: version is empty")
	}
	seen := make(map[string]bool)
	for _, e := range c.Endpoints {
		if e.ID == "" {
			return fmt.Errorf("contract: endpoint with empty id (path %q)", e.Path)
		}
		if seen[e.ID] {
			return fmt.Errorf("contract: duplicate endpoint id %q", e.ID)
		}
		seen[e.ID] = true
		if e.Method == "" {
			return fmt.Errorf("contract: endpoint %q has empty method", e.ID)
		}
		if e.RequestType != "" && !c.knownType(e.RequestType) {
			return fmt.Errorf("contract: endpoint %q references unknown request type %q", e.ID, e.RequestType)
		}
		for code, ref := range e.Responses {
			if ref != "" && !c.knownType(ref) {
				return fmt.Errorf("contract: endpoint %q response %s references unknown type %q", e.ID, code, ref)
			}
		}
	}
	for name, t := range c.Types {
		for fname, f := range t.Fields {
			if f.TypeRef != "" && !c.knownType(f.TypeRef) {
				return fmt.Errorf("contract: type %q field %q references unknown type %q", name, fname, f.TypeRef)
			}
		}
	}
	return nil
}

func (c *Contract) knownType(ref string) bool {
	switch ref {
	case "string", "integer", "number", "boolean", "object":
		return true
	}
	_, ok := c.Types[ref]
	return ok
}

// SortedTypeNames returns type names in deterministic order.
func (c *Contract) SortedTypeNames() []string {
	names := make([]string, 0, len(c.Types))
	for n := range c.Types {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Load reads and parses a contract from a JSON file at path, normalizes it and
// validates it. Any error is wrapped with the source path for diagnosability.
func Load(path string) (*Contract, error) {
	raw, err := os.ReadFile(path)
