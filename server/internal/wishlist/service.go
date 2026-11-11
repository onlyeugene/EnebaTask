package wishlist

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/products"
)

var ErrProductNotFound = errors.New("product not found")

// ErrFull is returned when an add would go past MaxItems.
var ErrFull = errors.New("wishlist is full")

// ValidationError lists broken field rules; the handler sends it as 422 with "fields".
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return fmt.Sprintf("validation failed: %v", e.Fields) }

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository, now func() time.Time) *Service {
	return &Service{repo: repo, now: now}
}

func (s *Service) Get(ctx context.Context, userID bson.ObjectID) (View, error) {
	w, err := s.repo.Get(ctx, userID)
	if err != nil {
		return View{}, err
	}
	return s.view(ctx, w.Items)
}

// Add saves a product. Adding one that is already there changes nothing (PUT is idempotent).
func (s *Service) Add(ctx context.Context, userID bson.ObjectID, productHex string) (View, error) {
	id, err := s.existingProduct(ctx, productHex)
	if err != nil {
		return View{}, err
	}
	now := s.now()
	full := false
	w, err := s.repo.Update(ctx, userID, now, func(w *Wishlist) {
		full = false // reset: the callback can run again on retry
		if indexOf(w.Items, id) >= 0 {
			return
		}
		if len(w.Items) >= MaxItems {
			full = true
			return
		}
		w.Items = append(w.Items, Item{ProductID: id, AddedAt: now})
	})
	if err != nil {
		return View{}, err
	}
	if full {
		return View{}, ErrFull
	}
	return s.view(ctx, w.Items)
}

// Remove deletes a product. Removing one that isn't there is not an error.
func (s *Service) Remove(ctx context.Context, userID bson.ObjectID, productHex string) (View, error) {
	id, err := bson.ObjectIDFromHex(productHex)
	if err != nil {
		return View{}, ErrProductNotFound
	}
	w, err := s.repo.Update(ctx, userID, s.now(), func(w *Wishlist) {
		w.Items = slices.DeleteFunc(w.Items, func(it Item) bool { return it.ProductID == id })
	})
	if err != nil {
		return View{}, err
	}
	return s.view(ctx, w.Items)
}

// Merge adds a guest wishlist at login: the union of both lists, each product once.
// A product already saved keeps its own date, so running the merge twice changes nothing.
// Gone products are skipped, and so is anything past MaxItems.
func (s *Service) Merge(ctx context.Context, userID bson.ObjectID, in []MergeInput) (View, error) {
	guest, err := parseMerge(in)
	if err != nil {
		return View{}, err
	}
	found, err := s.repo.Products(ctx, productIDs(guest))
	if err != nil {
		return View{}, err
	}
	now := s.now()
	w, err := s.repo.Update(ctx, userID, now, func(w *Wishlist) {
		mergeItems(w, guest, found, now)
	})
	if err != nil {
		return View{}, err
	}
	return s.view(ctx, w.Items)
}

func mergeItems(w *Wishlist, guest []Item, found map[bson.ObjectID]products.Product, now time.Time) {
	for _, g := range guest {
		if _, ok := found[g.ProductID]; !ok || indexOf(w.Items, g.ProductID) >= 0 || len(w.Items) >= MaxItems {
			continue
		}
		// The date comes from the browser, so a missing or future one becomes now.
		added := g.AddedAt
		if added.IsZero() || added.After(now) {
			added = now
		}
		w.Items = append(w.Items, Item{ProductID: g.ProductID, AddedAt: added})
	}
}

// parseMerge checks a merge body. A product listed twice is kept once (it is a union).
func parseMerge(in []MergeInput) ([]Item, error) {
	if len(in) > MaxItems {
		return nil, &ValidationError{Fields: map[string]string{"items": fmt.Sprintf("at most %d items", MaxItems)}}
	}
	fields := map[string]string{}
	items := make([]Item, 0, len(in))
	for i, it := range in {
		id, err := bson.ObjectIDFromHex(it.ProductID)
		if err != nil {
			fields[fmt.Sprintf("items[%d].productId", i)] = "is not a valid id"
			continue
		}
		if indexOf(items, id) >= 0 {
			continue
		}
		item := Item{ProductID: id}
		if it.AddedAt != nil {
			item.AddedAt = it.AddedAt.UTC()
		}
		items = append(items, item)
	}
	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}
	return items, nil
}

func (s *Service) existingProduct(ctx context.Context, productHex string) (bson.ObjectID, error) {
	id, err := bson.ObjectIDFromHex(productHex)
	if err != nil {
		return id, ErrProductNotFound
	}
	found, err := s.repo.Products(ctx, []bson.ObjectID{id})
	if err != nil {
		return id, err
	}
	if _, ok := found[id]; !ok {
		return id, ErrProductNotFound
	}
	return id, nil
}

// view turns stored items into cards. A product taken down since it was saved is left out
// of the answer but kept in the doc, so it comes back if the product is put back up.
func (s *Service) view(ctx context.Context, items []Item) (View, error) {
	found, err := s.repo.Products(ctx, productIDs(items))
	if err != nil {
		return View{}, err
	}
	v := View{Items: make([]Entry, 0, len(items))} // [] not null in JSON
	for _, it := range items {
		if p, ok := found[it.ProductID]; ok {
			// UTC: a date just set from time.Now() is still in the server's local zone.
			v.Items = append(v.Items, Entry{Product: products.ToCard(p), AddedAt: it.AddedAt.UTC()})
		}
	}
	return v, nil
}

func indexOf(items []Item, id bson.ObjectID) int {
	return slices.IndexFunc(items, func(it Item) bool { return it.ProductID == id })
}

func productIDs(items []Item) []bson.ObjectID {
	ids := make([]bson.ObjectID, len(items))
	for i, it := range items {
		ids[i] = it.ProductID
	}
	return ids
}
