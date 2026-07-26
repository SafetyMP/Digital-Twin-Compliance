package immudb

import (
	"strings"
)

// IsKeyNotFound reports whether err is an immudb missing-key failure.
// Immudb surfaces this over gRPC as a message containing "key not found"
// (see embedded/store.ErrKeyNotFound); typed errors.Is does not cross gRPC.
func IsKeyNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "key not found")
}
