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

// JSON renders the report as indented, deterministic JSON with a trailing
// newline. Map iteration is avoided in the pipeline so ordering is stable.
func JSON(r *impact.Report) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(r); err != nil {
		return nil, fmt.Errorf("report: encoding JSON: %w", err)
