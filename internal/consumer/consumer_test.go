package consumer

import "testing"

func TestCriticalityWeight(t *testing.T) {
	cases := map[string]int{"high": 3, "medium": 2, "low": 1, "": 1, "bogus": 1}
