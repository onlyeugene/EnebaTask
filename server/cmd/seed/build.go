package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"eugene/internal/home"
	"eugene/internal/keys"
	"eugene/internal/products"
)

type dataset struct {
	sellers  []products.Seller
	products []products.Product
	offers   []products.Offer
	keys     []keys.Key
	blocks   []home.Block
}

// build turns the lists in data.go into documents.
// The random source has a fixed seed, so every run gives the same titles, prices and stock
// (only the ObjectIDs and timestamps differ).
func build(now time.Time) (dataset, error) {
	rng := rand.New(rand.NewPCG(42, 2026))
	var d dataset

	for _, s := range sellerSeeds {
		d.sellers = append(d.sellers, products.Seller{
			ID: bson.NewObjectID(), Slug: s.slug, Name: s.name, Rating: s.rating,
			SalesCount: s.sales, Color: s.color, CreatedAt: now,
		})
	}

	for i, g := range gameSeeds {
		platform, region := pickPlatformRegion(rng)
		if i == 0 {
			// The hero banner links to this exact slug.
			platform, region = "steam", "global"
		}
		studio := studios[rng.IntN(len(studios))]
		p := newProduct(rng, now, g.title, "game", platform, region,
			[]int64{1999, 2999, 3999, 4999, 5999, 6999}[rng.IntN(6)])
		p.Genres = g.genres
		p.Developer, p.Publisher = studio, studios[rng.IntN(len(studios))]
		p.Activation = activationText[platform]
		p.Description = fmt.Sprintf("%s is a %s game from %s. Buy a digital key, redeem it on %s and start playing right away.",
			g.title, strings.Join(g.genres, " and "), studio, platformName(platform))
		d.addProduct(rng, now, p, 35, 95)
	}

	for _, g := range giftCardSeeds {
		title := fmt.Sprintf("%s Gift Card €%d", platformName(g.platform), g.valueEUR)
		p := newProduct(rng, now, title, "gift_card", g.platform, g.region, int64(g.valueEUR)*100)
		// Gift cards have their own name pattern: there is no "(PC)" edition label.
		p.Name = fmt.Sprintf("%s Gift Card %d EUR Key %s", platformName(g.platform), g.valueEUR, regionLabels[g.region])
		p.Slug = slugify(p.Name)
		value := fmt.Sprintf("€%d", g.valueEUR)
		p.GiftCardValue = &value
		p.Developer, p.Publisher = platformName(g.platform), platformName(g.platform)
		p.Activation = activationText[g.platform]
		p.Description = fmt.Sprintf("Adds €%d to your %s wallet. Works only for accounts in the %s region.",
			g.valueEUR, platformName(g.platform), strings.ToLower(regionLabels[g.region]))
		// Gift cards sell close to face value, sometimes a little above it.
		d.addProduct(rng, now, p, 90, 105)
	}

	for _, s := range softwareSeeds {
		p := newProduct(rng, now, s.title, "software", "windows", "global", s.msrpCents)
		studio := studios[rng.IntN(len(studios))]
		p.Developer, p.Publisher = studio, studio
		p.Activation = s.activation
		p.Description = fmt.Sprintf("%s by %s. One key activates one Windows PC.", s.title, studio)
		d.addProduct(rng, now, p, 30, 80)
	}

	for _, b := range blockSeeds {
		raw, err := bson.Marshal(b.config)
		if err != nil {
			return d, fmt.Errorf("block %s config: %w", b.key, err)
		}
		d.blocks = append(d.blocks, home.Block{
			ID: bson.NewObjectID(), Key: b.key, Type: b.typ, Title: b.title,
			Order: b.order, Active: true, Config: raw,
		})
	}
	return d, nil
}

// newProduct fills the fields every product type shares. Callers set the rest.
func newProduct(rng *rand.Rand, now time.Time, title, typ, platform, region string, msrp int64) products.Product {
	name := fmt.Sprintf("%s %s %s", title, platformLabels[platform], regionLabels[region])
	// createdAt is spread over the last 180 days so "New releases" has a real order.
	created := now.Add(-time.Duration(rng.IntN(180*24)) * time.Hour)
	return products.Product{
		ID:                 bson.NewObjectID(),
		Slug:               slugify(title + " " + platform + " " + region),
		Name:               name,
		Type:               typ,
		Platform:           platform,
		Region:             region,
		Genres:             []string{}, // empty array, not null, so filters never trip on a missing field
		ReleaseDate:        now.AddDate(0, 0, -30-rng.IntN(2000)).Truncate(24 * time.Hour),
		Cover:              coverFor(rng, title),
		OriginalPriceCents: msrp,
		CashbackPct:        3 + rng.IntN(10),
		SoldCount:          rng.IntN(1500),
		Active:             true,
		CreatedAt:          created,
		UpdatedAt:          created,
	}
}

