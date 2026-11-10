package products

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrNotFound = errors.New("product not found")

// Repository reads products and their offers. One interface for both collections:
// the products feature always reads them together, so splitting it would only add wiring.
type Repository interface {
	// List returns one page of products (cover SVG left out) and the total that match.
	List(ctx context.Context, q ListQuery) ([]Product, int64, error)
	Facets(ctx context.Context) (Facets, error)
	// BySlug returns an active product without its cover SVG, or ErrNotFound.
	BySlug(ctx context.Context, slug string) (Product, error)
	// CoverSVG returns only the cover markup, or ErrNotFound.
	CoverSVG(ctx context.Context, slug string) (string, error)
	// Offers returns a product's active offers with their seller, cheapest first.
	Offers(ctx context.Context, productID bson.ObjectID) ([]OfferWithSeller, error)
}

// StockCounter gives live stock from the keys collection. keys.MongoRepo implements it.
type StockCounter interface {
	CountClaimable(ctx context.Context, offerIDs []bson.ObjectID, now time.Time) (map[bson.ObjectID]int, error)
}
