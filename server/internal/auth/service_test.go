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
	// beforeInsert, if set, runs once just before the next Insert. Tests use it to make
	// another request land between Refresh's atomic Use and its Insert.
	beforeInsert func()
}

func (m *memRefresh) Insert(_ context.Context, t RefreshToken) error {
	if h := m.beforeInsert; h != nil {
		m.beforeInsert = nil
		h()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens = append(m.tokens, t)
	return nil
}

// Use holds the lock for both the check and the write, like the atomic FindOneAndUpdate.
func (m *memRefresh) Use(_ context.Context, hash string, now time.Time, replacedBy bson.ObjectID) (RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, t := range m.tokens {
		if t.TokenHash == hash && t.UsedAt == nil && t.ExpiresAt.After(now) {
			m.tokens[i].UsedAt, m.tokens[i].ReplacedBy = &now, &replacedBy
			return t, nil
		}
	}
	return RefreshToken{}, ErrRefreshNotFound
}

func (m *memRefresh) ByHash(_ context.Context, hash string) (RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tokens {
		if t.TokenHash == hash {
			return t, nil
		}
	}
	return RefreshToken{}, ErrRefreshNotFound
}

func (m *memRefresh) DeleteUser(_ context.Context, userID bson.ObjectID) error {
	return m.deleteWhere(func(t RefreshToken) bool { return t.UserID == userID })
}

func (m *memRefresh) DeleteFamily(_ context.Context, userID, familyID bson.ObjectID) error {
	return m.deleteWhere(func(t RefreshToken) bool { return t.UserID == userID && t.FamilyID == familyID })
}

func (m *memRefresh) deleteWhere(match func(RefreshToken) bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.tokens[:0]
	for _, t := range m.tokens {
		if !match(t) {
			kept = append(kept, t)
		}
	}
	m.tokens = kept
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

func (m *memRevoked) Revoke(_ context.Context, jti string, _ bson.ObjectID, _ time.Time) error {
	m.revoke(jti)
	return nil
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

func TestRefreshRotation(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	a := env.register(t, "eugene@example.com")

	*env.clock = env.clock.Add(time.Minute) // so the new access token differs from the old one
	b, err := env.svc.Refresh(ctx, a.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if b.RefreshToken == a.RefreshToken || b.AccessToken == a.AccessToken {
		t.Error("refresh must return a new pair")
	}
	if _, err := env.svc.Authenticate(ctx, b.AccessToken); err != nil {
		t.Errorf("new access token: %v", err)
	}

	// The old token is kept, marked used, and points at its successor in the same family.
	old, _ := env.refresh.ByHash(ctx, hashToken(a.RefreshToken))
	next, _ := env.refresh.ByHash(ctx, hashToken(b.RefreshToken))
	if old.UsedAt == nil || old.ReplacedBy == nil || *old.ReplacedBy != next.ID {
		t.Errorf("old token must be marked used and point at the new one: %+v", old)
	}
	if next.FamilyID != old.FamilyID {
		t.Error("rotation must keep the familyId")
	}

	// The new refresh token works in turn.
	if _, err := env.svc.Refresh(ctx, b.RefreshToken); err != nil {
		t.Errorf("refresh with the new token: %v", err)
	}
}

func TestRefreshReuseSignsOutEverySession(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	a := env.register(t, "eugene@example.com")
	other, err := env.svc.Login(ctx, "eugene@example.com", "hunter22!") // a second device
	if err != nil {
		t.Fatal(err)
	}
	stranger := env.register(t, "stranger@example.com")

	b, err := env.svc.Refresh(ctx, a.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.svc.Refresh(ctx, a.RefreshToken); !errors.Is(err, ErrRefreshReused) {
		t.Fatalf("second use of A: got %v, want ErrRefreshReused", err)
	}
	if _, err := env.svc.Refresh(ctx, b.RefreshToken); !errors.Is(err, ErrRefreshInvalid) {
		t.Errorf("B after reuse: got %v, want ErrRefreshInvalid", err)
	}
	if _, err := env.svc.Refresh(ctx, other.RefreshToken); !errors.Is(err, ErrRefreshInvalid) {
		t.Errorf("other session after reuse: got %v, want ErrRefreshInvalid", err)
	}
	if _, err := env.svc.Refresh(ctx, stranger.RefreshToken); err != nil {
		t.Errorf("another user's session must survive: %v", err)
	}
}

// Many parallel refreshes with one token: at most one gets a new pair, and since the losers
// count as reuse, no refresh token survives. Zero winners is fine: if a loser's cleanup lands
// before the winner's final check, the winner is refused too. Which one happens is up to timing.
func TestRefreshParallelOnlyOneWins(t *testing.T) {
	env := newTestEnv(t)
	a := env.register(t, "eugene@example.com")

	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for range 10 {
		wg.Go(func() {
			if _, err := env.svc.Refresh(context.Background(), a.RefreshToken); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if wins > 1 {
		t.Errorf("%d refreshes succeeded, want at most 1", wins)
	}
	if n := len(env.refresh.tokens); n != 0 {
		t.Errorf("%d refresh tokens survived the race, want 0", n)
	}
}

func TestRefreshUnknownOrExpired(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	a := env.register(t, "eugene@example.com")

	for _, tok := range []string{"", "not-a-real-token"} {
		if _, err := env.svc.Refresh(ctx, tok); !errors.Is(err, ErrRefreshInvalid) {
			t.Errorf("%q: got %v, want ErrRefreshInvalid", tok, err)
		}
	}
	*env.clock = env.clock.Add(721 * time.Hour) // TTL is 720h
	if _, err := env.svc.Refresh(ctx, a.RefreshToken); !errors.Is(err, ErrRefreshInvalid) {
		t.Errorf("expired: got %v, want ErrRefreshInvalid", err)
	}
}

func TestLogout(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	a := env.register(t, "eugene@example.com")
	other, _ := env.svc.Login(ctx, "eugene@example.com", "hunter22!")

	// A thief rotated a stolen copy of A before the user logs out with A: the thief's B
	// is in the same family, so logout must end it too.
	b, err := env.svc.Refresh(ctx, a.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := env.svc.Authenticate(ctx, a.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.svc.Logout(ctx, claims, a.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}

	if _, err := env.svc.Authenticate(ctx, a.AccessToken); !errors.Is(err, ErrTokenRevoked) {
		t.Errorf("access token after logout: got %v, want ErrTokenRevoked", err)
	}
	if _, err := env.svc.Refresh(ctx, b.RefreshToken); !errors.Is(err, ErrRefreshInvalid) {
		t.Errorf("same-family token after logout: got %v, want ErrRefreshInvalid", err)
	}
	if _, err := env.svc.Refresh(ctx, other.RefreshToken); err != nil {
		t.Errorf("the other session must survive logout: %v", err)
	}
}

// Logout with someone else's refresh token deletes nothing of theirs.
func TestLogoutIgnoresOtherUsersToken(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	me := env.register(t, "eugene@example.com")
	victim := env.register(t, "victim@example.com")

	claims, _ := env.svc.Authenticate(ctx, me.AccessToken)
	if err := env.svc.Logout(ctx, claims, victim.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := env.svc.Refresh(ctx, victim.RefreshToken); err != nil {
		t.Errorf("victim's session must survive: %v", err)
	}
}

// The narrow race: the winner marks A used, then, before it inserts B, a replay of A runs
// the reuse cleanup. B is written after the cleanup, so without a check it would survive
// and a thief who won the race would keep a 30-day session.
func TestRefreshReuseCleanupBetweenUseAndInsert(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	a := env.register(t, "eugene@example.com")

	var replayErr error
	env.refresh.beforeInsert = func() {
		_, replayErr = env.svc.Refresh(ctx, a.RefreshToken)
	}
	won, err := env.svc.Refresh(ctx, a.RefreshToken)

	if !errors.Is(replayErr, ErrRefreshReused) {
		t.Fatalf("replay: got %v, want ErrRefreshReused", replayErr)
	}
	if err == nil {
		t.Error("winner must not get a pair once its chain was wiped")
	}
	// Counted before the next check, which would itself rotate and add a token.
	if n := len(env.refresh.tokens); n != 0 {
		t.Errorf("%d refresh tokens left for the user, want 0", n)
	}
	if won.RefreshToken != "" {
		if _, err := env.svc.Refresh(ctx, won.RefreshToken); err == nil {
			t.Error("winner's new refresh token must not work")
		}
	}
}
