package cart

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoRepo struct {
	client *mongo.Client
	carts  *mongo.Collection
	offers *mongo.Collection
}

func NewMongoRepo(database *mongo.Database) *MongoRepo {
	return &MongoRepo{
		client: database.Client(),
		carts:  database.Collection("carts"),
		offers: database.Collection("offers"),
	}
}

func (r *MongoRepo) Get(ctx context.Context, userID bson.ObjectID) (Cart, error) {
	var c Cart
	err := r.carts.FindOne(ctx, bson.M{"userId": userID}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Cart{UserID: userID, Items: []Item{}}, nil
	}
	if err != nil {
		return Cart{}, fmt.Errorf("find cart: %w", err)
	}
	return c, nil
}

// Update works because Mongo checks writes inside a transaction: if another transaction
// changed this cart after ours read it, our write fails with a write conflict, and
// WithTransaction runs the whole callback again on the fresh cart.
func (r *MongoRepo) Update(ctx context.Context, userID bson.ObjectID, now time.Time, change func(*Cart) error) (Cart, error) {
	sess, err := r.client.StartSession()
	if err != nil {
		return Cart{}, fmt.Errorf("start session: %w", err)
	}
	defer sess.EndSession(ctx)

	res, err := sess.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		c, err := r.Get(ctx, userID)
		if err != nil {
			return nil, err
		}
		if err := change(&c); err != nil {
			return nil, err
		}
		if c.Items == nil {
			c.Items = []Item{} // stored as [] not null
		}
		c.UpdatedAt = now
		_, err = r.carts.UpdateOne(ctx, bson.M{"userId": userID},
			bson.M{"$set": bson.M{"items": c.Items, "updatedAt": c.UpdatedAt}},
			options.UpdateOne().SetUpsert(true))
		if err != nil {
			return nil, fmt.Errorf("save cart: %w", err)
		}
		return c, nil
	})
	if err != nil {
		return Cart{}, err
	}
	return res.(Cart), nil
}

// Offers joins each offer with its product and seller. The $unwind stages drop an offer
// whose product or seller is missing, and the second $match drops inactive products,
// so a missing id in the result always means "can't be bought".
func (r *MongoRepo) Offers(ctx context.Context, offerIDs []bson.ObjectID) (map[bson.ObjectID]OfferInfo, error) {
	out := make(map[bson.ObjectID]OfferInfo, len(offerIDs))
	if len(offerIDs) == 0 {
		return out, nil
	}
	cur, err := r.offers.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": offerIDs}, "active": true}}},
		{{Key: "$lookup", Value: bson.M{"from": "products", "localField": "productId", "foreignField": "_id", "as": "product"}}},
		{{Key: "$unwind", Value: "$product"}},
		{{Key: "$match", Value: bson.M{"product.active": true}}},
		{{Key: "$lookup", Value: bson.M{"from": "sellers", "localField": "sellerId", "foreignField": "_id", "as": "seller"}}},
		{{Key: "$unwind", Value: "$seller"}},
		{{Key: "$project", Value: bson.M{
			"productId": 1, "priceCents": 1, "sellerId": 1,
			"slug": "$product.slug", "name": "$product.name", "platform": "$product.platform",
			"region": "$product.region", "cashbackPct": "$product.cashbackPct",
			"sellerName": "$seller.name",
		}}},
	})
	if err != nil {
		return nil, fmt.Errorf("find cart offers: %w", err)
	}
	var rows []OfferInfo
	if err := cur.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("decode cart offers: %w", err)
	}
	for _, row := range rows {
		out[row.OfferID] = row
	}
	return out, nil
}

// RemoveOffers drops these offers' lines from the user's cart. Orders calls it inside the pay
// transaction, so only what was just bought leaves the cart; a line added after checkout started stays.
func (r *MongoRepo) RemoveOffers(ctx context.Context, userID bson.ObjectID, offerIDs []bson.ObjectID, now time.Time) error {
	_, err := r.carts.UpdateOne(ctx, bson.M{"userId": userID}, bson.M{
		"$pull": bson.M{"items": bson.M{"offerId": bson.M{"$in": offerIDs}}},
		"$set":  bson.M{"updatedAt": now},
	})
	if err != nil {
		return fmt.Errorf("remove paid lines from cart: %w", err)
	}
	return nil
}
