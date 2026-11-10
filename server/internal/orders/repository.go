package orders

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/cart"
	"eugene/internal/products"
)

var ErrNotFound = errors.New("order not found")

// Repository is the orders collection, plus the transaction that the service runs its steps in.
type Repository interface {
	// Transaction runs fn in one Mongo transaction. Every repo call made with the ctx that fn gets
	// (keys, offers, products, carts, orders) joins it. On a write conflict the whole of fn runs
	// again, so fn must start from fresh reads each time.
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	Insert(ctx context.Context, o Order) error
	// Save writes the whole order back.
	Save(ctx context.Context, o Order) error
	// ByID returns any user's order, or ErrNotFound. The service checks the owner.
	ByID(ctx context.Context, id bson.ObjectID) (Order, error)
	// List returns the user's orders, newest first.
	List(ctx context.Context, userID bson.ObjectID) ([]Order, error)
	// Pending returns the user's pending_payment orders, whether or not their hold has run out.
	Pending(ctx context.Context, userID bson.ObjectID) ([]Order, error)
	// Stale returns ids of pending_payment orders whose hold ran out before now, for the cleanup job.
	Stale(ctx context.Context, now time.Time, limit int) ([]bson.ObjectID, error)
}

// Keys is what orders needs from the keys collection. keys.MongoRepo implements it.
type Keys interface {
	CountClaimable(ctx context.Context, offerIDs []bson.ObjectID, now time.Time) (map[bson.ObjectID]int, error)
	Claim(ctx context.Context, offerID, orderID bson.ObjectID, now, until time.Time) (bson.ObjectID, error)
	Sell(ctx context.Context, orderID bson.ObjectID, now time.Time) (int, error)
	Codes(ctx context.Context, orderID bson.ObjectID) (map[bson.ObjectID]string, error)
	Release(ctx context.Context, orderID bson.ObjectID) error
}

// Catalog writes the cached stock and price summary. products.MongoRepo implements it.
type Catalog interface {
	StockInputs(ctx context.Context, productID bson.ObjectID) (products.Product, []products.Offer, error)
	SaveStock(ctx context.Context, p products.Product, offers []products.Offer, now time.Time) error
	AddSold(ctx context.Context, productID bson.ObjectID, qty int) error
}

// Carts reads the cart an order is made from, and removes paid lines. cart.MongoRepo implements it.
type Carts interface {
	Get(ctx context.Context, userID bson.ObjectID) (cart.Cart, error)
	Offers(ctx context.Context, offerIDs []bson.ObjectID) (map[bson.ObjectID]cart.OfferInfo, error)
	RemoveOffers(ctx context.Context, userID bson.ObjectID, offerIDs []bson.ObjectID, now time.Time) error
}
