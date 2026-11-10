package keys

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ErrNoneLeft means no key of the offer can be claimed right now.
var ErrNoneLeft = errors.New("no claimable key left")

// Repository is what other features need from the keys collection.
// Every method takes the ctx it is given, so inside a transaction it joins that transaction.
type Repository interface {
	// CountClaimable returns the live stock per offer. Offers with no claimable keys are missing from the map.
	CountClaimable(ctx context.Context, offerIDs []bson.ObjectID, now time.Time) (map[bson.ObjectID]int, error)
	// Claim reserves one claimable key of the offer for the order until `until`, or returns ErrNoneLeft.
	Claim(ctx context.Context, offerID, orderID bson.ObjectID, now, until time.Time) (bson.ObjectID, error)
	// Sell marks the order's keys sold, but only those it still holds (reserved, hold not run out).
	// It returns how many it sold, so the caller can tell if some were lost to expiry.
	Sell(ctx context.Context, orderID bson.ObjectID, now time.Time) (int, error)
	// Codes returns the codes of the order's sold keys, by key id.
	Codes(ctx context.Context, orderID bson.ObjectID) (map[bson.ObjectID]string, error)
	// Release makes the keys the order still holds available again.
	Release(ctx context.Context, orderID bson.ObjectID) error
}

// claimable matches keys that a buyer can take right now: available ones, and reserved ones whose
// hold has run out (PLAN.md section 5). Every stock count and every claim uses this one filter,
// so they can never disagree.
func claimable(now time.Time) bson.M {
	return bson.M{"$or": bson.A{
		bson.M{"status": StatusAvailable},
		bson.M{"status": StatusReserved, "reservedUntil": bson.M{"$lt": now}},
	}}
}
