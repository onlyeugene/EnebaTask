package auth

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrDuplicate       = errors.New("email already in use")
	ErrRefreshNotFound = errors.New("refresh token not found")
)

type UserRepo interface {
	// Insert returns ErrDuplicate when the email is taken (unique index on email).
	Insert(ctx context.Context, u User) error
	// ByEmail and ByID return ErrUserNotFound when there is no such user.
	ByEmail(ctx context.Context, email string) (User, error)
	ByID(ctx context.Context, id bson.ObjectID) (User, error)
}

type RefreshRepo interface {
	Insert(ctx context.Context, t RefreshToken) error
	// Use marks the token used and points it at its successor, in one atomic step, but only if it
	// is unused and not expired. Two calls with the same token can't both succeed: the loser gets
	// ErrRefreshNotFound.
	Use(ctx context.Context, hash string, now time.Time, replacedBy bson.ObjectID) (RefreshToken, error)
	// ByHash finds a token whatever its state; ErrRefreshNotFound if there is none.
	ByHash(ctx context.Context, hash string) (RefreshToken, error)
	// DeleteUser deletes every refresh token of a user (reuse detected: log out all sessions).
	DeleteUser(ctx context.Context, userID bson.ObjectID) error
	// DeleteFamily deletes one login session's chain, only if it belongs to userID.
	DeleteFamily(ctx context.Context, userID, familyID bson.ObjectID) error
}

type RevokedRepo interface {
	// IsRevoked reports whether an access token's jti was revoked by logout.
	IsRevoked(ctx context.Context, jti string) (bool, error)
	// Revoke stores the jti until expiresAt (the token's own exp). Revoking twice is not an error.
	Revoke(ctx context.Context, jti string, userID bson.ObjectID, expiresAt time.Time) error
}
