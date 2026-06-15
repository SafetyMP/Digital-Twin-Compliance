package store

import "testing"

func TestNumericField(t *testing.T) {
	t.Parallel()

	row := map[string]any{
		"f64": float64(123.45),
		"f32": float32(7.5),
		"str": "99",
		"nil": nil,
	}
	if got := numericField(row, "f64"); got != "123.45" {
		t.Fatalf("f64 = %q", got)
	}
	if got := numericField(row, "f32"); got != "7.5" {
		t.Fatalf("f32 = %q", got)
	}
	if got := numericField(row, "str"); got != "99" {
		t.Fatalf("str = %q", got)
	}
	if got := numericField(row, "missing"); got != "" {
		t.Fatalf("missing = %q", got)
	}
}

func TestDebeziumDaysToDate(t *testing.T) {
	t.Parallel()

	got := debeziumDaysToDate(0)
	if got != "1970-01-01" {
		t.Fatalf("epoch day = %q", got)
	}
}

func TestDateFieldNumericDays(t *testing.T) {
	t.Parallel()

	row := map[string]any{"maturity_date": float64(20000)}
	got := dateField(row, "maturity_date")
	if got == nil || *got != debeziumDaysToDate(20000) {
		t.Fatalf("date = %v", got)
	}
}
