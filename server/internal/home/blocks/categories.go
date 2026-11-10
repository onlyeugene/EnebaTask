package blocks

import (
	"context"
	"fmt"
	"net/url"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/home"
	"eugene/internal/products"
)

// Categories returns the category tiles from config, each with a link to the store filter
// and its product count.
type Categories struct {
	catalog Catalog
}

func (c Categories) Resolve(ctx context.Context, b home.Block) (any, error) {
	var cfg home.CategoryConfig
	if err := bson.Unmarshal(b.Config, &cfg); err != nil {
		return nil, fmt.Errorf("category config: %w", err)
	}

	// The store's filter panel already counts in-stock products per type, platform and genre
	// in one aggregation, so the tile counts match what the store shows after the click.
	f, err := c.catalog.Filters(ctx)
	if err != nil {
		return nil, err
	}
	counts := map[string][]products.FilterOption{
		"type": f.Types, "platform": f.Platforms, "region": f.Regions, "genre": f.Genres,
	}

	out := make([]home.Category, 0, len(cfg.Categories))
	for _, link := range cfg.Categories {
		out = append(out, home.Category{
			Slug: link.Slug, Label: link.Label, Icon: link.Icon,
			Href:  "/store?" + url.Values{link.Param: {link.Value}}.Encode(),
			Count: countOf(counts[link.Param], link.Value),
		})
	}
	return map[string]any{"categories": out}, nil
}

// countOf is 0 for a value the filter list doesn't have (no in-stock products).
func countOf(options []products.FilterOption, value string) int {
	for _, o := range options {
		if o.Value == value {
			return o.Count
		}
	}
	return 0
}
