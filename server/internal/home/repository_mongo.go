package home

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoRepo struct {
	blocks *mongo.Collection
}

func NewMongoRepo(database *mongo.Database) *MongoRepo {
	return &MongoRepo{blocks: database.Collection("homepage_blocks")}
}

func (r *MongoRepo) Active(ctx context.Context) ([]Block, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "order", Value: 1}}).
		SetProjection(bson.M{"config": 0}) // the layout never needs block content
	cur, err := r.blocks.Find(ctx, bson.M{"active": true}, opts)
	if err != nil {
		return nil, fmt.Errorf("find blocks: %w", err)
	}
	var out []Block
	if err := cur.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("decode blocks: %w", err)
	}
	return out, nil
}

func (r *MongoRepo) ByID(ctx context.Context, id bson.ObjectID) (Block, error) {
	var b Block
	err := r.blocks.FindOne(ctx, bson.M{"_id": id, "active": true}).Decode(&b)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Block{}, ErrNotFound
	}
	if err != nil {
		return Block{}, fmt.Errorf("find block %s: %w", id.Hex(), err)
	}
	return b, nil
}
