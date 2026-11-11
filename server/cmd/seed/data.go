package main

import "eugene/internal/home"

// All titles, studios and sellers are made up. Platform names appear as plain text only.

type sellerSeed struct {
	slug, name string
	rating     float64
	sales      int
	color      string
}

var sellerSeeds = []sellerSeed{
	{"lootcodes", "LootCodes", 4.8, 12093, "#84e916"},
	{"keyvault", "KeyVault", 4.6, 8740, "#fad318"},
	{"pixeltrade", "PixelTrade", 4.9, 21388, "#63e3c2"},
	{"gamenest", "GameNest", 4.3, 3911, "#e76a8c"},
	{"codeharbor", "CodeHarbor", 4.5, 6402, "#7a4edf"},
	{"quickkeys", "QuickKeys", 4.2, 1877, "#23c299"},
}

type gameSeed struct {
	title  string
	genres []string
}

// 48 games. Platform, region, price and studio are picked by the seeded random source in build.go.
var gameSeeds = []gameSeed{
	{"Starfall Odyssey", []string{"rpg", "adventure"}},
	{"Iron Tide", []string{"strategy"}},
	{"Neon Drift", []string{"racing"}},
	{"Hollow Crown", []string{"rpg", "action"}},
	{"Last Lantern", []string{"horror", "adventure"}},
	{"Skyforge Tactics", []string{"strategy", "rpg"}},
	{"Velvet Circuit", []string{"racing", "indie"}},
	{"Ember and Ash", []string{"action", "rpg"}},
	{"Quiet Harbor", []string{"simulation", "indie"}},
	{"Rogue Meridian", []string{"shooter", "action"}},
	{"Pixel Pilgrims", []string{"puzzle", "indie"}},
	{"Frostline Siege", []string{"strategy"}},
	{"Crimson Relay", []string{"sports"}},
	{"Deep Archive", []string{"puzzle", "adventure"}},
	{"Wildgrove Farm", []string{"simulation"}},
	{"Titan Breakers", []string{"action", "shooter"}},
	{"Moonlit Courier", []string{"adventure", "indie"}},
	{"Gravity Bloom", []string{"puzzle"}},
	{"Shadow Ledger", []string{"action", "adventure"}},
	{"Sunken Kingdoms", []string{"rpg"}},
	{"Overclock Arena", []string{"shooter"}},
	{"Paper Galaxy", []string{"indie", "strategy"}},
	{"Riftwalker", []string{"action", "rpg"}},
	{"Copper Valley", []string{"simulation"}},
	{"Stormborne Legends", []string{"rpg"}},
	{"Midnight Garage", []string{"racing", "simulation"}},
	{"Hex and Hollow", []string{"strategy", "rpg"}},
	{"Signal Lost", []string{"horror"}},
	{"Glass Orchard", []string{"puzzle", "adventure"}},
	{"Thunderpeak Rally", []string{"racing"}},
	{"Arcane Foundry", []string{"strategy", "simulation"}},
	{"Dust Runners", []string{"action", "shooter"}},
	{"Lumen Drift", []string{"indie", "adventure"}},
	{"Bastion of Thorns", []string{"strategy"}},
	{"Coral Depths", []string{"adventure"}},
	{"Haunt Alley", []string{"horror", "action"}},
	{"Pocket Kingdoms", []string{"strategy", "indie"}},
	{"Velocity Cup", []string{"sports", "racing"}},
	{"Ironclad Saga", []string{"rpg"}},
	{"Static Bloom", []string{"puzzle", "indie"}},
	{"Nightfall Protocol", []string{"shooter"}},
	{"Orchard Lane", []string{"simulation"}},
	{"Wyrmspire", []string{"rpg", "action"}},
	{"Cosmic Courier Co", []string{"simulation", "indie"}},
	{"Frozen Echo", []string{"adventure"}},
	{"Brawl Circuit", []string{"sports", "action"}},
	{"Saltwind Sailors", []string{"adventure", "simulation"}},
	{"Clockwork Keep", []string{"puzzle", "strategy"}},
}

var studios = []string{
	"Ninefold Studio", "Brightmoor Games", "Lowtide Interactive", "Paper Lantern Works",
	"Redfern Labs", "Glasshouse Games", "Northgate Studio", "Saltmarsh Software",
}

type giftCardSeed struct {
	platform, region string
	valueEUR         int
}

var giftCardSeeds = []giftCardSeed{
	{"steam", "europe", 20},
	{"steam", "europe", 50},
	{"psn", "europe", 25},
	{"psn", "united_kingdom", 50},
	{"xbox", "europe", 25},
	{"nintendo", "europe", 35},
}

type softwareSeed struct {
	title      string
	msrpCents  int64
	activation string
}

var softwareSeeds = []softwareSeed{
	{"Orbit Office Suite", 9999, "Enter the key in Orbit Office under Account > Activate."},
	{"Shieldwall Antivirus 1 Year", 3999, "Install Shieldwall, open Settings > License and paste the key."},
	{"Pixelsmith Photo Editor", 5999, "Start Pixelsmith and enter the key when asked on first launch."},
	{"Cleanline PC Optimizer", 2999, "Open Cleanline, choose Register and paste the key."},
	{"Vaultkey Password Manager 1 Year", 2499, "Sign in to Vaultkey and redeem the key under Billing."},
	{"Tunnelfox VPN 1 Year", 4999, "Sign in to the Tunnelfox app and enter the key under Subscription."},
}

