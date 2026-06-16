package decision

import (
	"testing"
)

func TestRequiresAudit(t *testing.T) {
	cases := []struct {
		outcome string
		want    bool
	}{
		{OutcomeAllow, false},
		{OutcomeDeny, true},
		{OutcomeFlag, true},
		{OutcomeEscalate, true},
	}
	for _, tc := range cases {
		if got := RequiresAudit(tc.outcome); got != tc.want {
			t.Errorf("RequiresAudit(%q) = %v, want %v", tc.outcome, got, tc.want)
		}
	}
}

func TestRegimeForRuleCode(t *testing.T) {
	if got := RegimeForRuleCode("BASEL-R001"); got != "Basel" {
		t.Fatalf("got %q", got)
	}
	if got := RegimeForRuleCode("INT-R001"); got != "Internal" {
		t.Fatalf("got %q", got)
	}
}

func TestInputHash(t *testing.T) {
	hash, err := InputHash(map[string]any{"velocity": float64(55)})
	if err != nil {
		t.Fatal(err)
	}
	if hash[:7] != "sha256:" {
		t.Fatalf("hash = %q", hash)
	}
}
