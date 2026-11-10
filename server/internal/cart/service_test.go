package cart

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// In-memory repo and stock, so these tests need no database.

type memRepo struct {
	mu     sync.Mutex
	carts  map[bson.ObjectID]Cart
	offers map[bson.ObjectID]OfferInfo
}

func (m *memRepo) Get(_ context.Context, userID bson.ObjectID) (Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.carts[userID]
	c.Items = slices.Clone(c.Items)
	return c, nil
}

// Update holds the lock for read, change and write, like the Mongo transaction.
func (m *memRepo) Update(_ context.Context, userID bson.ObjectID, now time.Time, change func(*Cart) error) (Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.carts[userID]
	c.UserID, c.Items = userID, slices.Clone(c.Items)
	if err := change(&c); err != nil {
		return Cart{}, err
	}
	c.UpdatedAt = now
	m.carts[userID] = c
	return c, nil
}

func (m *memRepo) Offers(_ context.Context, ids []bson.ObjectID) (map[bson.ObjectID]OfferInfo, error) {
	out := map[bson.ObjectID]OfferInfo{}
	for _, id := range ids {
		if o, ok := m.offers[id]; ok {
			out[id] = o
		}
	}
	return out, nil
}

type memStock map[bson.ObjectID]int

func (s memStock) CountClaimable(_ context.Context, ids []bson.ObjectID, _ time.Time) (map[bson.ObjectID]int, error) {
	out := map[bson.ObjectID]int{}
	for _, id := range ids {
		if n := s[id]; n > 0 {
			out[id] = n
		}
	}
	return out, nil
}

type fixture struct {
	svc   *Service
	repo  *memRepo
	stock memStock
	user  bson.ObjectID
}

func newFixture() *fixture {
	repo := &memRepo{carts: map[bson.ObjectID]Cart{}, offers: map[bson.ObjectID]OfferInfo{}}
	stock := memStock{}
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	return &fixture{
		svc:  NewService(repo, stock, func() time.Time { return now }),
		repo: repo, stock: stock, user: bson.NewObjectID(),
	}
}

// offer adds a buyable offer with the given stock and returns its id.
func (f *fixture) offer(productID bson.ObjectID, stock int) bson.ObjectID {
	id := bson.NewObjectID()
	f.repo.offers[id] = OfferInfo{OfferID: id, ProductID: productID, PriceCents: 1000, Name: "Test", SellerID: bson.NewObjectID()}
	f.stock[id] = stock
	return id
}

func (f *fixture) put(t *testing.T, offerID bson.ObjectID, qty int) {
	t.Helper()
	if _, err := f.svc.SetQuantity(context.Background(), f.user, offerID.Hex(), qty); err != nil {
		t.Fatalf("SetQuantity: %v", err)
	}
}

func (f *fixture) merge(t *testing.T, items ...ItemInput) (View, []Dropped) {
	t.Helper()
	v, dropped, err := f.svc.Merge(context.Background(), f.user, items)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	return v, dropped
}

func qtyOf(v View, offerID bson.ObjectID) int {
	for _, l := range v.Items {
		if l.OfferID == offerID.Hex() {
			return l.Quantity
		}
	}
	return 0
}

func TestMergeSameOfferTakesLargerQuantity(t *testing.T) {
	f := newFixture()
	a, b := f.offer(bson.NewObjectID(), 50), f.offer(bson.NewObjectID(), 50)
	f.put(t, a, 2)
	f.put(t, b, 5)

	v, _ := f.merge(t, ItemInput{a.Hex(), 3}, ItemInput{b.Hex(), 1})
	if got := qtyOf(v, a); got != 3 {
		t.Errorf("guest larger: qty = %d, want 3", got)
	}
	if got := qtyOf(v, b); got != 5 {
		t.Errorf("server larger: qty = %d, want 5", got)
	}
}

func TestMergeTwiceGivesSameCart(t *testing.T) {
	f := newFixture()
	a, b := f.offer(bson.NewObjectID(), 50), f.offer(bson.NewObjectID(), 50)
	f.put(t, a, 2)
	guest := []ItemInput{{a.Hex(), 4}, {b.Hex(), 1}}

	first, _ := f.merge(t, guest...)
	second, _ := f.merge(t, guest...)
	if len(second.Items) != 2 || qtyOf(second, a) != 4 || qtyOf(second, b) != 1 {
		t.Fatalf("second merge changed the cart: %+v", second.Items)
	}
	if first.ItemCount != second.ItemCount || first.TotalCents != second.TotalCents {
		t.Errorf("totals differ: %d/%d vs %d/%d", first.ItemCount, first.TotalCents, second.ItemCount, second.TotalCents)
	}
}

func TestMergeCapsAtLiveStock(t *testing.T) {
	f := newFixture()
	a, b := f.offer(bson.NewObjectID(), 3), f.offer(bson.NewObjectID(), 2)
	f.put(t, b, 6) // over stock already; merge lowers it because b is in the guest cart too

	v, _ := f.merge(t, ItemInput{a.Hex(), 7}, ItemInput{b.Hex(), 1})
	if got := qtyOf(v, a); got != 3 {
		t.Errorf("new line: qty = %d, want 3 (stock)", got)
	}
	if got := qtyOf(v, b); got != 2 {
		t.Errorf("existing line: qty = %d, want 2 (stock)", got)
	}
}

func TestMergeDropsGoneAndSoldOutKeepsOtherSeller(t *testing.T) {
	f := newFixture()
	product := bson.NewObjectID()
	sellerA, sellerB := f.offer(product, 5), f.offer(product, 5) // same game, two sellers
	soldOut := f.offer(bson.NewObjectID(), 0)
	gone := bson.NewObjectID() // not in the offers map
	f.put(t, sellerA, 1)

	v, dropped := f.merge(t, ItemInput{sellerB.Hex(), 2}, ItemInput{soldOut.Hex(), 1}, ItemInput{gone.Hex(), 1})
	if qtyOf(v, sellerA) != 1 || qtyOf(v, sellerB) != 2 || len(v.Items) != 2 {
		t.Errorf("want both sellers' lines, got %+v", v.Items)
	}
	want := []Dropped{{soldOut.Hex(), DropOutOfStock}, {gone.Hex(), DropOfferGone}}
	if !slices.Equal(dropped, want) {
		t.Errorf("dropped = %+v, want %+v", dropped, want)
	}
}

func TestMergeStopsAtMaxLines(t *testing.T) {
	f := newFixture()
	for range MaxLines {
		f.put(t, f.offer(bson.NewObjectID(), 5), 1)
	}
	extra := f.offer(bson.NewObjectID(), 5)

	v, dropped := f.merge(t, ItemInput{extra.Hex(), 1})
	if len(v.Items) != MaxLines {
		t.Errorf("lines = %d, want %d", len(v.Items), MaxLines)
	}
	if len(dropped) != 1 || dropped[0].Reason != DropCartFull {
		t.Errorf("dropped = %+v, want one CART_FULL", dropped)
	}
}
