package cart

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Repository interface {
	// Get returns the user's cart, or an empty one if they have none yet.
	Get(ctx context.Context, userID bson.ObjectID) (Cart, error)
	// Update reads the cart, lets change edit it, and saves the result, all in one transaction,
	// so two tabs writing at once can't wipe each other's lines (the loser retries).
	// change may run more than once on retry, so it must only depend on the cart it is given.
	Update(ctx context.Context, userID bson.ObjectID, now time.Time, change func(*Cart) error) (Cart, error)
	// Offers returns the offers that can still be bought, keyed by offer id. Missing ids are gone.
	Offers(ctx context.Context, offerIDs []bson.ObjectID) (map[bson.ObjectID]OfferInfo, error)
}

// StockCounter gives live stock from the keys collection. keys.MongoRepo implements it.
type StockCounter interface {
	CountClaimable(ctx context.Context, offerIDs []bson.ObjectID, now time.Time) (map[bson.ObjectID]int, error)
}
