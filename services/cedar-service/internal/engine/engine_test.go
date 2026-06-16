package engine_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/digital-twin/platform/services/cedar-service/internal/engine"
)

func policyDir(t *testing.T) string {
	t.Helper()
	dir := engine.PolicyDirFromRepoRoot()
	if _, err := os.Stat(filepath.Join(dir, "schema.cedarschema")); err != nil {
		t.Fatalf("policies missing at %s: %v", dir, err)
	}
	return dir
}

func newEngine(t *testing.T) *engine.Engine {
	t.Helper()
	eng, err := engine.New(policyDir(t))
	if err != nil {
		t.Fatal(err)
	}
	return eng
}

func TestLoaderStatus(t *testing.T) {
	t.Parallel()
	eng := newEngine(t)
	st := eng.Status()
	if !st.Loaded || !st.SchemaLoaded || st.PoliciesLoaded != 5 {
		t.Fatalf("status = %#v", st)
	}
}

func TestINT_R003_DenyWithoutRole(t *testing.T) {
	t.Parallel()
	eng := newEngine(t)
	got, err := eng.Evaluate(engine.EvaluateRequest{
		RuleCode:  "INT-R003",
		Principal: engine.PrincipalInput{ID: "u1"},
		Resource: engine.ResourceInput{
			Type: "TwinData",
			ID:   "t1",
			Attrs: map[string]any{
				"sensitivity": "high",
				"tenantId":    "00000000-0000-0000-0000-000000000001",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Outcome != "Deny" {
		t.Fatalf("outcome = %q want Deny", got.Outcome)
	}
	if got.InputHash == "" {
		t.Fatal("missing inputHash")
	}
}

func TestINT_R003_AllowWithReporter(t *testing.T) {
	t.Parallel()
	eng := newEngine(t)
	got, err := eng.Evaluate(engine.EvaluateRequest{
		RuleCode:  "INT-R003",
		Principal: engine.PrincipalInput{ID: "u1", Roles: []string{"Reporter"}},
		Resource: engine.ResourceInput{
			Type: "TwinData",
			ID:   "t1",
			Attrs: map[string]any{
				"sensitivity": "high",
				"tenantId":    "00000000-0000-0000-0000-000000000001",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Outcome != "Allow" {
		t.Fatalf("outcome = %q want Allow", got.Outcome)
	}
}

func TestINT_R004_DenyLargePayment(t *testing.T) {
	t.Parallel()
	eng := newEngine(t)
	got, err := eng.Evaluate(engine.EvaluateRequest{
		RuleCode:  "INT-R004",
		Principal: engine.PrincipalInput{ID: "u1", Roles: []string{"Analyst"}},
		Resource: engine.ResourceInput{
			Type: "Payment",
			ID:   "pay-1",
			Attrs: map[string]any{
				"amountEur": int64(600000),
				"tenantId":  "00000000-0000-0000-0000-000000000001",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Outcome != "Deny" {
		t.Fatalf("outcome = %q want Deny", got.Outcome)
	}
}

func TestCOREP_R005_AllowCapitalManager(t *testing.T) {
	t.Parallel()
	eng := newEngine(t)
	got, err := eng.Evaluate(engine.EvaluateRequest{
		RuleCode:  "COREP-R005",
		Principal: engine.PrincipalInput{ID: "u1", Roles: []string{"CapitalManager"}},
		Resource: engine.ResourceInput{
			Type:  "CapitalAdjustment",
			ID:    "adj-1",
			Attrs: map[string]any{"tenantId": "00000000-0000-0000-0000-000000000001"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Outcome != "Allow" {
		t.Fatalf("outcome = %q want Allow", got.Outcome)
	}
}

func TestDORA_R001_AllowNonCritical(t *testing.T) {
	t.Parallel()
	eng := newEngine(t)
	got, err := eng.Evaluate(engine.EvaluateRequest{
		RuleCode:  "DORA-R001",
		Principal: engine.PrincipalInput{ID: "u1"},
		Resource: engine.ResourceInput{
			Type: "ICTContract",
			ID:   "ict-1",
			Attrs: map[string]any{
				"criticality": "standard",
				"tenantId":    "00000000-0000-0000-0000-000000000001",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Outcome != "Allow" {
		t.Fatalf("outcome = %q want Allow", got.Outcome)
	}
}

func TestUnknownRuleCode(t *testing.T) {
	t.Parallel()
	eng := newEngine(t)
	_, err := eng.Evaluate(engine.EvaluateRequest{
		RuleCode:  "UNKNOWN",
		Principal: engine.PrincipalInput{ID: "u1"},
		Resource:  engine.ResourceInput{Type: "TwinData", ID: "t1"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
