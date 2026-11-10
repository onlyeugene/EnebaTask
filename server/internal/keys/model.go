package keys

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// A key moves available → reserved → sold, or back to available (PLAN.md section 5).
const (
	StatusAvailable = "available"
	StatusReserved  = "reserved"
	StatusSold      = "sold"
)

// Key is one game code. One document per key, so a single atomic update can claim it.
type Key struct {
	ID            bson.ObjectID  `bson:"_id"`
	OfferID       bson.ObjectID  `bson:"offerId"`
	ProductID     bson.ObjectID  `bson:"productId"`
	Code          string         `bson:"code"`
	Status        string         `bson:"status"`
	OrderID       *bson.ObjectID `bson:"orderId"`       // set when reserved
	ReservedUntil *time.Time     `bson:"reservedUntil"` // once in the past, the key is claimable again
	SoldAt        *time.Time     `bson:"soldAt"`
	CreatedAt     time.Time      `bson:"createdAt"`
}
