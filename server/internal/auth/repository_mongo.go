package auth

import (
	"context"
	"errors"
	"fmt"

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
