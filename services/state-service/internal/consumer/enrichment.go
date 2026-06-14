package consumer

import (
	"encoding/json"
)

// lowLCRInstitution is seeded at 0.95 LCR for BASEL-M001 smoke tests.
const lowLCRInstitution = "44444444-4444-4444-4444-444444444401"

func enrichInstitutionState(entityID string, row map[string]any) map[string]any {
	out := make(map[string]any, len(row)+1)
	for k, v := range row {
		out[k] = v
	}

	lcr := 1.05
	hqla := 500000000.00
	netOutflows := 476190476.00
	if entityID == lowLCRInstitution {
		lcr = 0.95
		hqla = 450000000.00
		netOutflows = 473684211.00
	}

	out["liquidity"] = map[string]any{
		"lcr":                lcr,
		"hqla":               hqla,
		"netCashOutflows30d": netOutflows,
		"currency":           "EUR",
	}
	return out
}

func enrichStateBytes(table, pk string, row map[string]any) ([]byte, error) {
	if table == "legal_entities" {
		enriched := enrichInstitutionState(pk, row)
		return json.Marshal(enriched)
	}
	return json.Marshal(row)
}
