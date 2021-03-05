package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/seismic-labs/contractfault/internal/impact"
)

func TestExitCodeRupture(t *testing.T) {
	r := &impact.Report{Summary: impact.Summary{Breaking: 1}}
	if got := ExitCode(r, false); got != 2 {
		t.Fatalf("breaking should exit 2, got %d", got)
	}
}

func TestExitCodeShaken(t *testing.T) {
	r := &impact.Report{Summary: impact.Summary{Behavioral: 1}}
	if got := ExitCode(r, false); got != 1 {
