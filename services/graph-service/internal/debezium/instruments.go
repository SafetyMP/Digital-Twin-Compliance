package debezium

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type InstrumentPayload struct {
	After map[string]any `json:"after"`
	Op    string         `json:"op"`
}

func ParseInstrumentMessage(data []byte) (map[string]any, error) {
	var p InstrumentPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if p.Op == "d" || p.After == nil {
		return nil, nil
	}
	return normalizeInstrumentRow(p.After), nil
}

func normalizeInstrumentRow(row map[string]any) map[string]any {
	out := make(map[string]any, len(row))
	for k, v := range row {
		out[k] = v
	}
	if v, ok := out["notional_amount"]; ok {
		out["notional_amount"] = decodeDecimal(v, 2)
	}
	return out
}

func StringField(row map[string]any, key string) string {
	v, ok := row[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case map[string]any:
		if s, ok := t["string"].(string); ok {
			return s
		}
	}
	return fmt.Sprint(v)
}

func FloatField(row map[string]any, key string) float64 {
	v, ok := row[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err == nil {
			return f
		}
	}
	return 0
}

func decodeDecimal(v any, scale int) any {
	switch t := v.(type) {
	case float64:
		return t
	case string:
		if t == "" || strings.Contains(t, ".") {
			if f, err := strconv.ParseFloat(t, 64); err == nil {
				return f
			}
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
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
		rat := new(big.Rat).SetFrac(unscaled, divisor)
		f, _ := rat.Float64()
		return f
	default:
		return 0
	}
}
