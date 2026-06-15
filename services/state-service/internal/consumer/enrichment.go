package consumer

import (
	"encoding/json"
	"strconv"
)

var institutionLiquidityColumns = map[string]struct{}{
	"lcr":                   {},
	"hqla":                  {},
	"net_cash_outflows_30d": {},
	"liquidity_currency":    {},
}

func institutionLiquidityFromRow(row map[string]any) (map[string]any, bool) {
	lcr, ok := floatField(row, "lcr")
	if !ok {
		return nil, false
	}
	hqla, ok := floatField(row, "hqla")
	if !ok {
		return nil, false
	}
	netOutflows, ok := floatField(row, "net_cash_outflows_30d")
	if !ok {
		return nil, false
	}
	currency := stringField(row, "liquidity_currency")
	if currency == "" {
		currency = "EUR"
	}
	return map[string]any{
		"lcr":                lcr,
		"hqla":               hqla,
		"netCashOutflows30d": netOutflows,
		"currency":           currency,
	}, true
}

var instrumentNumericColumns = map[string]struct{}{
	"notional_amount": {},
}

func enrichInstrumentState(row map[string]any) map[string]any {
	out := make(map[string]any, len(row))
	for k, v := range row {
		if _, numeric := instrumentNumericColumns[k]; numeric {
			if f, ok := floatField(row, k); ok {
				out[k] = f
				continue
			}
		}
		out[k] = v
	}
	return out
}

func enrichInstitutionState(row map[string]any) map[string]any {
	out := make(map[string]any, len(row)+1)
	for k, v := range row {
		if _, skip := institutionLiquidityColumns[k]; skip {
			continue
		}
		out[k] = v
	}
	if liq, ok := institutionLiquidityFromRow(row); ok {
		out["liquidity"] = liq
	}
	return out
}

func enrichStateBytes(table, pk string, row map[string]any) ([]byte, error) {
	switch table {
	case "legal_entities":
		return json.Marshal(enrichInstitutionState(row))
	case "instruments":
		return json.Marshal(enrichInstrumentState(row))
	default:
		return json.Marshal(row)
	}
}

func floatField(row map[string]any, key string) (float64, bool) {
	v, ok := row[key]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case string:
		if t == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	default:
		return 0, false
	}
}
