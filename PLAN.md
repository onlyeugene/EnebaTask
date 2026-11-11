# Plan

An Eneba-style marketplace for digital game keys. Go + MongoDB Atlas API on Render, Next.js frontend on Vercel. Endpoints are in [API.md](API.md).

App name: **Eugene**. Logo is a plain text wordmark next to a small lucide icon (`KeyRound`), in the brand colors.

---

## 0. Architecture at a glance

```
Browser (Next.js on Vercel)
  │  fetch + Authorization: Bearer <access JWT>
  ▼
Go API (chi, on Render, Docker)
  │  mongo-go-driver v2
  ▼
MongoDB Atlas M0
  ├─ db eneba_dev   (local)
  ├─ db eneba_prod  (Render)
  └─ Atlas Search index "products_search" (one per db)
```

- The browser talks to the Go API directly. Next.js serves pages only: no API routes, no rewrites, no server-side data fetching.
- Libraries: `go-chi/chi/v5`, `go-chi/cors`, `go-chi/httprate`, `go.mongodb.org/mongo-driver/v2`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, `joho/godotenv` (local `.env` only). Client: Next.js (App Router), TanStack Query, Zustand, React Hook Form, Zod, lucide-react, Tailwind v4.
- Money is stored and sent as integer cents (EUR only). Floats can't hold 0.10 exactly; cents can.

---

## 1. Collections and document shapes

`_id` is an `ObjectId` unless noted. Times are UTC `Date`.

### products

```js
{
  _id, slug: "starfall-odyssey-steam-global",   // unique, used in URLs
  name: "Starfall Odyssey (PC) Steam Key GLOBAL",
  type: "game",                                  // game | gift_card | software
  platform: "steam",                             // steam | psn | xbox | nintendo | windows
  region: "global",                              // global | europe | united_states | united_kingdom
  genres: ["rpg", "adventure"],
  description, developer, publisher, releaseDate, activation,
  giftCardValue: null,                           // "€50" for gift cards
  cover: { svg: "<svg…>", palette: ["#4618ac", "#63e3c2"] },
  originalPriceCents: 2999,

  // summary of its offers, kept in sync inside the purchase transaction
  minPriceCents: 1549, discountPct: 48, offerCount: 3, inStock: true,
  cashbackPct: 11,
  soldCount: 412,                                 // drives "best sellers"
  active: true, createdAt, updatedAt
}
```

### sellers

```js
{ _id, slug: "lootcodes", name: "LootCodes", rating: 4.8, salesCount: 12093, color: "#84e916", createdAt }
```

### offers

```js
{ _id, productId, sellerId, priceCents: 1549, stock: 14, active: true, createdAt, updatedAt }
```

`stock` is a cached count used for listings and sorting. It is recomputed from `keys` inside every transaction that claims or releases keys, never `$inc`'d. The source of truth is always the `keys` collection (see section 5).

### keys

```js
{
  _id, offerId, productId,
  code: "7QX4M-9KPLD-2WZ8A",
  status: "available",            // available | reserved | sold
  orderId: null,                  // set when reserved
  reservedUntil: null,            // set when reserved; once in the past, the key counts as available again
  soldAt: null, createdAt
}
```

### users

```js
{ _id, email: "eugene@example.com", passwordHash: "$2a$12$…", name: "Eugene", createdAt, updatedAt }
```

### carts

```js
{ _id, userId, items: [ { offerId, productId, quantity: 2, addedAt } ], updatedAt }
```

### orders

```js
{
  _id, number: "ORD-7K2M9Q", userId,
  status: "pending_payment",      // pending_payment | paid | failed | expired
  items: [{
    offerId, productId, sellerId,
    name, slug, platform, region, sellerName,     // snapshot at purchase time
    unitPriceCents: 1549, quantity: 2,
    keyIds: [ObjectId, ObjectId],
    keys: [ { code: "7QX4M-…" } ]                  // filled when paid
  }],
  totalCents: 3098, cashbackCents: 340,
  reservedUntil, payment: { last4: "4242", outcome: "success" },
  createdAt, paidAt, updatedAt
}
```

### homepage_blocks

```js
{
  _id, key: "hero-main", type: "hero",            // hero | category_row | category_grid | bestsellers | deals | new_releases | gift_cards
  title: "", order: 10, active: true,
  config: { /* per type: banners for hero, limit/filters for rails */ }
}
```

### refresh_tokens

```js
{
  _id, tokenHash: "sha256 hex",   // the raw token is never stored
  userId, familyId,               // familyId: same for every token in one login session's rotation chain
  createdAt, expiresAt,           // +30 days
  usedAt: null,                   // set when rotated; a second use = theft
  replacedBy: null                // _id of the next token in the chain
}
```

### revoked_tokens

```js
{ _id: "<jti>", userId, expiresAt }   // expiresAt = the access token's own exp
```

### Embed vs reference

| Data | Choice | Why |
|---|---|---|
| Offers → product, seller | **Separate collection, reference by id** | A product has several offers, and price/stock change on every sale. Embedding would rewrite the product doc on every purchase and make every buyer of any offer contend on one document. |
| Keys → offer | **Separate collection, one doc per key** | This is the concurrency guard: each key can be claimed with one atomic `findOneAndUpdate`. In an array inside the offer, every buyer would fight over the same doc, and arrays grow without bound. |
| Seller in offer | **Reference + `$lookup`** | Few sellers, and a product page reads at most a handful of offers. Not worth copying seller data into every offer. |
| Product summary (`minPriceCents`, `discountPct`, `offerCount`, `inStock`) | **Copied onto product** | The store page sorts and filters by price. Computing a min over offers for every listing query would mean a `$lookup` per row. We copy it and update it in the same transaction that changes stock. |
| Cart items | **Embedded in cart** | Always read and written together, small, capped at ~20 lines. |
| Order items | **Embedded, with a snapshot of name/price/seller** | An order is a receipt. It must show what was paid even if the product is renamed or repriced later. |
| Delivered keys | **Copied into the order item on payment** (ids kept too) | "Order and key" page reads one document. The key status in `keys` stays the source of truth for stock. |
| Block config | **Embedded in block** | Owned by that block alone. |
| Cover SVG | **Embedded in product, excluded by projection** | Kept with the product it belongs to; every list query projects it out and the `cover.svg` route serves it with cache headers. Keeps list responses small. |

---

## 2. Indexes

Created at API startup (idempotent `CreateMany`) and by the seed script.

| Collection | Index | Purpose |
|---|---|---|
| users | `{ email: 1 }` unique | Login lookup, no duplicate accounts |
| sellers | `{ slug: 1 }` unique | |
| products | `{ slug: 1 }` unique | Product page |
| products | `{ active: 1, inStock: 1, soldCount: -1 }` | Default store sort, best sellers |
| products | `{ active: 1, inStock: 1, minPriceCents: 1 }` | Price sort and price range |
| products | `{ active: 1, inStock: 1, discountPct: -1 }` | Deals block |
| products | `{ active: 1, createdAt: -1 }` | New releases |
| products | `{ type: 1, platform: 1, region: 1 }` | Filter combos without `q` |
| offers | `{ productId: 1, active: 1, priceCents: 1 }` | Offers on product page, summary recompute |
| keys | `{ offerId: 1, status: 1, reservedUntil: 1 }` | Claim next key (`available`, or `reserved` with `reservedUntil < now`) and live stock count |
| keys | `{ status: 1, reservedUntil: 1 }` | Cleanup job finds expired reservations |
| keys | `{ orderId: 1 }` | Release/sell keys of one order |
| carts | `{ userId: 1 }` unique | One cart per user |
| orders | `{ userId: 1, createdAt: -1 }` | Order history |
| orders | `{ status: 1, reservedUntil: 1 }` | Sweeper |
| orders | `{ number: 1 }` unique | |
| homepage_blocks | `{ active: 1, order: 1 }` | Layout |
| refresh_tokens | `{ tokenHash: 1 }` unique | Refresh lookup |
| refresh_tokens | `{ userId: 1 }` | Revoke all on reuse |
| refresh_tokens | `{ expiresAt: 1 }` **TTL, `expireAfterSeconds: 0`** | Mongo deletes the doc once `expiresAt` passes |
| revoked_tokens | `_id` = jti (built in) | Revocation check on every request |
| revoked_tokens | `{ expiresAt: 1 }` **TTL, `expireAfterSeconds: 0`** | A revoked jti only matters until the token would have expired anyway |

TTL note: Mongo's TTL monitor runs about once a minute, so a doc can live a little past `expiresAt`. Code never relies on TTL for correctness; it always checks `expiresAt` too.

### Atlas Search index `products_search` (on `products`)

M0 allows 3 search indexes per cluster; we need one per database (2 total). Stored as JSON in `server/atlas/products_search.json` (embedded into Go by `server/atlas/atlas.go`) and created by the seed script with `SearchIndexes().CreateOne`, or updated with `UpdateOne` if it exists; the seed then waits until Atlas reports it `READY` (fallback: paste the JSON in the Atlas UI). The API checks the index at startup and only logs a warning if it is missing, because `$search` on a missing index returns nothing and no error (checked on `eneba_dev`: 0 rows, nil error). `/healthz` also reports `"search": "ok" | "missing"` but stays `200` when only search is missing; only db down is `503`, so Render never restarts a server that otherwise works.

`_id` is mapped so search sorts can end in `_id` like the plain list sorts, so pages never overlap.

```json
{
  "name": "products_search",
  "definition": {
    "mappings": {
      "dynamic": false,
      "fields": {
        "_id":  { "type": "objectId" },
        "name": [
          { "type": "string", "analyzer": "lucene.standard" },
          { "type": "autocomplete", "tokenization": "edgeGram", "minGrams": 2, "maxGrams": 15, "foldDiacritics": true }
        ],
        "type":          { "type": "token" },
        "platform":      { "type": "token" },
        "region":        { "type": "token" },
        "genres":        { "type": "token" },
        "active":        { "type": "boolean" },
        "inStock":       { "type": "boolean" },
        "minPriceCents": { "type": "number" },
        "discountPct":   { "type": "number" },
        "soldCount":     { "type": "number" },
        "createdAt":     { "type": "date" }
      }
    }
  }
}
```

**Store search** (`/api/products?q=`):

```js
{ $search: {
    index: "products_search",
    compound: {
      should: [ // two clauses per word of q (first 8 words)
        { text:         { query: word, path: "name", fuzzy: { maxEdits: textEdits(word), prefixLength: 1 }, score: { boost: { value: 3 } } } },
        { autocomplete: { query: word, path: "name", fuzzy: { maxEdits: prefixEdits(word), prefixLength: 1 } } }
      ],
      minimumShouldMatch: 1,
      filter: [ /* equals active/inStock, in type/platform/region/genres, range minPriceCents */ ]
    },
    sort: { inStock: -1, <score | minPriceCents | soldCount | …>, _id: 1 },
    count: { type: "total" }
} },
{ $facet: { items: [ { $skip }, { $limit }, { $project: { cover: 0 } } ],
            meta:  [ { $replaceWith: "$$SEARCH_META" }, { $limit: 1 } ] } }
```

