package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPrincipalFromRequestRejectsExpired(t *testing.T) {
	secret := "test-secret-min-32-chars-long!!!!"
	t.Setenv("CEDAR_SERVICE_JWT_SECRET", secret)
	token, err := SignTokenTTL(secret, "user-1", []string{"Approver"}, -time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if _, _, err := PrincipalFromRequest(req); err == nil {
		t.Fatal("expected expired token to fail")
	}
}

func TestPrincipalFromRequestAcceptsFresh(t *testing.T) {
	secret := "test-secret-min-32-chars-long!!!!"
	t.Setenv("CEDAR_SERVICE_JWT_SECRET", secret)
	token, err := SignToken(secret, "user-1", []string{"Approver"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	sub, roles, err := PrincipalFromRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if sub != "user-1" || len(roles) != 1 || roles[0] != "Approver" {
		t.Fatalf("got sub=%q roles=%v", sub, roles)
	}
}

func TestPrincipalFromRequestRejectsMissingExp(t *testing.T) {
	secret := "test-secret-min-32-chars-long!!!!"
	t.Setenv("CEDAR_SERVICE_JWT_SECRET", secret)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body, _ := json.Marshal(map[string]any{"sub": "u1", "roles": []string{"Approver"}})
	payload := base64.RawURLEncoding.EncodeToString(body)
	sigInput := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sigInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	token := sigInput + "." + sig

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if _, _, err := PrincipalFromRequest(req); err == nil {
		t.Fatal("expected missing exp to fail")
	}
}