// Cover palettes: [base color, accent]. Bases lean on the brand purples so covers fit the page.
var palettes = [][2]string{
	{"#4618ac", "#63e3c2"},
	{"#351183", "#fad318"},
	{"#7a4edf", "#84e916"},
	{"#d32252", "#fad318"},
	{"#23c299", "#1f0a4d"},
	{"#3c1395", "#e76a8c"},
	{"#99001f", "#63e3c2"},
	{"#1f0a4d", "#84e916"},
}

var platformLabels = map[string]string{
	"steam":    "(PC) Steam Key",
	"psn":      "(PS5) PSN Key",
	"xbox":     "(Xbox Series X|S) Xbox Live Key",
	"nintendo": "(Nintendo Switch) eShop Key",
	"windows":  "(PC) Key",
}

var regionLabels = map[string]string{
	"global":         "GLOBAL",
	"europe":         "EUROPE",
	"united_states":  "UNITED STATES",
	"united_kingdom": "UNITED KINGDOM",
}

var activationText = map[string]string{
	"steam":    "Open Steam, go to Games > Activate a Product on Steam and enter the key.",
	"psn":      "Sign in to PlayStation Store, choose Redeem Codes and enter the key.",
	"xbox":     "Go to the Microsoft Store redeem page or press Redeem a code on your console.",
	"nintendo": "Open Nintendo eShop, choose Redeem Code and enter the key.",
}

// Homepage layout. Order leaves gaps of 10 so a block can be slotted in later without renumbering.
var heroConfig = home.HeroConfig{
	Main: home.Banner{
		Title: "Starfall Odyssey", Subtitle: "Up to 48% off on the year's biggest space RPG",
		CTALabel: "Buy now", Href: "/product/starfall-odyssey-steam-global", Theme: "purple",
	},
	Tiles: []home.Banner{
		{Title: "Top deals", Subtitle: "Discounts of 20% and more", CTALabel: "Shop deals", Href: "/store?sort=discount", Theme: "yellow"},
		{Title: "Gift cards", Subtitle: "Top up any wallet", CTALabel: "Browse", Href: "/store?type=gift_card", Theme: "mint"},
		{Title: "New releases", Subtitle: "Fresh keys every week", CTALabel: "See what's new", Href: "/store?sort=newest", Theme: "pink"},
	},
}

var categoryConfig = home.CategoryConfig{
	Categories: []home.CategoryLink{
		{Slug: "steam", Label: "Steam", Icon: "Gamepad2", Param: "platform", Value: "steam"},
		{Slug: "psn", Label: "PlayStation", Icon: "Gamepad", Param: "platform", Value: "psn"},
		{Slug: "xbox", Label: "Xbox", Icon: "Joystick", Param: "platform", Value: "xbox"},
		{Slug: "nintendo", Label: "Nintendo", Icon: "Puzzle", Param: "platform", Value: "nintendo"},
		{Slug: "gift-cards", Label: "Gift cards", Icon: "Gift", Param: "type", Value: "gift_card"},
		{Slug: "software", Label: "Software", Icon: "Monitor", Param: "type", Value: "software"},
		{Slug: "rpg", Label: "RPG", Icon: "Swords", Param: "genre", Value: "rpg"},
		{Slug: "racing", Label: "Racing", Icon: "Car", Param: "genre", Value: "racing"},
	},
}

// Genre tiles for the "Categories" grid at the bottom of the homepage. Every genre in the seed
// has a tile; the client shows one row and a "Load more" button for the rest.
var genreGridConfig = home.CategoryConfig{
	Categories: []home.CategoryLink{
		{Slug: "action", Label: "Action", Icon: "Zap", Param: "genre", Value: "action"},
		{Slug: "adventure", Label: "Adventure", Icon: "Compass", Param: "genre", Value: "adventure"},
		{Slug: "rpg", Label: "RPG", Icon: "Swords", Param: "genre", Value: "rpg"},
		{Slug: "strategy", Label: "Strategy", Icon: "Castle", Param: "genre", Value: "strategy"},
		{Slug: "indie", Label: "Indie", Icon: "Sparkles", Param: "genre", Value: "indie"},
		{Slug: "simulation", Label: "Simulation", Icon: "Plane", Param: "genre", Value: "simulation"},
		{Slug: "puzzle", Label: "Puzzle", Icon: "Puzzle", Param: "genre", Value: "puzzle"},
		{Slug: "racing", Label: "Racing", Icon: "Car", Param: "genre", Value: "racing"},
		{Slug: "shooter", Label: "Shooter", Icon: "Crosshair", Param: "genre", Value: "shooter"},
		{Slug: "horror", Label: "Horror", Icon: "Ghost", Param: "genre", Value: "horror"},
		{Slug: "sports", Label: "Sports", Icon: "Trophy", Param: "genre", Value: "sports"},
	},
}

type blockSeed struct {
	key, typ, title string
	order           int
	config          any
}

var blockSeeds = []blockSeed{
	{"hero-main", home.TypeHero, "", 10, heroConfig},
	{"categories", home.TypeCategoryRow, "Browse categories", 20, categoryConfig},
	{"bestsellers", home.TypeBestsellers, "Best sellers", 30, home.RailConfig{Limit: 10}},
	{"deals", home.TypeDeals, "Top deals", 40, home.RailConfig{Limit: 10, MinDiscountPct: 20}},
	{"new-releases", home.TypeNewReleases, "New releases", 50, home.RailConfig{Limit: 10}},
	{"gift-cards", home.TypeGiftCards, "Gift cards", 60, home.RailConfig{Limit: 10}},
	{"genres", home.TypeCategoryGrid, "Categories", 70, genreGridConfig},
}
