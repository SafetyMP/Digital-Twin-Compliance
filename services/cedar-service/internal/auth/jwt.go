package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
)

type Claims struct {
	Sub   string   `json:"sub"`
	Roles []string `json:"roles"`
}

// PrincipalFromRequest extracts verified principal id + roles from Authorization bearer JWT.
// Body/header role claims are ignored by callers once this returns.
func PrincipalFromRequest(r *http.Request) (id string, roles []string, err error) {
	secret := strings.TrimSpace(os.Getenv("CEDAR_SERVICE_JWT_SECRET"))
	if secret == "" {
		return "", nil, errors.New("CEDAR_SERVICE_JWT_SECRET is not configured")
	}
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", nil, errors.New("missing bearer token")
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", nil, errors.New("invalid token format")
	}
	sigInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigInput))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return "", nil, errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, errors.New("invalid token payload")
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", nil, errors.New("invalid token claims")
	}
	if strings.TrimSpace(claims.Sub) == "" {
		return "", nil, errors.New("token missing sub")
	}
	return claims.Sub, claims.Roles, nil
}

// SignToken creates an HS256-style JWT for tests and internal callers.
func SignToken(secret, sub string, roles []string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body, err := json.Marshal(Claims{Sub: sub, Roles: roles})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(body)
	sigInput := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return sigInput + "." + sig, nil
}
