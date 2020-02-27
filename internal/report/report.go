// Package report renders an impact.Report into the two output formats
// contractfault supports: deterministic pretty JSON for machines and the
// TypeScript viewer, and a colorless, aligned text seismograph for humans and
// CI logs. Both renderers are pure functions of the report so output is
// reproducible byte-for-byte across runs.
package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/seismic-labs/contractfault/internal/analyze"
	"github.com/seismic-labs/contractfault/internal/impact"
)

