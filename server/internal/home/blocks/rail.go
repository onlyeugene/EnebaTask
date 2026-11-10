package blocks

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/home"
	"eugene/internal/products"
)

const defaultRailLimit = 10

// Rail is a row of product cards: best sellers, deals, new releases, gift cards.
// It runs the same query as the store page its "see all" link opens, cut to the first few.
type Rail struct {
	catalog Catalog
	sort    string
	types   []string // e.g. ["gift_card"]; nil means every type
	seeAll  string
}

func (rl Rail) Resolve(ctx context.Context, b home.Block) (any, error) {
	var cfg home.RailConfig
	if err := bson.Unmarshal(b.Config, &cfg); err != nil {
		return nil, fmt.Errorf("rail config: %w", err)
	}
	limit := cfg.Limit
	if limit < 1 || limit > products.MaxPageSize {
		limit = defaultRailLimit
	}

	res, err := rl.catalog.List(ctx, products.ListQuery{
		Types: rl.types, InStockOnly: true, Sort: rl.sort, Page: 1, PageSize: limit,
	})
	if err != nil {
		return nil, err
	}

	cards := res.Items
	if cfg.MinDiscountPct > 0 {
		// Only the deals rail sets this, and it sorts by discount (highest first), so dropping
		// the cards below the minimum gives the same result as filtering in the query.
		cards = make([]products.ProductCard, 0, len(res.Items))
		for _, c := range res.Items {
			if c.DiscountPct >= cfg.MinDiscountPct {
				cards = append(cards, c)
			}
		}
	}
	return map[string]any{"products": cards, "seeAllHref": rl.seeAll}, nil
}
