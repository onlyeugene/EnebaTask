package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// signAccess makes an HS256 JWT with sub, jti, iat, exp, iss (PLAN.md section 8).
// The jti lets logout revoke this one token without touching the user's other sessions.
func signAccess(secret []byte, issuer string, userID bson.ObjectID, now time.Time, ttl time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.Hex(),
		ID:        randomHex(16),
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// parseAccess checks the signature, alg, issuer and expiry, and returns the user id and jti.
// WithValidMethods pins HS256, so a token with "alg: none" or another algorithm is refused
// before its signature is even looked at.
func parseAccess(secret []byte, issuer, token string, now time.Time) (Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	var claims jwt.RegisteredClaims
	_, err := parser.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) { return secret, nil })
	if errors.Is(err, jwt.ErrTokenExpired) {
		return Claims{}, ErrTokenExpired
	}
	if err != nil {
		return Claims{}, ErrTokenInvalid
	}
	userID, err := bson.ObjectIDFromHex(claims.Subject)
	if err != nil || claims.ID == "" {
		return Claims{}, ErrTokenInvalid
	}
	return Claims{UserID: userID, JTI: claims.ID}, nil
}

// newRefreshToken returns 32 random bytes as base64url (sent to the client once) and its sha256 hex
// (stored). SHA-256 and not bcrypt: the token is already 256 random bits, so there is nothing to
// brute force, and the lookup needs an exact match on an index.
func newRefreshToken() (raw, hash string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error (it panics instead)
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, hashToken(raw)
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
