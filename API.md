# API

Base URL: `NEXT_PUBLIC_API_URL` (local `http://localhost:8080`). All bodies are JSON. All money is **integer cents in EUR** (`1549` = €15.49).

**Stock:** `minPriceCents`, `inStock`, `offerCount`, and `discountPct` on products are recomputed from the `keys` collection inside every transaction that claims, releases, or sells keys — the same request that changes stock. Store listings are therefore accurate immediately after a purchase, decline, or payment.

Auth column: **—** public, **Bearer** needs `Authorization: Bearer <accessToken>`.

## Conventions

**Error body** (every non-2xx):

```json
{ "error": { "code": "OUT_OF_STOCK", "message": "Only 2 keys left for this offer", "fields": { "email": "already taken" } } }
```

`fields` only appears on `422 VALIDATION_FAILED`.

| Status | Code examples | When |
|---|---|---|
| 400 | `BAD_REQUEST` | Malformed JSON, bad query param |
| 401 | `UNAUTHORIZED`, `TOKEN_EXPIRED`, `TOKEN_REVOKED`, `TOKEN_REUSED` | Missing/invalid/expired/revoked access token, bad credentials, refresh token used twice |
| 402 | `PAYMENT_DECLINED` | Simulated card declined |
| 403 | `FORBIDDEN` | Order belongs to another user |
| 404 | `NOT_FOUND` | |
| 409 | `EMAIL_TAKEN`, `OUT_OF_STOCK`, `ORDER_NOT_PAYABLE` | State conflict |
| 422 | `VALIDATION_FAILED` | Field rules broken |
| 429 | `RATE_LIMITED` | Auth endpoints, per IP |

**Pagination**: `?page=1&pageSize=20` (max 60). Response carries `page`, `pageSize`, `total`.

## Shared shapes

```ts
type User = { id: string; email: string; name: string; createdAt: string };

type ProductCard = {
  id: string; slug: string; name: string;
  type: "game" | "gift_card" | "software";
  platform: "steam" | "psn" | "xbox" | "nintendo" | "windows";
  region: "global" | "europe" | "united_states" | "united_kingdom";
  coverUrl: string;            // "/api/products/{slug}/cover.svg" (prefix with API URL)
  giftCardValue?: string;      // "€50" for gift cards, shown as the corner tag
  originalPriceCents: number;  // MSRP, the struck-through price
  minPriceCents: number | null;// cheapest in-stock offer; null = sold out
  discountPct: number;         // 0..100, from minPrice vs MSRP
  cashbackPct: number;         // display only
  offerCount: number;
};

type ProductDetail = ProductCard & {
  description: string; genres: string[]; developer: string; publisher: string;
  releaseDate: string; activation: string; // short "how to activate" text
};

type Offer = {
  id: string; productId: string; priceCents: number; stock: number;
  seller: { id: string; slug: string; name: string; rating: number; salesCount: number };
};

type CartLine = {
  offerId: string; productId: string; slug: string; name: string;
  platform: string; region: string; coverUrl: string;
  seller: { id: string; name: string };
  unitPriceCents: number; quantity: number; lineTotalCents: number;
  cashbackCents: number;
  available: number;           // current stock
  problem?: "OUT_OF_STOCK" | "NOT_ENOUGH_STOCK" | "OFFER_GONE";
};

type Cart = { items: CartLine[]; itemCount: number; totalCents: number; cashbackCents: number };

type OrderItem = {
  offerId: string; productId: string; name: string; slug: string;
  platform: string; region: string; sellerName: string;
  unitPriceCents: number; quantity: number;
  keys?: { code: string }[];   // only present when order.status == "paid"
};

type Order = {
  id: string; number: string;  // "ORD-7K2M9Q", human-friendly
  status: "pending_payment" | "paid" | "failed" | "expired";
  items: OrderItem[]; totalCents: number; cashbackCents: number;
  createdAt: string; reservedUntil: string; paidAt?: string;
  payment?: { last4: string; outcome: "success" | "declined" };
};
```

---

## Health

| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/healthz` | — | `200 { "status": "ok", "db": "ok", "search": "ok" }` (pings Mongo; Render health check) |

- Search index missing or not queryable yet → still `200`, body `{ "status": "degraded", "db": "ok", "search": "missing" }`. Render restarts a service that fails its health check, and a restart can't fix a missing index, while everything except search still works.
- Database unreachable → `503 { "status": "degraded", "db": "down" }`. The search check is skipped, since it needs the database.

## Auth

| Method | Path | Auth | Request | Response |
|---|---|---|---|---|
| POST | `/api/auth/register` | — | `{ email, password, name? }` | `201 { accessToken, refreshToken, user }` |
| POST | `/api/auth/login` | — | `{ email, password }` | `200 { accessToken, refreshToken, user }` |
| POST | `/api/auth/refresh` | — | `{ refreshToken }` | `200 { accessToken, refreshToken }` |
| POST | `/api/auth/logout` | Bearer | `{ refreshToken }` | `204` |
| GET | `/api/auth/me` | Bearer | — | `200 { user }` |

Rules:
- Password: 8–72 bytes (bcrypt ignores bytes past 72), at least one digit or symbol. Email lowercased and trimmed. `name` optional, at most 50 chars, stored as `""` when left out.
- Broken rules → `422 VALIDATION_FAILED` with `fields` (e.g. `{ "email": "is not a valid email", "password": "must contain a digit or a symbol" }`). Email already registered → `409 EMAIL_TAKEN`.
- Login returns the same `401 UNAUTHORIZED "Invalid email or password"` for unknown email and wrong password, so it doesn't reveal which emails exist. An empty email or password is a `422`.
- Refresh: each refresh token works once; the response carries a new pair. Unknown, empty or expired token → `401 UNAUTHORIZED`. Already-used token → `401 TOKEN_REUSED` and **every refresh token for that user is deleted** (all sessions must log in again).
- Logout: adds the access token's `jti` to `revoked_tokens` (that access token stops working at once) and deletes the given refresh token along with every token rotated from the same login. Other logins of the same user stay signed in. An unknown, empty or someone else's `refreshToken` is skipped, still `204`. The client clears local tokens even if this call fails.
- Middleware errors: `401 TOKEN_EXPIRED` (client should refresh), `401 TOKEN_REVOKED` / `401 UNAUTHORIZED` (client should also try refresh once; if that fails, log out). Missing header, bad signature, wrong `alg` or issuer → `UNAUTHORIZED`.
- `/register` and `/login` share one rate limit per IP: 10 requests per minute across both, then `429 RATE_LIMITED`. `/refresh` has its own limit: 60 per minute per IP, counted apart from login, so failed logins never block a refresh. `/me` and `/logout` are not rate limited.
- Malformed JSON body → `400 BAD_REQUEST`.

## Catalog and search

| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/api/products` | — | `200 { items: ProductCard[], page, pageSize, total }` |
| GET | `/api/products/filters` | — | `200 { types, platforms, regions, genres: {value,label,count}[], price: {minCents, maxCents} }` |
| GET | `/api/products/{slug}` | — | `200 { product: ProductDetail, offers: Offer[] }` (offers sorted by price, in-stock first) |
| GET | `/api/products/{slug}/cover.svg` | — | `200 image/svg+xml`, `Cache-Control: public, max-age=86400` |
| GET | `/api/search/autocomplete?q=` | — | `200 { suggestions: { slug, name, platform, region, minPriceCents, coverUrl }[] }` (max 8) |

`GET /api/products` query params (all optional, combine freely):

| Param | Example | Notes |
|---|---|---|
| `q` | `skyforje` | When set, query goes through Atlas Search (typos + partial words). Trimmed; over 100 chars → `400`. |
| `type` | `game,gift_card` | Comma list = OR within a filter, AND across filters |
| `platform` | `steam,psn` | |
| `region` | `global` | |
| `genre` | `rpg` | |
| `minPrice`, `maxPrice` | `500`, `3000` | cents |
| `inStock` | `true` | default `true` |
| `sort` | `relevance` \| `price_asc` \| `price_desc` \| `bestselling` \| `discount` \| `newest` | default `relevance` with `q`, else `bestselling` |
| `page`, `pageSize` | `2`, `20` | |

Any bad value (unknown `type`/`platform`/`region`, non-number price, `minPrice` above `maxPrice`, unknown `sort`, `pageSize` outside 1–60) → `400 BAD_REQUEST` with the param name in `message`. Unknown genres are not an error; they just match nothing. Sorted lists always put in-stock products first and break ties by id, so pages never overlap.

