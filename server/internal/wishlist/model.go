package wishlist

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/products"
)

// MaxItems caps one wishlist. It keeps the doc small and bounds the product lookup per read.
const MaxItems = 100

// Wishlist is the document in the wishlists collection: one per user.
// It stores only product ids and dates. Names and prices are read live on every request.
type Wishlist struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"userId"`
	Items     []Item        `bson:"items"`
	UpdatedAt time.Time     `bson:"updatedAt"`
}

type Item struct {
	ProductID bson.ObjectID `bson:"productId"`
	AddedAt   time.Time     `bson:"addedAt"`
}

// MergeInput is one guest entry in a merge body. AddedAt is when the guest saved it,
// so "date: newest first" keeps the guest's order after login.
type MergeInput struct {
	ProductID string     `json:"productId"`
	AddedAt   *time.Time `json:"addedAt"`
}

// ---- JSON shapes from API.md ----

// View is the API's Wishlist.
type View struct {
	Items []Entry `json:"items"`
}

type Entry struct {
	Product products.ProductCard `json:"product"`
	AddedAt time.Time            `json:"addedAt"`
}
