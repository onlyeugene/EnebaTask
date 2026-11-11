package wishlist

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/products"
)

// In-memory repo, so these tests need no database.

type memRepo struct {
	mu        sync.Mutex
	wishlists map[bson.ObjectID]Wishlist
	products  map[bson.ObjectID]products.Product
}

func (m *memRepo) Get(_ context.Context, userID bson.ObjectID) (Wishlist, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w := m.wishlists[userID]
	w.Items = slices.Clone(w.Items)
	return w, nil
}

// Update holds the lock for read, change and write, like the Mongo transaction.
func (m *memRepo) Update(_ context.Context, userID bson.ObjectID, now time.Time, change func(*Wishlist)) (Wishlist, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w := m.wishlists[userID]
	w.UserID, w.Items = userID, slices.Clone(w.Items)
	change(&w)
	w.UpdatedAt = now
	m.wishlists[userID] = w
	return w, nil
}

func (m *memRepo) Products(_ context.Context, ids []bson.ObjectID) (map[bson.ObjectID]products.Product, error) {
	out := map[bson.ObjectID]products.Product{}
	for _, id := range ids {
		if p, ok := m.products[id]; ok {
			out[id] = p
		}
	}
	return out, nil
}

var now = time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)

type fixture struct {
	svc  *Service
	repo *memRepo
	user bson.ObjectID
}

func newFixture() *fixture {
	repo := &memRepo{wishlists: map[bson.ObjectID]Wishlist{}, products: map[bson.ObjectID]products.Product{}}
	return &fixture{svc: NewService(repo, func() time.Time { return now }), repo: repo, user: bson.NewObjectID()}
}

func (f *fixture) product() bson.ObjectID {
	id := bson.NewObjectID()
	f.repo.products[id] = products.Product{ID: id, Slug: id.Hex(), Name: "Test", Active: true}
	return id
}

func (f *fixture) add(t *testing.T, id bson.ObjectID) {
	t.Helper()
	if _, err := f.svc.Add(context.Background(), f.user, id.Hex()); err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func (f *fixture) merge(t *testing.T, in ...MergeInput) View {
	t.Helper()
	v, err := f.svc.Merge(context.Background(), f.user, in)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	return v
}

func ids(v View) []string {
	out := make([]string, len(v.Items))
	for i, e := range v.Items {
		out[i] = e.Product.ID
	}
	return out
}

func at(t time.Time) *time.Time { return &t }

func TestMergeIsUnionWithoutDuplicates(t *testing.T) {
	f := newFixture()
	a, b, c := f.product(), f.product(), f.product()
	f.add(t, a) // saved on the server at `now`

	earlier := now.Add(-time.Hour)
	v := f.merge(t,
		MergeInput{ProductID: a.Hex(), AddedAt: at(earlier)}, // in both lists
		MergeInput{ProductID: b.Hex(), AddedAt: at(earlier)},
		MergeInput{ProductID: b.Hex()}, // listed twice in the body
		MergeInput{ProductID: c.Hex()},
	)

	want := []string{a.Hex(), b.Hex(), c.Hex()}
	if got := ids(v); !slices.Equal(got, want) {
		t.Fatalf("items = %v, want %v", got, want)
	}
	if !v.Items[0].AddedAt.Equal(now) {
		t.Errorf("server item kept its own date? got %v, want %v", v.Items[0].AddedAt, now)
	}
	if !v.Items[1].AddedAt.Equal(earlier) {
		t.Errorf("guest date lost: got %v, want %v", v.Items[1].AddedAt, earlier)
	}
	if !v.Items[2].AddedAt.Equal(now) {
		t.Errorf("missing date should become now, got %v", v.Items[2].AddedAt)
	}
}

func TestMergeTwiceGivesSameList(t *testing.T) {
	f := newFixture()
	a, b := f.product(), f.product()
	f.add(t, a)
	guest := []MergeInput{{ProductID: a.Hex()}, {ProductID: b.Hex(), AddedAt: at(now.Add(-time.Minute))}}

	first := f.merge(t, guest...)
	second := f.merge(t, guest...)
	if !slices.Equal(ids(first), ids(second)) || len(second.Items) != 2 {
		t.Fatalf("second merge changed the list: %v → %v", ids(first), ids(second))
	}
	for i := range first.Items {
		if !first.Items[i].AddedAt.Equal(second.Items[i].AddedAt) {
			t.Errorf("item %d date changed: %v → %v", i, first.Items[i].AddedAt, second.Items[i].AddedAt)
		}
	}
}

func TestMergeSkipsGoneProductsAndFutureDates(t *testing.T) {
	f := newFixture()
	a := f.product()
	gone := bson.NewObjectID() // not in the products map

	v := f.merge(t, MergeInput{ProductID: gone.Hex()}, MergeInput{ProductID: a.Hex(), AddedAt: at(now.Add(time.Hour))})
	if got := ids(v); !slices.Equal(got, []string{a.Hex()}) {
		t.Fatalf("items = %v, want only %s", got, a.Hex())
	}
	if !v.Items[0].AddedAt.Equal(now) {
		t.Errorf("future date not clamped: %v", v.Items[0].AddedAt)
	}
	if n := len(f.repo.wishlists[f.user].Items); n != 1 {
		t.Errorf("stored %d items, want 1 (gone product not saved)", n)
	}
}

func TestMergeStopsAtMaxItems(t *testing.T) {
	f := newFixture()
	for range MaxItems {
		f.add(t, f.product())
	}
	v := f.merge(t, MergeInput{ProductID: f.product().Hex()})
	if len(v.Items) != MaxItems {
		t.Errorf("items = %d, want %d", len(v.Items), MaxItems)
	}
}