`/api/products/filters` counts active, in-stock products (the store's default view). Each list only holds values with at least one product: `types`, `platforms`, `regions` in a fixed order, `genres` alphabetical. `price` is the lowest and highest `minPriceCents` among them.

`/api/products/{slug}`: `offers[].stock` is counted live from the `keys` collection, and the product's `minPriceCents`, `offerCount`, `inStock`, `discountPct` are rebuilt from those live counts, so the page never shows a stale price. Inactive products → `404 NOT_FOUND`.

`/api/products/{slug}/cover.svg` also sends `Content-Security-Policy: default-src 'none'` so an SVG opened in a tab can never run script.

Search with `q`: each word matches whole (with typos: none for words of 1–3 letters, 1 for 4–6, 2 for 7+) or as the start of a word (`frost` → Frostline). A product matching any word is returned; matching more words ranks higher. Filters, `sort` and paging work the same as without `q`; `sort=relevance` is only used when `q` is set.

`/api/search/autocomplete`: active products only, sold out included (`minPriceCents: null`). Every word must match the start of a word in the name (`neon dr` → Neon Drift); words of 5+ letters allow 1 typo (`thundr` → Thunderpeak). `q` under 2 chars → `{ "suggestions": [] }`; over 100 chars → `400`.

The header search box calls `/api/search/autocomplete` as you type (debounced 200 ms, min 2 chars). Pressing Enter opens `/store?q=...`, which calls `/api/products?q=...`.

## Homepage

| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/api/homepage` | — | `200 { blocks: { id, type, title }[] }` (active blocks, in order) |
| GET | `/api/homepage/blocks/{id}` | — | `200 { id, type, title, data }` — `data` depends on `type` |

`data` by block type:

```ts
// type: "hero"
{ main: Banner; tiles: Banner[] }      // Banner = { title, subtitle, ctaLabel, href, theme: "purple"|"yellow"|"mint"|"pink" }

// type: "category_row"
{ categories: { slug: string; label: string; icon: string; count: number }[] } // icon = lucide name

// types: "bestsellers" | "deals" | "new_releases" | "gift_cards"
{ products: ProductCard[]; seeAllHref: string }   // e.g. "/store?sort=discount"
```

A block whose service fails returns `503 { error: { code: "BLOCK_UNAVAILABLE" } }`. Only that block shows an error; the layout and other blocks are unaffected.

Dev only (`APP_ENV=development`): `GET /api/homepage/blocks/{id}?fail=1` forces a 503, to demo block isolation.

## Cart

Guests keep the cart in the browser (Zustand). Signed-in users have a server cart; the guest cart merges into it on login.

| Method | Path | Auth | Request | Response |
|---|---|---|---|---|
| POST | `/api/cart/quote` | — | `{ items: { offerId, quantity }[] }` | `200 Cart` — prices and stock for a guest cart, nothing saved |
| GET | `/api/cart` | Bearer | — | `200 Cart` |
| PUT | `/api/cart/items/{offerId}` | Bearer | `{ quantity }` (1–10) | `200 Cart` |
| DELETE | `/api/cart/items/{offerId}` | Bearer | — | `200 Cart` |
| POST | `/api/cart/merge` | Bearer | `{ items: { offerId, quantity }[] }` | `200 { cart: Cart, dropped: { offerId, reason }[] }` |

Merge rules (lines are keyed by `offerId`; full reasoning in PLAN.md section 5b):
- Same offer in both carts → the **larger** quantity, not the sum. Running the merge twice gives the same cart.
- Same product from a different seller → separate line.
- Capped at 10 and at live stock. Offers that are gone or inactive go in `dropped`.

Cart responses always use live prices and stock, never stored ones. A line that can't be bought carries `problem` and checkout is blocked until it's fixed.

## Orders and payment

| Method | Path | Auth | Request | Response |
|---|---|---|---|---|
| POST | `/api/orders` | Bearer | `{}` (uses the server cart) | `201 { order }` status `pending_payment`, keys reserved until `reservedUntil` |
| POST | `/api/orders/{id}/pay` | Bearer | `{ cardNumber, expiry, cvc, name }` | `200 { order }` status `paid` with keys, or `402 PAYMENT_DECLINED` + `{ order }` status `failed` |
| GET | `/api/orders` | Bearer | — | `200 { items: Order[] }` (newest first, no key codes) |
| GET | `/api/orders/{id}` | Bearer | — | `200 { order }` (key codes included when paid) |

- `POST /api/orders` → `409 OUT_OF_STOCK { error, details: { offerId, available } }` if any line can't be reserved. Nothing is reserved in that case (the transaction rolls back).
- Simulated card: any valid-looking number pays, **`4000 0000 0000 0002` is declined**. Card data is checked for format and then dropped; only `last4` is stored.
- Decline releases the keys and marks the order `failed`. The cart is left as is, so the user can try again.
- `pay` on an order that is not `pending_payment` or past `reservedUntil` → `409 ORDER_NOT_PAYABLE`. Once `reservedUntil` passes, the order's keys can be claimed by other buyers right away, so a late payment is always refused and the order becomes `expired`.
- `GET /api/orders` and `GET /api/orders/{id}` report a `pending_payment` order past `reservedUntil` as `expired`, even before the cleanup job has updated it.
- A paid order clears the server cart.
- Someone else's order → `404` (not 403, so order IDs can't be probed).
