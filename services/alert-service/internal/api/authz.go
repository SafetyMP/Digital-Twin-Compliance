package api

import (
	"net/http"
	"os"
	"strings"
)

// withOptionalInternalToken requires X-Internal-Token (or Bearer) when INTERNAL_API_TOKEN is set.
// Health checks remain open. Empty env keeps the historical open-dev posture.
func withOptionalInternalToken(next http.Handler) http.Handler {
	token := strings.TrimSpace(os.Getenv("INTERNAL_API_TOKEN"))
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" || strings.HasPrefix(r.URL.Path, "/ws/") {
			next.ServeHTTP(w, r)
			return
		}
		got := strings.TrimSpace(r.Header.Get("X-Internal-Token"))
		if got == "" {
			auth := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if strings.HasPrefix(auth, prefix) {
				got = strings.TrimSpace(strings.TrimPrefix(auth, prefix))
			}
		}
		if got != token {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "internal token required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
