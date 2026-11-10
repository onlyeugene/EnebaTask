package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// These go through the real routes, to check the status and error code the client switches on.

func serve(t *testing.T, router http.Handler, method, path, token, body string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var out struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&out)
	return rec.Code, out.Error.Code
}

func TestMeErrorCodes(t *testing.T) {
	env := newTestEnv(t)
	router := NewHandler(env.svc).Routes()
	s := env.register(t, "eugene@example.com")

	if status, _ := serve(t, router, "GET", "/me", s.AccessToken, ""); status != 200 {
		t.Errorf("valid token: status %d, want 200", status)
	}
	if status, code := serve(t, router, "GET", "/me", "", ""); status != 401 || code != "UNAUTHORIZED" {
		t.Errorf("no token: %d %s, want 401 UNAUTHORIZED", status, code)
	}

	claims, _ := env.svc.Authenticate(t.Context(), s.AccessToken)
	env.revoked.revoke(claims.JTI)
	if status, code := serve(t, router, "GET", "/me", s.AccessToken, ""); status != 401 || code != "TOKEN_REVOKED" {
		t.Errorf("revoked: %d %s, want 401 TOKEN_REVOKED", status, code)
	}

	fresh := env.register(t, "other@example.com")
	*env.clock = env.clock.Add(16 * time.Minute)
	if status, code := serve(t, router, "GET", "/me", fresh.AccessToken, ""); status != 401 || code != "TOKEN_EXPIRED" {
		t.Errorf("expired: %d %s, want 401 TOKEN_EXPIRED", status, code)
	}
}

func TestLoginRateLimit(t *testing.T) {
	env := newTestEnv(t)
	router := NewHandler(env.svc).Routes()
	body := `{"email":"nobody@example.com","password":"wrong-pass1"}`

	for i := 1; i <= 10; i++ {
		if status, _ := serve(t, router, "POST", "/login", "", body); status != 401 {
			t.Fatalf("attempt %d: status %d, want 401", i, status)
		}
	}
	if status, code := serve(t, router, "POST", "/login", "", body); status != 429 || code != "RATE_LIMITED" {
		t.Errorf("11th attempt: %d %s, want 429 RATE_LIMITED", status, code)
	}
}

// /me is called on every protected page and in the header, so it must never hit the login limit,
// not even from an IP that has already used up its login attempts.
func TestMeNotRateLimited(t *testing.T) {
	env := newTestEnv(t)
	router := NewHandler(env.svc).Routes()
	s := env.register(t, "eugene@example.com") // counts against the limit, like a real sign-up

	for i := 0; i < 10; i++ {
		serve(t, router, "POST", "/login", "", `{"email":"nobody@example.com","password":"wrong-pass1"}`)
	}
	for i := 1; i <= 30; i++ {
		if status, _ := serve(t, router, "GET", "/me", s.AccessToken, ""); status != 200 {
			t.Fatalf("/me call %d: status %d, want 200", i, status)
		}
	}
}
