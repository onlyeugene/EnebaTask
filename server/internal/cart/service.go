package cart

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrOfferNotFound = errors.New("offer not found")

// ValidationError lists broken field rules; the handler sends it as 422 with "fields".
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return fmt.Sprintf("validation failed: %v", e.Fields) }

type Service struct {
	repo  Repository
	stock StockCounter
	now   func() time.Time
}

func NewService(repo Repository, stock StockCounter, now func() time.Time) *Service {
	return &Service{repo: repo, stock: stock, now: now}
}

// Quote prices a guest cart with live prices and stock. Nothing is saved.
func (s *Service) Quote(ctx context.Context, in []ItemInput) (View, error) {
	items, err := parseItems(in)
	if err != nil {
		return View{}, err
	}
	return s.view(ctx, items)
}

func (s *Service) Get(ctx context.Context, userID bson.ObjectID) (View, error) {
	c, err := s.repo.Get(ctx, userID)
	if err != nil {
		return View{}, err
	}
	return s.view(ctx, c.Items)
}

// SetQuantity adds the offer to the cart or changes its quantity. It is not capped at stock:
// the user picked this number, so the cart shows NOT_ENOUGH_STOCK instead of changing it silently.
func (s *Service) SetQuantity(ctx context.Context, userID bson.ObjectID, offerHex string, qty int) (View, error) {
	offerID, err := bson.ObjectIDFromHex(offerHex)
	if err != nil {
		return View{}, ErrOfferNotFound
	}
	if qty < 1 || qty > MaxQuantity {
		return View{}, &ValidationError{Fields: map[string]string{"quantity": fmt.Sprintf("must be between 1 and %d", MaxQuantity)}}
	}
	offers, err := s.repo.Offers(ctx, []bson.ObjectID{offerID})
	if err != nil {
		return View{}, err
	}
	info, ok := offers[offerID]
	if !ok {
		return View{}, ErrOfferNotFound
	}

	now := s.now()
	c, err := s.repo.Update(ctx, userID, now, func(c *Cart) error {
		if i := indexOf(c.Items, offerID); i >= 0 {
			c.Items[i].Quantity = qty
			return nil
		}
		if len(c.Items) >= MaxLines {
			return &ValidationError{Fields: map[string]string{"offerId": fmt.Sprintf("cart holds at most %d different offers", MaxLines)}}
		}
		c.Items = append(c.Items, Item{OfferID: offerID, ProductID: info.ProductID, Quantity: qty, AddedAt: now})
		return nil
	})
	if err != nil {
		return View{}, err
	}
	return s.view(ctx, c.Items)
}

// Remove deletes a line. Removing an offer that isn't in the cart is not an error.
func (s *Service) Remove(ctx context.Context, userID bson.ObjectID, offerHex string) (View, error) {
	offerID, err := bson.ObjectIDFromHex(offerHex)
	if err != nil {
		return View{}, ErrOfferNotFound
	}
	c, err := s.repo.Update(ctx, userID, s.now(), func(c *Cart) error {
		c.Items = slices.DeleteFunc(c.Items, func(it Item) bool { return it.OfferID == offerID })
		return nil
	})
	if err != nil {
		return View{}, err
	}
	return s.view(ctx, c.Items)
}

// Merge adds a guest cart to the user's server cart at login (rules in PLAN.md section 5b).
// Offers and stock are read before the transaction, so the merge step inside it is pure
// and gives the same result when the transaction retries it.
func (s *Service) Merge(ctx context.Context, userID bson.ObjectID, in []ItemInput) (View, []Dropped, error) {
	guest, err := parseItems(in)
	if err != nil {
		return View{}, nil, err
	}
	ids := offerIDs(guest)
	offers, err := s.repo.Offers(ctx, ids)
	if err != nil {
		return View{}, nil, err
	}
	now := s.now()
	stock, err := s.stock.CountClaimable(ctx, ids, now)
	if err != nil {
		return View{}, nil, err
	}

	var dropped []Dropped
	c, err := s.repo.Update(ctx, userID, now, func(c *Cart) error {
		dropped = mergeItems(c, guest, offers, stock, now)
		return nil
	})
	if err != nil {
		return View{}, nil, err
	}
	v, err := s.view(ctx, c.Items)
	return v, dropped, err
}

