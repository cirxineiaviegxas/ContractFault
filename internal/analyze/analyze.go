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
	Detail string `json:"detail"`
	// Migration is an actionable hint for consumers.
	Migration string `json:"migration"`
}

// severityRank exposes the numeric severity for sorting and scoring.
func severityRank(s string) int {
	switch s {
	case "critical":
		return 4
	case "major":
		return 3
	case "minor":
		return 2
	default:
		return 1
	}
}

// Diff holds the full set of classified changes between two versions.
type Diff struct {
	FromVersion string
	ToVersion   string
	Service     string
	Changes     []Change
}

// Compare classifies every difference between the old and new contract.
// The two contracts must describe the same service.
func Compare(oldC, newC *contract.Contract) (*Diff, error) {
	if oldC.Service != newC.Service {
		return nil, fmt.Errorf("analyze: service mismatch %q vs %q", oldC.Service, newC.Service)
	}
	d := &Diff{
		FromVersion: oldC.Version,
		ToVersion:   newC.Version,
		Service:     newC.Service,
	}
	d.diffEndpoints(oldC, newC)
	d.diffTypes(oldC, newC)
	d.sortChanges()
	return d, nil
}

func (d *Diff) add(c Change) { d.Changes = append(d.Changes, c) }

// diffEndpoints classifies endpoint-level differences.
func (d *Diff) diffEndpoints(oldC, newC *contract.Contract) {
	oldEps := oldC.EndpointByID()
	newEps := newC.EndpointByID()

	for id, oe := range oldEps {
		ne, ok := newEps[id]
		if !ok {
			d.add(Change{
				Code:      "endpoint.removed",
				Category:  Breaking,
				Severity:  Critical.String(),
				Location:  id,
				Endpoint:  id,
				Detail:    fmt.Sprintf("endpoint %s %s (%s) was removed", oe.Method, oe.Path, id),
				Migration: "Stop calling this endpoint; migrate to a replacement operation before upgrading.",
			})
			continue
		}
		d.diffEndpointPair(oe, ne)
	}
	for id, ne := range newEps {
		if _, ok := oldEps[id]; !ok {
			d.add(Change{
				Code:      "endpoint.added",
				Category:  Additive,
				Severity:  Info.String(),
				Location:  id,
				Endpoint:  id,
				Detail:    fmt.Sprintf("endpoint %s %s (%s) was added", ne.Method, ne.Path, id),
				Migration: "No action required; new capability is available to adopt.",
			})
		}
	}
}

// diffEndpointPair classifies differences within a single correlated endpoint.
func (d *Diff) diffEndpointPair(oe, ne contract.Endpoint) {
	if oe.Method != ne.Method {
		d.add(Change{
			Code:      "endpoint.method.changed",
			Category:  Breaking,
			Severity:  Critical.String(),
			Location:  ne.ID,
			Endpoint:  ne.ID,
			Detail:    fmt.Sprintf("method changed %s -> %s", oe.Method, ne.Method),
			Migration: fmt.Sprintf("Update the HTTP verb from %s to %s.", oe.Method, ne.Method),
		})
	}
	if oe.Path != ne.Path {
		d.add(Change{
			Code:      "endpoint.path.changed",
			Category:  Breaking,
			Severity:  Major.String(),
			Location:  ne.ID,
			Endpoint:  ne.ID,
			Detail:    fmt.Sprintf("path changed %s -> %s", oe.Path, ne.Path),
			Migration: fmt.Sprintf("Update the request URL template to %s.", ne.Path),
		})
	}
	if !oe.Deprecated && ne.Deprecated {
		d.add(Change{
			Code:      "endpoint.deprecated",
			Category:  Behavioral,
			Severity:  Minor.String(),
			Location:  ne.ID,
			Endpoint:  ne.ID,
			Detail:    "endpoint marked deprecated",
			Migration: "Plan migration off this endpoint; it may be removed in a future version.",
		})
	}
	if oe.Idempotent != ne.Idempotent {
		d.add(Change{
			Code:      "endpoint.idempotency.changed",
			Category:  Behavioral,
			Severity:  Major.String(),
			Location:  ne.ID,
			Endpoint:  ne.ID,
			Detail:    fmt.Sprintf("idempotency changed %v -> %v", oe.Idempotent, ne.Idempotent),
			Migration: "Review retry logic: repeated calls may no longer be safe.",
		})
	}
	if oe.RequestType != ne.RequestType {
		d.add(Change{
			Code:      "endpoint.requestType.changed",
			Category:  Breaking,
			Severity:  Major.String(),
			Location:  ne.ID,
			Endpoint:  ne.ID,
			Detail:    fmt.Sprintf("request body type changed %q -> %q", oe.RequestType, ne.RequestType),
			Migration: "Rebuild the request body to match the new type.",
		})
	}
	d.diffParams(oe, ne)
	d.diffResponses(oe, ne)
}

