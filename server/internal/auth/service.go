package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

// Errors the handler and middleware turn into API.md codes.
var (
	ErrEmailTaken         = errors.New("email taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenExpired       = errors.New("access token expired")
	ErrTokenInvalid       = errors.New("access token invalid")
	ErrTokenRevoked       = errors.New("access token revoked")
	ErrRefreshInvalid     = errors.New("refresh token unknown or expired")
	ErrRefreshReused      = errors.New("refresh token used twice")
)

// ValidationError lists broken field rules; the handler sends it as 422 with "fields".
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return fmt.Sprintf("validation failed: %v", e.Fields) }

// Settings are the auth values from config. A separate struct keeps NewService's argument list short.
type Settings struct {
	Secret     []byte
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	BcryptCost int
}

type Service struct {
	users   UserRepo
	refresh RefreshRepo
	revoked RevokedRepo
	set     Settings
	now     func() time.Time
	// dummyHash is compared against when the email is unknown, so a login for a missing
	// account takes as long as a wrong password and the timing doesn't reveal which emails exist.
	dummyHash []byte
}

func NewService(users UserRepo, refresh RefreshRepo, revoked RevokedRepo, set Settings, now func() time.Time) (*Service, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte("not-a-real-password-1"), set.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("dummy hash: %w", err)
	}
	return &Service{users: users, refresh: refresh, revoked: revoked, set: set, now: now, dummyHash: dummy}, nil
}

const (
	minPasswordLen = 8
	maxPasswordLen = 72 // bcrypt ignores (and x/crypto refuses) bytes past 72
	maxNameLen     = 50
)

func (s *Service) Register(ctx context.Context, email, password, name string) (Session, error) {
	email = normalizeEmail(email)
	name = strings.TrimSpace(name)

	fields := map[string]string{}
	if msg := checkEmail(email); msg != "" {
		fields["email"] = msg
	}
	if msg := checkPassword(password); msg != "" {
		fields["password"] = msg
	}
	if len([]rune(name)) > maxNameLen {
		fields["name"] = fmt.Sprintf("must be at most %d characters", maxNameLen)
	}
	if len(fields) > 0 {
		return Session{}, &ValidationError{Fields: fields}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.set.BcryptCost)
	if err != nil {
		return Session{}, fmt.Errorf("hash password: %w", err)
	}
	now := s.now()
	u := User{ID: bson.NewObjectID(), Email: email, PasswordHash: string(hash), Name: name, CreatedAt: now, UpdatedAt: now}
	if err := s.users.Insert(ctx, u); err != nil {
		if errors.Is(err, ErrDuplicate) {
			return Session{}, ErrEmailTaken
		}
		return Session{}, err
	}
	return s.newSession(ctx, u)
}

// Login returns ErrInvalidCredentials for both an unknown email and a wrong password,
// so the response doesn't reveal which emails have accounts.
func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	email = normalizeEmail(email)
	fields := map[string]string{}
	if email == "" {
		fields["email"] = "is required"
	}
	if password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		return Session{}, &ValidationError{Fields: fields}
	}

	u, err := s.users.ByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password)) // same work as a real check
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return Session{}, ErrInvalidCredentials
	}
	return s.newSession(ctx, u)
}

// Authenticate checks an access token for RequireAuth: signature, alg, issuer, expiry, then the
// revoked list. The last check is a db read on every request; it is what makes logout end the
// access token at once instead of when it expires.
func (s *Service) Authenticate(ctx context.Context, token string) (Claims, error) {
	c, err := parseAccess(s.set.Secret, s.set.Issuer, token, s.now())
	if err != nil {
		return Claims{}, err
	}
	revoked, err := s.revoked.IsRevoked(ctx, c.JTI)
	if err != nil {
		return Claims{}, err
	}
	if revoked {
		return Claims{}, ErrTokenRevoked
	}
	return c, nil
}

// Me returns the signed-in user. A valid token for a deleted user counts as an invalid token.
func (s *Service) Me(ctx context.Context, userID bson.ObjectID) (UserView, error) {
	u, err := s.users.ByID(ctx, userID)
	if errors.Is(err, ErrUserNotFound) {
		return UserView{}, ErrTokenInvalid
	}
	if err != nil {
		return UserView{}, err
	}
	return toView(u), nil
}

