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
