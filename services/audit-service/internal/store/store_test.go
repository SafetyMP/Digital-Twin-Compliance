package store

import "testing"

func TestExtractRuleCode(t *testing.T) {
	t.Parallel()

	code := ExtractRuleCode([]byte(`{"alertId":"a1","ruleCode":"INT-M001"}`))
	if code != "INT-M001" {
		t.Fatalf("ruleCode = %q", code)
	}
	if ExtractRuleCode([]byte(`invalid`)) != "" {
		t.Fatal("expected empty for invalid json")
	}
}
