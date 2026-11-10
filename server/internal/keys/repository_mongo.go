package keys

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoRepo struct {
	keys *mongo.Collection
}

func NewMongoRepo(database *mongo.Database) *MongoRepo {
	return &MongoRepo{keys: database.Collection("keys")}
}

func (r *MongoRepo) CountClaimable(ctx context.Context, offerIDs []bson.ObjectID, now time.Time) (map[bson.ObjectID]int, error) {
	counts := make(map[bson.ObjectID]int, len(offerIDs))
	if len(offerIDs) == 0 {
		return counts, nil
	}

	match := claimable(now)
	match["offerId"] = bson.M{"$in": offerIDs}
	cur, err := r.keys.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.M{"_id": "$offerId", "count": bson.M{"$sum": 1}}}},
	})
	if err != nil {
		return nil, fmt.Errorf("count claimable keys: %w", err)
	}
	var rows []struct {
		OfferID bson.ObjectID `bson:"_id"`
		Count   int           `bson:"count"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("count claimable keys: %w", err)
	}
	for _, row := range rows {
		counts[row.OfferID] = row.Count
	}
	return counts, nil
}

// Claim is one atomic findOneAndUpdate: once a key is reserved with a future reservedUntil it no
// longer matches the filter, so two buyers can never get the same key. Setting orderId also takes
// an expired key away from the order that held it before: that order filters on its own orderId
// when it pays or releases, so it can't touch this key any more.
func (r *MongoRepo) Claim(ctx context.Context, offerID, orderID bson.ObjectID, now, until time.Time) (bson.ObjectID, error) {
	filter := claimable(now)
	filter["offerId"] = offerID
	var k Key
	err := r.keys.FindOneAndUpdate(ctx, filter, bson.M{"$set": bson.M{
		"status": StatusReserved, "orderId": orderID, "reservedUntil": until,
	}}).Decode(&k)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return bson.ObjectID{}, ErrNoneLeft
	}
	if err != nil {
		return bson.ObjectID{}, fmt.Errorf("claim key: %w", err)
	}
	return k.ID, nil
}

func (r *MongoRepo) Sell(ctx context.Context, orderID bson.ObjectID, now time.Time) (int, error) {
	res, err := r.keys.UpdateMany(ctx,
		bson.M{"orderId": orderID, "status": StatusReserved, "reservedUntil": bson.M{"$gt": now}},
		bson.M{"$set": bson.M{"status": StatusSold, "soldAt": now}})
	if err != nil {
		return 0, fmt.Errorf("sell keys: %w", err)
	}
	return int(res.ModifiedCount), nil
}

func (r *MongoRepo) Codes(ctx context.Context, orderID bson.ObjectID) (map[bson.ObjectID]string, error) {
	cur, err := r.keys.Find(ctx, bson.M{"orderId": orderID, "status": StatusSold})
	if err != nil {
		return nil, fmt.Errorf("find sold keys: %w", err)
	}
	var sold []Key
	if err := cur.All(ctx, &sold); err != nil {
		return nil, fmt.Errorf("find sold keys: %w", err)
	}
	codes := make(map[bson.ObjectID]string, len(sold))
	for _, k := range sold {
		codes[k.ID] = k.Code
	}
	return codes, nil
}

// Release filters on orderId, so it only frees keys this order still owns. A key another order
// has claimed since (after this one's hold ran out) carries the other orderId and is left alone.
func (r *MongoRepo) Release(ctx context.Context, orderID bson.ObjectID) error {
	_, err := r.keys.UpdateMany(ctx,
		bson.M{"orderId": orderID, "status": StatusReserved},
		bson.M{"$set": bson.M{"status": StatusAvailable, "orderId": nil, "reservedUntil": nil}})
	if err != nil {
		return fmt.Errorf("release keys: %w", err)
	}
	return nil
}