// mergeItems applies each guest line to the cart and returns the lines it left out.
// Same offer in both → the larger quantity, never the sum, so a retried merge can't double it.
// Every merged quantity is capped at MaxQuantity and at live stock.
func mergeItems(c *Cart, guest []Item, offers map[bson.ObjectID]OfferInfo, stock map[bson.ObjectID]int, now time.Time) []Dropped {
	dropped := []Dropped{}
	for _, g := range guest {
		info, ok := offers[g.OfferID]
		if !ok {
			dropped = append(dropped, Dropped{OfferID: g.OfferID.Hex(), Reason: DropOfferGone})
			continue
		}
		available := stock[g.OfferID]
		if available == 0 {
			dropped = append(dropped, Dropped{OfferID: g.OfferID.Hex(), Reason: DropOutOfStock})
			continue
		}
		if i := indexOf(c.Items, g.OfferID); i >= 0 {
			c.Items[i].Quantity = min(max(c.Items[i].Quantity, g.Quantity), MaxQuantity, available)
			continue
		}
		if len(c.Items) >= MaxLines {
			dropped = append(dropped, Dropped{OfferID: g.OfferID.Hex(), Reason: DropCartFull})
			continue
		}
		c.Items = append(c.Items, Item{
			OfferID: g.OfferID, ProductID: info.ProductID,
			Quantity: min(g.Quantity, MaxQuantity, available), AddedAt: now,
		})
	}
	return dropped
}

// view builds the API cart from stored lines, with live prices and stock.
func (s *Service) view(ctx context.Context, items []Item) (View, error) {
	ids := offerIDs(items)
	offers, err := s.repo.Offers(ctx, ids)
	if err != nil {
		return View{}, err
	}
	stock, err := s.stock.CountClaimable(ctx, ids, s.now())
	if err != nil {
		return View{}, err
	}

	v := View{Items: make([]Line, 0, len(items))} // [] not null in JSON
	for _, it := range items {
		v.ItemCount += it.Quantity
		info, ok := offers[it.OfferID]
		if !ok {
			v.Items = append(v.Items, Line{OfferID: it.OfferID.Hex(), Quantity: it.Quantity, Problem: ProblemOfferGone})
			continue
		}
		line := toLine(info, it.Quantity, stock[it.OfferID])
		v.TotalCents += line.LineTotalCents
		v.CashbackCents += line.CashbackCents
		v.Items = append(v.Items, line)
	}
	return v, nil
}

func toLine(o OfferInfo, qty, available int) Line {
	total := o.PriceCents * int64(qty)
	l := Line{
		OfferID: o.OfferID.Hex(), ProductID: o.ProductID.Hex(),
		Slug: o.Slug, Name: o.Name, Platform: o.Platform, Region: o.Region,
		CoverURL:       "/api/products/" + o.Slug + "/cover.svg",
		Seller:         LineSeller{ID: o.SellerID.Hex(), Name: o.SellerName},
		UnitPriceCents: o.PriceCents, Quantity: qty, LineTotalCents: total,
		// Whole cents, rounded down: cashback is display only (PLAN.md section 9).
		CashbackCents: total * int64(o.CashbackPct) / 100,
		Available:     available,
	}
	switch {
	case available == 0:
		l.Problem = ProblemOutOfStock
	case qty > available:
		l.Problem = ProblemNotEnoughStock
	}
	return l
}

// parseItems checks a quote or merge body. The client keys its guest cart by offerId,
// so a repeated offer means a broken request and is refused rather than guessed at.
func parseItems(in []ItemInput) ([]Item, error) {
	fields := map[string]string{}
	if len(in) > MaxLines {
		fields["items"] = fmt.Sprintf("at most %d items", MaxLines)
		return nil, &ValidationError{Fields: fields}
	}
	items := make([]Item, 0, len(in))
	for i, it := range in {
		id, err := bson.ObjectIDFromHex(it.OfferID)
		switch {
		case err != nil:
			fields[fmt.Sprintf("items[%d].offerId", i)] = "is not a valid id"
		case indexOf(items, id) >= 0:
			fields[fmt.Sprintf("items[%d].offerId", i)] = "is listed twice"
		}
		if it.Quantity < 1 || it.Quantity > MaxQuantity {
			fields[fmt.Sprintf("items[%d].quantity", i)] = fmt.Sprintf("must be between 1 and %d", MaxQuantity)
		}
		items = append(items, Item{OfferID: id, Quantity: it.Quantity})
	}
	if len(fields) > 0 {
		return nil, &ValidationError{Fields: fields}
	}
	return items, nil
}

func indexOf(items []Item, offerID bson.ObjectID) int {
	return slices.IndexFunc(items, func(it Item) bool { return it.OfferID == offerID })
}

func offerIDs(items []Item) []bson.ObjectID {
	ids := make([]bson.ObjectID, len(items))
	for i, it := range items {
		ids[i] = it.OfferID
	}
	return ids
}
