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
		t.Fatalf("behavioral should exit 1, got %d", got)
	}
}

func TestExitCodeStable(t *testing.T) {
	r := &impact.Report{Summary: impact.Summary{Additive: 3}}
	if got := ExitCode(r, false); got != 0 {
		t.Fatalf("additive-only should exit 0, got %d", got)
	}
}

func TestExitCodeFailOnBehavioral(t *testing.T) {
	r := &impact.Report{Summary: impact.Summary{Behavioral: 1}}
	if got := ExitCode(r, true); got != 2 {
		t.Fatalf("fail-on-behavioral should exit 2, got %d", got)
	}
}

func TestJSONIsValidAndDeterministic(t *testing.T) {
	r := &impact.Report{Schema: "contractfault/v1", Service: "svc",
		Summary: impact.Summary{Magnitude: 5.5, Verdict: "shaken"}}
	a, err := JSON(r)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := JSON(r)
	if string(a) != string(b) {
		t.Fatal("JSON output not deterministic")
	}
	var back impact.Report
	if err := json.Unmarshal(a, &back); err != nil {
		t.Fatalf("emitted JSON does not round-trip: %v", err)
	}
	if back.Service != "svc" {
		t.Fatalf("round-trip lost service: %q", back.Service)
	}
}

func TestTextContainsSeismograph(t *testing.T) {
	r := &impact.Report{Service: "svc", FromVersion: "1", ToVersion: "2",
