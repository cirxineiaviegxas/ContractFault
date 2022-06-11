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
