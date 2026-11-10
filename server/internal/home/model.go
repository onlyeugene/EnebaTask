package home

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Block types. The client picks a component by type; the server picks a service by type.
const (
	TypeHero        = "hero"
	TypeCategoryRow = "category_row"
	TypeBestsellers = "bestsellers"
	TypeDeals       = "deals"
	TypeNewReleases = "new_releases"
	TypeGiftCards   = "gift_cards"
)

// Block is one homepage section. Config differs per type, so it is kept raw
// and each block service decodes it into its own config struct below.
type Block struct {
	ID     bson.ObjectID `bson:"_id"`
	Key    string        `bson:"key"` // stable human name, e.g. "hero-main"
	Type   string        `bson:"type"`
	Title  string        `bson:"title"`
	Order  int           `bson:"order"`
	Active bool          `bson:"active"`
	Config bson.Raw      `bson:"config"`
}

type Banner struct {
	Title    string `bson:"title" json:"title"`
	Subtitle string `bson:"subtitle" json:"subtitle"`
	CTALabel string `bson:"ctaLabel" json:"ctaLabel"`
	Href     string `bson:"href" json:"href"`
	Theme    string `bson:"theme" json:"theme"` // purple | yellow | mint | pink
}

type HeroConfig struct {
	Main  Banner   `bson:"main"`
	Tiles []Banner `bson:"tiles"`
}

// CategoryLink points at a store filter: Param and Value become /store?{param}={value},
// and the same pair is used to count matching products.
type CategoryLink struct {
	Slug  string `bson:"slug"`
	Label string `bson:"label"`
	Icon  string `bson:"icon"`  // lucide icon name
	Param string `bson:"param"` // type | platform | genre
	Value string `bson:"value"`
}

type CategoryConfig struct {
	Categories []CategoryLink `bson:"categories"`
}

// RailConfig is shared by bestsellers, deals, new_releases and gift_cards.
type RailConfig struct {
	Limit          int `bson:"limit"`
	MinDiscountPct int `bson:"minDiscountPct,omitempty"` // deals only
}
