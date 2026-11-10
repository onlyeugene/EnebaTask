package blocks

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/home"
)

// Hero returns the banners stored in the block's config. No other data source.
type Hero struct{}

func (Hero) Resolve(_ context.Context, b home.Block) (any, error) {
	var cfg home.HeroConfig
	if err := bson.Unmarshal(b.Config, &cfg); err != nil {
		return nil, fmt.Errorf("hero config: %w", err)
	}
	if cfg.Tiles == nil {
		cfg.Tiles = []home.Banner{} // [] not null in JSON
	}
	return cfg, nil
}
