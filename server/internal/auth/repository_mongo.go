package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoUsers struct{ users *mongo.Collection }

func NewMongoUsers(database *mongo.Database) *MongoUsers {
	return &MongoUsers{users: database.Collection("users")}
}

func (r *MongoUsers) Insert(ctx context.Context, u User) error {
	_, err := r.users.InsertOne(ctx, u)
	// The unique index on email decides, not a lookup first: two sign-ups with the same
	// email at the same moment would both pass a lookup.
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicate
	}
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *MongoUsers) ByEmail(ctx context.Context, email string) (User, error) {
	return r.findOne(ctx, bson.M{"email": email})
}

func (r *MongoUsers) ByID(ctx context.Context, id bson.ObjectID) (User, error) {
	return r.findOne(ctx, bson.M{"_id": id})
}

func (r *MongoUsers) findOne(ctx context.Context, filter bson.M) (User, error) {
	var u User
	err := r.users.FindOne(ctx, filter).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

type MongoRefresh struct{ tokens *mongo.Collection }

func NewMongoRefresh(database *mongo.Database) *MongoRefresh {
	return &MongoRefresh{tokens: database.Collection("refresh_tokens")}
}

func (r *MongoRefresh) Insert(ctx context.Context, t RefreshToken) error {
	if _, err := r.tokens.InsertOne(ctx, t); err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

// Use is one FindOneAndUpdate, so the check (unused, not expired) and the write (mark used) can't
// be split by another request with the same token. expiresAt is checked here as well because the
// TTL monitor only runs about once a minute.
func (r *MongoRefresh) Use(ctx context.Context, hash string, now time.Time, replacedBy bson.ObjectID) (RefreshToken, error) {
	filter := bson.M{"tokenHash": hash, "usedAt": nil, "expiresAt": bson.M{"$gt": now}}
	update := bson.M{"$set": bson.M{"usedAt": now, "replacedBy": replacedBy}}
	var t RefreshToken
	err := r.tokens.FindOneAndUpdate(ctx, filter, update).Decode(&t)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return RefreshToken{}, ErrRefreshNotFound
	}
	if err != nil {
		return RefreshToken{}, fmt.Errorf("use refresh token: %w", err)
	}
	return t, nil
}

func (r *MongoRefresh) ByHash(ctx context.Context, hash string) (RefreshToken, error) {
	var t RefreshToken
	err := r.tokens.FindOne(ctx, bson.M{"tokenHash": hash}).Decode(&t)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return RefreshToken{}, ErrRefreshNotFound
	}
	if err != nil {
		return RefreshToken{}, fmt.Errorf("find refresh token: %w", err)
	}
	return t, nil
}

func (r *MongoRefresh) DeleteUser(ctx context.Context, userID bson.ObjectID) error {
	if _, err := r.tokens.DeleteMany(ctx, bson.M{"userId": userID}); err != nil {
		return fmt.Errorf("delete user refresh tokens: %w", err)
	}
	return nil
}

func (r *MongoRefresh) DeleteFamily(ctx context.Context, userID, familyID bson.ObjectID) error {
	if _, err := r.tokens.DeleteMany(ctx, bson.M{"userId": userID, "familyId": familyID}); err != nil {
		return fmt.Errorf("delete refresh family: %w", err)
	}
	return nil
}

type MongoRevoked struct{ revoked *mongo.Collection }

func NewMongoRevoked(database *mongo.Database) *MongoRevoked {
	return &MongoRevoked{revoked: database.Collection("revoked_tokens")}
}

// IsRevoked is a lookup on _id (the jti), so it uses the built-in _id index.
func (r *MongoRevoked) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := r.revoked.CountDocuments(ctx, bson.M{"_id": jti})
	if err != nil {
		return false, fmt.Errorf("check revoked token: %w", err)
	}
	return n > 0, nil
}

// Revoke inserts the jti as _id. A duplicate means two logouts with the same token raced;
// the token is revoked either way, so that is not an error.
func (r *MongoRevoked) Revoke(ctx context.Context, jti string, userID bson.ObjectID, expiresAt time.Time) error {
	_, err := r.revoked.InsertOne(ctx, RevokedToken{JTI: jti, UserID: userID, ExpiresAt: expiresAt})
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("revoke access token: %w", err)
	}
	return nil
}
