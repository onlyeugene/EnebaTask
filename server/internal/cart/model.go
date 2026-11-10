package cart

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	// MaxQuantity is the most keys of one offer a line can hold (API.md).
	MaxQuantity = 10
	// MaxLines keeps a cart small (PLAN.md section 1). It also caps how many offers
	// one quote or merge request can make us look up.
	MaxLines = 20
)

// Problem values on a line that can't be bought as it is (API.md CartLine.problem).
const (
	ProblemOutOfStock     = "OUT_OF_STOCK"
	ProblemNotEnoughStock = "NOT_ENOUGH_STOCK"
	ProblemOfferGone      = "OFFER_GONE"
)

// Reasons a guest line is left out of a merge.
const (
	DropOfferGone  = "OFFER_GONE"
	DropOutOfStock = "OUT_OF_STOCK"
	DropCartFull   = "CART_FULL"
)

// Cart is the document in the carts collection: one per user.
// It stores only what the user chose. Prices and stock are looked up live on every read.
type Cart struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"userId"`
	Items     []Item        `bson:"items"`
	UpdatedAt time.Time     `bson:"updatedAt"`
}

// Item is one stored line. Lines are keyed by offerId: the same game from two sellers is two lines.
type Item struct {
	OfferID   bson.ObjectID `bson:"offerId"`
	ProductID bson.ObjectID `bson:"productId"`
	Quantity  int           `bson:"quantity"`
	AddedAt   time.Time     `bson:"addedAt"`
}

// OfferInfo is what a cart line shows about its offer, read from offers + products + sellers.
// Only offers that can still be bought are returned: offer active, product active, seller present.
type OfferInfo struct {
	OfferID     bson.ObjectID `bson:"_id"`
	ProductID   bson.ObjectID `bson:"productId"`
	PriceCents  int64         `bson:"priceCents"`
	Slug        string        `bson:"slug"`
	Name        string        `bson:"name"`
	Platform    string        `bson:"platform"`
	Region      string        `bson:"region"`
	CashbackPct int           `bson:"cashbackPct"`
	SellerID    bson.ObjectID `bson:"sellerId"`
	SellerName  string        `bson:"sellerName"`
}

// ItemInput is one { offerId, quantity } from a quote or merge body.
type ItemInput struct {
	OfferID  string `json:"offerId"`
	Quantity int    `json:"quantity"`
}

// ---- JSON shapes from API.md ----

// View is the API's Cart.
type View struct {
	Items         []Line `json:"items"`
	ItemCount     int    `json:"itemCount"`     // sum of quantities, for the header badge
	TotalCents    int64  `json:"totalCents"`    // sum of lineTotalCents
	CashbackCents int64  `json:"cashbackCents"` // sum of the lines' cashbackCents
}

// Line is the API's CartLine. For OFFER_GONE only offerId, quantity and problem are filled:
// the offer can't be read any more, so there is nothing true to show.
type Line struct {
	OfferID        string     `json:"offerId"`
	ProductID      string     `json:"productId"`
	Slug           string     `json:"slug"`
	Name           string     `json:"name"`
	Platform       string     `json:"platform"`
	Region         string     `json:"region"`
	CoverURL       string     `json:"coverUrl"`
	Seller         LineSeller `json:"seller"`
	UnitPriceCents int64      `json:"unitPriceCents"`
	Quantity       int        `json:"quantity"`
	LineTotalCents int64      `json:"lineTotalCents"`
	CashbackCents  int64      `json:"cashbackCents"`
	Available      int        `json:"available"` // live stock
	Problem        string     `json:"problem,omitempty"`
}

type LineSeller struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Dropped is a guest line the merge left out, so the client can show a one-time notice.
type Dropped struct {
	OfferID string `json:"offerId"`
	Reason  string `json:"reason"`
}
