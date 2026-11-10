package search

import (
	"context"

	"eugene/internal/products"
)

// Repository runs Atlas Search pipelines on the products collection.
type Repository interface {
	// Search returns one page of products matching q.Q plus the filters, and the total that match.
	Search(ctx context.Context, q products.ListQuery) ([]products.Product, int64, error)
	// Autocomplete returns up to limit active products whose name starts like text, best match first.
	Autocomplete(ctx context.Context, text string, limit int) ([]products.Product, error)
}
