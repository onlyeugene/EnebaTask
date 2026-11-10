package keys

import (
	"context"
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
