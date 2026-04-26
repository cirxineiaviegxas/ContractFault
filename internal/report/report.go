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
	}
	return buf.Bytes(), nil
}

// Text renders a human-readable seismograph report.
func Text(r *impact.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "contractfault seismic report\n")
	fmt.Fprintf(&b, "service : %s\n", r.Service)
	fmt.Fprintf(&b, "shift   : %s -> %s\n", r.FromVersion, r.ToVersion)
	fmt.Fprintf(&b, "magnitude %.1f  verdict %s\n", r.Summary.Magnitude, strings.ToUpper(r.Summary.Verdict))
	b.WriteString(seismograph(r.Summary.Magnitude))
	b.WriteString("\n")
	fmt.Fprintf(&b, "breaking=%d  additive=%d  behavioral=%d  affected-consumers=%d\n\n",
		r.Summary.Breaking, r.Summary.Additive, r.Summary.Behavioral, r.Summary.AffectedConsumers)

	if len(r.Changes) == 0 {
		b.WriteString("no changes detected; the fault line is quiet.\n")
		return b.String()
	}

	b.WriteString("CHANGES (sorted by severity)\n")
	for _, ch := range r.Changes {
		marker := categoryMarker(ch.Category)
		fmt.Fprintf(&b, "  %s [%-8s] %-28s %s\n", marker, ch.Severity, ch.Location, ch.Detail)
		fmt.Fprintf(&b, "      code: %s\n", ch.Code)
		if len(ch.Consumers) > 0 {
			fmt.Fprintf(&b, "      hits: %s\n", strings.Join(ch.Consumers, ", "))
		}
		fmt.Fprintf(&b, "      fix : %s\n", ch.Migration)
	}

	b.WriteString("\nCONSUMER BLAST RADIUS\n")
	for _, c := range r.Consumers {
		total := c.Breaking + c.Additive + c.Behavioral
		if total == 0 {
			fmt.Fprintf(&b, "  %-18s %-6s  unaffected\n", c.Name, c.Criticality)
			continue
		}
		fmt.Fprintf(&b, "  %-18s %-6s  score=%.1f  breaking=%d additive=%d behavioral=%d\n",
			c.Name, c.Criticality, c.Score, c.Breaking, c.Additive, c.Behavioral)
	}
	return b.String()
}

// categoryMarker returns a compact ascii glyph for a change category.
func categoryMarker(cat analyze.Category) string {
	switch cat {
	case analyze.Breaking:
		return "!!"
	case analyze.Behavioral:
		return "~~"
	default:
		return "++"
	}
}

// seismograph draws a small ascii magnitude bar on a 0-10 scale.
func seismograph(mag float64) string {
	const width = 40
	filled := int((mag / 10.0) * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}

// ExitCode maps a report to a CI exit code:
//
//	0 = stable (only additive changes, no consumers broken)
//	1 = shaken (behavioral changes or additive breaks a consumer)
//	2 = rupture (breaking changes present)
//
// The failOnBehavioral flag promotes behavioral-only shifts to a failure for
// stricter pipelines.
func ExitCode(r *impact.Report, failOnBehavioral bool) int {
	if r.Summary.Breaking > 0 {
		return 2
	}
	if failOnBehavioral && r.Summary.Behavioral > 0 {
		return 2
	}
	if r.Summary.Behavioral > 0 {
		return 1
	}
	return 0
}
