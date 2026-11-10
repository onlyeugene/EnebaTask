package orders

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/cart"
	"eugene/internal/keys"
	"eugene/internal/products"
)

var (
	ErrCartEmpty  = errors.New("cart is empty")
	ErrNotPayable = errors.New("order is not payable")
	ErrDeclined   = errors.New("payment declined")

	// errHoldLost means the order's hold ran out (or some of its keys went to another order).
	// Pay returns it from inside the transaction to roll the sale back, then expires the order.
	errHoldLost = errors.New("reservation ran out")
)

// OutOfStockError says which cart line couldn't be reserved and how many keys it could get.
type OutOfStockError struct {
	OfferID   bson.ObjectID
	Available int
}

func (e *OutOfStockError) Error() string {
	return fmt.Sprintf("offer %s: only %d keys left", e.OfferID.Hex(), e.Available)
}

// ValidationError lists broken field rules; the handler sends it as 422 with "fields".
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return fmt.Sprintf("validation failed: %v", e.Fields) }

type Service struct {
	repo    Repository
	keys    Keys
	catalog Catalog
	carts   Carts
	ttl     time.Duration // how long an order holds its keys
	now     func() time.Time
}

func NewService(repo Repository, keys Keys, catalog Catalog, carts Carts, ttl time.Duration, now func() time.Time) *Service {
	return &Service{repo: repo, keys: keys, catalog: catalog, carts: carts, ttl: ttl, now: now}
}

// Create turns the user's server cart into a pending_payment order and reserves one key per unit,
// all in one transaction (PLAN.md section 5, step 1). If any line can't get enough keys, the
// whole transaction rolls back and nothing stays reserved.
func (s *Service) Create(ctx context.Context, userID bson.ObjectID) (View, error) {
	c, err := s.carts.Get(ctx, userID)
	if err != nil {
		return View{}, err
	}
	if len(c.Items) == 0 {
		return View{}, ErrCartEmpty
	}
	ids := make([]bson.ObjectID, len(c.Items))
	for i, it := range c.Items {
		ids[i] = it.OfferID
	}
	offers, err := s.carts.Offers(ctx, ids)
	if err != nil {
		return View{}, err
	}
	for _, it := range c.Items {
		if _, ok := offers[it.OfferID]; !ok { // offer or product taken down: nothing to buy
			return View{}, &OutOfStockError{OfferID: it.OfferID}
		}
	}

	now := s.now()
	o := Order{
		ID: bson.NewObjectID(), Number: newNumber(), UserID: userID, Status: StatusPendingPayment,
		ReservedUntil: now.Add(s.ttl), CreatedAt: now, UpdatedAt: now,
	}
	err = s.repo.Transaction(ctx, func(ctx context.Context) error {
		// A new checkout replaces the user's earlier ones. Without this, a user who goes back from
		// checkout and starts again holds two sets of keys, and on low stock their own first order
		// makes the second one fail. Released first, so the new order can claim those same keys.
		replaced, err := s.expirePending(ctx, userID, now)
		if err != nil {
			return err
		}
		// Built fresh on every run: a retried transaction must not keep key ids from the failed run.
		o.Items, o.TotalCents, o.CashbackCents = newItems(c.Items, offers)
		for i := range o.Items {
			it := &o.Items[i]
			for n := range it.Quantity {
				keyID, err := s.keys.Claim(ctx, it.OfferID, o.ID, now, o.ReservedUntil)
				if errors.Is(err, keys.ErrNoneLeft) {
					// n keys were claimable when this one failed, so that is the stock left.
					return &OutOfStockError{OfferID: it.OfferID, Available: n}
				}
				if err != nil {
					return err
				}
				it.KeyIDs = append(it.KeyIDs, keyID)
			}
		}
		// The replaced orders may hold other products, whose stock just went up.
		if err := s.refreshStock(ctx, append(replaced, o.Items...), now); err != nil {
			return err
		}
		return s.repo.Insert(ctx, o)
	})
	if err != nil {
		return View{}, err
	}
	return toView(o, now, true), nil
}

