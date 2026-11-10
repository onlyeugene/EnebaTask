// Package blocks holds one service per homepage block type.
package blocks

import (
	"context"

	"eugene/internal/home"
	"eugene/internal/products"
)

// Catalog is what the product blocks need from the products feature. *products.Service implements it,
// so home rails return the same cards, sort and in-stock rules as the store.
type Catalog interface {
	List(ctx context.Context, q products.ListQuery) (products.ListResult, error)
	Filters(ctx context.Context) (products.Filters, error)
}

// Registry maps each block type to its service. A new block type needs one entry here
// (and, on the client, one component). A new block of an existing type needs only a db doc.
func Registry(catalog Catalog) map[string]home.BlockService {
	return map[string]home.BlockService{
		home.TypeHero:        Hero{},
		home.TypeCategoryRow: Categories{catalog},
		// The four rails differ only in sort, type filter and "see all" link.
		home.TypeBestsellers: Rail{catalog, products.SortBestselling, nil, "/store?sort=bestselling"},
		home.TypeDeals:       Rail{catalog, products.SortDiscount, nil, "/store?sort=discount"},
		home.TypeNewReleases: Rail{catalog, products.SortNewest, nil, "/store?sort=newest"},
		home.TypeGiftCards:   Rail{catalog, products.SortBestselling, []string{"gift_card"}, "/store?type=gift_card"},
	}
}
