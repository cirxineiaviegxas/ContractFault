package analyze

import (
	"testing"

	"github.com/seismic-labs/contractfault/internal/contract"
)

// build constructs a minimal contract from endpoints and types for tests.
func build(version string, types map[string]contract.Type, eps ...contract.Endpoint) *contract.Contract {
	c := &contract.Contract{Service: "svc", Version: version, Types: types, Endpoints: eps}
	return c
}
