package analyze

import (
	"testing"

	"github.com/seismic-labs/contractfault/internal/contract"
)

// build constructs a minimal contract from endpoints and types for tests.
func build(version string, types map[string]contract.Type, eps ...contract.Endpoint) *contract.Contract {
	c := &contract.Contract{Service: "svc", Version: version, Types: types, Endpoints: eps}
	return c
}

func find(d *Diff, code string) *Change {
	for i := range d.Changes {
		if d.Changes[i].Code == code {
			return &d.Changes[i]
		}
	}
	return nil
}

func TestEndpointRemovedIsBreakingCritical(t *testing.T) {
	oldC := build("1", nil, contract.Endpoint{ID: "a", Method: "GET", Path: "/a"})
	newC := build("2", nil)
	d, err := Compare(oldC, newC)
	if err != nil {
		t.Fatal(err)
	}
	c := find(d, "endpoint.removed")
	if c == nil {
		t.Fatal("missing endpoint.removed")
	}
	if c.Category != Breaking || c.Severity != "critical" {
		t.Errorf("got %s/%s", c.Category, c.Severity)
	}
}

func TestEndpointAddedIsAdditive(t *testing.T) {
	oldC := build("1", nil)
	newC := build("2", nil, contract.Endpoint{ID: "a", Method: "GET", Path: "/a"})
	d, _ := Compare(oldC, newC)
	if c := find(d, "endpoint.added"); c == nil || c.Category != Additive {
		t.Fatalf("endpoint.added not additive: %+v", c)
	}
}

func TestRequiredFieldAddedIsBreaking(t *testing.T) {
	oldT := map[string]contract.Type{"T": {Kind: "object", Fields: map[string]contract.Field{}}}
	newT := map[string]contract.Type{"T": {Kind: "object", Fields: map[string]contract.Field{
		"x": {TypeRef: "string", Required: true},
	}}}
	d, _ := Compare(build("1", oldT), build("2", newT))
	c := find(d, "field.added")
	if c == nil || c.Category != Breaking {
		t.Fatalf("required field.added should be breaking: %+v", c)
	}
}

func TestOptionalFieldAddedIsAdditive(t *testing.T) {
	oldT := map[string]contract.Type{"T": {Kind: "object", Fields: map[string]contract.Field{}}}
	newT := map[string]contract.Type{"T": {Kind: "object", Fields: map[string]contract.Field{
		"x": {TypeRef: "string"},
	}}}
	d, _ := Compare(build("1", oldT), build("2", newT))
	if c := find(d, "field.added"); c == nil || c.Category != Additive {
		t.Fatalf("optional field.added should be additive: %+v", c)
	}
}

func TestNullableAddedIsBehavioral(t *testing.T) {
	oldT := map[string]contract.Type{"T": {Kind: "object", Fields: map[string]contract.Field{
		"x": {TypeRef: "string"},
	}}}
	newT := map[string]contract.Type{"T": {Kind: "object", Fields: map[string]contract.Field{
		"x": {TypeRef: "string", Nullable: true},
	}}}
	d, _ := Compare(build("1", oldT), build("2", newT))
	if c := find(d, "field.nullable.added"); c == nil || c.Category != Behavioral {
		t.Fatalf("nullable added should be behavioral: %+v", c)
	}
}

func TestEnumRemovedIsBreaking(t *testing.T) {
	oldT := map[string]contract.Type{"S": {Kind: "string", Enum: []string{"a", "b", "c"}}}
	newT := map[string]contract.Type{"S": {Kind: "string", Enum: []string{"a", "b"}}}
	d, _ := Compare(build("1", oldT), build("2", newT))
	c := find(d, "type.enum.removed")
	if c == nil || c.Category != Breaking {
		t.Fatalf("enum removal should be breaking: %+v", c)
	}
}

func TestMethodChangeIsBreaking(t *testing.T) {
	oldC := build("1", nil, contract.Endpoint{ID: "a", Method: "POST", Path: "/a"})
	newC := build("2", nil, contract.Endpoint{ID: "a", Method: "DELETE", Path: "/a"})
	d, _ := Compare(oldC, newC)
	if c := find(d, "endpoint.method.changed"); c == nil || c.Category != Breaking {
		t.Fatalf("method change should be breaking: %+v", c)
	}
}

func TestIdempotencyChangeIsBehavioral(t *testing.T) {
	oldC := build("1", nil, contract.Endpoint{ID: "a", Method: "POST", Path: "/a", Idempotent: true})
	newC := build("2", nil, contract.Endpoint{ID: "a", Method: "POST", Path: "/a", Idempotent: false})
	d, _ := Compare(oldC, newC)
	if c := find(d, "endpoint.idempotency.changed"); c == nil || c.Category != Behavioral {
		t.Fatalf("idempotency change should be behavioral: %+v", c)
	}
}

func TestRequiredParamAddedIsBreaking(t *testing.T) {
	oldC := build("1", nil, contract.Endpoint{ID: "a", Method: "GET", Path: "/a"})
	newC := build("2", nil, contract.Endpoint{ID: "a", Method: "GET", Path: "/a",
		Params: []contract.Param{{Name: "q", In: "query", TypeRef: "string", Required: true}}})
	d, _ := Compare(oldC, newC)
	if c := find(d, "param.added"); c == nil || c.Category != Breaking {
		t.Fatalf("required param add should be breaking: %+v", c)
	}
}

func TestServiceMismatchErrors(t *testing.T) {
	a := &contract.Contract{Service: "x", Version: "1"}
	b := &contract.Contract{Service: "y", Version: "2"}
	if _, err := Compare(a, b); err == nil {
		t.Fatal("expected service mismatch error")
	}
}

func TestSortDeterministicBySeverity(t *testing.T) {
	oldT := map[string]contract.Type{"T": {Kind: "object", Fields: map[string]contract.Field{
		"keep": {TypeRef: "string"},
	}}}
	newT := map[string]contract.Type{"T": {Kind: "object", Fields: map[string]contract.Field{
		"keep": {TypeRef: "string"}, "opt": {TypeRef: "string"},
	}}}
	oldC := build("1", oldT, contract.Endpoint{ID: "z", Method: "GET", Path: "/z"})
	newC := build("2", newT)
	d, _ := Compare(oldC, newC)
	// The critical endpoint removal must sort before the info field addition.
	if len(d.Changes) < 2 {
		t.Fatalf("expected multiple changes, got %d", len(d.Changes))
	}
	if severityRank(d.Changes[0].Severity) < severityRank(d.Changes[len(d.Changes)-1].Severity) {
		t.Fatal("changes not sorted by severity descending")
	}
}
