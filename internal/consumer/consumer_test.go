package consumer

import "testing"

func TestCriticalityWeight(t *testing.T) {
	cases := map[string]int{"high": 3, "medium": 2, "low": 1, "": 1, "bogus": 1}
	for label, want := range cases {
		m := Manifest{Criticality: label}
		if got := m.CriticalityWeight(); got != want {
			t.Errorf("weight(%q) = %d, want %d", label, got, want)
		}
	}
}

func TestValidateRejectsBadCriticality(t *testing.T) {
	m := Manifest{Name: "c", Criticality: "extreme"}
	if err := m.Validate(); err == nil {
		t.Fatal("expected invalid criticality error")
	}
}

func TestValidateDefaultsEmptyCriticality(t *testing.T) {
	m := Manifest{Name: "c"}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
