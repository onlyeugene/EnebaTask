package orders

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
	orders *mongo.Collection
}

func NewMongoRepo(database *mongo.Database) *MongoRepo {
	return &MongoRepo{client: database.Client(), orders: database.Collection("orders")}
}

// Transaction uses WithTransaction, which retries fn on a write conflict (another transaction
// changed a doc we touched) and on a commit that may not have gone through. The ctx it passes
// to fn carries the session; the other repos use the same client, so their calls join in.
func (r *MongoRepo) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	sess, err := r.client.StartSession()
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	defer sess.EndSession(ctx)
	_, err = sess.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		return nil, fn(ctx)
	})
	return err
}

func (r *MongoRepo) Insert(ctx context.Context, o Order) error {
	if _, err := r.orders.InsertOne(ctx, o); err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	return nil
}

func (r *MongoRepo) Save(ctx context.Context, o Order) error {
	if _, err := r.orders.ReplaceOne(ctx, bson.M{"_id": o.ID}, o); err != nil {
		return fmt.Errorf("save order: %w", err)
	}
	return nil
}

func (r *MongoRepo) ByID(ctx context.Context, id bson.ObjectID) (Order, error) {
	var o Order
	err := r.orders.FindOne(ctx, bson.M{"_id": id}).Decode(&o)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Order{}, ErrNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("find order: %w", err)
	}
	return o, nil
}

// List has no paging: a demo user has a handful of orders. The list view never shows codes,
// so they are left out here rather than read and dropped.
func (r *MongoRepo) List(ctx context.Context, userID bson.ObjectID) ([]Order, error) {
	cur, err := r.orders.Find(ctx, bson.M{"userId": userID}, options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetProjection(bson.M{"items.keys": 0}))
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	out := []Order{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	return out, nil
}

// Pending uses the { userId, createdAt } index; a user has at most a few pending orders.
func (r *MongoRepo) Pending(ctx context.Context, userID bson.ObjectID) ([]Order, error) {
	cur, err := r.orders.Find(ctx, bson.M{"userId": userID, "status": StatusPendingPayment})
	if err != nil {
		return nil, fmt.Errorf("find pending orders: %w", err)
	}
	var out []Order
	if err := cur.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("find pending orders: %w", err)
	}
	return out, nil
}

func (r *MongoRepo) Stale(ctx context.Context, now time.Time, limit int) ([]bson.ObjectID, error) {
	cur, err := r.orders.Find(ctx,
		bson.M{"status": StatusPendingPayment, "reservedUntil": bson.M{"$lt": now}},
		options.Find().SetProjection(bson.M{"_id": 1}).SetLimit(int64(limit)))
	if err != nil {
		return nil, fmt.Errorf("find stale orders: %w", err)
	}
	var rows []struct {
		ID bson.ObjectID `bson:"_id"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("find stale orders: %w", err)
	}
	ids := make([]bson.ObjectID, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return ids, nil
}
