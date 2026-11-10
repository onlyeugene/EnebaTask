//go:build integration

package orders

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"eugene/internal/cart"
	"eugene/internal/keys"
	"eugene/internal/platform/db"
	"eugene/internal/products"
)

// These tests prove PLAN.md section 5 against a real Atlas replica set. Run with:
//   MONGODB_TEST_URI="mongodb+srv://..." go test -tags integration ./internal/orders/... -v -count=1
// The cleanup job never runs on its own here; test 4 calls CleanupOnce by hand.

var goodCard = Card{CardNumber: "4242 4242 4242 4242", Expiry: "12/99", CVC: "123", Name: "Test Buyer"}
var declineCard = Card{CardNumber: "4000 0000 0000 0002", Expiry: "12/99", CVC: "123", Name: "Test Buyer"}

// clock is the service's now(). Tests move it forward instead of sleeping.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) Now() time.Time          { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) Advance(d time.Duration) { c.mu.Lock(); defer c.mu.Unlock(); c.t = c.t.Add(d) }

type fixture struct {
	db        *mongo.Database
	svc       *Service
	clock     *clock
	productID bson.ObjectID
	offerID   bson.ObjectID
}

// setup makes a fresh eneba_test db with one product, one offer and n available keys.
// The db name is fixed here, never read from env, so this can't touch eneba_dev or eneba_prod.
func setup(t *testing.T, n int) *fixture {
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
	if err := db.EnsureIndexes(ctx, database); err != nil {
		t.Fatalf("indexes: %v", err)
	}

	// Mongo keeps milliseconds, so the clock does too; times then compare equal after a round trip.
	clk := &clock{t: time.Now().UTC().Truncate(time.Millisecond)}
	f := &fixture{db: database, clock: clk, productID: bson.NewObjectID(), offerID: bson.NewObjectID()}
	sellerID := bson.NewObjectID()

	mustInsert(t, database, "sellers", products.Seller{ID: sellerID, Slug: "seller", Name: "Seller", CreatedAt: clk.t})
	mustInsert(t, database, "products", products.Product{
		ID: f.productID, Slug: "test-game", Name: "Test Game", Type: "game", Platform: "steam", Region: "global",
		OriginalPriceCents: 2000, CashbackPct: 5, Active: true, CreatedAt: clk.t, UpdatedAt: clk.t,
	})
	mustInsert(t, database, "offers", products.Offer{
		ID: f.offerID, ProductID: f.productID, SellerID: sellerID, PriceCents: 1000, Stock: n, Active: true,
		CreatedAt: clk.t, UpdatedAt: clk.t,
	})
	for i := range n {
		mustInsert(t, database, "keys", keys.Key{
			ID: bson.NewObjectID(), OfferID: f.offerID, ProductID: f.productID,
			Code: fmt.Sprintf("CODE-%02d", i), Status: keys.StatusAvailable, CreatedAt: clk.t,
		})
	}

	f.svc = NewService(NewMongoRepo(database), keys.NewMongoRepo(database), products.NewMongoRepo(database),
		cart.NewMongoRepo(database), 10*time.Minute, clk.Now)
	return f
}

func mustInsert(t *testing.T, database *mongo.Database, coll string, doc any) {
	t.Helper()
	if _, err := database.Collection(coll).InsertOne(context.Background(), doc); err != nil {
		t.Fatalf("insert %s: %v", coll, err)
	}
}

// buyer makes a new user whose cart holds qty keys of the test offer.
func (f *fixture) buyer(t *testing.T, qty int) bson.ObjectID {
	t.Helper()
	user := bson.NewObjectID()
	mustInsert(t, f.db, "carts", cart.Cart{ID: bson.NewObjectID(), UserID: user, UpdatedAt: f.clock.Now(),
		Items: []cart.Item{{OfferID: f.offerID, ProductID: f.productID, Quantity: qty, AddedAt: f.clock.Now()}}})
	return user
}

func (f *fixture) order(t *testing.T, id string) Order {
	t.Helper()
	oid, _ := bson.ObjectIDFromHex(id)
	o, err := NewMongoRepo(f.db).ByID(context.Background(), oid)
	if err != nil {
		t.Fatalf("read order: %v", err)
	}
	return o
}

// checkStock asserts the cached offers.stock and product inStock both match the live key count.
func (f *fixture) checkStock(t *testing.T, want int) {
	t.Helper()
	ctx := context.Background()
	live, err := keys.NewMongoRepo(f.db).CountClaimable(ctx, []bson.ObjectID{f.offerID}, f.clock.Now())
	if err != nil {
		t.Fatal(err)
	}
	var offer products.Offer
	var product products.Product
	_ = f.db.Collection("offers").FindOne(ctx, bson.M{"_id": f.offerID}).Decode(&offer)
	_ = f.db.Collection("products").FindOne(ctx, bson.M{"_id": f.productID}).Decode(&product)
	if live[f.offerID] != want || offer.Stock != want || product.InStock != (want > 0) {
		t.Errorf("stock: live %d, offers.stock %d, product.inStock %v; want %d", live[f.offerID], offer.Stock, product.InStock, want)
	}
}

