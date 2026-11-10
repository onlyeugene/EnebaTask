package auth

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrDuplicate    = errors.New("email already in use")
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
}

type RevokedRepo interface {
	// IsRevoked reports whether an access token's jti was revoked by logout.
	IsRevoked(ctx context.Context, jti string) (bool, error)
}
