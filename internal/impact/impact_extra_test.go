package impact

import (
	"testing"

	"github.com/seismic-labs/contractfault/internal/analyze"
)

func TestEmptyDiffProducesNoChanges(t *testing.T) {
	r := Build(mkDiff(), nil)
	if len(r.Changes) != 0 {
		t.Fatalf("expected no changes for an empty diff, got %d", len(r.Changes))
	}
}

func TestEndpointRemovalWithNoConsumers(t *testing.T) {
	d := mkDiff(analyze.Change{
		Code: "endpoint.removed", Category: analyze.Breaking, Severity: "critical",
		Location: "getInvoice", Endpoint: "getInvoice",
	})
	r := Build(d, nil)
	if len(r.Changes) != 1 || len(r.Changes[0].Consumers) != 0 {
		t.Fatalf("expected one change with no consumers, got %+v", r.Changes)
	}
}

func TestUnknownChangeCodeDoesNotPanic(t *testing.T) {
	d := mkDiff(analyze.Change{
		Code: "surface.deprecated", Category: analyze.Behavioral, Severity: "info",
		Location: "Order.legacyId",
	})
	r := Build(d, nil)
	if len(r.Changes) != 1 {
		t.Fatalf("unknown-but-typed change must still be reported, got %d", len(r.Changes))
	}
}
