package db

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EnsureIndexes creates every index from PLAN.md section 2.
// CreateMany is a no-op for an index that already exists with the same keys and options,
// so this is safe to run on every startup and from the seed script.
func EnsureIndexes(ctx context.Context, database *mongo.Database) error {
	const (
		plain = iota
		unique
		ttl // expireAfterSeconds 0: Mongo deletes the doc once the date in the field has passed
	)

	// idx is the only place index names are set. Each call builds its own options:
	// the driver writes the index name into the options it is given, so one builder
	// shared by several indexes would hand them all the first index's name.
	idx := func(keys bson.D, kind ...int) mongo.IndexModel {
		opts := options.Index().SetName(indexName(keys))
		if len(kind) > 0 && kind[0] == unique {
			opts.SetUnique(true)
		}
		if len(kind) > 0 && kind[0] == ttl {
			opts.SetExpireAfterSeconds(0)
		}
		return mongo.IndexModel{Keys: keys, Options: opts}
	}

	all := map[string][]mongo.IndexModel{
		"users":   {idx(bson.D{{Key: "email", Value: 1}}, unique)},
		"sellers": {idx(bson.D{{Key: "slug", Value: 1}}, unique)},
		"products": {
			idx(bson.D{{Key: "slug", Value: 1}}, unique),
			idx(bson.D{{Key: "active", Value: 1}, {Key: "inStock", Value: 1}, {Key: "soldCount", Value: -1}}),
			idx(bson.D{{Key: "active", Value: 1}, {Key: "inStock", Value: 1}, {Key: "minPriceCents", Value: 1}}),
			idx(bson.D{{Key: "active", Value: 1}, {Key: "inStock", Value: 1}, {Key: "discountPct", Value: -1}}),
			idx(bson.D{{Key: "active", Value: 1}, {Key: "createdAt", Value: -1}}),
			idx(bson.D{{Key: "type", Value: 1}, {Key: "platform", Value: 1}, {Key: "region", Value: 1}}),
		},
		"offers": {
			idx(bson.D{{Key: "productId", Value: 1}, {Key: "active", Value: 1}, {Key: "priceCents", Value: 1}}),
		},
		"keys": {
			idx(bson.D{{Key: "offerId", Value: 1}, {Key: "status", Value: 1}, {Key: "reservedUntil", Value: 1}}),
			idx(bson.D{{Key: "status", Value: 1}, {Key: "reservedUntil", Value: 1}}),
			idx(bson.D{{Key: "orderId", Value: 1}}),
		},
		"carts": {idx(bson.D{{Key: "userId", Value: 1}}, unique)},
		"orders": {
			idx(bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}),
			idx(bson.D{{Key: "status", Value: 1}, {Key: "reservedUntil", Value: 1}}),
			idx(bson.D{{Key: "number", Value: 1}}, unique),
		},
		"homepage_blocks": {
			idx(bson.D{{Key: "active", Value: 1}, {Key: "order", Value: 1}}),
		},
		"refresh_tokens": {
			idx(bson.D{{Key: "tokenHash", Value: 1}}, unique),
			idx(bson.D{{Key: "userId", Value: 1}}),
			idx(bson.D{{Key: "expiresAt", Value: 1}}, ttl),
		},
		// _id is the jti, so the lookup index comes for free.
		"revoked_tokens": {idx(bson.D{{Key: "expiresAt", Value: 1}}, ttl)},
	}

	for coll, models := range all {
		if _, err := database.Collection(coll).Indexes().CreateMany(ctx, models); err != nil {
			return fmt.Errorf("indexes on %s: %w", coll, err)
		}
	}
	return nil
}

// indexName follows Mongo's default style, field_direction joined by "_":
// {active: 1, createdAt: -1} → "active_1_createdAt_-1".
func indexName(keys bson.D) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s_%v", k.Key, k.Value))
	}
	return strings.Join(parts, "_")
}
