package contract

import (
	"strings"
	"testing"
)

const minimal = `{
  "service": "svc",
  "version": "1.0.0",
  "types": {
    "Widget": { "kind": "object", "fields": { "id": { "type": "string", "required": true } } }
  },
  "endpoints": [
    { "id": "getWidget", "method": "get", "path": "/w/{id}", "responses": { "200": "Widget" } }
  ]
}`

func TestParseNormalizesMethodAndTypeName(t *testing.T) {
	c, err := Parse([]byte(minimal), "test")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got := c.Endpoints[0].Method; got != "GET" {
		t.Errorf("method not upper-cased: %q", got)
	}
	if got := c.Types["Widget"].Name; got != "Widget" {
		t.Errorf("type name not mirrored: %q", got)
	}
}

func TestValidateRejectsUnknownRequestType(t *testing.T) {
	bad := `{
      "service": "svc", "version": "1.0.0", "types": {},
      "endpoints": [ { "id": "x", "method": "POST", "path": "/x", "requestType": "Nope" } ]
    }`
	_, err := Parse([]byte(bad), "test")
	if err == nil || !strings.Contains(err.Error(), "unknown request type") {
		t.Fatalf("expected unknown request type error, got %v", err)
	}
}

func TestValidateRejectsDuplicateEndpointID(t *testing.T) {
	dup := `{
      "service": "svc", "version": "1.0.0", "types": {},
      "endpoints": [
        { "id": "x", "method": "GET", "path": "/a" },
        { "id": "x", "method": "GET", "path": "/b" }
      ]
    }`
	_, err := Parse([]byte(dup), "test")
	if err == nil || !strings.Contains(err.Error(), "duplicate endpoint id") {
		t.Fatalf("expected duplicate id error, got %v", err)
	}
}

func TestParseRejectsUnknownFields(t *testing.T) {
	extra := `{ "service": "svc", "version": "1.0.0", "types": {}, "endpoints": [], "bogus": 1 }`
	if _, err := Parse([]byte(extra), "test"); err == nil {
		t.Fatal("expected error on unknown field")
	}
}

func TestSortedTypeNamesDeterministic(t *testing.T) {
	c := &Contract{Types: map[string]Type{"b": {}, "a": {}, "c": {}}}
	got := c.SortedTypeNames()
	want := []string{"a", "b", "c"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sorted names = %v, want %v", got, want)
		}
	}
}

func TestParamKey(t *testing.T) {
	p := Param{Name: "status", In: "query"}
	if p.Key() != "query:status" {
		t.Fatalf("key = %q", p.Key())
	}
}
