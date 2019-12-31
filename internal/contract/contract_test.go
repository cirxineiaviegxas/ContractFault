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
