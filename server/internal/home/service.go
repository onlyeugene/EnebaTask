package home

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// blockTimeout bounds one block. A slow block becomes an error for that block only,
// instead of holding up the page.
const blockTimeout = 2 * time.Second

type Service struct {
	repo     Repository
	registry map[string]BlockService // block type → the service that builds its data
}

// NewService takes the registry from outside (main.go builds it with blocks.Registry),
// because the blocks package imports this one; building it here would be an import cycle.
func NewService(repo Repository, registry map[string]BlockService) *Service {
	return &Service{repo: repo, registry: registry}
}

// Layout lists active blocks in order: id, type and title only.
func (s *Service) Layout(ctx context.Context) ([]LayoutEntry, error) {
	blocks, err := s.repo.Active(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]LayoutEntry, 0, len(blocks)) // [] not null when there are no blocks
	for _, b := range blocks {
		out = append(out, toEntry(b))
	}
	return out, nil
}

// Resolve loads one block and asks its type's service for the data.
// Returns ErrNotFound for a bad or unknown id; any other error means the block is unavailable.
func (s *Service) Resolve(ctx context.Context, id string) (Resolved, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Resolved{}, ErrNotFound
	}

	ctx, cancel := context.WithTimeout(ctx, blockTimeout)
	defer cancel()

	b, err := s.repo.ByID(ctx, oid)
	if err != nil {
		return Resolved{}, err
	}
	svc, ok := s.registry[b.Type]
	if !ok {
		return Resolved{}, fmt.Errorf("block %s: no service for type %q", id, b.Type)
	}
	data, err := svc.Resolve(ctx, b)
	if err != nil {
		return Resolved{}, fmt.Errorf("block %s (%s): %w", id, b.Type, err)
	}
	return Resolved{LayoutEntry: toEntry(b), Data: data}, nil
}

func toEntry(b Block) LayoutEntry {
	return LayoutEntry{ID: b.ID.Hex(), Type: b.Type, Title: b.Title}
}
