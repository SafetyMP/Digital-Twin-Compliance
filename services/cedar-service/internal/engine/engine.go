package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cedar "github.com/cedar-policy/cedar-go"
	"github.com/digital-twin/platform/services/cedar-service/internal/decision"
)

const policyVersion = "1.0.0"

var rulePolicyFile = map[string]string{
	"INT-R003":   "int-r003.cedar",
	"INT-R004":   "int-r004.cedar",
	"COREP-R005": "corep-r005.cedar",
	"EMIR-R004":  "emir-r004.cedar",
	"DORA-R001":  "dora-r001.cedar",
}

var ruleDefaultAction = map[string]string{
	"INT-R003":   "view",
	"INT-R004":   "approve",
	"COREP-R005": "adjust",
	"EMIR-R004":  "report",
	"DORA-R001":  "modify",
}

var ruleDefaultResource = map[string]string{
	"INT-R003":   "TwinData",
	"INT-R004":   "Payment",
	"COREP-R005": "CapitalAdjustment",
	"EMIR-R004":  "TradeReport",
	"DORA-R001":  "ICTContract",
}

type EvaluateRequest struct {
	RuleCode  string         `json:"ruleCode"`
	Principal PrincipalInput `json:"principal"`
	Action    string         `json:"action"`
	Resource  ResourceInput  `json:"resource"`
	Context   map[string]any `json:"context"`
}

type PrincipalInput struct {
	ID    string   `json:"id"`
	Roles []string `json:"roles"`
}

type ResourceInput struct {
	Type  string         `json:"type"`
	ID    string         `json:"id"`
	Attrs map[string]any `json:"attrs"`
}

