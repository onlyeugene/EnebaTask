package products

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Product is the document in the products collection (PLAN.md section 1).
type Product struct {
	ID            bson.ObjectID `bson:"_id"`
	Slug          string        `bson:"slug"`
	Name          string        `bson:"name"`
	Type          string        `bson:"type"`     // game | gift_card | software
	Platform      string        `bson:"platform"` // steam | psn | xbox | nintendo | windows
	Region        string        `bson:"region"`   // global | europe | united_states | united_kingdom
	Genres        []string      `bson:"genres"`
	Description   string        `bson:"description"`
	Developer     string        `bson:"developer"`
	Publisher     string        `bson:"publisher"`
	ReleaseDate   time.Time     `bson:"releaseDate"`
	Activation    string        `bson:"activation"`
	GiftCardValue *string       `bson:"giftCardValue"` // "€50" for gift cards, null otherwise
	Cover         Cover         `bson:"cover"`

	OriginalPriceCents int64 `bson:"originalPriceCents"`

	// Summary of the product's offers. Copied here so the store can sort and filter
	// without a $lookup per row; kept in sync in the same transaction that changes stock.
	MinPriceCents *int64 `bson:"minPriceCents"` // null when sold out
	DiscountPct   int    `bson:"discountPct"`
	OfferCount    int    `bson:"offerCount"`
	InStock       bool   `bson:"inStock"`

	CashbackPct int       `bson:"cashbackPct"`
	SoldCount   int       `bson:"soldCount"`
	Active      bool      `bson:"active"`
	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
}

type Cover struct {
	SVG     string    `bson:"svg"`
	Palette [2]string `bson:"palette"`
}

type Seller struct {
	ID         bson.ObjectID `bson:"_id"`
	Slug       string        `bson:"slug"`
	Name       string        `bson:"name"`
	Rating     float64       `bson:"rating"`
	SalesCount int           `bson:"salesCount"`
	Color      string        `bson:"color"`
	CreatedAt  time.Time     `bson:"createdAt"`
}

// Offer is one seller's listing of a product. Stock is a cached count of claimable keys;
// the keys collection is the source of truth.
type Offer struct {
	ID         bson.ObjectID `bson:"_id"`
	ProductID  bson.ObjectID `bson:"productId"`
	SellerID   bson.ObjectID `bson:"sellerId"`
	PriceCents int64         `bson:"priceCents"`
	Stock      int           `bson:"stock"`
	Active     bool          `bson:"active"`
	CreatedAt  time.Time     `bson:"createdAt"`
	UpdatedAt  time.Time     `bson:"updatedAt"`
}
