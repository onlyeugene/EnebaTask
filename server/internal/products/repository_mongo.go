package products

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoRepo struct {
	products *mongo.Collection
	offers   *mongo.Collection
}

func NewMongoRepo(database *mongo.Database) *MongoRepo {
	return &MongoRepo{products: database.Collection("products"), offers: database.Collection("offers")}
}

// The cover SVG is a few KB per product; lists and detail never need it.
var noCover = bson.M{"cover": 0}

func (r *MongoRepo) List(ctx context.Context, q ListQuery) ([]Product, int64, error) {
	filter := listFilter(q)
	total, err := r.products.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count products: %w", err)
	}

	opts := options.Find().
		SetSort(listSort(q.Sort)).
		SetSkip(int64((q.Page - 1) * q.PageSize)).
		SetLimit(int64(q.PageSize)).
		SetProjection(noCover)
	cur, err := r.products.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("find products: %w", err)
	}
	var out []Product
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, fmt.Errorf("decode products: %w", err)
	}
	return out, total, nil
}

func listFilter(q ListQuery) bson.M {
	f := bson.M{"active": true}
	if q.InStockOnly {
		f["inStock"] = true
	}
	in := func(field string, values []string) {
		if len(values) > 0 {
			f[field] = bson.M{"$in": values}
		}
	}
	in("type", q.Types)
	in("platform", q.Platforms)
	in("region", q.Regions)
	in("genres", q.Genres) // genres is an array: $in matches if any genre is in the list

	if q.MinPriceCents != nil || q.MaxPriceCents != nil {
		price := bson.M{}
		if q.MinPriceCents != nil {
			price["$gte"] = *q.MinPriceCents
		}
		if q.MaxPriceCents != nil {
			price["$lte"] = *q.MaxPriceCents
		}
		// Sold-out products have minPriceCents null, which never matches a range, so they drop out here.
		f["minPriceCents"] = price
	}
	return f
}

// listSort puts in-stock products first, then the chosen order, then _id.
// In-stock first matters for price_asc with inStock=false: Mongo sorts null before numbers,
// so sold-out products would otherwise lead the list. The _id tie-break keeps pages stable:
// without it, products with the same soldCount could swap between page 1 and page 2.
func listSort(sort string) bson.D {
	var by bson.E
	switch sort {
	case SortPriceAsc:
		by = bson.E{Key: "minPriceCents", Value: 1}
	case SortPriceDesc:
		by = bson.E{Key: "minPriceCents", Value: -1}
	case SortDiscount:
		by = bson.E{Key: "discountPct", Value: -1}
	case SortNewest:
		by = bson.E{Key: "createdAt", Value: -1}
	default:
		by = bson.E{Key: "soldCount", Value: -1}
	}
	return bson.D{{Key: "inStock", Value: -1}, by, {Key: "_id", Value: 1}}
}

// Facets counts in-stock active products per filter value in one aggregation.
// It counts the same set the store shows by default (inStock=true).
func (r *MongoRepo) Facets(ctx context.Context) (Facets, error) {
	group := func(field string) bson.A {
		return bson.A{bson.M{"$group": bson.M{"_id": field, "count": bson.M{"$sum": 1}}}}
	}
	cur, err := r.products.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"active": true, "inStock": true}}},
		{{Key: "$facet", Value: bson.M{
			"types":     group("$type"),
			"platforms": group("$platform"),
			"regions":   group("$region"),
			"genres":    append(bson.A{bson.M{"$unwind": "$genres"}}, group("$genres")...),
			"price": bson.A{bson.M{"$group": bson.M{
				"_id": nil, "min": bson.M{"$min": "$minPriceCents"}, "max": bson.M{"$max": "$minPriceCents"},
			}}},
		}}},
	})
	if err != nil {
		return Facets{}, fmt.Errorf("facets: %w", err)
	}

	type countRow struct {
		Value string `bson:"_id"`
		Count int    `bson:"count"`
	}
	var rows []struct {
		Types     []countRow `bson:"types"`
		Platforms []countRow `bson:"platforms"`
		Regions   []countRow `bson:"regions"`
		Genres    []countRow `bson:"genres"`
		Price     []struct {
			Min int64 `bson:"min"`
			Max int64 `bson:"max"`
		} `bson:"price"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		return Facets{}, fmt.Errorf("decode facets: %w", err)
	}

	toMap := func(rs []countRow) map[string]int {
		m := make(map[string]int, len(rs))
		for _, row := range rs {
			m[row.Value] = row.Count
		}
		return m
	}
	// $facet always returns exactly one document.
	row := rows[0]
	f := Facets{
		Types: toMap(row.Types), Platforms: toMap(row.Platforms),
		Regions: toMap(row.Regions), Genres: toMap(row.Genres),
	}
	if len(row.Price) > 0 { // empty when nothing is in stock
		f.MinPriceCents, f.MaxPriceCents = row.Price[0].Min, row.Price[0].Max
	}
	return f, nil
}

func (r *MongoRepo) BySlug(ctx context.Context, slug string) (Product, error) {
	var p Product
	err := r.products.FindOne(ctx, bson.M{"slug": slug, "active": true}, options.FindOne().SetProjection(noCover)).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("find product %q: %w", slug, err)
	}
	return p, nil
}

// CoverSVG does not filter on active: a cart line or old link can still point at a product
// that was taken down, and a cover image leaks nothing.
func (r *MongoRepo) CoverSVG(ctx context.Context, slug string) (string, error) {
	var doc struct {
		Cover struct {
			SVG string `bson:"svg"`
		} `bson:"cover"`
	}
	err := r.products.FindOne(ctx, bson.M{"slug": slug}, options.FindOne().SetProjection(bson.M{"cover.svg": 1})).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("find cover %q: %w", slug, err)
	}
	return doc.Cover.SVG, nil
}

func (r *MongoRepo) Offers(ctx context.Context, productID bson.ObjectID) ([]OfferWithSeller, error) {
	cur, err := r.offers.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"productId": productID, "active": true}}},
		{{Key: "$sort", Value: bson.D{{Key: "priceCents", Value: 1}, {Key: "_id", Value: 1}}}},
		{{Key: "$lookup", Value: bson.M{"from": "sellers", "localField": "sellerId", "foreignField": "_id", "as": "seller"}}},
		// $unwind also drops an offer whose seller is missing, so the page never shows a blank seller.
		{{Key: "$unwind", Value: "$seller"}},
	})
	if err != nil {
		return nil, fmt.Errorf("find offers: %w", err)
	}
	var out []OfferWithSeller
	if err := cur.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("decode offers: %w", err)
	}
	return out, nil
}
