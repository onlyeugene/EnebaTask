package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User is the users collection document. PasswordHash never leaves the server: the API sends UserView.
type User struct {
	ID           bson.ObjectID `bson:"_id"`
	Email        string        `bson:"email"`
	PasswordHash string        `bson:"passwordHash"`
	Name         string        `bson:"name"`
	CreatedAt    time.Time     `bson:"createdAt"`
	UpdatedAt    time.Time     `bson:"updatedAt"`
}

// RefreshToken is the refresh_tokens document (PLAN.md section 1). Only the sha256 of the token is stored.
// UsedAt and ReplacedBy are set when the token is rotated. The used token is kept, not deleted,
// so a second use can be caught as theft.
type RefreshToken struct {
	ID         bson.ObjectID  `bson:"_id"`
	TokenHash  string         `bson:"tokenHash"`
	UserID     bson.ObjectID  `bson:"userId"`
	FamilyID   bson.ObjectID  `bson:"familyId"` // same for every token in one login's rotation chain
	CreatedAt  time.Time      `bson:"createdAt"`
	ExpiresAt  time.Time      `bson:"expiresAt"`
	UsedAt     *time.Time     `bson:"usedAt"`
	ReplacedBy *bson.ObjectID `bson:"replacedBy"`
}

// RevokedToken is the revoked_tokens document. The TTL index on expiresAt deletes it once the
// access token would have expired anyway.
type RevokedToken struct {
	JTI       string        `bson:"_id"`
	UserID    bson.ObjectID `bson:"userId"`
	ExpiresAt time.Time     `bson:"expiresAt"`
}

// UserView is the API.md User shape.
type UserView struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// Tokens is what /refresh returns.
type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Session is what register and login return. Tokens is embedded, so its fields sit at
// the top level of the JSON next to "user".
type Session struct {
	Tokens
	User UserView `json:"user"`
}

// Claims is what RequireAuth puts in the request context after checking the access token.
type Claims struct {
	UserID    bson.ObjectID
	JTI       string
	ExpiresAt time.Time // logout keeps the revoked jti until this time
}

func toView(u User) UserView {
	return UserView{ID: u.ID.Hex(), Email: u.Email, Name: u.Name, CreatedAt: u.CreatedAt}
}
