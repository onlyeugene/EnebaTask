package wishlist

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/products"
)

type Repository interface {
	// Get returns the user's wishlist, or an empty one if they have none yet.
	Get(ctx context.Context, userID bson.ObjectID) (Wishlist, error)
	// Update reads the wishlist, lets change edit it, and saves it in one transaction,
	// like the cart, so two tabs writing at once can't lose each other's items.
	// change may run more than once on retry, so it must only depend on the list it is given.
	Update(ctx context.Context, userID bson.ObjectID, now time.Time, change func(*Wishlist)) (Wishlist, error)
	// Products returns the active products among ids (cover left out), keyed by id.
	// A missing id means the product is gone or taken down.
	Products(ctx context.Context, ids []bson.ObjectID) (map[bson.ObjectID]products.Product, error)
}
