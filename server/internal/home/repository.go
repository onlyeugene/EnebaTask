package home

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrNotFound = errors.New("block not found")

// Repository reads the homepage_blocks collection.
type Repository interface {
	// Active returns active blocks sorted by order, without their config.
	Active(ctx context.Context) ([]Block, error)
	// ByID returns one active block with its config, or ErrNotFound.
	ByID(ctx context.Context, id bson.ObjectID) (Block, error)
}

// BlockService builds the data for one block type. The registry in blocks/ maps each type to one.
type BlockService interface {
	Resolve(ctx context.Context, block Block) (any, error)
}
