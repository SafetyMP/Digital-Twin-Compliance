package api

import (
	"net/http"
	"os"
	"strings"
)

func withOptionalInternalToken(next http.Handler) http.Handler {
	token := strings.TrimSpace(os.Getenv("INTERNAL_API_TOKEN"))
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" {
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
