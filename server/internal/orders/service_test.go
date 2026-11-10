package orders

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// memRepo is an in-memory orders store for the read paths. Transaction just runs fn: the
// paths that need a real transaction are covered by claim_integration_test.go.
type memRepo struct {
	orders map[bson.ObjectID]Order
}

func (m *memRepo) Transaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
func (m *memRepo) Insert(_ context.Context, o Order) error { m.orders[o.ID] = o; return nil }
func (m *memRepo) Save(_ context.Context, o Order) error   { m.orders[o.ID] = o; return nil }
func (m *memRepo) ByID(_ context.Context, id bson.ObjectID) (Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return Order{}, ErrNotFound
	}
	return o, nil
}
func (m *memRepo) List(_ context.Context, userID bson.ObjectID) ([]Order, error) {
	var out []Order
	for _, o := range m.orders {
		if o.UserID == userID {
			out = append(out, o)
		}
	}
	return out, nil
}
func (m *memRepo) Pending(context.Context, bson.ObjectID) ([]Order, error)        { return nil, nil }
func (m *memRepo) Stale(context.Context, time.Time, int) ([]bson.ObjectID, error) { return nil, nil }

var now = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

// newTestService has no keys, catalog or carts: the read paths and the checks that run
// before the transaction don't use them.
func newTestService(orders ...Order) *Service {
	repo := &memRepo{orders: map[bson.ObjectID]Order{}}
	for _, o := range orders {
		repo.orders[o.ID] = o
	}
	return NewService(repo, nil, nil, nil, 10*time.Minute, func() time.Time { return now })
}

func testOrder(user bson.ObjectID, status string, reservedUntil time.Time) Order {
	return Order{
		ID: bson.NewObjectID(), Number: "ORD-TEST22", UserID: user, Status: status, ReservedUntil: reservedUntil,
		Items: []Item{{OfferID: bson.NewObjectID(), Quantity: 1, Keys: []KeyCode{{Code: "AAAAA-BBBBB"}}}},
	}
}

func TestCheckCard(t *testing.T) {
	good := Card{CardNumber: "4242 4242 4242 4242", Expiry: "12/30", CVC: "123", Name: "Eugene"}
	tests := []struct {
		name  string
		edit  func(c *Card)
		field string // "" = valid
	}{
		{"valid with spaces", func(c *Card) {}, ""},
		{"valid with dashes", func(c *Card) { c.CardNumber = "4242-4242-4242-4242" }, ""},
		{"expires this month", func(c *Card) { c.Expiry = "11/26" }, ""},
		{"4-digit cvc", func(c *Card) { c.CVC = "1234" }, ""},
		{"too short", func(c *Card) { c.CardNumber = "4242 4242" }, "cardNumber"},
		{"letters", func(c *Card) { c.CardNumber = "4242 4242 4242 424X" }, "cardNumber"},
		{"expired last month", func(c *Card) { c.Expiry = "10/26" }, "expiry"},
		{"month 13", func(c *Card) { c.Expiry = "13/30" }, "expiry"},
		{"no slash", func(c *Card) { c.Expiry = "1230" }, "expiry"},
		{"short cvc", func(c *Card) { c.CVC = "12" }, "cvc"},
		{"no name", func(c *Card) { c.Name = "  " }, "name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := good
			tt.edit(&c)
			_, err := checkCard(c, now)
			var v *ValidationError
			if tt.field == "" {
				if err != nil {
					t.Fatalf("want valid, got %v", err)
				}
				return
			}
			if !errors.As(err, &v) || v.Fields[tt.field] == "" {
				t.Fatalf("want error on %q, got %v", tt.field, err)
			}
		})
	}
}

func TestSimulateDeclinesOnlyTheMagicNumber(t *testing.T) {
	number, err := checkCard(Card{CardNumber: "4000 0000 0000 0002", Expiry: "12/30", CVC: "123", Name: "E"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if got := simulate(number); got != OutcomeDeclined {
		t.Errorf("decline card: got %q", got)
	}
	if got := simulate("4242424242424242"); got != OutcomeSuccess {
		t.Errorf("good card: got %q", got)
	}
}

// The cleanup job may not have run (Render sleeps), so reads must show a stale order as expired.
func TestReadShowsStalePendingOrderAsExpired(t *testing.T) {
	user := bson.NewObjectID()
	stale := testOrder(user, StatusPendingPayment, now.Add(-time.Second))
	live := testOrder(user, StatusPendingPayment, now.Add(time.Minute))
	svc := newTestService(stale, live)

	got, err := svc.Get(context.Background(), user, stale.ID.Hex())
	if err != nil || got.Status != StatusExpired {
		t.Fatalf("stale order: status %q, err %v; want expired", got.Status, err)
	}
	got, err = svc.Get(context.Background(), user, live.ID.Hex())
	if err != nil || got.Status != StatusPendingPayment {
		t.Fatalf("live order: status %q, err %v; want pending_payment", got.Status, err)
	}
	list, err := svc.List(context.Background(), user)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range list {
		if v.ID == stale.ID.Hex() && v.Status != StatusExpired {
			t.Errorf("list: stale order is %q, want expired", v.Status)
		}
	}
}

func TestOtherUsersOrderIsNotFound(t *testing.T) {
	owner, stranger := bson.NewObjectID(), bson.NewObjectID()
	o := testOrder(owner, StatusPendingPayment, now.Add(time.Minute))
	svc := newTestService(o)
	good := Card{CardNumber: "4242424242424242", Expiry: "12/30", CVC: "123", Name: "E"}

	if _, err := svc.Get(context.Background(), stranger, o.ID.Hex()); !errors.Is(err, ErrNotFound) {
		t.Errorf("get: want ErrNotFound, got %v", err)
	}
	if _, err := svc.Pay(context.Background(), stranger, o.ID.Hex(), good); !errors.Is(err, ErrNotFound) {
		t.Errorf("pay: want ErrNotFound, got %v", err)
	}
	if _, err := svc.Get(context.Background(), owner, "not-an-id"); !errors.Is(err, ErrNotFound) {
		t.Errorf("bad id: want ErrNotFound, got %v", err)
	}
}

func TestCodesOnlyOnPaidDetail(t *testing.T) {
	user := bson.NewObjectID()
	paid := testOrder(user, StatusPaid, now.Add(-time.Hour))
	failed := testOrder(user, StatusFailed, now.Add(time.Minute))
	svc := newTestService(paid, failed)

	got, _ := svc.Get(context.Background(), user, paid.ID.Hex())
	if len(got.Items[0].Keys) != 1 {
		t.Errorf("paid detail: want the code, got %v", got.Items[0].Keys)
	}
	got, _ = svc.Get(context.Background(), user, failed.ID.Hex())
	if len(got.Items[0].Keys) != 0 {
		t.Errorf("failed detail: want no codes, got %v", got.Items[0].Keys)
	}
	list, _ := svc.List(context.Background(), user)
	for _, v := range list {
		if len(v.Items[0].Keys) != 0 {
			t.Errorf("list: want no codes, got %v", v.Items[0].Keys)
		}
	}
}

func TestPayRefusesOrderThatIsNotPending(t *testing.T) {
	user := bson.NewObjectID()
	paid := testOrder(user, StatusPaid, now.Add(time.Minute))
	svc := newTestService(paid)
	good := Card{CardNumber: "4242424242424242", Expiry: "12/30", CVC: "123", Name: "E"}
	if _, err := svc.Pay(context.Background(), user, paid.ID.Hex(), good); !errors.Is(err, ErrNotPayable) {
		t.Errorf("want ErrNotPayable, got %v", err)
	}
}