func keyIDs(o Order) []bson.ObjectID {
	var ids []bson.ObjectID
	for _, it := range o.Items {
		ids = append(ids, it.KeyIDs...)
	}
	return ids
}

// raceCreate starts n buyers of 1 key each at the same moment and returns the orders that succeeded.
func raceCreate(t *testing.T, f *fixture, n int) []View {
	t.Helper()
	users := make([]bson.ObjectID, n)
	for i := range users {
		users[i] = f.buyer(t, 1)
	}
	start := make(chan struct{})
	views := make([]View, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i, u := range users {
		wg.Go(func() {
			<-start
			views[i], errs[i] = f.svc.Create(context.Background(), u)
		})
	}
	close(start)
	wg.Wait()

	var ok []View
	for i, err := range errs {
		var oos *OutOfStockError
		switch {
		case err == nil:
			ok = append(ok, views[i])
		case errors.As(err, &oos):
		default:
			t.Fatalf("buyer %d: unexpected error %v", i, err)
		}
	}
	return ok
}

// assertDistinctKeys checks no key went to two orders.
func assertDistinctKeys(t *testing.T, f *fixture, views []View) {
	t.Helper()
	seen := map[bson.ObjectID]bool{}
	for _, v := range views {
		for _, id := range keyIDs(f.order(t, v.ID)) {
			if seen[id] {
				t.Fatalf("key %s was reserved by two orders", id.Hex())
			}
			seen[id] = true
		}
	}
}

// 1. Race: 5 keys, 20 buyers at once → exactly 5 orders, 5 distinct keys, stock 0.
func TestRaceTwentyBuyersFiveKeys(t *testing.T) {
	f := setup(t, 5)
	ok := raceCreate(t, f, 20)
	if len(ok) != 5 {
		t.Fatalf("%d orders succeeded, want 5", len(ok))
	}
	assertDistinctKeys(t, f, ok)
	f.checkStock(t, 0)
}

// 2. A's hold runs out, B takes the same keys, A's late payment is refused, B pays.
func TestExpiredReservationIsReclaimed(t *testing.T) {
	f := setup(t, 5)
	ctx := context.Background()
	a, b := f.buyer(t, 5), f.buyer(t, 5)

	orderA, err := f.svc.Create(ctx, a)
	if err != nil {
		t.Fatalf("A create: %v", err)
	}
	f.clock.Advance(11 * time.Minute)
	orderB, err := f.svc.Create(ctx, b)
	if err != nil {
		t.Fatalf("B create: %v", err)
	}
	if !sameKeys(keyIDs(f.order(t, orderA.ID)), keyIDs(f.order(t, orderB.ID))) {
		t.Fatalf("B did not get A's expired keys")
	}

	if _, err := f.svc.Pay(ctx, a, orderA.ID, goodCard); !errors.Is(err, ErrNotPayable) {
		t.Fatalf("A pay: want ErrNotPayable, got %v", err)
	}
	if got := f.order(t, orderA.ID).Status; got != StatusExpired {
		t.Errorf("A status in db = %q, want expired", got)
	}

	paid, err := f.svc.Pay(ctx, b, orderB.ID, goodCard)
	if err != nil {
		t.Fatalf("B pay: %v", err)
	}
	if paid.Status != StatusPaid || len(paid.Items[0].Keys) != 5 {
		t.Fatalf("B: status %q with %d codes, want paid with 5", paid.Status, len(paid.Items[0].Keys))
	}
	f.checkStock(t, 0)
}

// 3. Race on expired keys: 3 expired holds + 2 available, 10 buyers → exactly 5 orders, all distinct.
func TestRaceOnExpiredKeys(t *testing.T) {
	f := setup(t, 5)
	ctx := context.Background()
	past := f.clock.Now().Add(-time.Minute)
	oldOrder := bson.NewObjectID()
	cur, err := f.db.Collection("keys").Find(ctx, bson.M{}, options.Find().SetLimit(3))
	if err != nil {
		t.Fatal(err)
	}
	var three []keys.Key
	if err := cur.All(ctx, &three); err != nil {
		t.Fatal(err)
	}
	for _, k := range three {
		_, err := f.db.Collection("keys").UpdateOne(ctx, bson.M{"_id": k.ID}, bson.M{"$set": bson.M{
			"status": keys.StatusReserved, "orderId": oldOrder, "reservedUntil": past,
		}})
		if err != nil {
			t.Fatal(err)
		}
	}

	ok := raceCreate(t, f, 10)
	if len(ok) != 5 {
		t.Fatalf("%d orders succeeded, want 5", len(ok))
	}
	assertDistinctKeys(t, f, ok)
	f.checkStock(t, 0)
}

// 4. Cleanup expires A but leaves B's keys (claimed from A after A's hold ran out) alone.
func TestCleanupDoesNotSteal(t *testing.T) {
	f := setup(t, 5)
	ctx := context.Background()
	a, b := f.buyer(t, 5), f.buyer(t, 5)
	orderA, err := f.svc.Create(ctx, a)
	if err != nil {
		t.Fatal(err)
	}
	f.clock.Advance(11 * time.Minute)
	orderB, err := f.svc.Create(ctx, b)
	if err != nil {
		t.Fatal(err)
	}

	n, err := f.svc.CleanupOnce(ctx)
	if err != nil || n != 1 {
		t.Fatalf("cleanup: expired %d, err %v; want 1, nil", n, err)
	}
	if got := f.order(t, orderA.ID).Status; got != StatusExpired {
		t.Errorf("A status in db = %q, want expired", got)
	}
	bID, _ := bson.ObjectIDFromHex(orderB.ID)
	held, err := f.db.Collection("keys").CountDocuments(ctx, bson.M{"orderId": bID, "status": keys.StatusReserved})
	if err != nil || held != 5 {
		t.Errorf("B holds %d keys, want 5 (err %v)", held, err)
	}
	f.checkStock(t, 0)
}

// 5. Checkout: a declined card puts the keys back and recounts stock in the same request;
// the cart stays. Then a good card sells the keys, fills the codes and empties the cart.
func TestDeclineReleasesThenPaySells(t *testing.T) {
	f := setup(t, 5)
	ctx := context.Background()
	user := f.buyer(t, 2)

	first, err := f.svc.Create(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	f.checkStock(t, 3)

	declined, err := f.svc.Pay(ctx, user, first.ID, declineCard)
	if !errors.Is(err, ErrDeclined) || declined.Status != StatusFailed {
		t.Fatalf("decline: status %q, err %v; want failed, ErrDeclined", declined.Status, err)
	}
	if declined.Payment == nil || declined.Payment.Last4 != "0002" {
		t.Errorf("decline payment = %+v, want last4 0002", declined.Payment)
	}
	f.checkStock(t, 5)
	c, _ := cart.NewMongoRepo(f.db).Get(ctx, user)
	if len(c.Items) != 1 {
		t.Errorf("cart after decline has %d lines, want 1", len(c.Items))
	}

	second, err := f.svc.Create(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	paid, err := f.svc.Pay(ctx, user, second.ID, goodCard)
	if err != nil {
		t.Fatalf("pay: %v", err)
	}
	if paid.Status != StatusPaid || len(paid.Items[0].Keys) != 2 || paid.PaidAt == nil {
		t.Fatalf("pay: %+v, want paid with 2 codes", paid)
	}
	f.checkStock(t, 3)
	c, _ = cart.NewMongoRepo(f.db).Get(ctx, user)
	if len(c.Items) != 0 {
		t.Errorf("cart after pay has %d lines, want 0", len(c.Items))
	}
	var product products.Product
	_ = f.db.Collection("products").FindOne(ctx, bson.M{"_id": f.productID}).Decode(&product)
	if product.SoldCount != 2 {
		t.Errorf("soldCount = %d, want 2", product.SoldCount)
	}
	if _, err := f.svc.Pay(ctx, user, second.ID, goodCard); !errors.Is(err, ErrNotPayable) {
		t.Errorf("paying twice: want ErrNotPayable, got %v", err)
	}
}

// 6. Starting checkout again replaces the user's earlier pending order: its keys go to the new
// order instead of being held twice. 3 keys: B holds 1, A holds the other 2 and checks out again.
// Without the fix, A's second order would get OUT_OF_STOCK on A's own keys.
func TestNewOrderReplacesUsersPendingOrder(t *testing.T) {
	f := setup(t, 3)
	ctx := context.Background()
	a, b := f.buyer(t, 2), f.buyer(t, 1)

	orderB, err := f.svc.Create(ctx, b)
	if err != nil {
		t.Fatalf("B create: %v", err)
	}
	first, err := f.svc.Create(ctx, a)
	if err != nil {
		t.Fatalf("A first create: %v", err)
	}
	second, err := f.svc.Create(ctx, a)
	if err != nil {
		t.Fatalf("A second create: %v", err)
	}

	if got := f.order(t, first.ID).Status; got != StatusExpired {
		t.Errorf("A's first order is %q in db, want expired", got)
	}
	if !sameKeys(keyIDs(f.order(t, first.ID)), keyIDs(f.order(t, second.ID))) {
		t.Errorf("A's second order did not get the keys the first one released")
	}
	if got := f.order(t, orderB.ID).Status; got != StatusPendingPayment {
		t.Errorf("B's order is %q, want pending_payment (another user's order must be left alone)", got)
	}
	f.checkStock(t, 0)

	if _, err := f.svc.Pay(ctx, a, first.ID, goodCard); !errors.Is(err, ErrNotPayable) {
		t.Errorf("paying the replaced order: want ErrNotPayable, got %v", err)
	}
	if _, err := f.svc.Pay(ctx, a, second.ID, goodCard); err != nil {
		t.Errorf("paying the new order: %v", err)
	}
}

func sameKeys(a, b []bson.ObjectID) bool {
	if len(a) != len(b) {
		return false
	}
	set := map[bson.ObjectID]bool{}
	for _, id := range a {
		set[id] = true
	}
	for _, id := range b {
		if !set[id] {
			return false
		}
	}
	return true
}