// Pay runs the simulated payment, then in one transaction either sells the held keys and copies
// their codes into the order, or (declined) releases them (PLAN.md section 5, step 2).
// A declined payment returns the failed order together with ErrDeclined.
func (s *Service) Pay(ctx context.Context, userID bson.ObjectID, orderHex string, card Card) (View, error) {
	id, err := bson.ObjectIDFromHex(orderHex)
	if err != nil {
		return View{}, ErrNotFound
	}
	// Owner check first, so a stranger gets 404 whatever card they send.
	if _, err := s.owned(ctx, id, userID); err != nil {
		return View{}, err
	}
	now := s.now()
	number, err := checkCard(card, now)
	if err != nil {
		return View{}, err
	}
	payment := Payment{Last4: number[len(number)-4:], Outcome: simulate(number)}

	var o Order
	err = s.repo.Transaction(ctx, func(ctx context.Context) error {
		var err error
		if o, err = s.owned(ctx, id, userID); err != nil {
			return err
		}
		if o.Status != StatusPendingPayment {
			return ErrNotPayable
		}
		if !now.Before(o.ReservedUntil) {
			return errHoldLost
		}
		o.Payment, o.UpdatedAt = &payment, now

		if payment.Outcome == OutcomeDeclined {
			if err := s.keys.Release(ctx, o.ID); err != nil {
				return err
			}
			o.Status = StatusFailed
			if err := s.repo.Save(ctx, o); err != nil {
				return err
			}
			return s.refreshStock(ctx, o.Items, now)
		}

		// Sell only touches keys this order still holds. Fewer than expected means some
		// expired and another order took them; returning an error rolls back the ones sold here.
		sold, err := s.keys.Sell(ctx, o.ID, now)
		if err != nil {
			return err
		}
		if sold < keyCount(o) {
			return errHoldLost
		}
		codes, err := s.keys.Codes(ctx, o.ID)
		if err != nil {
			return err
		}
		paidOffers := make([]bson.ObjectID, len(o.Items))
		for i := range o.Items {
			it := &o.Items[i]
			it.Keys = make([]KeyCode, 0, len(it.KeyIDs))
			for _, keyID := range it.KeyIDs {
				code, ok := codes[keyID]
				if !ok {
					return fmt.Errorf("order %s: key %s not sold", o.ID.Hex(), keyID.Hex())
				}
				it.Keys = append(it.Keys, KeyCode{Code: code})
			}
			if err := s.catalog.AddSold(ctx, it.ProductID, it.Quantity); err != nil {
				return err
			}
			paidOffers[i] = it.OfferID
		}
		o.Status, o.PaidAt = StatusPaid, &now
		if err := s.repo.Save(ctx, o); err != nil {
			return err
		}
		if err := s.carts.RemoveOffers(ctx, userID, paidOffers, now); err != nil {
			return err
		}
		return s.refreshStock(ctx, o.Items, now)
	})
	if errors.Is(err, errHoldLost) {
		if err := s.expire(ctx, id); err != nil {
			return View{}, err
		}
		return View{}, ErrNotPayable
	}
	if err != nil {
		return View{}, err
	}
	if o.Status == StatusFailed {
		return toView(o, now, true), ErrDeclined
	}
	return toView(o, now, true), nil
}

func (s *Service) Get(ctx context.Context, userID bson.ObjectID, orderHex string) (View, error) {
	id, err := bson.ObjectIDFromHex(orderHex)
	if err != nil {
		return View{}, ErrNotFound
	}
	o, err := s.owned(ctx, id, userID)
	if err != nil {
		return View{}, err
	}
	return toView(o, s.now(), true), nil
}

// List returns the user's orders, newest first, without key codes.
func (s *Service) List(ctx context.Context, userID bson.ObjectID) ([]View, error) {
	list, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	views := make([]View, len(list))
	for i, o := range list {
		views[i] = toView(o, now, false)
	}
	return views, nil
}

// expire marks a pending order whose hold ran out as expired, frees the keys it still owns and
// recounts stock. Used by Pay (late payment) and by the cleanup job. Nothing depends on it for
// correctness: expired keys are claimable anyway; this only tidies the order and the cached stock.
func (s *Service) expire(ctx context.Context, id bson.ObjectID) error {
	now := s.now()
	return s.repo.Transaction(ctx, func(ctx context.Context) error {
		o, err := s.repo.ByID(ctx, id)
		if err != nil {
			return err
		}
		if o.Status != StatusPendingPayment { // paid, declined or expired meanwhile
			return nil
		}
		// Release filters on this order's id, so a key another order has claimed since stays theirs.
		if err := s.keys.Release(ctx, o.ID); err != nil {
			return err
		}
		o.Status, o.UpdatedAt = StatusExpired, now
		if err := s.repo.Save(ctx, o); err != nil {
			return err
		}
		return s.refreshStock(ctx, o.Items, now)
	})
}

// expirePending marks the user's pending orders expired and frees the keys they still own.
// It runs inside Create's transaction and returns those orders' items, so Create can recount
// their stock. If the user pays an old order in another tab at the same moment, the two
// transactions touch the same order and keys, so Mongo makes one of them retry: either the
// payment wins and the order is no longer pending here, or it is expired and the payment gets 409.
func (s *Service) expirePending(ctx context.Context, userID bson.ObjectID, now time.Time) ([]Item, error) {
	pending, err := s.repo.Pending(ctx, userID)
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, p := range pending {
		if err := s.keys.Release(ctx, p.ID); err != nil {
			return nil, err
		}
		p.Status, p.UpdatedAt = StatusExpired, now
		if err := s.repo.Save(ctx, p); err != nil {
			return nil, err
		}
		items = append(items, p.Items...)
	}
	return items, nil
}

