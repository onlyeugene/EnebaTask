package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"eugene/atlas"
)

// SearchIndexName is the Atlas Search index on products (PLAN.md section 2).
const SearchIndexName = "products_search"

type searchIndexFile struct {
	Name       string `bson:"name"`
	Definition bson.M `bson:"definition"`
}

type searchIndexInfo struct {
	Name      string `bson:"name"`
	Status    string `bson:"status"` // PENDING, BUILDING, READY, FAILED, ...
	Queryable bool   `bson:"queryable"`
}

// EnsureSearchIndex creates the products search index from atlas/products_search.json,
// or updates it if it already exists, then waits until Atlas reports it READY.
// Atlas builds search indexes in the background, so without the wait the first searches
// after a seed would quietly return nothing.
func EnsureSearchIndex(ctx context.Context, database *mongo.Database) error {
	var file searchIndexFile
	if err := bson.UnmarshalExtJSON(atlas.ProductsSearch, false, &file); err != nil {
		return fmt.Errorf("parse products_search.json: %w", err)
	}

	view := database.Collection("products").SearchIndexes()
	existing, err := findSearchIndex(ctx, database, file.Name)
	if err != nil {
		return err
	}
	if existing == nil {
		model := mongo.SearchIndexModel{Definition: file.Definition, Options: options.SearchIndexes().SetName(file.Name)}
		if _, err := view.CreateOne(ctx, model); err != nil {
			return fmt.Errorf("create search index: %w", err)
		}
	} else if err := view.UpdateOne(ctx, file.Name, file.Definition); err != nil {
		return fmt.Errorf("update search index: %w", err)
	}

	// Poll every 2s. The first check waits too: right after an update Atlas can still
	// report the old build as READY for a moment.
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("search index %s not ready in time: %w", file.Name, ctx.Err())
		case <-time.After(2 * time.Second):
		}
		info, err := findSearchIndex(ctx, database, file.Name)
		if err != nil {
			return err
		}
		if info != nil && info.Status == "FAILED" {
			return errors.New("search index build FAILED; check it in the Atlas UI")
		}
		if info != nil && info.Status == "READY" {
			return nil
		}
	}
}

// CheckSearchIndex returns an error if the search index is missing or can't be queried yet.
// The API calls it at startup only to log a warning: $search on a missing index returns
// no results and no error, which would look like "nothing matches".
func CheckSearchIndex(ctx context.Context, database *mongo.Database) error {
	info, err := findSearchIndex(ctx, database, SearchIndexName)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("search index %s not found; run the seed or create it from server/atlas/products_search.json", SearchIndexName)
	}
	if !info.Queryable {
		return fmt.Errorf("search index %s is %s, not queryable yet", SearchIndexName, info.Status)
	}
	return nil
}

// findSearchIndex returns nil, nil when no index has that name.
func findSearchIndex(ctx context.Context, database *mongo.Database, name string) (*searchIndexInfo, error) {
	cur, err := database.Collection("products").SearchIndexes().List(ctx, options.SearchIndexes().SetName(name))
	if err != nil {
		return nil, fmt.Errorf("list search indexes: %w", err)
	}
	var found []searchIndexInfo
	if err := cur.All(ctx, &found); err != nil {
		return nil, fmt.Errorf("decode search indexes: %w", err)
	}
	if len(found) == 0 {
		return nil, nil
	}
	return &found[0], nil
}
