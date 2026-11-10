package orders

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	StatusPendingPayment = "pending_payment"
	StatusPaid           = "paid"
	StatusFailed         = "failed"
	StatusExpired        = "expired"
)

const (
	OutcomeSuccess  = "success"
	OutcomeDeclined = "declined"
)

// Order is the document in the orders collection (PLAN.md section 1).
// Items keep a copy of name, price and seller: an order is a receipt and must show what was
// paid even if the product is renamed or repriced later.
type Order struct {
	ID            bson.ObjectID `bson:"_id"`
	Number        string        `bson:"number"` // "ORD-7K2M9Q"
	UserID        bson.ObjectID `bson:"userId"`
	Status        string        `bson:"status"`
	Items         []Item        `bson:"items"`
	TotalCents    int64         `bson:"totalCents"`
	CashbackCents int64         `bson:"cashbackCents"`
	ReservedUntil time.Time     `bson:"reservedUntil"`
	Payment       *Payment      `bson:"payment"`
	CreatedAt     time.Time     `bson:"createdAt"`
	PaidAt        *time.Time    `bson:"paidAt"`
	UpdatedAt     time.Time     `bson:"updatedAt"`
}

type Item struct {
	OfferID        bson.ObjectID   `bson:"offerId"`
	ProductID      bson.ObjectID   `bson:"productId"`
	SellerID       bson.ObjectID   `bson:"sellerId"`
	Name           string          `bson:"name"`
	Slug           string          `bson:"slug"`
	Platform       string          `bson:"platform"`
	Region         string          `bson:"region"`
	SellerName     string          `bson:"sellerName"`
	UnitPriceCents int64           `bson:"unitPriceCents"`
	Quantity       int             `bson:"quantity"`
	KeyIDs         []bson.ObjectID `bson:"keyIds"`
	Keys           []KeyCode       `bson:"keys"` // filled when paid
}

type KeyCode struct {
	Code string `bson:"code" json:"code"`
}

// Payment keeps only the last 4 digits. The rest of the card is checked and dropped.
type Payment struct {
	Last4   string `bson:"last4" json:"last4"`
	Outcome string `bson:"outcome" json:"outcome"`
}

// Card is the pay request body. Never stored.
type Card struct {
	CardNumber string `json:"cardNumber"`
	Expiry     string `json:"expiry"` // "MM/YY"
	CVC        string `json:"cvc"`
	Name       string `json:"name"`
}

// ---- JSON shapes from API.md ----

type View struct {
	ID            string     `json:"id"`
	Number        string     `json:"number"`
	Status        string     `json:"status"`
	Items         []ItemView `json:"items"`
	TotalCents    int64      `json:"totalCents"`
	CashbackCents int64      `json:"cashbackCents"`
	CreatedAt     time.Time  `json:"createdAt"`
	ReservedUntil time.Time  `json:"reservedUntil"`
	PaidAt        *time.Time `json:"paidAt,omitempty"`
	Payment       *Payment   `json:"payment,omitempty"`
}

type ItemView struct {
	OfferID        string    `json:"offerId"`
	ProductID      string    `json:"productId"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Platform       string    `json:"platform"`
	Region         string    `json:"region"`
	SellerName     string    `json:"sellerName"`
	UnitPriceCents int64     `json:"unitPriceCents"`
	Quantity       int       `json:"quantity"`
	Keys           []KeyCode `json:"keys,omitempty"` // only on a paid order's detail
}
