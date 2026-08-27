package impact

import (
	"testing"

	"github.com/seismic-labs/contractfault/internal/analyze"
	"github.com/seismic-labs/contractfault/internal/consumer"
)

func mkDiff(changes ...analyze.Change) *analyze.Diff {
	return &analyze.Diff{Service: "svc", FromVersion: "1", ToVersion: "2", Changes: changes}
}

func TestFieldChangeOnlyHitsReaders(t *testing.T) {
	d := mkDiff(analyze.Change{
		Code: "field.removed", Category: analyze.Breaking, Severity: "major",
		Location: "Order.couponCode", FieldPath: "Order.couponCode",
	})
	consumers := []consumer.Manifest{
		{Name: "reader", Criticality: "high", Uses: []consumer.Usage{
			{Endpoint: "getOrder", ReadsFields: []string{"Order.couponCode"}}}},
		{Name: "nonreader", Criticality: "high", Uses: []consumer.Usage{
			{Endpoint: "getOrder", ReadsFields: []string{"Order.id"}}}},
	}
	r := Build(d, consumers)
	if len(r.Changes[0].Consumers) != 1 || r.Changes[0].Consumers[0] != "reader" {
		t.Fatalf("expected only reader affected, got %v", r.Changes[0].Consumers)
	}
}

func TestEndpointChangeHitsAllCallers(t *testing.T) {
	d := mkDiff(analyze.Change{
		Code: "endpoint.removed", Category: analyze.Breaking, Severity: "critical",
		Location: "getReceipt", Endpoint: "getReceipt",
	})
	consumers := []consumer.Manifest{
		{Name: "a", Criticality: "low", Uses: []consumer.Usage{{Endpoint: "getReceipt"}}},
		{Name: "b", Criticality: "low", Uses: []consumer.Usage{{Endpoint: "other"}}},
	}
	r := Build(d, consumers)
	if len(r.Changes[0].Consumers) != 1 || r.Changes[0].Consumers[0] != "a" {
		t.Fatalf("expected only a, got %v", r.Changes[0].Consumers)
	}
	if r.Summary.AffectedConsumers != 1 {
		t.Fatalf("affected consumers = %d", r.Summary.AffectedConsumers)
	}
}

func TestMagnitudeScaleBounded(t *testing.T) {
	if m := magnitudeScale(0); m != 0 {
		t.Fatalf("zero raw should give 0, got %v", m)
	}
	if m := magnitudeScale(1000); m >= 10.0 {
		t.Fatalf("magnitude must stay below 10, got %v", m)
	}
	if magnitudeScale(100) <= magnitudeScale(10) {
		t.Fatal("magnitude should be monotonic increasing")
	}
}

func TestCriticalityAmplifiesScore(t *testing.T) {
	ch := analyze.Change{Code: "endpoint.removed", Category: analyze.Breaking,
		Severity: "critical", Endpoint: "e"}
	high := Build(mkDiff(ch), []consumer.Manifest{
		{Name: "h", Criticality: "high", Uses: []consumer.Usage{{Endpoint: "e"}}}})
	low := Build(mkDiff(ch), []consumer.Manifest{
		{Name: "l", Criticality: "low", Uses: []consumer.Usage{{Endpoint: "e"}}}})
	if high.Consumers[0].Score <= low.Consumers[0].Score {
		t.Fatalf("high criticality should score higher: %v vs %v",
			high.Consumers[0].Score, low.Consumers[0].Score)
