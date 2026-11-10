// Command seed fills a database with sellers, products, offers, keys and homepage blocks.
//
//	go run ./cmd/seed            # refuses if products already exist
//	go run ./cmd/seed --reset    # clears catalog, carts and orders first
//
// After inserting, it creates (or updates) the Atlas Search index from atlas/products_search.json
// and waits until it is READY.
//
// It reads MONGODB_URI and MONGODB_DB from server/.env or the environment.
// MONGODB_DB=eneba_prod also needs --allow-prod.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"eugene/internal/platform/db"
)

func main() {
	reset := flag.Bool("reset", false, "delete catalog, carts and orders before seeding")
	allowProd := flag.Bool("allow-prod", false, "allow seeding eneba_prod")
	flag.Parse()

	if err := run(*reset, *allowProd); err != nil {
		slog.Error("seed", "err", err)
		os.Exit(1)
	}
}

func run(reset, allowProd bool) error {
	// Same rule as the API: a missing .env is fine, a broken one is not.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf(".env: %w", err)
	}
	uri, dbName := os.Getenv("MONGODB_URI"), os.Getenv("MONGODB_DB")
	if uri == "" || dbName == "" {
		return errors.New("MONGODB_URI and MONGODB_DB are required")
	}
	if dbName == "eneba_prod" && !allowProd {
		return errors.New("refusing to seed eneba_prod without --allow-prod")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client, database, err := db.Connect(ctx, uri, dbName)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	if reset {
		if err := clearCatalog(ctx, database); err != nil {
			return err
		}
	} else if n, err := database.Collection("products").CountDocuments(ctx, bson.D{}); err != nil {
		return err
	} else if n > 0 {
		return fmt.Errorf("%s already has %d products; run with --reset to replace them", dbName, n)
	}

	if err := db.EnsureIndexes(ctx, database); err != nil {
		return err
	}

	d, err := build(time.Now().UTC())
	if err != nil {
		return err
	}
	inserts := []struct {
		coll string
		docs any
	}{
		{"sellers", d.sellers},
		{"products", d.products},
		{"offers", d.offers},
		{"keys", d.keys},
		{"homepage_blocks", d.blocks},
	}
	for _, in := range inserts {
		if _, err := database.Collection(in.coll).InsertMany(ctx, in.docs); err != nil {
			return fmt.Errorf("insert %s: %w", in.coll, err)
		}
	}

	slog.Info("seeded", "db", dbName, "sellers", len(d.sellers), "products", len(d.products),
		"offers", len(d.offers), "keys", len(d.keys), "blocks", len(d.blocks))

	// Its own timeout: Atlas builds the index in the background, which on M0 can take a minute or two.
	slog.Info("creating Atlas Search index and waiting until READY", "index", db.SearchIndexName)
	searchCtx, cancelSearch := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelSearch()
	if err := db.EnsureSearchIndex(searchCtx, database); err != nil {
		return err
	}
	slog.Info("search index ready", "index", db.SearchIndexName)
	return nil
}

// clearCatalog empties the catalog and everything that points at it. Carts and orders go too,
// because they hold offer and key ids that would no longer exist. Users and tokens stay,
// so a reset doesn't log anyone out.
//
// DeleteMany, not Drop: dropping a collection also drops its indexes, including the
// Atlas Search index on products, which takes a while to rebuild.
func clearCatalog(ctx context.Context, database *mongo.Database) error {
	for _, coll := range []string{"keys", "offers", "products", "sellers", "homepage_blocks", "carts", "orders"} {
		if _, err := database.Collection(coll).DeleteMany(ctx, bson.D{}); err != nil {
			return fmt.Errorf("clear %s: %w", coll, err)
		}
	}
	return nil
}