// addProduct creates 2–4 offers from different sellers, 5–20 keys per offer,
// then fills in the product's offer summary and stores everything.
// Offer prices are minPct–maxPct percent of the MSRP.
func (d *dataset) addProduct(rng *rand.Rand, now time.Time, p products.Product, minPct, maxPct int) {
	count := 2 + rng.IntN(3)
	var offers []products.Offer
	for _, si := range rng.Perm(len(d.sellers))[:count] {
		pct := int64(minPct + rng.IntN(maxPct-minPct+1))
		// Prices end in 9 cents, like real store prices.
		price := p.OriginalPriceCents*pct/100/10*10 + 9
		o := products.Offer{
			ID: bson.NewObjectID(), ProductID: p.ID, SellerID: d.sellers[si].ID,
			PriceCents: price, Active: true, CreatedAt: now, UpdatedAt: now,
		}
		n := 5 + rng.IntN(16)
		for range n {
			d.keys = append(d.keys, keys.Key{
				ID: bson.NewObjectID(), OfferID: o.ID, ProductID: p.ID,
				Code: keyCode(rng), Status: keys.StatusAvailable, CreatedAt: now,
			})
		}
		o.Stock = n
		offers = append(offers, o)
	}
	summarize(&p, offers)
	d.offers = append(d.offers, offers...)
	d.products = append(d.products, p)
}

// summarize sets minPriceCents, discountPct, offerCount and inStock from the offers.
// The purchase transaction (phase 9) must use the same rules when it recounts.
func summarize(p *products.Product, offers []products.Offer) {
	p.MinPriceCents, p.OfferCount, p.DiscountPct, p.InStock = nil, 0, 0, false
	for _, o := range offers {
		if !o.Active || o.Stock == 0 {
			continue
		}
		p.OfferCount++
		if p.MinPriceCents == nil || o.PriceCents < *p.MinPriceCents {
			price := o.PriceCents
			p.MinPriceCents = &price
		}
	}
	if p.MinPriceCents == nil {
		return
	}
	p.InStock = true
	// Only a price below MSRP counts as a discount; gift cards above face value show 0%.
	if saved := p.OriginalPriceCents - *p.MinPriceCents; saved > 0 {
		p.DiscountPct = int(saved * 100 / p.OriginalPriceCents)
	}
}

// pickPlatformRegion leans toward Steam, like the real store. Steam keys are mostly global;
// console keys are usually tied to one region.
func pickPlatformRegion(rng *rand.Rand) (string, string) {
	platforms := []string{"steam", "steam", "steam", "steam", "psn", "psn", "xbox", "xbox", "nintendo"}
	platform := platforms[rng.IntN(len(platforms))]
	if platform == "steam" {
		return platform, []string{"global", "global", "global", "europe", "united_states"}[rng.IntN(5)]
	}
	return platform, []string{"europe", "europe", "united_states", "united_kingdom", "global"}[rng.IntN(5)]
}

func coverFor(rng *rand.Rand, title string) products.Cover {
	palette := palettes[rng.IntN(len(palettes))]
	return products.Cover{SVG: products.CoverSVG(title, palette), Palette: palette}
}

func platformName(platform string) string {
	return map[string]string{
		"steam": "Steam", "psn": "PSN", "xbox": "Xbox", "nintendo": "Nintendo eShop", "windows": "Windows",
	}[platform]
}

// keyCode makes a fake code like "7QX4M-9KPLD-2WZ8A". No 0/O or 1/I, which are easy to misread.
func keyCode(rng *rand.Rand) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var b strings.Builder
	for i := range 15 {
		if i > 0 && i%5 == 0 {
			b.WriteByte('-')
		}
		b.WriteByte(chars[rng.IntN(len(chars))])
	}
	return b.String()
}

// slugify lowercases and turns every run of non-letters/digits into one dash.
func slugify(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}
