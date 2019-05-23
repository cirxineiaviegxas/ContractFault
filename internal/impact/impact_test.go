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
