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
// UsedAt and ReplacedBy are filled by rotation in phase 6.
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

// UserView is the API.md User shape.
type UserView struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// Session is what register and login return.
type Session struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	User         UserView `json:"user"`
}

// Claims is what RequireAuth puts in the request context after checking the access token.
type Claims struct {
	UserID bson.ObjectID
	JTI    string
}

func toView(u User) UserView {
	return UserView{ID: u.ID.Hex(), Email: u.Email, Name: u.Name, CreatedAt: u.CreatedAt}
}
