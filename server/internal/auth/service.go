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

// newSession issues an access token and a refresh token. Each login starts a new refresh
// family; phase 6 rotation keeps the familyId along the chain.
func (s *Service) newSession(ctx context.Context, u User) (Session, error) {
	now := s.now()
	access, err := signAccess(s.set.Secret, s.set.Issuer, u.ID, now, s.set.AccessTTL)
	if err != nil {
		return Session{}, fmt.Errorf("sign access token: %w", err)
	}
	raw, hash := newRefreshToken()
	err = s.refresh.Insert(ctx, RefreshToken{
		ID: bson.NewObjectID(), TokenHash: hash, UserID: u.ID, FamilyID: bson.NewObjectID(),
		CreatedAt: now, ExpiresAt: now.Add(s.set.RefreshTTL),
	})
	if err != nil {
		return Session{}, err
	}
	return Session{AccessToken: access, RefreshToken: raw, User: toView(u)}, nil
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