// Refresh swaps a refresh token for a new pair (PLAN.md section 8). Each refresh token works once.
//
//  1. Mark the token used, in one atomic step that only matches an unused, unexpired token.
//  2. Matched: issue a new refresh token in the same family and a new access token.
//  3. Not matched: if the token exists and was already used, someone is replaying it. Either the
//     user or a thief holds a copy, and we can't tell which, so every session of that user ends.
func (s *Service) Refresh(ctx context.Context, raw string) (Tokens, error) {
	if raw == "" {
		return Tokens{}, ErrRefreshInvalid
	}
	now := s.now()
	hash := hashToken(raw)
	nextID := bson.NewObjectID() // known before the update, so step 1 can also set replacedBy

	old, err := s.refresh.Use(ctx, hash, now, nextID)
	if errors.Is(err, ErrRefreshNotFound) {
		return Tokens{}, s.refreshFailed(ctx, hash)
	}
	if err != nil {
		return Tokens{}, err
	}
	// If anything below fails, the old token is already used and the user has to log in again.
	// Accepted: the other order (insert first) can't work, as we only learn the userId in step 1.
	tokens, err := s.issue(ctx, nextID, old.UserID, old.FamilyID, now)
	if err != nil {
		return Tokens{}, err
	}

	// A replay of the old token may have run the reuse cleanup between our Use and our insert.
	// Then the new token was written after the cleanup and would survive it, so a thief who won
	// the race would keep a session. Used tokens are kept, never deleted, so if the old one is
	// gone, a cleanup (or a logout of this family) ran: delete the new token too and refuse.
	if _, err := s.refresh.ByHash(ctx, hash); errors.Is(err, ErrRefreshNotFound) {
		if err := s.refresh.DeleteFamily(ctx, old.UserID, old.FamilyID); err != nil {
			return Tokens{}, err
		}
		return Tokens{}, ErrRefreshInvalid
	} else if err != nil {
		return Tokens{}, err
	}
	return tokens, nil
}

// refreshFailed tells a used token (theft: revoke all) from an unknown or expired one.
func (s *Service) refreshFailed(ctx context.Context, hash string) error {
	t, err := s.refresh.ByHash(ctx, hash)
	if errors.Is(err, ErrRefreshNotFound) {
		return ErrRefreshInvalid
	}
	if err != nil {
		return err
	}
	if t.UsedAt == nil {
		return ErrRefreshInvalid // exists, unused, so it failed on expiresAt
	}
	if err := s.refresh.DeleteUser(ctx, t.UserID); err != nil {
		return err
	}
	return ErrRefreshReused
}

// Logout ends one session: the access token's jti goes on the revoked list (so it stops working
// now, not in up to 15 minutes), and the refresh token's whole family is deleted. Deleting the
// family and not just the given token matters if a thief already rotated a stolen copy: their
// newer token is in the same family and dies too. An unknown, empty or someone else's refresh
// token is skipped without an error, since the client clears its tokens whatever we answer.
func (s *Service) Logout(ctx context.Context, c Claims, rawRefresh string) error {
	if err := s.revoked.Revoke(ctx, c.JTI, c.UserID, c.ExpiresAt); err != nil {
		return err
	}
	if rawRefresh == "" {
		return nil
	}
	t, err := s.refresh.ByHash(ctx, hashToken(rawRefresh))
	if errors.Is(err, ErrRefreshNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.refresh.DeleteFamily(ctx, c.UserID, t.FamilyID) // filters on userId, so only our own
}

// newSession is register and login: a new refresh family per login.
func (s *Service) newSession(ctx context.Context, u User) (Session, error) {
	tokens, err := s.issue(ctx, bson.NewObjectID(), u.ID, bson.NewObjectID(), s.now())
	if err != nil {
		return Session{}, err
	}
	return Session{Tokens: tokens, User: toView(u)}, nil
}

// issue signs an access token and stores a new refresh token with the given id and family.
func (s *Service) issue(ctx context.Context, id, userID, familyID bson.ObjectID, now time.Time) (Tokens, error) {
	access, err := signAccess(s.set.Secret, s.set.Issuer, userID, now, s.set.AccessTTL)
	if err != nil {
		return Tokens{}, fmt.Errorf("sign access token: %w", err)
	}
	raw, hash := newRefreshToken()
	err = s.refresh.Insert(ctx, RefreshToken{
		ID: id, TokenHash: hash, UserID: userID, FamilyID: familyID,
		CreatedAt: now, ExpiresAt: now.Add(s.set.RefreshTTL),
	})
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: access, RefreshToken: raw}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// checkEmail returns "" when the email is fine. mail.ParseAddress also accepts
// `Name <a@b.c>`, so the parsed address must equal the input.
func checkEmail(email string) string {
	if email == "" {
		return "is required"
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || len(email) > 254 {
		return "is not a valid email"
	}
	return ""
}

// checkPassword counts bytes, not letters: the 72 limit is bcrypt's, and bcrypt works on bytes.
func checkPassword(p string) string {
	if len(p) < minPasswordLen || len(p) > maxPasswordLen {
		return fmt.Sprintf("must be %d to %d characters", minPasswordLen, maxPasswordLen)
	}
	for _, r := range p {
		if unicode.IsDigit(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return ""
		}
	}
	return "must contain a digit or a symbol"
}
