package consumer

import (
	"testing"
	"time"
)

func TestParseTimestampRFC3339(t *testing.T) {
	t.Parallel()

	got, err := parseTimestamp("2026-06-13T18:45:00Z")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 6, 13, 18, 45, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseTimestampDebeziumMicros(t *testing.T) {
	t.Parallel()

	got, err := parseTimestamp(float64(1718294700000000))
	if err != nil {
		t.Fatal(err)
	}
	if got.Unix() != 1718294700 {
		t.Fatalf("unix = %d", got.Unix())
	}
}

func TestParseTimestampMissing(t *testing.T) {
	t.Parallel()

	_, err := parseTimestamp(nil)
	if err == nil {
		t.Fatal("expected error for nil updated_at")
	}
}

func TestParseTimestampGarbage(t *testing.T) {
	t.Parallel()

	_, err := parseTimestamp("not-a-timestamp")
	if err == nil {
		t.Fatal("expected error for garbage updated_at")
	}
}
