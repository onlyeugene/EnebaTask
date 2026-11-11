package wishlist

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"eugene/internal/products"
)

type MongoRepo struct {
	client    *mongo.Client
	wishlists *mongo.Collection
	products  *mongo.Collection
}

func NewMongoRepo(database *mongo.Database) *MongoRepo {
	return &MongoRepo{
		client:    database.Client(),
		wishlists: database.Collection("wishlists"),
		products:  database.Collection("products"),
	}
}

func (r *MongoRepo) Get(ctx context.Context, userID bson.ObjectID) (Wishlist, error) {
	var w Wishlist
	err := r.wishlists.FindOne(ctx, bson.M{"userId": userID}).Decode(&w)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Wishlist{UserID: userID, Items: []Item{}}, nil
	}
	if err != nil {
		return Wishlist{}, fmt.Errorf("find wishlist: %w", err)
	}
	return w, nil
}

// Update is the cart's pattern (cart/repository_mongo.go): a write conflict with another
// transaction makes WithTransaction run the callback again on the fresh list.
func (r *MongoRepo) Update(ctx context.Context, userID bson.ObjectID, now time.Time, change func(*Wishlist)) (Wishlist, error) {
	sess, err := r.client.StartSession()
	if err != nil {
		return Wishlist{}, fmt.Errorf("start session: %w", err)
	}
	defer sess.EndSession(ctx)

	res, err := sess.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		w, err := r.Get(ctx, userID)
		if err != nil {
			return nil, err
		}
		change(&w)
		if w.Items == nil {
			w.Items = []Item{} // stored as [] not null
		}
		w.UpdatedAt = now
		_, err = r.wishlists.UpdateOne(ctx, bson.M{"userId": userID},
			bson.M{"$set": bson.M{"items": w.Items, "updatedAt": w.UpdatedAt}},
			options.UpdateOne().SetUpsert(true))
		if err != nil {
			return nil, fmt.Errorf("save wishlist: %w", err)
		}
		return w, nil
	})
	if err != nil {
		return Wishlist{}, err
	}
	return res.(Wishlist), nil
}

func (r *MongoRepo) Products(ctx context.Context, ids []bson.ObjectID) (map[bson.ObjectID]products.Product, error) {
	out := make(map[bson.ObjectID]products.Product, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	cur, err := r.products.Find(ctx, bson.M{"_id": bson.M{"$in": ids}, "active": true},
		options.Find().SetProjection(bson.M{"cover": 0}))
	if err != nil {
		return nil, fmt.Errorf("find wishlist products: %w", err)
	}
	var rows []products.Product
	if err := cur.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("decode wishlist products: %w", err)
	}
	for _, p := range rows {
		out[p.ID] = p
	}
	return out, nil
}