// diffParams classifies parameter differences within an endpoint.
func (d *Diff) diffParams(oe, ne contract.Endpoint) {
	oldP := map[string]contract.Param{}
	for _, p := range oe.Params {
		oldP[p.Key()] = p
	}
	newP := map[string]contract.Param{}
	for _, p := range ne.Params {
		newP[p.Key()] = p
	}
	for k, op := range oldP {
		np, ok := newP[k]
		if !ok {
			cat, sev := Additive, Minor
			mig := "Parameter removed; stop sending it (ignored if still present)."
			if op.Required {
				cat, sev = Breaking, Major
				mig = "A required parameter was removed; update calls that relied on it."
			}
			d.add(Change{
				Code:      "param.removed",
				Category:  cat,
				Severity:  sev.String(),
				Location:  ne.ID + "#" + k,
				Endpoint:  ne.ID,
				ParamKey:  k,
				Detail:    fmt.Sprintf("parameter %s removed", k),
				Migration: mig,
			})
			continue
		}
		if !op.Required && np.Required {
			d.add(Change{
				Code:      "param.required.added",
				Category:  Breaking,
				Severity:  Major.String(),
				Location:  ne.ID + "#" + k,
				Endpoint:  ne.ID,
				ParamKey:  k,
				Detail:    fmt.Sprintf("parameter %s became required", k),
				Migration: fmt.Sprintf("Always supply %s; requests without it will be rejected.", np.Name),
			})
		}
		if op.Required && !np.Required {
			d.add(Change{
				Code:      "param.required.relaxed",
				Category:  Additive,
				Severity:  Info.String(),
				Location:  ne.ID + "#" + k,
				Endpoint:  ne.ID,
				ParamKey:  k,
				Detail:    fmt.Sprintf("parameter %s became optional", k),
				Migration: "No action required; the parameter is now optional.",
			})
		}
		if removed := removedEnum(op.Enum, np.Enum); len(removed) > 0 {
			d.add(Change{
				Code:      "param.enum.removed",
				Category:  Breaking,
				Severity:  Major.String(),
				Location:  ne.ID + "#" + k,
				Endpoint:  ne.ID,
				ParamKey:  k,
				Detail:    fmt.Sprintf("parameter %s dropped enum values %s", k, strings.Join(removed, ",")),
				Migration: "Stop sending the removed values; choose a still-valid option.",
			})
		}
	}
	for k, np := range newP {
		if _, ok := oldP[k]; ok {
			continue
		}
		cat, sev := Additive, Info
		mig := "Optional parameter added; adopt when useful."
		if np.Required {
			cat, sev = Breaking, Major
			mig = fmt.Sprintf("A required parameter %s was added; all callers must supply it.", np.Name)
		}
		d.add(Change{
			Code:      "param.added",
			Category:  cat,
			Severity:  sev.String(),
			Location:  ne.ID + "#" + k,
			Endpoint:  ne.ID,
			ParamKey:  k,
			Detail:    fmt.Sprintf("parameter %s added (required=%v)", k, np.Required),
			Migration: mig,
		})
	}
}

// diffResponses classifies response status/type differences.
func (d *Diff) diffResponses(oe, ne contract.Endpoint) {
	for code, oref := range oe.Responses {
		nref, ok := ne.Responses[code]
		if !ok {
			d.add(Change{
				Code:      "response.removed",
				Category:  Breaking,
				Severity:  Major.String(),
				Location:  ne.ID + "#" + code,
				Endpoint:  ne.ID,
				Detail:    fmt.Sprintf("response %s removed", code),
				Migration: fmt.Sprintf("Stop relying on the %s response; handle the remaining status codes.", code),
			})
			continue
		}
		if oref != nref {
			d.add(Change{
				Code:      "response.type.changed",
				Category:  Breaking,
				Severity:  Major.String(),
				Location:  ne.ID + "#" + code,
				Endpoint:  ne.ID,
				Detail:    fmt.Sprintf("response %s body type changed %q -> %q", code, oref, nref),
				Migration: "Update response parsing to the new body type.",
			})
		}
	}
	for code, nref := range ne.Responses {
		if _, ok := oe.Responses[code]; !ok {
			d.add(Change{
				Code:      "response.added",
				Category:  Additive,
				Severity:  Info.String(),
				Location:  ne.ID + "#" + code,
				Endpoint:  ne.ID,
				Detail:    fmt.Sprintf("response %s (%s) added", code, nref),
				Migration: "Optionally handle the new status code.",
			})
		}
	}
}

