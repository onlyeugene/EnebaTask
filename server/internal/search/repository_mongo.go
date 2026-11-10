package search

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"eugene/internal/platform/db"
	"eugene/internal/products"
)

type MongoRepo struct {
	products *mongo.Collection
}

func NewMongoRepo(database *mongo.Database) *MongoRepo {
	return &MongoRepo{products: database.Collection("products")}
}

// maxWords caps how many words of the query get their own clauses. A product name has
// at most ~8 words, so later words would not change the result much, only slow it down.
const maxWords = 8

// textEdits is how many typos a whole-word match allows, by word length.
// One fixed maxEdits for all words is too loose on short words: with 2 edits, "star" matches
// "steam", so every Steam product turned up for "star". Same steps as Elasticsearch's AUTO.
func textEdits(word string) int {
	switch n := len([]rune(word)); {
	case n <= 3:
		return 0
	case n <= 6:
		return 1
	default:
		return 2
	}
}

// prefixEdits is the typo allowance for a partial-word match. Short prefixes get none:
// "star" with 1 edit would match the prefix "stea" of "Steam".
func prefixEdits(word string) int {
	if len([]rune(word)) >= 5 {
		return 1
	}
	return 0
}

// nameClauses builds, for each word, a whole-word text clause and a partial-word
// autocomplete clause, each with a typo allowance that fits the word's length.
// prefixLength 1 means the first letter must be right, which cuts noise and keeps fuzzy fast.
func nameClauses(query string, withText bool) bson.A {
	words := strings.Fields(query)
	if len(words) > maxWords {
		words = words[:maxWords]
	}
	var clauses bson.A
	for _, w := range words {
		if withText {
			text := bson.M{"query": w, "path": "name", "score": bson.M{"boost": bson.M{"value": 3}}}
			if e := textEdits(w); e > 0 {
				text["fuzzy"] = bson.M{"maxEdits": e, "prefixLength": 1}
			}
			clauses = append(clauses, bson.M{"text": text})
		}
		auto := bson.M{"query": w, "path": "name"}
		if e := prefixEdits(w); e > 0 {
			auto["fuzzy"] = bson.M{"maxEdits": e, "prefixLength": 1}
		}
		clauses = append(clauses, bson.M{"autocomplete": auto})
	}
	return clauses
}

// Search is the store query from PLAN.md section 2. Per word, two clauses score the name:
//   - text + fuzzy catches typos ("skyforje" → "Skyforge").
//   - autocomplete catches partial words ("frost" → "Frostline").
//
// A product matches if any clause matches; the more words it matches, the higher it ranks.
// The text clause is boosted so a whole-word match beats a prefix match.
// Filters go in "filter", so they narrow the results without changing the score.
func (r *MongoRepo) Search(ctx context.Context, q products.ListQuery) ([]products.Product, int64, error) {
	stage := bson.M{
		"index": db.SearchIndexName,
		"compound": bson.M{
			"should":             nameClauses(q.Q, true),
			"minimumShouldMatch": 1,
			"filter":             searchFilters(q),
		},
		"sort":  searchSort(q.Sort),
		"count": bson.M{"type": "total"},
	}

	// $facet gives the page and the total in one round trip. The total comes from
	// $$SEARCH_META, which Atlas fills because of count: total above.
	cur, err := r.products.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$search", Value: stage}},
		{{Key: "$facet", Value: bson.M{
			"items": bson.A{
				bson.M{"$skip": (q.Page - 1) * q.PageSize},
				bson.M{"$limit": q.PageSize},
				bson.M{"$project": bson.M{"cover": 0}},
			},
			"meta": bson.A{
				bson.M{"$replaceWith": "$$SEARCH_META"},
				bson.M{"$limit": 1},
			},
		}}},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("search products: %w", err)
	}

	var rows []struct {
		Items []products.Product `bson:"items"`
		Meta  []struct {
			Count struct {
				Total int64 `bson:"total"`
			} `bson:"count"`
		} `bson:"meta"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		return nil, 0, fmt.Errorf("decode search: %w", err)
	}
	// $facet always returns one document; meta is empty when nothing matched.
	var total int64
	if len(rows[0].Meta) > 0 {
		total = rows[0].Meta[0].Count.Total
	}
	return rows[0].Items, total, nil
}

// searchFilters is the Atlas Search version of products.listFilter: same rules, other syntax.
func searchFilters(q products.ListQuery) bson.A {
	f := bson.A{bson.M{"equals": bson.M{"path": "active", "value": true}}}
	if q.InStockOnly {
		f = append(f, bson.M{"equals": bson.M{"path": "inStock", "value": true}})
	}
	in := func(path string, values []string) {
		if len(values) > 0 {
			f = append(f, bson.M{"in": bson.M{"path": path, "value": values}})
		}
	}
	in("type", q.Types)
	in("platform", q.Platforms)
	in("region", q.Regions)
	in("genres", q.Genres) // matches if any of the product's genres is in the list

	if q.MinPriceCents != nil || q.MaxPriceCents != nil {
		price := bson.M{"path": "minPriceCents"}
		if q.MinPriceCents != nil {
			price["gte"] = *q.MinPriceCents
		}
		if q.MaxPriceCents != nil {
			price["lte"] = *q.MaxPriceCents
		}
		// Sold-out products have minPriceCents null, which no range matches, same as the plain list.
		f = append(f, bson.M{"range": price})
	}
	return f
}

// searchSort mirrors products.listSort: in-stock first, then the chosen order, then _id so
// pages never overlap. "relevance" sorts by the search score; equal scores are common
// (two editions of one game), so it needs the _id tie-break too.
func searchSort(sort string) bson.D {
	var by bson.E
	switch sort {
	case products.SortPriceAsc:
		by = bson.E{Key: "minPriceCents", Value: 1}
	case products.SortPriceDesc:
		by = bson.E{Key: "minPriceCents", Value: -1}
	case products.SortDiscount:
		by = bson.E{Key: "discountPct", Value: -1}
	case products.SortNewest:
		by = bson.E{Key: "createdAt", Value: -1}
	case products.SortBestselling:
		by = bson.E{Key: "soldCount", Value: -1}
	default:
		by = bson.E{Key: "score", Value: bson.M{"$meta": "searchScore"}}
	}
	return bson.D{{Key: "inStock", Value: -1}, by, {Key: "_id", Value: 1}}
}

// Autocomplete uses only the partial-word clauses, active products only.
// Unlike the store, every word must match ("must"): while typing "neon dr" the dropdown
// should narrow to Neon Drift, not list everything with "neon" or "dr" in it.
// Sold-out products stay in: the dropdown shows them as sold out, which beats "no results".
func (r *MongoRepo) Autocomplete(ctx context.Context, text string, limit int) ([]products.Product, error) {
	cur, err := r.products.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$search", Value: bson.M{
			"index": db.SearchIndexName,
			"compound": bson.M{
				"must":   nameClauses(text, false),
				"filter": bson.A{bson.M{"equals": bson.M{"path": "active", "value": true}}},
			},
		}}},
		{{Key: "$limit", Value: limit}},
		{{Key: "$project", Value: bson.M{"slug": 1, "name": 1, "platform": 1, "region": 1, "minPriceCents": 1}}},
	})
	if err != nil {
		return nil, fmt.Errorf("autocomplete: %w", err)
	}
	var out []products.Product
	if err := cur.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("decode autocomplete: %w", err)
	}
	return out, nil
}
