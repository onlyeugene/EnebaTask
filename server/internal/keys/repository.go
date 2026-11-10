package keys

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Repository is what other features need from the keys collection.
// Phase 9 adds claim, sell and release here.
type Repository interface {
	// CountClaimable returns the live stock per offer. Offers with no claimable keys are missing from the map.
	CountClaimable(ctx context.Context, offerIDs []bson.ObjectID, now time.Time) (map[bson.ObjectID]int, error)
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
