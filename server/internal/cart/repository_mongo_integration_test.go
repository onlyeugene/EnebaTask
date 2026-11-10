//go:build integration

package cart

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"eugene/internal/platform/db"
)

// testDB connects to MONGODB_TEST_URI and returns a fresh eneba_test database with all indexes.
// The db name is fixed here, never read from env, so this can't touch eneba_dev or eneba_prod.
func testDB(t *testing.T) *mongo.Database {
	t.Helper()
	uri := os.Getenv("MONGODB_TEST_URI")
	if uri == "" {
		t.Fatal("MONGODB_TEST_URI is not set")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	ctx := context.Background()
	database := client.Database("eneba_test")
	if err := database.Drop(ctx); err != nil {
		t.Fatalf("drop: %v", err)
	}
	// The race below depends on the unique carts.userId index.
	if err := db.EnsureIndexes(ctx, database); err != nil {
		t.Fatalf("indexes: %v", err)
	}
	return database
}

// The first-cart race in its worst order: writer 1 has read "no cart" when writer 2 inserts
// the cart and commits, and only then does writer 1 write. writer 2 runs inside writer 1's
// change callback, which sits between the read and the write, so the order is fixed.
//
// One might expect writer 1's insert to fail on the unique userId index (a 500). It doesn't:
// inside a transaction Mongo reports it as a WriteConflict, which WithTransaction retries,
// so writer 1 runs again on the cart that now exists. This test pins that down.
func TestUpdateFirstCartWhenOtherWriterCommitsFirst(t *testing.T) {
	repo := NewMongoRepo(testDB(t))
	ctx := context.Background()
	user := bson.NewObjectID()
	first, second := bson.NewObjectID(), bson.NewObjectID()

	calls := 0
	_, err := repo.Update(ctx, user, time.Now(), func(c *Cart) error {
		calls++
		if calls == 1 {
			// writer 2, in its own goroutine and transaction, runs to commit before writer 1 writes.
			done := make(chan error)
			go func() {
				_, err := repo.Update(ctx, user, time.Now(), func(c *Cart) error {
					c.Items = append(c.Items, Item{OfferID: second, Quantity: 1})
					return nil
				})
				done <- err
			}()
			if err := <-done; err != nil {
				t.Errorf("writer 2: %v", err)
			}
		}
		c.Items = append(c.Items, Item{OfferID: first, Quantity: 1})
		return nil
	})
	if err != nil {
		t.Fatalf("writer 1: %v", err)
	}
	if calls != 2 {
		t.Errorf("writer 1's change ran %d times, want 2 (once, then the retry)", calls)
	}
	c, err := repo.Get(ctx, user)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(c.Items) != 2 || c.Items[0].OfferID != second || c.Items[1].OfferID != first {
		t.Errorf("cart = %+v, want [second, first]", c.Items)
	}
}

// Two writes for a user with no cart yet, released at the same moment. Both must succeed
// and both lines must end up in the one cart. Repeated over fresh users, because the timing
// differs from run to run.
func TestUpdateFirstCartInParallel(t *testing.T) {
	repo := NewMongoRepo(testDB(t))
	ctx := context.Background()

	const rounds = 20
	for round := range rounds {
		user := bson.NewObjectID()
		offers := []bson.ObjectID{bson.NewObjectID(), bson.NewObjectID()}

		start := make(chan struct{})
		errs := make([]error, len(offers))
		var wg sync.WaitGroup
		for i, offer := range offers {
			wg.Go(func() {
				<-start // release both writers together
				_, errs[i] = repo.Update(ctx, user, time.Now(), func(c *Cart) error {
					c.Items = append(c.Items, Item{OfferID: offer, Quantity: 1})
					return nil
				})
			})
		}
		close(start)
		wg.Wait()

		for i, err := range errs {
			if err != nil {
				t.Fatalf("round %d, writer %d: %v", round, i, err)
			}
		}
		c, err := repo.Get(ctx, user)
		if err != nil {
			t.Fatalf("round %d: get: %v", round, err)
		}
		if len(c.Items) != len(offers) {
			t.Fatalf("round %d: cart has %d lines, want %d", round, len(c.Items), len(offers))
		}
	}
}