// owned returns the order only if it belongs to the user. Someone else's order is ErrNotFound,
// not forbidden, so order ids can't be probed (API.md).
func (s *Service) owned(ctx context.Context, id, userID bson.ObjectID) (Order, error) {
	o, err := s.repo.ByID(ctx, id)
	if err != nil {
		return Order{}, err
	}
	if o.UserID != userID {
		return Order{}, ErrNotFound
	}
	return o, nil
}

// refreshStock recounts the cached offers.stock from the keys collection and rebuilds the product
// summary, for every product in the order. It counts all of a product's offers, not only the
// ones in the order, because the summary (cheapest in-stock offer) depends on all of them.
// Recount, never $inc: expired holds come back on their own, so +1/-1 bookkeeping would drift.
func (s *Service) refreshStock(ctx context.Context, items []Item, now time.Time) error {
	done := map[bson.ObjectID]bool{}
	for _, it := range items {
		if done[it.ProductID] {
			continue
		}
		done[it.ProductID] = true

		p, offers, err := s.catalog.StockInputs(ctx, it.ProductID)
		if errors.Is(err, products.ErrNotFound) { // product deleted: nothing to keep in sync
			continue
		}
		if err != nil {
			return err
		}
		ids := make([]bson.ObjectID, len(offers))
		for i, of := range offers {
			ids[i] = of.ID
		}
		live, err := s.keys.CountClaimable(ctx, ids, now)
		if err != nil {
			return err
		}
		for i := range offers {
			offers[i].Stock = live[offers[i].ID]
		}
		products.Summarize(&p, offers)
		if err := s.catalog.SaveStock(ctx, p, offers, now); err != nil {
			return err
		}
	}
	return nil
}

// newItems copies each cart line's offer into an order item, with the price as it is right now.
// Cashback is worked out the same way as in the cart: per line, rounded down.
func newItems(lines []cart.Item, offers map[bson.ObjectID]cart.OfferInfo) ([]Item, int64, int64) {
	items := make([]Item, len(lines))
	var total, cashback int64
	for i, l := range lines {
		of := offers[l.OfferID]
		items[i] = Item{
			OfferID: of.OfferID, ProductID: of.ProductID, SellerID: of.SellerID,
			Name: of.Name, Slug: of.Slug, Platform: of.Platform, Region: of.Region, SellerName: of.SellerName,
			UnitPriceCents: of.PriceCents, Quantity: l.Quantity,
		}
		lineTotal := of.PriceCents * int64(l.Quantity)
		total += lineTotal
		cashback += lineTotal * int64(of.CashbackPct) / 100
	}
	return items, total, cashback
}

func keyCount(o Order) int {
	n := 0
	for _, it := range o.Items {
		n += len(it.KeyIDs)
	}
	return n
}

// numberChars leaves out 0/O and 1/I, so a number read aloud or copied by hand can't be mistaken.
const numberChars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// newNumber makes a short order number like "ORD-7K2M9Q". 32^6 is about a billion, so a clash
// is very unlikely at this scale; if one ever happens, the unique index refuses the insert and
// the user gets a 500 and can try again.
func newNumber() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error
	for i := range b {
		b[i] = numberChars[int(b[i])%len(numberChars)] // 256 is a multiple of 32, so no bias
	}
	return "ORD-" + string(b)
}

// toView builds the API order. A pending order past its hold is shown as expired even if the
// cleanup job hasn't updated it yet (API.md). Codes are only shown on a paid order's detail.
func toView(o Order, now time.Time, withKeys bool) View {
	status := o.Status
	if status == StatusPendingPayment && !now.Before(o.ReservedUntil) {
		status = StatusExpired
	}
	v := View{
		ID: o.ID.Hex(), Number: o.Number, Status: status,
		Items:      make([]ItemView, len(o.Items)),
		TotalCents: o.TotalCents, CashbackCents: o.CashbackCents,
		CreatedAt: o.CreatedAt, ReservedUntil: o.ReservedUntil, PaidAt: o.PaidAt, Payment: o.Payment,
	}
	for i, it := range o.Items {
		v.Items[i] = ItemView{
			OfferID: it.OfferID.Hex(), ProductID: it.ProductID.Hex(),
			Name: it.Name, Slug: it.Slug, Platform: it.Platform, Region: it.Region, SellerName: it.SellerName,
			UnitPriceCents: it.UnitPriceCents, Quantity: it.Quantity,
		}
		if withKeys && status == StatusPaid {
			v.Items[i].Keys = it.Keys
		}
	}
	return v
}
