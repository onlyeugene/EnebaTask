package products

import (
	"context"
	"slices"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	repo  Repository
	stock StockCounter
	now   func() time.Time
}

func NewService(repo Repository, stock StockCounter, now func() time.Time) *Service {
	return &Service{repo: repo, stock: stock, now: now}
}

func (s *Service) List(ctx context.Context, q ListQuery) (ListResult, error) {
	found, total, err := s.repo.List(ctx, q)
	if err != nil {
		return ListResult{}, err
	}
	items := make([]ProductCard, 0, len(found)) // [] not null in JSON when nothing matches
	for _, p := range found {
		items = append(items, ToCard(p))
	}
	return ListResult{Items: items, Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

// Filters lists every filter value that has at least one in-stock product, with its count.
func (s *Service) Filters(ctx context.Context) (Filters, error) {
	f, err := s.repo.Facets(ctx)
	if err != nil {
		return Filters{}, err
	}
	var out Filters
	out.Types = fixedOptions(Types, f.Types)
	out.Platforms = fixedOptions(Platforms, f.Platforms)
	out.Regions = fixedOptions(Regions, f.Regions)

	out.Genres = []FilterOption{}
	for value, count := range f.Genres {
		out.Genres = append(out.Genres, FilterOption{Value: value, Label: genreLabel(value), Count: count})
	}
	// Sort on the lowercase value, not the label: "RPG" would sort before "Racing".
	slices.SortFunc(out.Genres, func(a, b FilterOption) int { return strings.Compare(a.Value, b.Value) })

	out.Price.MinCents, out.Price.MaxCents = f.MinPriceCents, f.MaxPriceCents
	return out, nil
}

// fixedOptions keeps the order of the allowed list and skips values with no products.
func fixedOptions(allowed []Option, counts map[string]int) []FilterOption {
	out := []FilterOption{}
	for _, o := range allowed {
		if n := counts[o.Value]; n > 0 {
			out = append(out, FilterOption{Value: o.Value, Label: o.Label, Count: n})
		}
	}
	return out
}

// genreLabel turns "rpg" into "RPG" and "adventure" into "Adventure".
// Genres are free-form in the seed data, so there is no fixed label list.
func genreLabel(g string) string {
	if g == "rpg" {
		return "RPG"
	}
	if g == "" {
		return g
	}
	return strings.ToUpper(g[:1]) + g[1:]
}

// Detail returns the product with its offers. Stock comes live from the keys collection, not the
// cached offers.stock, because the cache can lag behind an expired reservation (PLAN.md section 5).
// The price summary is rebuilt from that live stock so it always matches the offers shown below it.
func (s *Service) Detail(ctx context.Context, slug string) (ProductDetail, []OfferView, error) {
	p, err := s.repo.BySlug(ctx, slug)
	if err != nil {
		return ProductDetail{}, nil, err
	}
	withSeller, err := s.repo.Offers(ctx, p.ID)
	if err != nil {
		return ProductDetail{}, nil, err
	}

	ids := make([]bson.ObjectID, len(withSeller))
	for i, o := range withSeller {
		ids[i] = o.ID
	}
	live, err := s.stock.CountClaimable(ctx, ids, s.now())
	if err != nil {
		return ProductDetail{}, nil, err
	}
	offers := make([]Offer, len(withSeller))
	for i := range withSeller {
		withSeller[i].Stock = live[withSeller[i].ID]
		offers[i] = withSeller[i].Offer
	}
	Summarize(&p, offers)

	// The repo already sorts by price; a stable sort on "in stock" keeps that order inside each group.
	slices.SortStableFunc(withSeller, func(a, b OfferWithSeller) int {
		return boolRank(b.Stock > 0) - boolRank(a.Stock > 0)
	})
	views := make([]OfferView, 0, len(withSeller))
	for _, o := range withSeller {
		views = append(views, toOfferView(o))
	}
	return toDetail(p), views, nil
}

func boolRank(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (s *Service) CoverSVG(ctx context.Context, slug string) (string, error) {
	return s.repo.CoverSVG(ctx, slug)
}

// Summarize sets minPriceCents, discountPct, offerCount and inStock from the offers.
// Used by the seed, by the product page (with live stock) and by the purchase transaction,
// so all three always agree.
func Summarize(p *Product, offers []Offer) {
	p.MinPriceCents, p.OfferCount, p.DiscountPct, p.InStock = nil, 0, 0, false
	for _, o := range offers {
		if !o.Active || o.Stock == 0 {
			continue
		}
		p.OfferCount++
		if p.MinPriceCents == nil || o.PriceCents < *p.MinPriceCents {
			price := o.PriceCents
			p.MinPriceCents = &price
		}
	}
	if p.MinPriceCents == nil {
		return
	}
	p.InStock = true
	// Only a price below MSRP counts as a discount; gift cards above face value show 0%.
	if saved := p.OriginalPriceCents - *p.MinPriceCents; saved > 0 {
		p.DiscountPct = int(saved * 100 / p.OriginalPriceCents)
	}
}

// ToCard builds the API card. Exported for the search feature, so search results look exactly like store results.
func ToCard(p Product) ProductCard {
	return ProductCard{
		ID: p.ID.Hex(), Slug: p.Slug, Name: p.Name, Type: p.Type, Platform: p.Platform, Region: p.Region,
		CoverURL:           "/api/products/" + p.Slug + "/cover.svg",
		GiftCardValue:      p.GiftCardValue,
		OriginalPriceCents: p.OriginalPriceCents,
		MinPriceCents:      p.MinPriceCents,
		DiscountPct:        p.DiscountPct,
		CashbackPct:        p.CashbackPct,
		OfferCount:         p.OfferCount,
	}
}

func toDetail(p Product) ProductDetail {
	genres := p.Genres
	if genres == nil {
		genres = []string{}
	}
	return ProductDetail{
		ProductCard: ToCard(p),
		Description: p.Description, Genres: genres, Developer: p.Developer, Publisher: p.Publisher,
		ReleaseDate: p.ReleaseDate.Format(time.DateOnly),
		Activation:  p.Activation,
	}
}

func toOfferView(o OfferWithSeller) OfferView {
	return OfferView{
		ID: o.ID.Hex(), ProductID: o.ProductID.Hex(), PriceCents: o.PriceCents, Stock: o.Stock,
		Seller: SellerView{
			ID: o.Seller.ID.Hex(), Slug: o.Seller.Slug, Name: o.Seller.Name,
			Rating: o.Seller.Rating, SalesCount: o.Seller.SalesCount,
		},
	}
}
