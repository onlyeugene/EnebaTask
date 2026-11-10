package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

// In-memory repos, so these tests need no database.

type memUsers struct {
	mu    sync.Mutex
	users map[bson.ObjectID]User
}

func (m *memUsers) Insert(_ context.Context, u User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.users {
		if existing.Email == u.Email {
			return ErrDuplicate
		}
	}
	m.users[u.ID] = u
	return nil
}

func (m *memUsers) ByEmail(_ context.Context, email string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return User{}, ErrUserNotFound
}

func (m *memUsers) ByID(_ context.Context, id bson.ObjectID) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return u, nil
}

type memRefresh struct {
	mu     sync.Mutex
	tokens []RefreshToken
}

func (m *memRefresh) Insert(_ context.Context, t RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens = append(m.tokens, t)
	return nil
}

type memRevoked struct {
	mu  sync.Mutex
	ids map[string]bool
}

func (m *memRevoked) IsRevoked(_ context.Context, jti string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ids[jti], nil
}

func (m *memRevoked) revoke(jti string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ids[jti] = true
}

type testEnv struct {
	svc     *Service
	refresh *memRefresh
	revoked *memRevoked
	clock   *time.Time // move it forward to "wait" without sleeping
}

const testSecret = "test-secret-that-is-at-least-32-bytes-long"

func newTestEnv(t *testing.T) testEnv {
	t.Helper()
	clock := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	env := testEnv{
		refresh: &memRefresh{},
		revoked: &memRevoked{ids: map[string]bool{}},
		clock:   &clock,
	}
	svc, err := NewService(&memUsers{users: map[bson.ObjectID]User{}}, env.refresh, env.revoked, Settings{
		Secret: []byte(testSecret), Issuer: "eugene-api",
		AccessTTL: 15 * time.Minute, RefreshTTL: 720 * time.Hour,
		BcryptCost: bcrypt.MinCost, // the real cost (12) would make each test take seconds
	}, func() time.Time { return *env.clock })
	if err != nil {
		t.Fatal(err)
	}
	env.svc = svc
	return env
}

func (e testEnv) register(t *testing.T, email string) Session {
	t.Helper()
	s, err := e.svc.Register(context.Background(), email, "hunter22!", "Eugene")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return s
}

func TestRegisterThenLogin(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	s := env.register(t, "  Eugene@Example.COM ")
	if s.User.Email != "eugene@example.com" {
		t.Errorf("email not normalized: %q", s.User.Email)
	}
	if s.AccessToken == "" || s.RefreshToken == "" {
		t.Fatal("register must return both tokens")
	}

	// Only the hash of the refresh token is stored, never the token itself.
	stored := env.refresh.tokens[0]
	if stored.TokenHash != hashToken(s.RefreshToken) || stored.TokenHash == s.RefreshToken {
		t.Error("refresh token must be stored as its sha256 hash")
	}

	if _, err := env.svc.Login(ctx, "EUGENE@example.com", "hunter22!"); err != nil {
		t.Errorf("login with right password: %v", err)
	}
	if _, err := env.svc.Register(ctx, "eugene@example.com", "another1!", ""); !errors.Is(err, ErrEmailTaken) {
		t.Errorf("duplicate email: got %v, want ErrEmailTaken", err)
	}
}

func TestLoginSameErrorForWrongPasswordAndUnknownEmail(t *testing.T) {
	env := newTestEnv(t)
	env.register(t, "eugene@example.com")

	_, wrongPass := env.svc.Login(context.Background(), "eugene@example.com", "wrong-pass1")
	_, noUser := env.svc.Login(context.Background(), "nobody@example.com", "hunter22!")
	if !errors.Is(wrongPass, ErrInvalidCredentials) || !errors.Is(noUser, ErrInvalidCredentials) {
		t.Errorf("both must be ErrInvalidCredentials, got %v and %v", wrongPass, noUser)
	}
}

func TestRegisterValidation(t *testing.T) {
	cases := []struct {
		name, email, password, field string
	}{
		{"bad email", "not-an-email", "hunter22!", "email"},
		{"email with display name", "Eve <eve@example.com>", "hunter22!", "email"},
		{"short password", "a@example.com", "abc1", "password"},
		{"letters only", "a@example.com", "abcdefghij", "password"},
		{"over 72 bytes", "a@example.com", string(make([]byte, 73)) + "1", "password"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			env := newTestEnv(t)
			_, err := env.svc.Register(context.Background(), c.email, c.password, "")
			var v *ValidationError
			if !errors.As(err, &v) || v.Fields[c.field] == "" {
				t.Errorf("want a validation error on %q, got %v", c.field, err)
			}
		})
	}
}

func TestExpiredAccessToken(t *testing.T) {
	env := newTestEnv(t)
	s := env.register(t, "eugene@example.com")

	*env.clock = env.clock.Add(16 * time.Minute) // TTL is 15m
	if _, err := env.svc.Authenticate(context.Background(), s.AccessToken); !errors.Is(err, ErrTokenExpired) {
		t.Errorf("got %v, want ErrTokenExpired", err)
	}
}

func TestRevokedAccessToken(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	s := env.register(t, "eugene@example.com")

	claims, err := env.svc.Authenticate(ctx, s.AccessToken)
	if err != nil {
		t.Fatalf("fresh token: %v", err)
	}
	env.revoked.revoke(claims.JTI)
	if _, err := env.svc.Authenticate(ctx, s.AccessToken); !errors.Is(err, ErrTokenRevoked) {
		t.Errorf("got %v, want ErrTokenRevoked", err)
	}

	// Revoking one token is not revoking the user: a new login still works.
	again, err := env.svc.Login(ctx, "eugene@example.com", "hunter22!")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.svc.Authenticate(ctx, again.AccessToken); err != nil {
		t.Errorf("fresh token after revoke: %v", err)
	}
}

func TestForgedAccessTokens(t *testing.T) {
	env := newTestEnv(t)
	s := env.register(t, "eugene@example.com")
	claims, _ := env.svc.Authenticate(context.Background(), s.AccessToken)

	base := jwt.RegisteredClaims{
		Subject: claims.UserID.Hex(), ID: "abc", Issuer: "eugene-api",
		IssuedAt: jwt.NewNumericDate(*env.clock), ExpiresAt: jwt.NewNumericDate(env.clock.Add(time.Minute)),
	}
	sign := func(c jwt.RegisteredClaims, method jwt.SigningMethod, key any) string {
		tok, err := jwt.NewWithClaims(method, c).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return tok
	}
	wrongIssuer := base
	wrongIssuer.Issuer = "someone-else"
	noExpiry := base
	noExpiry.ExpiresAt = nil

	cases := map[string]string{
		"wrong secret": sign(base, jwt.SigningMethodHS256, []byte("another-secret-that-is-32-bytes-long!!")),
		"alg none":     sign(base, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType),
		"HS512":        sign(base, jwt.SigningMethodHS512, []byte(testSecret)),
		"wrong issuer": sign(wrongIssuer, jwt.SigningMethodHS256, []byte(testSecret)),
		"no expiry":    sign(noExpiry, jwt.SigningMethodHS256, []byte(testSecret)),
		"garbage":      "not.a.jwt",
	}
	for name, tok := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := env.svc.Authenticate(context.Background(), tok); !errors.Is(err, ErrTokenInvalid) {
				t.Errorf("got %v, want ErrTokenInvalid", err)
			}
		})
	}
}