func (r *ResourceInput) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type       string         `json:"type"`
		ID         string         `json:"id"`
		Attrs      map[string]any `json:"attrs"`
		Attributes map[string]any `json:"attributes"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Type = aux.Type
	r.ID = aux.ID
	r.Attrs = aux.Attrs
	if len(r.Attrs) == 0 {
		r.Attrs = aux.Attributes
	}
	return nil
}

type Status struct {
	Loaded         bool
	PolicyDir      string
	PolicyVersion  string
	SchemaLoaded   bool
	PoliciesLoaded int
	RuleCodes      []string
}

type Engine struct {
	policyDir    string
	schemaLoaded bool
	loaded       map[string]bool
}

func New(policyDir string) (*Engine, error) {
	abs, err := filepath.Abs(policyDir)
	if err != nil {
		return nil, err
	}
	schemaBytes, err := os.ReadFile(filepath.Join(abs, "schema.cedarschema"))
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}
	if len(schemaBytes) == 0 {
		return nil, fmt.Errorf("empty schema")
	}

	e := &Engine{policyDir: abs, schemaLoaded: true, loaded: map[string]bool{}}
	for code, file := range rulePolicyFile {
		data, err := os.ReadFile(filepath.Join(abs, file))
		if err != nil {
			return nil, fmt.Errorf("missing policy %s: %w", file, err)
		}
		if _, err := cedar.NewPolicySetFromBytes(file, data); err != nil {
			return nil, fmt.Errorf("parse policy %s: %w", file, err)
		}
		e.loaded[code] = true
	}
	return e, nil
}

func (e *Engine) Status() Status {
	codes := make([]string, 0, len(e.loaded))
	for code := range e.loaded {
		codes = append(codes, code)
	}
	return Status{
		Loaded:         len(e.loaded) == len(rulePolicyFile) && e.schemaLoaded,
		PolicyDir:      e.policyDir,
		PolicyVersion:  policyVersion,
		SchemaLoaded:   e.schemaLoaded,
		PoliciesLoaded: len(e.loaded),
		RuleCodes:      codes,
	}
}

func (e *Engine) Loaded() bool { return e.Status().Loaded }

func (e *Engine) Evaluate(req EvaluateRequest) (decision.RuleDecision, error) {
	code := strings.ToUpper(strings.TrimSpace(req.RuleCode))
	file, ok := rulePolicyFile[code]
	if !ok {
		return decision.RuleDecision{}, fmt.Errorf("unknown rule code %q", req.RuleCode)
	}

	policyBytes, err := os.ReadFile(filepath.Join(e.policyDir, file))
	if err != nil {
		return decision.RuleDecision{}, err
	}

	ps, err := cedar.NewPolicySetFromBytes(file, policyBytes)
	if err != nil {
		return decision.RuleDecision{}, fmt.Errorf("load policy: %w", err)
	}

	cedarReq, entities, err := e.buildRequest(code, req)
	if err != nil {
		return decision.RuleDecision{}, err
	}

	authDecision, diag := cedar.Authorize(ps, entities, cedarReq)
	outcome := "Deny"
	rationale := denyRationale(code)
	if authDecision == cedar.Allow {
		outcome = "Allow"
		rationale = fmt.Sprintf("%s evaluation permitted", code)
	} else if len(diag.Errors) > 0 {
		rationale = diag.Errors[0].Message
	}
	return decision.NewDecision(code, outcome, rationale, policyVersion, req), nil
}

func (e *Engine) buildRequest(ruleCode string, req EvaluateRequest) (cedar.Request, cedar.EntityMap, error) {
	principalID := strings.TrimSpace(req.Principal.ID)
	if principalID == "" {
		return cedar.Request{}, nil, fmt.Errorf("principal.id required")
	}
	resourceType := strings.TrimSpace(req.Resource.Type)
	if resourceType == "" {
		resourceType = ruleDefaultResource[ruleCode]
	}
	resourceID := strings.TrimSpace(req.Resource.ID)
	if resourceID == "" {
		return cedar.Request{}, nil, fmt.Errorf("resource.id required")
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		action = ruleDefaultAction[ruleCode]
	}

	principalUID := cedar.NewEntityUID(cedar.EntityType("DigitalTwin::User"), cedar.String(principalID))
	resourceUID := cedar.NewEntityUID(cedar.EntityType("DigitalTwin::"+resourceType), cedar.String(resourceID))
	actionUID := cedar.NewEntityUID(cedar.EntityType("DigitalTwin::Action"), cedar.String(action))

	entities := cedar.EntityMap{}
	roleUIDs := make([]cedar.EntityUID, 0, len(req.Principal.Roles))
	for _, role := range req.Principal.Roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		roleUID := cedar.NewEntityUID(cedar.EntityType("DigitalTwin::Role"), cedar.String(role))
		roleUIDs = append(roleUIDs, roleUID)
		entities[roleUID] = cedar.Entity{UID: roleUID, Attributes: cedar.NewRecord(cedar.RecordMap{})}
	}
	var parents cedar.EntityUIDSet
	if len(roleUIDs) > 0 {
		parents = cedar.NewEntityUIDSet(roleUIDs...)
	}
	entities[principalUID] = cedar.Entity{
		UID:        principalUID,
		Attributes: cedar.NewRecord(cedar.RecordMap{}),
		Parents:    parents,
	}
	entities[resourceUID] = cedar.Entity{
		UID:        resourceUID,
		Attributes: resourceAttrs(req.Resource),
	}

	ctxMap := cedar.RecordMap{}
	for k, v := range req.Context {
		ctxMap[cedar.String(k)] = anyToCedarValue(v)
	}

	return cedar.Request{
		Principal: principalUID,
		Action:    actionUID,
		Resource:  resourceUID,
		Context:   cedar.NewRecord(ctxMap),
	}, entities, nil
}

func resourceAttrs(res ResourceInput) cedar.Record {
	rec := cedar.RecordMap{}
	for k, v := range res.Attrs {
		rec[cedar.String(k)] = anyToCedarValue(v)
	}
	return cedar.NewRecord(rec)
}

func anyToCedarValue(v any) cedar.Value {
	switch t := v.(type) {
	case string:
		return cedar.String(t)
	case float64:
		return cedar.Long(int64(t))
	case int:
		return cedar.Long(int64(t))
	case int64:
		return cedar.Long(t)
	case bool:
		return cedar.Boolean(t)
	default:
		return cedar.String(fmt.Sprint(v))
	}
}

func denyRationale(ruleCode string) string {
	switch ruleCode {
	case "INT-R003":
		return "Sensitive twin data access denied: requires Analyst or Reporter role"
	case "INT-R004":
		return "Payment approval denied: amount exceeds €500K without Approver role"
	case "COREP-R005":
		return "Capital adjustment denied: requires CapitalManager role"
	case "EMIR-R004":
		return "Trade reporting denied: requires TradeReporter role"
	case "DORA-R001":
		return "Critical ICT contract change denied: requires ICTRiskManager role"
	default:
		return "Policy evaluation denied"
	}
}

func PolicyDirFromRepoRoot() string {
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(dir, "policies", "cedar")
		if _, err := os.Stat(filepath.Join(candidate, "schema.cedarschema")); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Join(wd, "policies", "cedar")
}