// diffTypes classifies type- and field-level differences.
func (d *Diff) diffTypes(oldC, newC *contract.Contract) {
	for _, name := range oldC.SortedTypeNames() {
		ot := oldC.Types[name]
		nt, ok := newC.Types[name]
		if !ok {
			d.add(Change{
				Code:      "type.removed",
				Category:  Breaking,
				Severity:  Major.String(),
				Location:  name,
				Detail:    fmt.Sprintf("type %s removed", name),
				Migration: "Remove references to this type; it no longer exists.",
			})
			continue
		}
		d.diffTypePair(ot, nt)
	}
	for _, name := range newC.SortedTypeNames() {
		if _, ok := oldC.Types[name]; !ok {
			d.add(Change{
				Code:      "type.added",
				Category:  Additive,
				Severity:  Info.String(),
				Location:  name,
				Detail:    fmt.Sprintf("type %s added", name),
				Migration: "New type available; adopt when needed.",
			})
		}
	}
}

// diffTypePair classifies differences within a correlated type.
func (d *Diff) diffTypePair(ot, nt contract.Type) {
	name := nt.Name
	if ot.Kind != nt.Kind {
		d.add(Change{
			Code:      "type.kind.changed",
			Category:  Breaking,
			Severity:  Critical.String(),
			Location:  name,
			Detail:    fmt.Sprintf("type %s kind changed %s -> %s", name, ot.Kind, nt.Kind),
			Migration: "This type's fundamental shape changed; rewrite all usages.",
		})
	}
	if removed := removedEnum(ot.Enum, nt.Enum); len(removed) > 0 {
		d.add(Change{
			Code:      "type.enum.removed",
			Category:  Breaking,
			Severity:  Major.String(),
			Location:  name,
			Detail:    fmt.Sprintf("type %s dropped enum values %s", name, strings.Join(removed, ",")),
			Migration: "Handle the removed enum values as invalid; they will no longer be produced.",
		})
	}
	if added := removedEnum(nt.Enum, ot.Enum); len(added) > 0 {
		d.add(Change{
			Code:      "type.enum.added",
			Category:  Behavioral,
			Severity:  Minor.String(),
			Location:  name,
			Detail:    fmt.Sprintf("type %s gained enum values %s", name, strings.Join(added, ",")),
			Migration: "Ensure consumers tolerate the new enum values.",
		})
	}
	for fname, of := range ot.Fields {
		path := name + "." + fname
		nf, ok := nt.Fields[fname]
		if !ok {
			d.add(Change{
				Code:      "field.removed",
				Category:  Breaking,
				Severity:  Major.String(),
				Location:  path,
				FieldPath: path,
				Detail:    fmt.Sprintf("field %s removed", path),
				Migration: fmt.Sprintf("Stop reading %s; it is no longer present.", path),
			})
			continue
		}
		d.diffFieldPair(path, of, nf)
	}
	for fname, nf := range nt.Fields {
		if _, ok := ot.Fields[fname]; ok {
			continue
		}
		path := name + "." + fname
		cat, sev := Additive, Info
		mig := "New optional field; read it when useful."
		if nf.Required {
			cat, sev = Breaking, Major
			mig = fmt.Sprintf("A required field %s was added; producers must populate it.", path)
		}
		d.add(Change{
			Code:      "field.added",
			Category:  cat,
			Severity:  sev.String(),
			Location:  path,
			FieldPath: path,
			Detail:    fmt.Sprintf("field %s added (required=%v)", path, nf.Required),
			Migration: mig,
		})
	}
}

// diffFieldPair classifies differences within a correlated field.
func (d *Diff) diffFieldPair(path string, of, nf contract.Field) {
	if of.TypeRef != nf.TypeRef {
		d.add(Change{
			Code:      "field.type.changed",
			Category:  Breaking,
