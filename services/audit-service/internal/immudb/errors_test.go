package immudb

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsKeyNotFound(t *testing.T) {
	t.Parallel()
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("key not found"), true},
		{errors.New("tbtree: key not found"), true},
		{fmt.Errorf("get: %w", errors.New("KEY NOT FOUND")), true},
		{errors.New("connection refused"), false},
		{errors.New("permission denied"), false},
	}
	for _, tc := range cases {
		if got := IsKeyNotFound(tc.err); got != tc.want {
			t.Fatalf("IsKeyNotFound(%v)=%v want %v", tc.err, got, tc.want)
		}
	}
}
