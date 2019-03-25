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