`text` + fuzzy catches typos ("skyforje" → "Skyforge"). `autocomplete` catches partial words ("frost" → "Frostline"). `prefixLength: 1` means the first letter must match, which cuts noise and speeds things up. Filters go in `filter` so they don't change the score. Total comes from `$$SEARCH_META.count.total` in the `$facet`. Relevance sorts on `{ $meta: "searchScore" }`, with the same in-stock-first and `_id` tie-break as the plain list.

Typo allowance depends on word length (same steps as Elasticsearch's `AUTO`): `textEdits` is 0 for 1–3 letters, 1 for 4–6, 2 for 7+; `prefixEdits` is 1 from 5 letters, else 0. The first version used one `maxEdits: 2` for the whole query, and `q=star` returned 25 products: "star" is 2 edits from "steam", and the fuzzy prefix "star" matched "stea", so every Steam product came back. Atlas can't vary edits inside one clause, hence one clause per word.

**Autocomplete** (`/api/search/autocomplete`): only the `autocomplete` clauses, in `must` (every word has to match, so "neon dr" narrows to Neon Drift), `filter active: true`, `$limit 8`, project a few fields. Under 2 chars returns `[]` without a query (the index's `minGrams` is 2).

Without `q`, the store uses a normal `find` with the B-tree indexes above. Cheaper and exact.

---

## 3. Endpoints

See [API.md](API.md). Summary of what needs a token:

| Public | Bearer |
|---|---|
| `/healthz`, auth register/login/refresh, products list/filters/detail/cover, autocomplete, homepage layout + blocks, cart quote | auth logout/me, cart get/put/delete/merge, orders create/pay/list/get |

---

## 4. Homepage

**Server**

- `GET /api/homepage` reads active `homepage_blocks` sorted by `order` and returns `{ id, type, title }` only. It never loads block content, so it can't fail because a block failed.
- `GET /api/homepage/blocks/{id}` loads the block doc, looks up its `type` in a registry, and calls that service:

```go
type BlockService interface {
    Resolve(ctx context.Context, block Block) (any, error)
}

// blocks/registry.go; catalog = *products.Service
registry := map[string]home.BlockService{
    "hero":         Hero{},                                       // banners from block.config
    "category_row": Categories{catalog},                          // counts from the store's filter aggregation
    "category_grid": Categories{catalog},                         // same data, drawn as genre tiles
    "bestsellers":  Rail{catalog, "bestselling", nil, "/store?sort=bestselling"},
    "deals":        Rail{catalog, "discount", nil, "/store?sort=discount"},     // + minDiscountPct 20 from config
    "new_releases": Rail{catalog, "newest", nil, "/store?sort=newest"},
    "gift_cards":   Rail{catalog, "bestselling", {"gift_card"}, "/store?type=gift_card"},
}
```

- Each service gets its own 2-second context timeout. An error or timeout becomes `503 BLOCK_UNAVAILABLE` for that block only.
- Adding a block = add a doc + (for a new type) one service. The client needs a new component only for a new `type`.

**Client**

- `app/page.tsx` is a server component. It renders the static page frame and one client component, `<HomeBlocks />`.
- `<HomeBlocks />` runs `useQuery(["homepage"])` for the layout, then maps each entry to a component by `type`:
  `hero → HeroBlock`, `category_row → CategoryRowBlock`, `category_grid → CategoryGridBlock`, the four rail types → `ProductRailBlock`.
- Each block component runs its own `useQuery(["homepage-block", id])` and renders its own skeleton (shaped like the screenshots: dark purple card placeholders), its own error box with a "Retry" button, and its own content.
- Block queries use `retry: 1`. TanStack Query's default is 3 retries with growing delays, so a broken block would sit on its skeleton for several seconds before showing the error. One retry covers a brief network blip; after that the user gets the error and the Retry button.
- Each block is also wrapped in a small React error boundary, so a render bug in one block doesn't blank the page.
- If the layout call itself fails, the page shows one error with retry (there is nothing else to show).
- Demo: in dev, `?fail=<blockId>` on the home URL passes `?fail=1` to that block's request, showing one broken block among working ones.

---

## 5. Two buyers, one key

**The rule:** a key moves `available → reserved → sold`, or back to `available`. A key is **claimable** when:

```js
{ offerId, $or: [ { status: "available" }, { status: "reserved", reservedUntil: { $lt: now } } ] }
```

So an abandoned reservation frees itself the moment `reservedUntil` passes. No background job has to run for that. This matters because Render's free tier sleeps, and a sleeping server runs no jobs.

**Step 1 — create order (reserve):** `POST /api/orders` runs `session.WithTransaction`:

```go
for each cart line:
    for i := 0; i < qty; i++ {
        key, err := keys.FindOneAndUpdate(sessCtx,
            claimable(line.OfferID, now),
            bson.M{"$set": bson.M{"status": "reserved", "orderId": orderID, "reservedUntil": now.Add(reservationTTL)}},
        )
        if errors.Is(err, mongo.ErrNoDocuments) { return ErrOutOfStock{...} } // aborts the whole transaction
        keyIDs = append(keyIDs, key.ID)
    }
    refreshOfferStock(sessCtx, line.OfferID, now)      // stock = count(claimable keys)
    refreshProductSummary(sessCtx, line.ProductID)     // minPriceCents, inStock, offerCount, discountPct
orders.InsertOne(sessCtx, order{status: "pending_payment", reservedUntil, items with keyIds})
```

Why this is safe:
- `findOneAndUpdate` is atomic on one document. Once a key is reserved with a future `reservedUntil`, it no longer matches the filter, so a second buyer can't take it.
- When an expired key is claimed again, its `orderId` changes to the new order. The old order no longer owns it, so the old order can't pay for it or release it (both steps filter on `orderId`).
- Inside a transaction, if two transactions touch the same key, Mongo raises a write conflict on one of them. `WithTransaction` retries that one from the start with a fresh read, and it picks another key or gets out of stock.
- If any line fails, the whole transaction rolls back: no half-reserved orders.
- Stock is **recomputed** from `keys`, never `$inc`'d. With expired keys coming back on their own, `+1/-1` bookkeeping would drift (a key would be counted out twice: once by the old order, once by the new). A recount can't drift.
- Atlas M0 is a replica set, so transactions work.

**Step 2 — pay:** `POST /api/orders/{id}/pay` runs a second transaction:
- Order must be `pending_payment`, owned by the user, and `reservedUntil > now`.
- Success: `keys.UpdateMany({ orderId, status: "reserved", reservedUntil: { $gt: now } }, → sold)`. If `ModifiedCount` is less than the number of keys in the order, some were lost to expiry: abort, mark the order `expired`, return `409 ORDER_NOT_PAYABLE`. Otherwise copy codes into the order, `status: "paid"`, `soldCount += qty`, remove the paid offers' lines from the cart. Then recount stock (see "Stock recount" in section 11).
- Decline: `keys.UpdateMany({ orderId, status: "reserved" }, → available, unset orderId/reservedUntil)`, then for each `offerId`: `refreshOfferStock` + `refreshProductSummary`, `status: "failed"`.

Both paths recount in the **same transaction**, so the store listing is accurate as soon as the response returns.

**Step 3 — cleanup job (tidy-up only):** a goroutine runs at startup and every 60 s. It finds `pending_payment` orders past `reservedUntil`, sets them to `expired`, resets their still-owned keys to `available` (same `orderId` filter, so it never touches a key another order has since claimed), then calls `refreshOfferStock` + `refreshProductSummary` for each affected offer. Nothing depends on it for correctness. Without it:
- No double sale and no stuck keys (the claim filter handles expiry).
- Old orders stay `pending_payment` in the db. `GET /api/orders*` reports any `pending_payment` order past `reservedUntil` as `expired` at read time, so the user never sees the stale status.
- Stock may lag for abandoned checkouts only: from when `reservedUntil` passes until the next request that touches that product (a purchase, decline, or payment on any offer for it) or until the cleanup job runs. On Render's free tier the server sleeps when idle, so the job may not run for a long time. Buying is never affected: the product page, cart, and checkout all count claimable keys live and ignore the cached value.

**Proof:** `internal/orders/claim_integration_test.go` — compiled only with the `integration` build tag, so `go test ./...` never runs it and never needs a database.

```
# run integration tests
MONGODB_TEST_URI="mongodb+srv://..." go test -tags integration ./internal/orders/... -v -count=1
```

The test database is always `eneba_test`. The suite drops and recreates it before each test; it never touches `eneba_dev` or `eneba_prod`. The service takes a `now func() time.Time` so tests control time without sleeping. The cleanup job is **off** in all four tests.

1. **Race:** offer with 5 keys, 20 goroutines each order 1 key → exactly 5 succeed, 15 get `OUT_OF_STOCK`, all 5 key ids distinct. After the race, `offers.stock` and product `inStock` match the live key count.
2. **Expired reservation is reclaimed:** buyer A reserves all 5 keys. Clock advances past `reservedUntil`. Buyer B orders 5 → succeeds with the same 5 keys. A pays → `409 ORDER_NOT_PAYABLE`, A's order is `expired`. B pays → `paid`, gets all 5 codes. Stock is 0 after B pays.
3. **Race on expired keys:** offer with 3 expired reservations + 2 available keys, 10 goroutines each order 1 → exactly 5 succeed, all distinct.
4. **Cleanup doesn't steal:** after test 2 setup (B holds keys, A is expired), run cleanup once → B's keys stay `reserved` to B, A's order is `expired` in db, `offers.stock` = 0 (B still holds them).

## 5b. Guest cart → server cart at login

Cart lines are keyed by **`offerId`**, not `productId`. The same game from two sellers is two separate lines, because they are different purchases at different prices.

On login or register, the client sends the guest cart to `POST /api/cart/merge` and clears the guest cart only after that call succeeds. If it fails, the guest lines stay, and the next page that shows the cart sends them again (safe, see below). Rules per guest line:

| Case | Result |
|---|---|
| Offer only in the guest cart | Added to the server cart |
| Offer only in the server cart | Kept as is |
| **Same offer in both** | Quantity = **the larger of the two**, not the sum |
| Same product, different seller | Both lines kept |
| Offer gone or inactive | Dropped, shown as a one-time notice |
| Quantity above live stock or above 10 | Capped |

Why the larger quantity and not the sum:
- The usual way to get the same offer in both carts is one person adding it twice: once while signed in on another visit, once as a guest. Adding the two would silently double their order.
- `max` gives the same result if the merge request runs twice (network retry, or the fetch helper retrying after a token refresh). A sum would add the guest items again on each retry.
- The user sees the merged cart and can raise the quantity in one click; they'd have to notice and lower it if we guessed high.

The merge runs in a transaction (read the cart, merge in Go, write it back), so two tabs logging in at once can't overwrite each other's lines; the loser gets a write conflict and retries.

---

## 6. Folder structure

```
/
├─ PLAN.md  API.md  CLAUDE.md  ASSIGNMENT.md  README.md
├─ design-reference/
│
├─ server/
│  ├─ cmd/
│  │  ├─ api/
│  │  │  ├─ main.go               # wire only: config → db → router → cleanup job → ListenAndServe
│  │  │  └─ routes.go             # global middleware, 404/405, /healthz, mounts each feature under /api
│  │  └─ seed/
│  │     ├─ main.go               # entry: flags, db connect, clear, insert
│  │     ├─ data.go               # fake titles, sellers, block configs, cover palettes
│  │     └─ build.go              # turns data.go into documents (fixed random seed, so runs repeat)
│  ├─ atlas/
│  │  ├─ products_search.json     # Atlas Search index definition
│  │  └─ atlas.go                 # go:embed of the JSON, used by db/search_index.go
│  ├─ internal/
│  │  ├─ platform/                # shared infrastructure, no business logic
│  │  │  ├─ config/
│  │  │  │  └─ config.go          # env parsing + startup validation
│  │  │  ├─ db/
│  │  │  │  ├─ connect.go         # Mongo client + ping
│  │  │  │  ├─ indexes.go         # EnsureIndexes called at startup
│  │  │  │  └─ search_index.go    # create/check Atlas Search index
│  │  │  ├─ httpx/
│  │  │  │  ├─ respond.go         # JSON helpers: OK, Created, NoContent
│  │  │  │  ├─ decode.go          # DecodeJSON: 1 MB cap, bad JSON → 400
│  │  │  │  └─ errors.go          # AppError type, error codes, WriteError
│  │  │  └─ middleware/
│  │  │     ├─ logger.go          # structured request logging
│  │  │     ├─ recoverer.go       # panic → 500
│  │  │     └─ cors.go            # chi/cors config from CORS_ORIGINS env
│  │  │
│  │  ├─ auth/
│  │  │  ├─ model.go              # User, RefreshToken, RevokedToken
│  │  │  ├─ repository.go         # interfaces: UserRepo, RefreshRepo, RevokedRepo
│  │  │  ├─ repository_mongo.go   # MongoDB implementations
│  │  │  ├─ tokens.go             # sign/parse JWT, random refresh token, sha256 hash
│  │  │  ├─ service.go            # Register, Login, Refresh, Logout — no Mongo calls here
│  │  │  ├─ service_test.go       # in-memory repos + fake clock; all auth test cases
│  │  │  ├─ middleware.go         # RequireAuth: verify sig → check jti → put userID in ctx
│  │  │  ├─ middleware_test.go    # status + error code through the real routes; rate limit
│  │  │  ├─ handler.go            # HTTP only: decode body, call service, write response
│  │  │  └─ routes.go             # Routes(): /register, /login, /refresh, /logout, /me → mounted at /api/auth
│  │  │
│  │  ├─ products/
│  │  │  ├─ model.go              # Product, ProductCard, Offer, Seller
│  │  │  ├─ repository.go         # Repository (products + offers, sellers via $lookup), StockCounter
│  │  │  ├─ repository_mongo.go
│  │  │  ├─ service.go            # List, Filters, Detail, Summarize — no Mongo calls here
│  │  │  ├─ cover.go              # deterministic SVG from name + palette (used by seed too)
│  │  │  ├─ handler.go            # list (hands ?q= to the search service), filters, detail, cover
│  │  │  └─ routes.go             # Routes(): /, /filters, /{slug}, /{slug}/cover.svg → mounted at /api/products
│  │  │
│  │  ├─ search/
│  │  │  ├─ model.go              # SearchResult, Suggestion
│  │  │  ├─ repository.go         # SearchRepo interface (Atlas Search pipelines)
│  │  │  ├─ repository_mongo.go
│  │  │  ├─ service.go            # Search, Autocomplete
│  │  │  ├─ handler.go            # autocomplete only; ?q= on /api/products goes through products/handler.go
│  │  │  └─ routes.go             # Routes(): /autocomplete → mounted at /api/search
│  │  │
│  │  ├─ home/
│  │  │  ├─ model.go              # Block, configs, Banner, Category, LayoutEntry, Resolved
│  │  │  ├─ repository.go         # Repository + BlockService interfaces
│  │  │  ├─ repository_mongo.go
│  │  │  ├─ service.go            # Layout, Resolve (2 s timeout) — dispatches by type through the registry
│  │  │  ├─ blocks/               # each implements home.BlockService
│  │  │  │  ├─ registry.go        # type → service; built in main.go (blocks imports home, so home can't build it)
│  │  │  │  ├─ hero.go
│  │  │  │  ├─ categories.go
│  │  │  │  └─ rail.go            # bestsellers, deals, new_releases, gift_cards: one service, four entries
│  │  │  ├─ handler.go            # layout, one block, ?fail=1 in dev
│  │  │  └─ routes.go             # Routes(): /, /blocks/{id} → mounted at /api/homepage
│  │  │
│  │  ├─ cart/
│  │  │  ├─ model.go              # Cart, Item (stored); View, Line (API's Cart, CartLine)
│  │  │  ├─ repository.go         # Repository (carts + offer lookup), StockCounter
│  │  │  ├─ repository_mongo.go
│  │  │  ├─ service.go            # Quote, Get, Upsert, Remove, Merge — no Mongo calls here
│  │  │  ├─ service_test.go       # merge rules: same offer max qty, idempotent, capped
│  │  │  ├─ handler.go
│  │  │  └─ routes.go             # Routes(): /quote public; /, /items/{offerId}, /merge behind RequireAuth → mounted at /api/cart
│  │  │
│  │  ├─ orders/
│  │  │  ├─ model.go              # Order, Item, Payment, Card; View, ItemView (API shapes)
│  │  │  ├─ repository.go         # Repository (orders + Transaction), Keys, Catalog, Carts interfaces
│  │  │  ├─ repository_mongo.go
│  │  │  ├─ service.go            # Create, Pay, Get, List, expire, refreshStock — runs the transaction steps
│  │  │  ├─ service_test.go       # card checks, expired-at-read-time, owner checks; no db
│  │  │  ├─ claim_integration_test.go  # //go:build integration — 6 tests against eneba_test db
│  │  │  ├─ payment.go            # checkCard (format only), simulate (decline magic number)
│  │  │  ├─ cleanup.go            # RunCleanup goroutine + CleanupOnce: expire stale pending_payment orders
│  │  │  ├─ handler.go
│  │  │  └─ routes.go             # Routes(): /, /{id}, /{id}/pay, all behind RequireAuth → mounted at /api/orders
│  │  │
│  │  └─ keys/
│  │     ├─ model.go              # Key (available | reserved | sold)
│  │     ├─ repository.go         # Repository: CountClaimable, Claim, Sell, Codes, Release; claimable() filter lives here
│  │     └─ repository_mongo.go
│  │
│  ├─ Dockerfile
│  ├─ .dockerignore
│  ├─ .env.example
│  └─ go.mod
│
└─ client/
   ├─ .env.example
   ├─ src/
   │  ├─ app/                     # routes and layouts only — no data fetching, no logic
   │  │  ├─ globals.css           # Tailwind v4 @theme tokens from notes.md
   │  │  ├─ layout.tsx            # server: html shell, Metropolis font, <Providers>
   │  │  ├─ (shop)/               # main header + footer shell (server layout)
   │  │  │  ├─ layout.tsx
   │  │  │  ├─ page.tsx           # home: renders <HomeView /> from features/home
   │  │  │  ├─ store/
   │  │  │  │  └─ page.tsx        # renders <StoreView />
   │  │  │  ├─ product/
   │  │  │  │  └─ [slug]/
   │  │  │  │     └─ page.tsx     # renders <ProductView />
   │  │  │  └─ orders/
   │  │  │     └─ page.tsx        # order history
   │  │  ├─ (checkout)/           # stepper header + short footer: cart, checkout, order and key
   │  │  │  ├─ layout.tsx
   │  │  │  ├─ cart/
   │  │  │  │  └─ page.tsx        # step 1
   │  │  │  ├─ checkout/
   │  │  │  │  └─ page.tsx        # step 2
   │  │  │  └─ orders/
   │  │  │     └─ [id]/
   │  │  │        └─ page.tsx     # step 3: order and key
   │  │  └─ (auth)/               # dark split layout (from login.png / signup.png)
   │  │     ├─ layout.tsx         # logo, "Hi! It's nice to see you!", card slot, demo footer
   │  │     ├─ login/
   │  │     │  └─ page.tsx
   │  │     └─ register/
   │  │        └─ page.tsx
   │  │
   │  ├─ features/                # feature slices: each owns its components, hooks, api calls
   │  │  ├─ home/
   │  │  │  ├─ components/        # home-view.tsx, home-blocks.tsx, hero-block.tsx, category-row-block.tsx, category-grid-block.tsx, category-icon.tsx, product-rail-block.tsx, block-shell.tsx, scroll-row.tsx
   │  │  │  ├─ hooks/             # use-homepage.ts, use-block.ts
   │  │  │  └─ api.ts             # getHomepageLayout, getBlock
   │  │  ├─ store/
   │  │  │  ├─ components/        # store-view.tsx, store-content.tsx, filter-panel.tsx, filter-section.tsx, check-list.tsx, price-range.tsx, active-filters.tsx, sort-select.tsx, product-results.tsx, pagination.tsx
   │  │  │  ├─ hooks/             # use-products.ts, use-filters.ts, use-store-params.ts (URL read/write)
   │  │  │  └─ api.ts             # getProducts, getFilters
   │  │  ├─ search/
   │  │  │  ├─ components/        # search-box.tsx (reads q from the URL), search-field.tsx, autocomplete-dropdown.tsx
   │  │  │  ├─ hooks/             # use-autocomplete.ts, use-debounce.ts
   │  │  │  └─ api.ts             # autocomplete
   │  │  ├─ product/
   │  │  │  ├─ components/        # product-view.tsx, breadcrumbs.tsx, product-hero.tsx, price-box.tsx, value-picker.tsx, offer-list.tsx, offer-row.tsx, offer-price.tsx, add-button.tsx, also-viewed.tsx, product-details.tsx
   │  │  │  ├─ hooks/             # use-product.ts, use-offer-actions.ts (add, buy now), use-add-cheapest.ts (card "Add to cart"), use-gift-card-family.ts, use-also-viewed.ts
   │  │  │  └─ api.ts             # getProduct, listProducts
   │  │  ├─ cart/
   │  │  │  ├─ components/        # cart-popover.tsx, cart-page.tsx, cart-line.tsx, cart-summary.tsx, cart-states.tsx, added-to-cart-modal.tsx, merge-notice.tsx
   │  │  │  ├─ hooks/             # use-cart.ts (switches guest ↔ server cart by auth state), use-guest-cart-view.ts, use-server-cart.ts, cart-view.ts (shared type), use-add-to-cart.ts, use-guest-cart-sync.ts
   │  │  │  ├─ store.ts           # Zustand guest cart + localStorage persist
   │  │  │  ├─ added-modal-store.ts # which "Added to cart" message is open
   │  │  │  ├─ merge.ts           # mergeGuestCart: POST /merge, clear guest lines only on success
   │  │  │  ├─ merge-notice-store.ts # how many lines the last merge dropped (one-time notice)
   │  │  │  ├─ add-item.ts        # add one to the guest or server cart; used by every add button
   │  │  │  └─ api.ts             # quoteCart, getCart, upsertItem, removeItem, mergeCart
   │  │  ├─ checkout/
   │  │  │  ├─ components/        # checkout-view.tsx, checkout-line.tsx, payment-form.tsx, order-summary.tsx
   │  │  │  ├─ hooks/             # use-checkout.ts, use-show-order.ts, use-resume-order.ts
   │  │  │  ├─ held-order.ts      # sessionStorage: the order Pay should finish
   │  │  │  ├─ schemas.ts         # Zod: payment card
   │  │  │  └─ api.ts             # createOrder, payOrder
   │  │  ├─ orders/
   │  │  │  ├─ components/        # orders-view.tsx, order-list.tsx, order-detail.tsx, order-status.tsx, key-reveal.tsx
   │  │  │  ├─ hooks/             # use-orders.ts, use-order.ts
   │  │  │  └─ api.ts             # getOrders, getOrder
   │  │  └─ auth/
   │  │     ├─ components/        # login-form.tsx, register-form.tsx, auth-card.tsx, form-bits.tsx (show password, form error), password-rules.tsx, account-menu.tsx, require-auth.tsx
   │  │     ├─ hooks/             # use-me.ts, use-sign-in.ts (tokens → merge → next), use-logout.ts
   │  │     ├─ schemas.ts         # Zod: login, register, PASSWORD_RULES
   │  │     ├─ next-path.ts       # safeNext: only same-site paths for ?next=
   │  │     ├─ server-errors.ts   # API error → form field or form message
   │  │     └─ api.ts             # register, login, logout, getMe (refresh lives in lib/client.ts)
   │  │
   │  ├─ components/
   │  │  ├─ ui/                   # shared design system only
   │  │  │  ├─ button.tsx
   │  │  │  ├─ input.tsx          # auth field: dark box, placeholder, error line
   │  │  │  ├─ badge.tsx
   │  │  │  ├─ price.tsx          # cents → "€15.49" display
   │  │  │  ├─ product-card.tsx   # used in home, store, product page
   │  │  │  ├─ skeleton.tsx
   │  │  │  ├─ modal.tsx          # native <dialog>
   │  │  │  ├─ platform-icon.tsx  # lucide icon per platform (card, product page)
   │  │  │  ├─ error-state.tsx
   │  │  │  └─ error-boundary.tsx
   │  │  └─ layout/               # shell components (mostly server)
   │  │     ├─ logo.tsx           # server: wordmark linking home, used by both headers
   │  │     ├─ header.tsx         # server: logo + static nav links shell
   │  │     ├─ checkout-header.tsx # server: logo, steps, "English EU | EUR"
   │  │     ├─ checkout-steps.tsx # client: 3-step bar, step read from the URL
   │  │     ├─ checkout-footer.tsx # server: demo disclaimer line only
   │  │     ├─ nav-bar.tsx        # server: categories nav
   │  │     ├─ cart-button.tsx    # client: cart icon + count badge
   │  │     └─ footer.tsx         # server: demo disclaimer
   │  │
   │  ├─ lib/
   │  │  ├─ api-url.ts            # API base URL, so components/ui can build cover URLs without client.ts
   │  │  ├─ client.ts             # apiFetch + single-flight refresh across tabs
   │  │  ├─ tokens.ts             # read/write/clear tokens in localStorage
   │  │  ├─ query-client.ts       # TanStack Query client setup
   │  │  ├─ providers.tsx         # QueryClientProvider + other wrappers
   │  │  └─ format.ts             # formatCents, regionColor, discountLabel
   │  │
   │  └─ types/
   │     └─ api.ts                # all types from API.md (User, ProductCard, Order, …)
   │
   └─ next.config.ts  package.json  tsconfig.json
```

**Server notes:**
- `cmd/api/main.go` only wires: it imports features, connects them to chi routes, then calls `http.ListenAndServe`. No business logic.
- `internal/platform/` has zero business logic. It never imports any feature package.
- Handlers never import the `repository_mongo.go` or the mongo driver directly. They call service methods only.
- Services never import chi or net/http. They call repository interfaces only.
- Repositories receive a `*mongo.Database` and return domain types. No HTTP concepts inside.
- `keys/` has no service or handler — it is a dependency of `orders/` (and of `products/`, `cart/` for live stock), which inject `keys.MongoRepo` behind their own small interfaces.
- Each feature's `routes.go` has `func (h *Handler) Routes() chi.Router` with paths relative to its mount point. `cmd/api/routes.go` mounts them: `api.Mount("/auth", authHandler.Routes())`. A feature that needs auth on some routes takes the `RequireAuth` middleware as a handler dependency and applies it inside its own `Routes()`.
- Search has no route under `/api/products`: chi can't mount two routers on one prefix, so the products handler takes a search interface and calls it when `q` is set.

**Client notes:**
- `app/` page files are server components that render one view component from `features/`. No query calls or business logic in `app/`.
- `features/` owns everything for that slice: components that import this feature's `api.ts` and hooks are co-located here.
- `components/ui/` has no imports from `features/` or `lib/client.ts`. Pure display.
- `components/layout/` server components import nothing from `features/`; `cart-button.tsx` is the only client component there. The header's `SearchBox` lives in `features/search` and the `(shop)` layout passes it in as the header's `search` prop.

---

## 7. Environment variables

**server/.env.example**

| Var | Example | Notes |
|---|---|---|
| `APP_ENV` | `development` | `production` on Render; turns off `?fail=1` |
| `PORT` | `8080` | Render sets this itself |
| `MONGODB_URI` | `mongodb+srv://user:pass@cluster0.xxxx.mongodb.net/?retryWrites=true&w=majority` | Same cluster for both envs |
| `MONGODB_DB` | `eneba_dev` | `eneba_prod` on Render |
| `JWT_SECRET` | 64 random hex chars | `openssl rand -hex 32`; different per env |
| `JWT_ISSUER` | `eugene-api` | Checked on parse |
| `ACCESS_TOKEN_TTL` | `15m` | Go duration |
| `REFRESH_TOKEN_TTL` | `720h` | 30 days |
| `CORS_ORIGINS` | `http://localhost:3000,https://eugene.vercel.app` | Exact origins, comma list |
| `RESERVATION_TTL` | `10m` | How long checkout holds keys; after that they are claimable again |
| `BCRYPT_COST` | `12` | |

**client/.env.example**

| Var | Example |
|---|---|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` (Render URL in Vercel settings) |

The API refuses to start if `JWT_SECRET` is shorter than 32 bytes or `MONGODB_URI`/`MONGODB_DB` are missing. The seed script refuses `MONGODB_DB=eneba_prod` unless `--allow-prod` is passed.

Deploy notes:
- Atlas network access must allow `0.0.0.0/0`, because Render's free tier has no fixed outbound IP. The DB user has a strong password and access only to these two dbs.
- **Check before the first Render deploy:** Atlas → Security → Network Access lists `0.0.0.0/0` as Active. If it doesn't, the startup ping times out, the API exits with code 1, and Render restarts it in a loop. The log only says `mongo ping: ... server selection timeout`, not "IP not allowed".
- **Check after the first Render deploy:** the `ip` field in the request logs shows your own public IP, not Render's. `ClientIP` trusts only the last `X-Forwarded-For` entry; if Render ever puts a second proxy in front, that entry becomes the proxy's IP and every user shares one rate limit bucket.
- Vercel preview deploys get random URLs. CORS lists only the production URL, so previews won't reach the API. Accepted.
- Render free tier sleeps after ~15 min idle; the first request then takes up to a minute. Skeletons cover it; README says so.

---

## 8. Auth details

**Tokens**
- Access: HS256 JWT, claims `sub` (user id), `jti` (16 random bytes, hex), `iat`, `exp` (+15 min), `iss`.
- Refresh: 32 random bytes from `crypto/rand`, base64url. DB stores `sha256(token)` only. SHA-256 (not bcrypt) is fine here because the token is already 256 bits of randomness; there is nothing to brute force, and we need an exact-match index lookup.

**Refresh flow** (`POST /api/auth/refresh`):
1. `h = sha256(token)`.
2. Pick the new token's `_id`, then `FindOneAndUpdate({ tokenHash: h, usedAt: null, expiresAt: { $gt: now } }, { $set: { usedAt: now, replacedBy: newId } })`. Atomic, so two parallel calls with the same token can't both succeed.
3. Matched → insert the new token with that `_id` (same `familyId`), issue a new access token, return the pair.
4. Not matched → `FindOne({ tokenHash: h })`. If it exists and `usedAt` is set → **reuse**: `DeleteMany({ userId })`, return `401 TOKEN_REUSED`. Otherwise → `401 UNAUTHORIZED`.
5. After step 3's insert, check the old token still exists. Used tokens are never deleted except by a reuse cleanup or a logout of that family, so if it's gone, one of those ran between step 2 and the insert, and the new token was written after it. Delete the family and return `401 UNAUTHORIZED`. Without this, a thief whose replay won the race would keep a new token that the cleanup missed. Cost: one indexed read per refresh. Test: `TestRefreshReuseCleanupBetweenUseAndInsert`.

The used token is marked, not deleted, so a later replay can still be recognized. The TTL index removes it after 30 days.

Known gap: reuse detection deletes refresh tokens, but access tokens already issued stay valid until they expire (≤ 15 min). Closing that would need a per-user "tokens before T are invalid" check. Not worth it here; noted.

**Middleware** (`RequireAuth`): parse `Bearer` → verify signature, `alg == HS256`, `iss`, `exp` → `revoked_tokens.FindOne({ _id: jti })` → put `userID` and `jti` in context. Expired → `TOKEN_EXPIRED`, revoked → `TOKEN_REVOKED`.

**Client fetch helper** (`lib/client.ts`):
- `apiFetch(path, opts)` adds `Authorization` from localStorage.
- On `401` (and the call wasn't `/auth/login`, `/auth/register` or `/auth/refresh`): `await refreshOnce()`, then retry the original request once. `/auth/me` and `/auth/logout` do refresh: they use the access token like any other call, and `useMe` on a protected page would otherwise treat an expired token as logged out.
- Only a `401` from `/refresh` ends the session. A `429`, `5xx` or network error is thrown as an error and the tokens are kept, so a server blip never logs the user out.
- `refreshOnce()` keeps one shared promise. Calls that fail while it's pending await the same promise; it's cleared when settled. So 5 parallel 401s → 1 refresh.
- Across tabs: the refresh runs inside `navigator.locks.request("auth-refresh", …)`. After getting the lock, it re-reads localStorage; if another tab already rotated the token, it uses that one instead of refreshing again. Without this, two tabs could both send the same refresh token and trigger the reuse alarm, logging the user out.
- Refresh fails → clear tokens, clear user query cache, `window.location.href = "/login?next=<current>"`.

**Protected pages:** `RequireAuth` wraps `/checkout` and `/orders*`. It calls `useMe()` (`GET /api/auth/me`); while loading it shows a skeleton, on no user it does `router.replace("/login?next=…")`. This is client-side only because tokens live in localStorage, which the Next server can't see. The real protection is the API returning 401.

**Tests** (`service_test.go` and `middleware_test.go`, no database: stores are interfaces with in-memory fakes, and the service takes a `now func() time.Time`):
- Expired access token → middleware 401 `TOKEN_EXPIRED`.
- Revoked jti → 401 `TOKEN_REVOKED`; a fresh token for the same user still works.
- Rotation: refresh returns a new pair; old refresh token is marked used; new one works.
- Reuse: use token A → get B; use A again → `TOKEN_REUSED`, and B no longer works either.
- Also: wrong signature, `alg: none`, wrong issuer, logout then reuse access token.

Trade-off we accept: localStorage tokens can be read by any script that runs on the page (XSS). httpOnly cookies avoid that but need same-site setup across Vercel and Render. The spec chose localStorage; we keep the access token short-lived and add no third-party scripts.

---

## 9. Scope

### Go deep

| Area | What "deep" means here |
|---|---|
| **Purchase integrity** | Transactional reservation, per-key atomic claim, reservations that expire on their own (no job needed), concurrency tests proving no double sale, including on expired holds. This is what a key marketplace can't get wrong. |
| **Auth** | Rotation, reuse detection, jti revocation, TTL cleanup, single-flight refresh across requests and tabs, tests. |
| **Search** | Atlas Search with typo tolerance, partial match, autocomplete in the header, filters + sort combined with relevance, URL-driven store state (shareable links, back button works). |
| **Homepage blocks** | Layout + per-type services on the server; per-block query, skeleton, error, retry, and error boundary on the client; a dev switch to show one block failing. |
| **Eneba look** | Product card, header, rails, cart popover, added-to-cart modal, checkout stepper, auth split layout, store and product pages from their screenshots, all from tokens in notes.md, responsive at the 4 breakpoints. |

### Keep simple

| Area | Simple version | Why |
|---|---|---|
| Payment | Card form, format check, one magic number declines | Spec allows simulation; real provider adds keys, webhooks, no new learning about our domain |
| Currency/language | EUR, English only | Shown in header as static text like Eneba |
| Sellers | Seeded, no seller accounts | Seller onboarding is the stretch goal |
| Keys | Fake codes, stored as plain text | A real system would encrypt them at rest; noted, not built |
| Key delivery | Shown on the order page; no email | No email provider needed |
| Cashback, service fee | Display only (cashback = `cashbackPct` of price, fee "--") | Matches screenshots without a wallet system |
| Wishlist, reviews, social login, newsletter, "someone just purchased" toasts, countdown widget | Not built | Not in the assignment; time goes to the core flow |
| Admin | None; data changes go through the seed script | |
| Images | Generated SVG covers, lucide icons for platforms/categories | No copyrighted art, no brand logos |

### Stretch (only if time allows)
1. Seller offers: `POST /api/seller/offers` with key upload, a `role: "seller"` on users.
2. Wishlist (heart on cards, `wishlists` collection). See future improvement 14 for the card heart.
3. "Recently viewed" block (client-only, localStorage).

---

## 10. Build order

Each phase ends with a check you can run. Tag `✅` in this file when done.

| # | Phase | Done when |
|---|---|---|
| 0 ✅ | Verify `.gitignore` covers `.env*` (except `.env.example`), `.next/`, `bin/` | All listed paths are ignored |
| 1 ✅ | Server skeleton: `go mod init`, config, chi, CORS, logger, recoverer, JSON error helper, Mongo connect, `/healthz` | `curl localhost:8080/healthz` → `{"status":"ok","db":"ok"}` |
| 2 ✅ | Indexes + seed: sellers, ~60 products (fictional titles), 2–4 offers each, 5–20 keys per offer, SVG covers, homepage blocks | `go run ./cmd/seed --reset`; docs visible in Atlas `eneba_dev` |
| 3 ✅ | Catalog API: list with filters/sort/pagination (no `q`), filters, product detail + offers, cover route | curl each; open a `cover.svg` in the browser |
| 4 ✅ | Atlas Search: create index from JSON, `q` on `/api/products`, `/api/search/autocomplete` | Typo query and partial query return the right product |
| 5 ✅ | Auth part 1: users, register/login, JWT, middleware, `/me`, rate limit + tests for expired/revoked | `go test ./internal/auth/...` green; curl register → me |
| 6 ✅ | Auth part 2: refresh rotation, reuse detection, logout, TTL indexes + tests | Tests for rotation and reuse green |
| 7 ✅ | Homepage: layout endpoint, registry, 6 block services, `?fail=1` | Each block id returns data; `?fail=1` returns 503 |
| 8 ✅ | Cart API: quote, get, put, delete, merge | curl flow: merge → put → get shows live prices |
| 9 ✅ | Orders: reserve transaction with expiry-aware claim, simulated pay, cleanup job, list/detail, `claim_integration_test.go` | `go test ./internal/orders/...` green (unit only); `go test -tags integration ./internal/orders/...` → all 4 integration tests green (plus 2 more: decline then pay, and a new order replacing the user's pending one); decline puts stock back and recounts immediately |
| 10 ✅ | Client skeleton: create-next-app (it refuses a folder with unknown files, so move `client/.env.example` aside and put it back after), Tailwind tokens, Metropolis font, `(shop)` layout with Header/Footer (server), Providers, `apiFetch` with single-flight refresh | Page renders with Eneba header and footer; demo disclaimer in footer |
| 11 ✅ | Homepage client: `HomeBlocks`, hero, category row, product rails, `ProductCard` | Home matches screenshots; `?fail=<id>` breaks only one block |
| 12 ✅ | Store + search (match `store.png`, `store2.png`): filter panel, sort, pagination in URL, header `SearchBox` with autocomplete | Typo in header → suggestions → Enter → results page |
| 13 | Product page (match `product_details1–3.png`) + cart: offers table, add to cart, added-to-cart modal, guest Zustand cart, cart popover, cart page | Guest adds 2 items, reloads, cart is still there |
| 14 | Auth UI: login/register (RHF + Zod, live password rules), `AccountMenu`, logout, merge on login (section 5b), `RequireAuth` | Guest cart survives login; same offer in both carts ends with the larger quantity; `/orders` redirects when logged out |
| 15 | Checkout: stepper layout, create order, payment form, decline path, order + key page, order history. When pay returns `409 ORDER_NOT_PAYABLE`, reload the order with `GET /api/orders/{id}`: if it is already `paid`, show its keys instead of an error (the pay went through but its response was lost, e.g. a crash after commit or a dropped connection) | Full flow from home to key; decline card returns to cart with stock intact; a second pay on a paid order shows the keys, not an error |
| 16 | Deploy: Dockerfile, Render service, Atlas `eneba_prod` seed + search index, Vercel project, CORS | Full flow works on the public URLs |
| 17 | Polish + check: responsive pass, empty states, README (setup, decisions, limits), check every line of ASSIGNMENT.md | Every assignment item ticked in README |

| 15b | **Parked until after phase 15** (owner's call): "Cheap games" promo page on the block system. A `page` field on `homepage_blocks` (`"home"` by default) and `?page=` on the layout endpoint; route `/deals/cheap-games`; a `promo_hero` block (heading, text, yellow button, generated SVG collage of our covers); rails reusing the `Rail` service with a max price ("Games / Xbox games / Nintendo games / Gift cards under €10", each "Show all" to the same store filters); same band alternation; seeded blocks; nav "Cheap Games" points there. No Trustpilot or real brands. Ask the owner first if it changes existing homepage behaviour. | Page renders from seeded blocks; homepage unchanged |

If behind schedule, cut in this order: category row block → gift cards block → cart popover (keep cart page) → order history list (keep order page).

---

## 11. Decisions made

| Question | Decision |
|---|---|
| App name and logo | **Eugene**. Text wordmark + small lucide icon. |
| Product titles | Made-up game titles ("Starfall Odyssey"). Platform names (Steam, PSN, Xbox) as plain text only, no logos. |
| Signed-in cart | Server cart; guest cart merges in at login (section 5b). |
| Declined payment | Release the keys, mark the order `failed`, cart stays so the user can check out again. |
| Abandoned reservations | Expire on their own through `reservedUntil` in the claim filter. The background job is cleanup only (section 5). |
| Go module name | `eugene`. Short imports (`eugene/internal/...`); the module is never imported by another repo. |
| Loading `server/.env` | `joho/godotenv`. Go has no built-in loader, and `source .env` breaks on the `&` in the Mongo URI. It never overrides vars already set, so Render's values win. |
| Where routes live | Each feature's `internal/<feature>/routes.go` returns a `chi.Router`; `cmd/api/routes.go` mounts them under `/api` and holds only global middleware, 404/405 and `/healthz` (too small to be a feature). A feature's paths sit next to its handlers, and `main.go` stays wiring only. |
| Client IP | Own `ClientIP` middleware, not `chimw.RealIP`. RealIP trusts `X-Real-IP`, `True-Client-IP` and the first `X-Forwarded-For` entry, all of which the client sets, so a fake header per request would dodge the login rate limit. In production we take the last `X-Forwarded-For` entry (the one Render's proxy added); locally we ignore all those headers and use the TCP address. |
| Product page stock | `GET /api/products/{slug}` counts claimable keys live per offer and rebuilds the price summary from that count with `products.Summarize` (the same function the seed uses, and the purchase transaction will use). The cached `offers.stock` is only used by listings. |
| Bad store query params | `400 BAD_REQUEST` naming the param (unknown platform, `pageSize` over 60, `minPrice` above `maxPrice`), instead of ignoring them, so a broken link fails loudly. |
| Search `q` length | Trimmed, capped at 100 chars (`400` above). Fuzzy matching cost grows with every term, and no product name is that long. |
| Where `Searcher` lives | `products.Searcher` interface, implemented by `search.Service`. `search` imports `products` (for `ListQuery`, `ToCard`), so the interface has to sit in `products` to avoid an import cycle. |
| Unknown errors | `httpx.WriteError` logs them and returns `500 INTERNAL "Something went wrong"`, so driver messages never reach the browser. |
| Client hung up | When the browser drops a request (tab closed, reload, aborted fetch), Go cancels the request's context and the db call returns `context.Canceled`. `WriteError` logs that at debug level and returns `499` (nginx's "client closed request") with no body, instead of an error log and a 500 that would look like a server bug. Only when the request's own context is cancelled too: a `context.Canceled` from elsewhere in our code, with the client still waiting, stays a logged 500. Test: `TestWriteErrorClientClosed`. |
| Refresh token in phase 5 | Register and login already return and store a refresh token (hash only, new `familyId` per login), because API.md promises it. `/refresh` rotation, reuse detection and logout were added in phase 6. |
| Auth rate limit | One `httprate` limiter shared by `/register` and `/login`: 10 per minute per IP across both, not per route. `/me` is not limited: the client calls it on every protected page and in the header, and it already needs a valid token. `/refresh` (phase 6) gets its own, higher limiter, because every open tab refreshes every 15 minutes and must not be blocked by failed logins from the same IP. The key is `r.RemoteAddr` after our `ClientIP` middleware, not httprate's `KeyByRealIP`, which trusts headers the client sets. |
| Login timing | For an unknown email, login still runs one bcrypt compare against a dummy hash, so the response time doesn't reveal whether the email has an account. |
| Auth errors in the service | The service returns plain errors (`ErrTokenExpired`, `ValidationError`, …) and `auth.toHTTP` maps them to API.md codes, so the service never imports `net/http`. |
| Logout deletes the refresh family | Logout deletes every refresh token with the given token's `familyId` (and the caller's `userId`), not just that one token. If a thief rotated a stolen copy first, their newer token is in the same family, so logout ends it too. Other logins (other families) stay signed in. Same cost: one lookup, one `DeleteMany`. |
| Refresh write order | `/refresh` marks the old token used (and sets `replacedBy` to an id picked in advance) before inserting the new one, since the `userId` is only known after that first atomic step. If the insert fails, the user has to log in again. Rare, and safe. |
| `/refresh` rate limit | 60 per minute per IP, own counter. Each tab refreshes about once per 15 minutes, so this only stops floods. |
| Homepage rails | One `Rail` service, registered four times with a different sort, type filter and "see all" link, instead of four near-identical files. It calls `products.Service.List`, so a rail shows the same cards in the same order as the store page its link opens. The deals minimum (20%) is applied after the query: the rail sorts by discount, highest first, so dropping the tail equals filtering in the query, and `ListQuery` needs no field the store never uses. |
| Category counts | Taken from `products.Service.Filters` (one aggregation, in-stock only), so a tile's count matches what the store shows after the click. Each tile also gets an `href` built on the server from the block config, so the client never maps a category to a store filter itself. |
| Block errors | Bad or unknown block id → `404`. Any other failure (db error, timeout, bad config, type with no service) → `503 BLOCK_UNAVAILABLE`, logged with the block id. |
| Cart writes | Every write (put, delete, merge) goes through one `Repository.Update(userID, change)`: read the cart, run `change` in Go, save, inside one transaction. If another write changed the cart in between, Mongo raises a write conflict and the driver runs it again on the fresh cart. The service never touches Mongo, and the in-memory test repo does the same with a lock. Offers and stock are read before the transaction, so `change` only edits the cart it is given and gives the same result on retry. The first cart for a user is safe too: if two writes both try to insert it, Mongo reports the second as a `WriteConflict` (not a duplicate key on the unique `userId` index), so it is retried like any other clash. No extra duplicate-key handling, because it never fires; `cart/repository_mongo_integration_test.go` forces the worst order (other writer commits between our read and our write) and runs 20 parallel rounds to prove it. Run with `MONGODB_TEST_URI=... go test -tags integration ./internal/cart/... -v -count=1`. |
| Cart stores choices, not prices | The cart doc holds only `offerId`, `productId`, `quantity`. Price, name, seller and stock are looked up live on every read (one `$lookup` over at most 20 offers + one key count), so a cart can never show a stale price. |
| PUT doesn't cap at stock, merge does | `PUT` stores the number the user picked and the line shows `NOT_ENOUGH_STOCK`; changing it silently would hide that stock dropped. Merge caps, because there the user never picked the merged number. |
| Cart size | 20 lines, 10 per line. Protects the db and bounds the offer lookup a public `quote` can trigger. Merge drops extra guest lines with reason `CART_FULL`. |
| Repeated offer in a quote/merge body | `422`, not merged. The client keys its guest cart by `offerId`, so a repeat means a broken request. |
| Who runs the transaction | `orders.Service` holds the steps (claim, sell, release, recount, save); `orders.Repository.Transaction(ctx, fn)` wraps `WithTransaction`. In driver v2 the `ctx` handed to `fn` carries the session, and every repo uses the same client, so keys, products, carts and orders writes made with that `ctx` all join one transaction. The service stays free of `mongo` imports, and each repo stays a plain set of reads and writes. |
| Stock recount | After every claim, sale, release or expiry, `refreshStock` recounts **all** offers of each product in the order (not only the order's offers) from `keys`, then runs `products.Summarize` and saves `offers.stock` + the product summary. The summary (cheapest in-stock offer) depends on every offer, and the other offers' cached stock may be stale from an expired hold. A product has 2–4 offers, so this is a few small writes. |
| Order snapshot | Prices, names and sellers are read from the cart's live offers before the transaction and copied into the order. Items are rebuilt at the start of every transaction run, so a retried run never keeps key ids from the failed run. |
| Late payment | `Pay` checks `reservedUntil > now` before selling, and `Sell` only sells keys still held (`orderId`, `reserved`, future `reservedUntil`). If either fails, the transaction returns an error so nothing stays sold, and a second, small transaction marks the order `expired` and frees any keys it still owns. |
| Pay order of checks | Owner (404) → card format (422) → order state (409) → simulated payment. A stranger gets 404 whatever card they send, so order ids can't be probed. |
| Paid order and cart | Pay removes only the paid offers' lines (`$pull`), inside the pay transaction, instead of emptying the cart. A line added in another tab after checkout started is kept. |
| Card check | Format only: 12–19 digits, `MM/YY` not past its month, 3–4 digit CVC, a name. No Luhn check, because API.md promises "any valid-looking number pays" and testers type random digits. |
| Order number | `ORD-` + 6 chars from `crypto/rand`, from a 32-letter alphabet without 0/O/1/I. About a billion values; a clash would hit the unique index and give that one request a 500. Not retried: rare enough not to matter. |
| New order replaces pending ones | `Create` first expires the user's other `pending_payment` orders and releases their keys, in the same transaction, before claiming. Without it, a user who goes back from checkout and starts again holds two sets of keys, and on low stock their own first order makes the second fail with `OUT_OF_STOCK`. It also caps what one account can hold at one order's worth (20 lines × 10). Released first, so the new order can claim the same keys. If the user pays the old order in another tab at the same moment, both transactions write the same order and keys, so Mongo makes one retry: either the payment wins and the old order is no longer pending, or it is expired and the payment gets `409`. Test: `TestNewOrderReplacesUsersPendingOrder`. |
| Cleanup job | `go ordersSvc.RunCleanup(ctx, time.Minute)` in `main.go`: runs once at start, then every minute, at most 100 orders per run, and stops with the shutdown `ctx`. Each order goes through the same `expire` as a late payment. |
| `golang.org/x/crypto` version | Pinned at v0.54. Newer versions need Go 1.26, which would raise the `go` line in go.mod and the toolchain the Dockerfile needs. |
| Client tokens in one key | `localStorage["eugene.tokens"]` holds `{accessToken, refreshToken}` as one JSON value, so a read never gets the access token of one pair and the refresh token of another while another tab writes. |
| Tailwind tokens only | `globals.css` resets Tailwind's colors, fonts, text sizes, breakpoints and radii (`--color-*: initial` etc.) and defines only the notes.md values, so a default class like `bg-blue-500` or `rounded-lg` produces nothing. Breakpoints are Eneba's: `xs` 400, `sm` 600, `md` 930, `lg` 1200. |
| Header search before phase 12 | A plain `<form action="/store">` with an input named `q`. Enter already opens `/store?q=...`, and it works without JS. Phase 12 swaps in the autocomplete `SearchBox`. The header has no wishlist heart, since wishlist isn't built. |
| Header `SearchBox` | Lives in `features/search` (it calls the API), so the server `Header` can't import it. The `(shop)` layout passes `<SearchBox />` in as `Header`'s `search` prop. It is still a GET form to `/store`, so Enter works before JS loads. Suggestions: debounced 250 ms, min 2 chars, `retry: false` (the next key asks again). The dropdown copies Eneba's: rows with cover, a type pill, the name (2 lines), "From", the struck-through MSRP when discounted and the price, then "Show all results" (hidden when nothing matched, since it would open the same empty search). For that, autocomplete also returns `type`, `originalPriceCents` and `discountPct`. While open, a fixed overlay at z-40 dims the page and the form is lifted to z-50 above it. Up/Down and the mouse share one highlight, so they never light two rows; Up/Down wrap back to the text box. Enter opens the highlighted product, or `/store?q=` when none is highlighted; Escape, or a click anywhere else (the input's blur), closes it; the X clears the text and keeps focus. The input is `type="text"`, not `search`, so the browser's own clear button doesn't sit next to ours. The list's `mousedown` `preventDefault` keeps the input from blurring before a click lands. On phones the list runs edge to edge (`-inset-x-[17px]`: the page's 16px gutter plus the box's border). The box shows the URL's `q`, so `/store?q=...` opened directly has its text: `search-box.tsx` reads `q` inside its own Suspense boundary (Next needs one around `useSearchParams` on a static page; the fallback is the same box, empty) and renders `search-field.tsx` with `key={q}`, so the box starts fresh whenever `q` changes. |
| Store state in the URL | Filters, sort and page are only in the query string (`use-store-params.ts`), so links, reloads and Back all work. Each change is a `router.push`, so Back undoes it; any change except the page goes back to page 1. The query string is passed to `/api/products` as is, and a `400` (bad value in an old or edited link) shows its message with a "Clear filters" button instead of Retry. Defaults (`sort`, page 1) are dropped from the URL. |
| Store filter counts | The counts next to each option come from `/api/products/filters`, so they are for the whole in-stock store, not the current selection (ticking Steam doesn't change the genre counts). Faceted counts would need a `$facet` over the filtered list on every change; not worth it at 60 products. |
| Store filters vs Eneba | Region is a single select ("All regions"), like Eneba's "Country" select, though the API accepts a list. Type, platform and genre are checkboxes (genre shows 6, then "N more"). Price applies on Enter or when focus leaves both boxes, not per key; euros are turned into cents from the digits, never `parseFloat`. The chips above the grid ("Type (1)") remove that filter on click instead of opening Eneba's dropdown, since the panel already has the options. The filter panel has its own query: if it fails, the grid still works. |
| Store grid | Sidebar 192px + 4 cards of 224px with 20px gaps = 1168px, the page width inside its padding, so desktop shows 4 cards like store.png. Eneba's sidebar is wider, so "Price range (EUR)" is shortened to "Price (EUR)" to fit one line. Columns are `repeat(auto-fill, 224px)`, so narrower screens show fewer; below 930px the filters fold behind a "Filters" button. While a new filter loads, the old grid stays, dimmed (`keepPreviousData`). |
| Homepage `?fail=` | `/?fail=<blockId>` takes the block's Mongo id (from `GET /api/homepage`), as the API does. Only that block's request gets `?fail=1`; the API honors it only when `APP_ENV=development`, so the client doesn't check the env itself. |
| Block loading states | `BlockShell` owns the query, skeleton, error and Retry for every block type. While a Retry runs it shows the skeleton again, because TanStack keeps the old error until the new answer arrives and the error box would look stuck. A render bug is caught by an `ErrorBoundary` around each block, whose fallback is the same error box. |
| Platform strip below the cover | On Eneba the platform strip and cashback pill lie over the bottom of the cover. Our seeded SVG covers print the title there, so the strip is its own 30px row under the cover. Moving the title to the top would need a cover change and a reseed. |
| Category icons | The API sends a lucide icon name. The client imports only the icons the seed uses, in a small map, with `Tag` as fallback; importing lucide's full name map would ship every icon. |
| Product card hover | The card is a fixed 224×470 box with `overflow: hidden`. On hover the strip and info move up 116px (the button panel's height) and the "Add to cart" / "View offers" panel moves in from below, both with a 200ms `translate`, so nothing around the card shifts. The buttons are siblings of the card's link, not inside it, because a button or link inside a link is invalid HTML. The panel is `display: none` unless the device can hover (`can-hover` variant in `globals.css`), so on touch screens it can't be tapped and a tap opens the product page. Keyboard users get the same slide when focus enters the card. "Add to cart" takes an `onAddToCart` prop; home and store pass the "add cheapest offer" action (see "Card Add to cart" below). |
| Scroll arrows | One `ScrollRow` with two looks. Categories: thin white chevrons at the row's edges, always shown while the row overflows, dimmed and disabled at either end, and a fade on the right while there is more to scroll. Rails: white squares at 30% opacity over the cards, shown only while the mouse is on the rail's section (`group/rail`), and hidden at either end. Both scroll one visible width per click and exist only on devices that can hover; touch users swipe. With 8 categories the row fits from about 1030px up, so its arrows only show below that. |
| Homepage bands | Every block after the hero sits in a full-width band, alternating `brand-mid` (#5825cc) and `brand-darkest` (#1f0a4d); the content stays in the 1200px container. A `category_row` has no heading and leads into the block after it, so the two share one band (the row keeps only its top padding). The color comes from the block's place in the API layout, counting only blocks the client can draw and skipping the hero (which has its own colors), so adding, removing or reordering blocks in the db keeps the pattern. The 40px section spacing is padding inside each band, so the bands touch. Each band sets `--band` to its color, which the category row's edge fade uses. |
| Categories grid | Eneba's "Categories" section at the bottom is a new block type, `category_grid`. It reuses the `Categories` service (same tiles, counts and store links as `category_row`), so the server change is one registry line plus a seeded block of the 11 genres. The client draws 7 tiles (one desktop row); "Load more" draws 7 more from data it already has, with no extra request. Tiles are white at 10% so they show on either band color. |
| Product covers | Plain `<img>` of the API's `cover.svg`, not `next/image`, which refuses SVG unless `dangerouslyAllowSVG` is on. |
| Footer spacing | The footer has no top margin. Each page ends with its own 40px of space (the home page through its last band's padding), so a colored band runs right up to the footer. A margin there showed as a 40px strip of page background under the last band. Pages built later need their own bottom padding. |
| Guest cart storage | Zustand `persist` under `localStorage["eugene.guest-cart"]`, holding only `{offerId, quantity}` lines. Names, prices and stock come from `POST /api/cart/quote` on every change (the items are the query key), so a guest never sees a stale price. The same limits as the server (20 lines, 1–10 each) are kept on add, so the cart can always be quoted and, in phase 14, merged. `skipHydration` stops the store reading localStorage during the first render, which would differ from the server HTML; `useGuestCartSync` (mounted by the header's cart popover and the cart page) reads it after mount, and again on the `storage` event, so a second tab doesn't overwrite the first tab's additions with its old list. The header badge counts the local lines, so it changes at once without a request. |
| Card "Add to cart" | A card knows the product, not its offers, and cart lines are offers. The click loads `GET /api/products/{slug}` (fresh, `staleTime: 0`) and adds the first in-stock offer, which is the cheapest because the API sorts that way, so the price added is the card's "From" price. No in-stock offer left → the modal says "Sold out". No new endpoint needed. The modal opens at once with a spinner ("Adding to cart…") while that request runs, then the same dialog swaps to the result, so the box doesn't close and reopen or change size. If the user closes it during the wait, the add still happens but the modal doesn't pop back up. The product page's "+cart" knows its offer, so it adds at once with no loader. |
| Adding the same offer again | Raises its quantity by one (up to 10), like Eneba. A 21st line is refused and the modal says the cart is full. |
| Quantity buttons | +/− step from the saved quantity, not from the quote on screen, so quick clicks while a new quote loads all count. The old quote stays on screen, dimmed, until the new one arrives. |
| Product page "Buy now" | Adds one and opens `/cart`. Eneba goes to checkout; ours goes through the cart, whose "Checkout now" leads on (checkout needs sign-in). The "+cart" button next to it adds and shows the "Added to cart" modal. |
| Product page layout | Wider than the rest of the site, as on Eneba: `max-w-wide` (1600px, a `--container-wide` token next to `--container-page` 1200px). From 930px up: main column + 350px sidebar with a 30px gap. The sidebar cell spans both grid rows, so its `sticky` price box follows the whole main column. Below 930px it's one column, and grid order puts the price box between the hero and the offers. The price box shows the product's "From" price (old price, discount, cashback), rebuilt from live stock by the API, and a yellow "See offers" that scrolls to the list, where buying happens. "Recommended offers" holds the cheapest offer, "N other offers" the rest. Left out from Eneba: PayPal, Trustpilot, ScamAdviser, PC Mag and payment logos (real brands), wishlist heart and count, "people watching", region dropdown, "USD per €1" rates, trailer and reviews. The small "i" icons next to prices are left out too. Eneba's "Instant refunds" fact becomes "Key on your order page", since we don't refund. Per-offer discount uses the server's rule (`discountPct` in `format.ts`: whole percent, rounded down). Seller logos are a letter tile. |
| Offer row hover | With a mouse, the "+cart" button is 0 wide until the row is hovered (or focused by keyboard), then opens from its right edge so it seems to slide out from behind "Buy now", and "Buy now" turns solid yellow with dark text. Touch screens can't hover, so they always show both. "Buy now" has its own classes, not the shared outline style: that style's own `hover:` background beat the row's yellow while the pointer was on the button, leaving dark text on a dark box. |
| Gift card value tiles | Each value of a gift card is its own product, so the tiles are links. They come from `GET /api/products?type=gift_card&platform=…&region=…&inStock=false` (sold-out values keep their tile), sorted by MSRP, which for a gift card is its face value. Hidden for non-gift-cards, while loading, on error, and when the family has one value. The seed was widened to 14 cards in 4 families (Steam, PSN and Xbox Europe, Nintendo Europe), with PSN UK left as a single card to show the hidden case. The games come first in the seed's random order, so their data didn't change; the software's prices and stock did. |
| Gamers also viewed | No view history, so it shows other products with the same type and platform, best sellers first, through `/api/products`. It asks for `limit + 1` and drops the current product. "Load more" raises the limit by 5 (the old cards stay while the bigger list loads); it stops at 55 because `pageSize` caps at 60. "Is there more?" takes the current product off `total` when it's in stock, since the list is in-stock only. Hidden when empty or on error. |
| Added-to-cart modal | One `AddedToCartModal`, rendered by the cart popover, opened through a small Zustand store, so product cards on any page and the product page share it. Built on the native `<dialog>` (`components/ui/modal.tsx`): `showModal()` gives the dimmed page, focus kept inside, and Escape for free. |
| Cart popover | Opens on click (like Eneba), closes on the X, Escape, a click outside, or any link inside it. It has no `relative` parent of its own: it's placed against the header row (`right-4`, max width = screen − 32px), so on a phone it stays on screen instead of hanging off the icon. |
| Cart page | Uses the checkout layout (see "Checkout layout"). Its body has no Eneba screenshot, so it reuses the popover's line and totals pieces: lines on the left, totals box on the right (stacked below 930px). A line with a `problem` shows why, and "Checkout now" is replaced by a note until it's fixed (API.md, Cart). A `422` from the quote means the saved cart itself is broken, so the error offers "Empty cart" instead of Retry. |
| Checkout layout | `/cart`, `/checkout` and `/orders/{id}` sit in the `(checkout)` group, as on Eneba: logo, a 3-step bar and "English EU \| EUR", no search or nav bar, and a footer with only the demo line, so nothing pulls the buyer away mid-purchase. The order history list (`/orders`) stays in `(shop)`. The step comes from the URL (`usePathname` in `checkout-steps.tsx`: `/checkout` 2, `/orders/…` 3, else 1), so pages pass nothing. Done steps: mint circle with a check; current: mint circle and mint label; later: purple circle, dimmed. Below 600px the labels and the language text hide, leaving circles and lines. This layout has no header cart popover, so the cart page calls `useGuestCartSync` itself to load the saved guest cart. |
| Auth pages | `(auth)` layout from login.png / signup.png: brand-darkest page, logo top left, "Hi! It's nice to see you!" (a `text-display` token, 80px, measured from the screenshots; hidden below 930px), brand card on the right. Login is email + password instead of Eneba's magic link. Left out: Google/Facebook/Apple/Steam/Discord/Twitch/X buttons (real brands, and social login isn't built), the offers checkbox (no newsletter), and the terms line (replaced by a demo note). The footer is the checkout pages' demo line, since Eneba's links go to pages we don't have. Forms use React Hook Form + Zod with the server's rules (8–72 bytes, a digit or symbol by the same unicode classes), and the signup rules tick live as you type. Server errors go under their field (`fields`, `EMAIL_TAKEN`) or above the button (401, 429). `?next=` is read by the page (server) and passed in, so the form needs no `useSearchParams`; only paths on this site are kept (`//evil.com` → `/`). `safeNext` parses the value with `new URL` against a stand-in origin and keeps it only if the origin is unchanged, because a text check misses `/%09/evil.com`: the URL parser deletes tabs and newlines, which turns it into `//evil.com`. Found and fixed in phase 15. |
| Who is signed in | `useMe` (`["me"]`, `GET /api/auth/me`), `null` with no request when no tokens are saved. `staleTime: Infinity`: only login, logout and the other-tab listener change it. Click handlers (add to cart) check the saved tokens instead, because they need the answer at once. While `/me` loads, the header keeps the account menu's space empty, so a signed-in user doesn't see "Log in" flash. |
| After login | `useSignIn`: save tokens, drop any cached cart, set `["me"]`, `await` the merge, then `router.replace(next)`, so the page the user lands on already shows the merged cart. A failed merge doesn't block the login: the guest lines stay in localStorage, and `useGuestCartSync` (cart popover, cart page) sends them again when a signed-in page finds them. Repeating is safe because the server takes the larger quantity. On success only the lines that were sent are removed, so a line added in another tab while the request ran is kept for the next merge. Dropped lines show one notice (in memory only, so once). |
| Signed-in cart on screen | `useCart` returns the server cart when signed in, the guest cart otherwise, with the same shape, so the popover, cart page and modal didn't change. +/− change the cached cart at once, then send `PUT`; the writes share a TanStack mutation `scope`, so they go one at a time in click order (`PUT` sets a number, so an older one landing last would win), and only the last answer is put on screen. Every add button goes through `addItem`, which for a signed-in user reads the cart fresh (`staleTime: 0`), then `PUT`s quantity + 1 (capped at 10); a `422` means the cart is full. |
| Login or logout in another tab | `lib/providers.tsx` listens for the `storage` event on the tokens key and refetches `["me"]` and `["cart"]` when it goes from empty to set or back, so other tabs switch between guest and account cart without a reload. A token refresh changes the value without emptying it, so it's ignored. Known gap: logging into a different account in another tab without logging out first isn't noticed until a reload. |
| Logout | Calls `/logout` (failure ignored), clears tokens, then a full page load to `/`, which drops every cached query of that user. |
| `/orders` in phase 14 | Exists only behind `RequireAuth` (a guest goes to `/login?next=/orders`); phase 15 adds the list. |
| Checkout page | From checkout.png: "My cart" on the left; "Where to deliver the order?", a card panel and "Summary" on the right, stacked below 930px. Eneba's checkout.png is its cart step; ours is step 2 (payment), so the card panel is added and the delivery panel shows the account email read only. The lines are read only (Eneba has +/−), with an "Edit cart" link to step 1: a retry pays the order already reserved, so the cart must not change under it. Left out: Trustpilot and newsletter boxes, "Got discount code?", the Ukraine donation and "Inspired by your picks". Behind `RequireAuth`; it also runs `useGuestCartSync`, so guest lines a failed login merge left behind join before paying. |
| When keys are reserved | On "Pay", not when `/checkout` opens: the click sends `POST /api/orders`, then `POST /pay`. Opening the page reserves nothing, so a visitor who only looks holds no keys. The order id is kept (see "Held order") while the outcome is unknown (network error, 5xx) or the card failed the format check (422), so the next click pays the same order instead of reserving again. It is dropped once the order is final (paid, declined, expired) or on another 4xx. |
| Held order | The order id Pay should finish lives in `sessionStorage["eugene.held-order"]` (`held-order.ts`), not in memory, so a refresh mid-payment keeps it. It is set when `POST /api/orders` returns and cleared when the order is paid, declined, expired or a 4xx. When `/checkout` or `/cart` opens with one (and tokens are saved), `useResumeOrder` loads it once with `GET /api/orders/{id}`: paid → its key page (the pay went through, its answer was lost); `pending_payment` → kept, so Pay finishes it; failed, expired or 404 → cleared; network error or 5xx → kept for next time. No UI while it checks. sessionStorage, not localStorage: it is one tab's checkout, and closing the tab ends it. **Any server cart write clears it too** (quantity change, remove, add button, guest merge): the held order was reserved from the old cart, and paying it would charge for something no longer on screen. The next Pay reserves a new order, and the server expires the old one in the same transaction. Checked in headless Chrome: paid → key page from `/checkout` and `/cart`; pending kept; failed and unknown ids cleared; "+" on the cart clears it. |
| Pay returns 409 | `ORDER_NOT_PAYABLE` → `GET /api/orders/{id}`. Paid → open the order page with its keys: an earlier pay went through but its answer was lost, and the retry above is what finds it. Not paid → the hold ran out; the message says so and the next click reserves again. |
| Declined card | `402` → `/cart?payment=declined`, whose page reads the param on the server (no `useSearchParams`) and shows "Your card was declined…" above the unchanged cart. The server already released the keys and recounted stock in the same request. |
| After paying | The paid order goes into the query cache under `["order", id]`, then `router.replace("/orders/{id}")`, so the keys show with no extra request and Back doesn't return to a used checkout. Cart and order queries are marked stale without refetching (`refetchType: "none"`); refetching at once would flash "Your cart is empty" on the checkout page before the redirect. |
| Order page | `/orders/{id}` is step 3 in the `(checkout)` group. No screenshot, so it uses the checkout panels: number, status tag, date, card's last 4, total, then each item with its keys and a Copy button (the code is also selectable text, for browsers that block the clipboard). Failed, expired and unpaid orders say what happened and link back to the cart. Another user's order is a 404, shown as "Order not found". `/orders` (in `(shop)`) lists every order newest first, each row linking there. |
| 404 page | Eneba's "Highscore! 404": heading (`text-display`, `text-3xl` below 600px), one line, an outlined "Continue browsing" to `/`, inside the shop layout. The background is our own: pixel keys, gems, hearts and plus signs drawn as SVG from small text grids (`pixel-shapes.tsx`), in the accent tokens, floating with `animate-float` (off under reduced motion). Eneba's ghosts are left out, since they look like a known game's characters. `(shop)/not-found.tsx` only catches `notFound()` calls, so a catch-all `(shop)/[...missing]` route calls it for unknown URLs (a real 404 status, with header and nav); every other route ranks above a catch-all. The product page calls `notFound()` when the API answers 404. It loads in the browser, so that status stays 200: accepted, since the data fetch rule keeps server components data-free. |
| Footer | Eneba's footer opens with a light strip of payment logos. Brand logos are off limits, so the strip lists "Secure checkout / Keys delivered at once / Simulated payments only" with lucide icons, then a dark section with the demo disclaimer. |

## 12. Future improvements

Not planned for the 3 days. Listed so the trade-offs are on record.

1. **Retry payment with the hold kept.** On decline, keep the keys reserved and let the user pay again until `reservedUntil`, with a "Cancel" that releases them. Needs a `payment_failed` order state and an attempt counter to stop endless retries on held keys.
2. **Cut off access tokens on refresh-token theft.** Store a per-user `tokensValidAfter` time and reject access tokens issued before it, so reuse detection also ends live sessions within seconds instead of ≤ 15 min.
3. **httpOnly cookies** instead of localStorage, to keep tokens away from page scripts.
4. **Encrypt key codes at rest** (AES-GCM, key from env), decrypt only for the owner's paid order.
5. **Live stock in listings.** Replace the cached `offers.stock` with a count at read time, or a change stream that updates it, so store listings never undercount.
6. **Seller accounts** that upload their own keys and set prices.
7. **Emailed keys** and order receipts.
8. **Panic after a partial response.** `Recoverer` writes a 500 error body after a panic. If the handler had already started writing, the headers are already sent, so the error JSON is appended to the half-written body and the client gets broken JSON. No current handler streams, so it can't happen yet. Fix: wrap the writer, and if bytes were already sent, abort the connection (`panic(http.ErrAbortHandler)`) instead of writing more.
9. **Store sorts that use their index.** Every store sort ends in `_id: 1` so pages stay stable, but no index holds `_id`, so Mongo can't read results in sort order. `explain` on `eneba_dev` shows that for every sort it scans all matching index entries, fetches every doc, sorts them in memory and keeps 20 (60 docs read for 20 returned). Without the tie-break, `platform=steam&sort=price_asc` reads the `minPriceCents` index in order and stops after 44. No cost at 60 products; at 100k it would read every in-stock product per page. Fix: add `_id` as the last field of the sort indexes (`{active, inStock, minPriceCents, _id}` and so on), and give the tie-break the main field's direction in `listSort` so `price_desc` can read the index backwards. `inStock=false` needs its own look: there `inStock` is not fixed by the filter, so the leading `inStock: -1` sort key has to line up with the index too, and `newest` uses `{active, createdAt}`, which has no `inStock` at all. Check with `explain` that the plan has no `SORT` stage.
10. **Two seeds at once.** Known limit, accepted: the seed is a dev tool only the owner runs. Without `--reset`, a second run fails on the unique `sellers.slug` index and inserts nothing. With `--reset` on both, the clear and inserts are separate steps, so they can interleave: both runs report success, but the db can hold orphan keys and every homepage block twice. A seed that dies halfway leaves partial data too. Re-running with `--reset` fixes both. Fix if it ever matters: run the clear and inserts in one transaction.
11. **Search fallback when Atlas Search fails.** Today a `$search` error is a `500`, so store search and autocomplete stop working while the rest of the store keeps going. Fix: when the aggregate errors, run a plain `find` with a case-insensitive, escaped prefix regex on `name` (`regexp.QuoteMeta` on each word) plus the same filters and sort. Partial names would still work; typos would not. It does not help when the index is missing, because then `$search` returns no results instead of an error; `/healthz` reports that case as `"search": "missing"`.
12. **Rate limit counts that survive restarts.** httprate keeps its counts in memory, so a deploy, crash or free-tier sleep resets every IP to zero, and a second instance would keep its own counts. Accepted: a reset gives at most 10 more attempts, and bcrypt is the real defence. Per-IP limits also let someone with many IPs get 10 tries on each. Fix: keep counts in a shared store (`httprate-redis`, or a Mongo collection with a TTL index), and add a per-email limit on failed logins.
13. **Rails skip the count.** Homepage rails call `products.Service.List`, which runs `CountDocuments` for a `total` the homepage never reads: 4 wasted queries per page load (about 16 db round trips in all). Accepted: cheap at 60 products, and reusing `List` keeps rails identical to the store. Fix: a `products.Service.Top(ctx, q)` that runs only the `Find`, used by `Rail`.
14. **Wishlist heart on product cards.** On Eneba a heart shows in the top right of a card on hover. Left out because wishlist isn't built (stretch goal 2), and a heart that does nothing would be a fake control. Add it with the wishlist: an `onToggleWishlist` prop on `ProductCard`, shown with the same `can-hover` rule as the hover buttons. The gift card tag sits in the same corner, so one of them has to move.
15. **Dark homepage band through a token.** The darker band is set to #3c1395 (kept by choice), and both its background and its `--band` use the raw value: `bg-[#3c1395] [--band:#3c1395]` in `home-blocks.tsx`. The two match, so the category row's edge fade blends into the band, but it breaks the tokens-only styling rule. #3c1395 is already the `brand-border` token. Fix: use `bg-brand-border [--band:var(--color-brand-border)]` in `BANDS`, change the comment above it ("darkest" → "darker"), note the second use of `brand-border` in `globals.css` and `design-reference/notes.md`, and change the "Homepage bands" row in section 11 from `brand-darkest` (#1f0a4d) to `brand-border` (#3c1395).
