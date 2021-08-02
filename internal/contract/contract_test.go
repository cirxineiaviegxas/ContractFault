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
