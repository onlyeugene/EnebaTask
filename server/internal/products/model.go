package products

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Product is the document in the products collection (PLAN.md section 1).
type Product struct {
	ID            bson.ObjectID `bson:"_id"`
	Slug          string        `bson:"slug"`
	Name          string        `bson:"name"`
	Type          string        `bson:"type"`     // game | gift_card | software
	Platform      string        `bson:"platform"` // steam | psn | xbox | nintendo | windows
	Region        string        `bson:"region"`   // global | europe | united_states | united_kingdom
	Genres        []string      `bson:"genres"`
	Description   string        `bson:"description"`
	Developer     string        `bson:"developer"`
	Publisher     string        `bson:"publisher"`
	ReleaseDate   time.Time     `bson:"releaseDate"`
	Activation    string        `bson:"activation"`
	GiftCardValue *string       `bson:"giftCardValue"` // "€50" for gift cards, null otherwise
	Cover         Cover         `bson:"cover"`

	OriginalPriceCents int64 `bson:"originalPriceCents"`

	// Summary of the product's offers. Copied here so the store can sort and filter
	// without a $lookup per row; kept in sync in the same transaction that changes stock.
	MinPriceCents *int64 `bson:"minPriceCents"` // null when sold out
	DiscountPct   int    `bson:"discountPct"`
	OfferCount    int    `bson:"offerCount"`
	InStock       bool   `bson:"inStock"`

	CashbackPct int       `bson:"cashbackPct"`
	SoldCount   int       `bson:"soldCount"`
	Active      bool      `bson:"active"`
	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
}

type Cover struct {
	SVG     string    `bson:"svg"`
	Palette [2]string `bson:"palette"`
}

type Seller struct {
	ID         bson.ObjectID `bson:"_id"`
	Slug       string        `bson:"slug"`
	Name       string        `bson:"name"`
	Rating     float64       `bson:"rating"`
	SalesCount int           `bson:"salesCount"`
	Color      string        `bson:"color"`
	CreatedAt  time.Time     `bson:"createdAt"`
}

// Offer is one seller's listing of a product. Stock is a cached count of claimable keys;
// the keys collection is the source of truth.
type Offer struct {
	ID         bson.ObjectID `bson:"_id"`
	ProductID  bson.ObjectID `bson:"productId"`
	SellerID   bson.ObjectID `bson:"sellerId"`
	PriceCents int64         `bson:"priceCents"`
	Stock      int           `bson:"stock"`
	Active     bool          `bson:"active"`
	CreatedAt  time.Time     `bson:"createdAt"`
	UpdatedAt  time.Time     `bson:"updatedAt"`
}

// OfferWithSeller is an offer joined with its seller by $lookup.
type OfferWithSeller struct {
	Offer  `bson:",inline"`
	Seller Seller `bson:"seller"`
}

// Sort values for GET /api/products (API.md).
const (
	SortRelevance   = "relevance" // only means something with q (phase 4); without q it falls back to bestselling
	SortPriceAsc    = "price_asc"
	SortPriceDesc   = "price_desc"
	SortBestselling = "bestselling"
	SortDiscount    = "discount"
	SortNewest      = "newest"
)

var Sorts = []string{SortRelevance, SortPriceAsc, SortPriceDesc, SortBestselling, SortDiscount, SortNewest}

const MaxPageSize = 60

// ListQuery is a parsed and checked GET /api/products request.
// Empty lists mean "no filter". Values in one list are OR'd; different lists are AND'd.
type ListQuery struct {
	Types, Platforms, Regions, Genres []string
	MinPriceCents, MaxPriceCents      *int64
	InStockOnly                       bool
	Sort                              string
	Page, PageSize                    int
}

// Option is one allowed filter value and the label the filter panel shows for it.
type Option struct{ Value, Label string }

// Allowed values, in the order the filter panel lists them.
var (
	Types = []Option{
		{"game", "Games"}, {"gift_card", "Gift cards"}, {"software", "Software"},
	}
	Platforms = []Option{
		{"steam", "Steam"}, {"psn", "PSN"}, {"xbox", "Xbox"}, {"nintendo", "Nintendo"}, {"windows", "Windows"},
	}
	Regions = []Option{
		{"global", "Global"}, {"europe", "Europe"}, {"united_states", "United States"}, {"united_kingdom", "United Kingdom"},
	}
)

// Facets is the raw per-value product count the repository returns for the filter panel.
type Facets struct {
	Types, Platforms, Regions, Genres map[string]int
	MinPriceCents, MaxPriceCents      int64
}

// ---- JSON shapes from API.md ----

type ProductCard struct {
	ID                 string  `json:"id"`
	Slug               string  `json:"slug"`
	Name               string  `json:"name"`
	Type               string  `json:"type"`
	Platform           string  `json:"platform"`
	Region             string  `json:"region"`
	CoverURL           string  `json:"coverUrl"`
	GiftCardValue      *string `json:"giftCardValue,omitempty"`
	OriginalPriceCents int64   `json:"originalPriceCents"`
	MinPriceCents      *int64  `json:"minPriceCents"` // null = sold out
	DiscountPct        int     `json:"discountPct"`
	CashbackPct        int     `json:"cashbackPct"`
	OfferCount         int     `json:"offerCount"`
}

type ProductDetail struct {
	ProductCard
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
	Developer   string   `json:"developer"`
	Publisher   string   `json:"publisher"`
	ReleaseDate string   `json:"releaseDate"` // "2024-03-18"
	Activation  string   `json:"activation"`
}

type OfferView struct {
	ID         string     `json:"id"`
	ProductID  string     `json:"productId"`
	PriceCents int64      `json:"priceCents"`
	Stock      int        `json:"stock"`
	Seller     SellerView `json:"seller"`
}

type SellerView struct {
	ID         string  `json:"id"`
	Slug       string  `json:"slug"`
	Name       string  `json:"name"`
	Rating     float64 `json:"rating"`
	SalesCount int     `json:"salesCount"`
}

type ListResult struct {
	Items    []ProductCard `json:"items"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	Total    int64         `json:"total"`
}

type FilterOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type Filters struct {
	Types     []FilterOption `json:"types"`
	Platforms []FilterOption `json:"platforms"`
	Regions   []FilterOption `json:"regions"`
	Genres    []FilterOption `json:"genres"`
	Price     struct {
		MinCents int64 `json:"minCents"`
		MaxCents int64 `json:"maxCents"`
	} `json:"price"`
}
