package consumer

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var tableNumericFields = map[string]map[string]int{
	"legal_entities": {
		"lcr":                   4,
		"hqla":                  2,
		"net_cash_outflows_30d": 2,
	},
	"instruments": {"notional_amount": 2},
}

var tableDateFields = map[string][]string{
	"instruments": {"maturity_date"},
}

func normalizeDebeziumRow(table string, row map[string]any) map[string]any {
	fields, ok := tableNumericFields[table]
	if !ok || row == nil {
		return row
	}
	out := make(map[string]any, len(row))
	for k, v := range row {
		out[k] = v
	}
	for field, scale := range fields {
		if v, exists := out[field]; exists {
			out[field] = decodeDebeziumDecimal(v, scale)
		}
	}
	for _, field := range tableDateFields[table] {
		if v, exists := out[field]; exists {
			out[field] = decodeDebeziumDate(v)
		}
	}
	return out
}

func decodeDebeziumDate(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case string:
		if t == "" {
			return t
		}
		return t
	case float64:
		return debeziumDaysToDate(int(t))
	case int:
		return debeziumDaysToDate(t)
	case int64:
		return debeziumDaysToDate(int(t))
	default:
		return fmt.Sprint(v)
	}
}

func debeziumDaysToDate(days int) string {
	return time.Unix(int64(days)*86400, 0).UTC().Format("2006-01-02")
}

func decodeDebeziumDecimal(v any, scale int) any {
	switch t := v.(type) {
	case nil:
		return nil
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		if t == "" {
			return t
		}
		if strings.Contains(t, ".") || isPlainDecimalString(t) {
			return t
		}
		decoded, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			return t
		}
		unscaled := new(big.Int).SetBytes(decoded)
		if len(decoded) > 0 && decoded[0]&0x80 != 0 {
			bitLen := uint(len(decoded) * 8)
			max := new(big.Int).Lsh(big.NewInt(1), bitLen)
			unscaled.Sub(unscaled, max)
		}
		if scale == 0 {
			return unscaled.String()
		}
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
		rat := new(big.Rat).SetFrac(unscaled, divisor)
		return rat.FloatString(scale)
	default:
		return fmt.Sprint(v)
	}
}

func isPlainDecimalString(s string) bool {
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r == '.' || r == '-':
			if r == '-' && i != 0 {
				return false
			}
		default:
			return false
		}
	}
	return s != "" && s != "-"
}
