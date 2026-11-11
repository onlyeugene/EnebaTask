# Eugene

A marketplace for digital game keys, built to look and work like [eneba.com](https://www.eneba.com). Buyers browse a store, search with typos, compare offers from several sellers, sign in, pay with a test card and get their key.

Go API in [server/](server/), Next.js frontend in [client/](client/), MongoDB Atlas.

Demo project for the Eneba intern assignment. Not affiliated with Eneba. No real money is involved: products, sellers, covers and keys are all made up.

## Live

| | URL |
|---|---|
| App | https://enebatasks.vercel.app |
| API | https://eugene-api-9tvy.onrender.com ([health](https://eugene-api-9tvy.onrender.com/healthz)) |
| Code | https://github.com/onlyeugene/EnebaTask |

**The first load can take up to a minute.** The API runs on Render's free tier, which sleeps after about 15 minutes with no traffic. The first request wakes it; the page shows loading skeletons until then. After that it's fast.

## Demo walkthrough

<!-- TODO(owner): write a short walkthrough here. Points to cover:
- Purchase: home → product → add to cart → sign up → checkout → pay → key.
- Test cards: any valid-looking number pays (e.g. 4242 4242 4242 4242, any future MM/YY, any 3-digit CVC).
  4000 0000 0000 0002 is declined; the keys go back to stock and the cart stays.
- Search: type "thundr" or "skyforje" in the header (typos), or "frost" (part of a word).
- One failing homepage block (local only, APP_ENV=development): get a block id from
  GET http://localhost:8080/api/homepage, then open http://localhost:3000/?fail=<blockId>.
  That block shows an error with Retry; the rest of the page loads.
-->

_Coming soon._

## Assignment checklist

Every line of [ASSIGNMENT.md](ASSIGNMENT.md), and where it's covered.

### Requirements

| Assignment line | Status | Where |
|---|---|---|
| Use AI; save the full agent chat log | ✅ | [chat-logs/](chat-logs/), one file per phase. See [How I used AI](#how-i-used-ai) |
| "We want to understand how you use AI" | ✅ | [How I used AI](#how-i-used-ai) |
| You must be able to explain your code | ✅ | Every decision that isn't obvious is in [PLAN.md section 11](PLAN.md#11-decisions-made) or a code comment. In most phases I asked the agent to walk me through how its code works; the answers are in the logs |
| Shortcuts allowed; decide the scope and justify it | ✅ | [PLAN.md section 9](PLAN.md#9-scope) (what goes deep, what stays simple, and why), [Known limits](#known-limits) |
| Marketplace in Eneba's style that matches eneba.com | ✅ | Every page was built from an Eneba screenshot. Left out on purpose: real brand logos, game art, social login, Trustpilot and similar badges ([PLAN.md section 11](PLAN.md#11-decisions-made) lists each one) |
| **Store**: browse every product | ✅ | `/store` → [features/store/](client/src/features/store/), `GET /api/products` in [internal/products/](server/internal/products/) |
| Store: filter by platform, region, price | ✅ | Plus type and genre. Filters live in the URL, so links, reload and Back work ([use-store-params.ts](client/src/features/store/hooks/use-store-params.ts)) |
| Store: change the order | ✅ | 6 sorts: relevance, price up/down, best selling, discount, newest |
| Store: search by name | ✅ | `/store?q=` uses Atlas Search ([internal/search/](server/internal/search/)) |
| **Homepage**: built from several blocks (banner, best sellers, deals) | ✅ | 7 blocks: hero, category row, best sellers, top deals, new releases, gift cards, categories grid |
| Homepage: each block gets its content from its own backend service | ✅ | `GET /api/homepage` returns only the layout; `GET /api/homepage/blocks/{id}` calls that block type's service from a registry ([internal/home/blocks/](server/internal/home/blocks/)) with a 2-second timeout |
| Homepage: the frontend only displays what it gets, nothing hardcoded | ✅ | Order, titles, banners and links all come from the `homepage_blocks` collection. Each block loads, fails and retries on its own ([home-blocks.tsx](client/src/features/home/components/home-blocks.tsx)) |
| **Search** from anywhere in the app | ✅ | Header search box on every shop page ([features/search/](client/src/features/search/)). The cart, checkout and sign-in pages have no search bar, as on Eneba |
| Search: see the matching results | ✅ | Suggestions while typing; Enter opens `/store?q=` with filters and sort |
| Search: cope with typos and partial names | ✅ | Fuzzy match per word plus word-start matching. `skyforje` → Skyforge, `frost` → Frostline, `thundr` → Thunderpeak ([PLAN.md section 2](PLAN.md#2-indexes)) |
| **Product page**: details of one product | ✅ | `/product/[slug]` → [features/product/](client/src/features/product/) |
| Product page: offers from several sellers to compare and choose | ✅ | 2–4 sellers per product, cheapest first, live stock per offer, "Buy now" or add to cart on each |
| **Checkout**: put products in a cart | ✅ | Guest cart in the browser, server cart once signed in, merged at login ([PLAN.md section 5b](PLAN.md#5b-guest-cart--server-cart-at-login)) |
| Checkout: pay (simulated or test mode) | ✅ | Simulated card form; one test number declines ([payment.go](server/internal/orders/payment.go)) |
| Checkout: receive the digital key | ✅ | `/orders/{id}` shows the keys with a Copy button; `/orders` lists past orders |
| **Authentication**: create an account and sign in | ✅ | `/register`, `/login` → [features/auth/](client/src/features/auth/), [internal/auth/](server/internal/auth/) |
| Auth: checkout and order history only for signed-in users | ✅ | The API returns 401 without a token; the client sends guests to `/login?next=…` ([require-auth.tsx](client/src/features/auth/components/require-auth.tsx)) |
| **Eneba styling**: colors, typography, spacing, components follow the design reference | ✅ | Tokens from [design-reference/notes.md](design-reference/notes.md) in [globals.css](client/src/app/globals.css). Tailwind's default colors and sizes are switched off, so only these tokens work. Checked at phone (375px), tablet (768px) and desktop widths |

### Purchase flow

| Assignment line | Status | Where |
|---|---|---|
| Home → product → cart → checkout → payment → order and key | ✅ | The `(checkout)` pages share a 3-step header: Cart, Payment, Get your product |
| Entry points: homepage and search | ✅ | Product cards on every rail, search suggestions and store results all open the product page |
| Checkout requires sign-in | ✅ | "Checkout now" sends a guest to sign in, then back to checkout with the cart merged |
| Eneba styling on every screen | ✅ | Includes sign-in, checkout, order and 404 pages |

### Technical requirements

| Assignment line | Status | Where |
|---|---|---|
| React, TypeScript preferred | ✅ | Next.js 16 (App Router), React 19, TypeScript |
| Backend in PHP or Go, REST or GraphQL | ✅ | Go 1.25, chi, REST. Contract in [API.md](API.md) |
| Any database; decide the data yourself | ✅ | MongoDB Atlas. See [Data model](#data-model) |
| Few features done deeply beat many done minimally | ✅ | Deep: buying keys safely under load, auth, search, homepage blocks ([PLAN.md section 9](PLAN.md#9-scope)) |

### Deliverables

| Assignment line | Status | Where |
|---|---|---|
| GitHub link | ✅ | https://github.com/onlyeugene/EnebaTask |
| Public URL on a free tier | ✅ | Vercel (app) + Render (API) + Atlas M0, all free |
| Full agent chat history | ✅ | [chat-logs/](chat-logs/) |

## Run it locally

You need Go 1.25+, Node 20+ and a MongoDB Atlas cluster (the free M0 works). Atlas is required, not a local Mongo: search uses Atlas Search, and orders use transactions.

### 1. API

```
cd server
cp .env.example .env        # fill in MONGODB_URI and a JWT_SECRET (openssl rand -hex 32)
go run ./cmd/seed --reset   # products, sellers, keys, homepage blocks, and the search index
go run ./cmd/api            # http://localhost:8080
```

The seed waits until Atlas reports the search index ready. It refuses to touch `eneba_prod` unless you pass `--allow-prod`.

For live reload, use [air](https://github.com/air-verse/air):

```
go install github.com/air-verse/air@latest   # once
cd server && air
```

Check it's up:

```
curl localhost:8080/healthz
# {"db":"ok","search":"ok","status":"ok"}
```

### 2. Client

```
cd client
cp .env.example .env.local   # NEXT_PUBLIC_API_URL=http://localhost:8080
npm install
npm run dev                  # http://localhost:3000
```

### Tests

```
cd server
go test ./...    # auth, cart merge, orders, payment checks, HTTP helpers; no database needed
```

The integration tests prove two buyers can never get the same key. They need a real Atlas cluster and always use (and wipe) a database named `eneba_test`:

```
MONGODB_TEST_URI="mongodb+srv://..." go test -tags integration ./internal/orders/... ./internal/cart/... -v -count=1
```

Deploying: [DEPLOY.md](DEPLOY.md) has the Render, Vercel and Atlas steps.

## Architecture

```
Browser ── Next.js on Vercel (pages only)
   │  fetch + Authorization: Bearer <access token>
   ▼
Go API on Render (Docker, chi)
   │  mongo-go-driver v2
   ▼
MongoDB Atlas M0
   ├─ eneba_dev / eneba_prod
   └─ Atlas Search index "products_search"
```

- **The browser calls the Go API directly.** Next.js only serves pages: no API routes, no server-side data fetching. Pages that show data are client components using TanStack Query.
- **The server is split by feature:** `auth`, `products`, `search`, `home`, `cart`, `orders`, `keys` under [server/internal/](server/internal/). Each has the same files: `model.go`, `repository.go` (interface), `repository_mongo.go`, `service.go` (logic), `handler.go` (HTTP only), `routes.go`. Handlers never touch Mongo; services never touch HTTP. That keeps the logic testable with in-memory fakes.
- **The client is split the same way:** [client/src/features/](client/src/features/) holds one folder per feature with its own components, hooks and API calls. [components/ui/](client/src/components/ui/) is the shared design system, and [app/](client/src/app/) holds only routes and layouts.

Full folder map: [PLAN.md section 6](PLAN.md#6-folder-structure). Every endpoint: [API.md](API.md).

## Data model

| Collection | Holds | Why this shape |
|---|---|---|
| `products` | Name, platform, region, genres, generated SVG cover, MSRP, and a copy of its cheapest in-stock price, discount and stock flag | The store sorts and filters by price, so the price summary lives on the product instead of being worked out per row |
| `sellers` | Name, rating, sales count | Few of them; offers point to them by id |
| `offers` | One seller's price for one product, and a cached stock count | Separate from products, because price and stock change on every sale |
| `keys` | **One document per key**: code, status (`available` / `reserved` / `sold`), order id, `reservedUntil` | Each key can be claimed with one atomic update, so two buyers can never get the same one |
| `users` | Email, bcrypt hash, name | |
| `carts` | Signed-in user's lines: offer id + quantity only | Prices and stock are looked up live on every read, so a cart never shows a stale price |
| `orders` | Lines with a copy of name, price and seller; key ids; codes once paid | An order is a receipt: it must show what was paid even if the product changes later |
| `homepage_blocks` | Type, title, order, active flag, per-type config | The homepage layout is data, not code |
| `refresh_tokens` | SHA-256 of each refresh token, its login "family", used flag | Lets the API spot a stolen token used twice. A TTL index deletes old ones |
| `revoked_tokens` | Access token ids revoked at logout | TTL index deletes each one when the token would have expired anyway |

All money is integer cents (EUR), never floats. Every document shape, every index and the search index: [PLAN.md sections 1–2](PLAN.md#1-collections-and-document-shapes).

## Key decisions and trade-offs

**Buying a key is safe under load.** Each key is its own document. Checkout claims keys one by one with an atomic `findOneAndUpdate`, inside a transaction, so if any line runs out, nothing is reserved. The integration test sends 20 buyers at 5 keys at once: exactly 5 succeed, with 5 different keys.
*Trade-off:* more documents and a transaction per checkout, against a stock counter on the offer that two buyers could both decrement.

**Reservations expire without a background job.** A claimed key carries `reservedUntil`. The claim query treats a key whose hold has passed as free, so an abandoned checkout frees its keys on time even when Render's free server is asleep and no job runs. A cleanup job still tidies old orders, but nothing depends on it.
*Trade-off:* the cached stock on store listings can lag after an abandoned checkout. The product page, cart and checkout always count live, so buying is never affected.

**Stock is recounted, never incremented.** After every claim, sale, decline or expiry, the API counts the claimable keys again in the same transaction. With holds that expire on their own, `+1/-1` counting would drift.

**Keys are reserved when you click Pay, not when checkout opens.** Someone who only looks at the checkout page holds nothing. The order id is kept in the tab's `sessionStorage`, so a refresh or a lost response mid-payment finishes the same order instead of reserving twice.

**Each homepage block is its own request.** The layout call returns only block ids and types. Each block then loads from its own service, with a 2-second timeout, its own skeleton, its own error with Retry, and a React error boundary. One broken block never blanks the page.
*Trade-off:* one request per block plus the layout, instead of one for the whole page.

**Search uses Atlas Search, with typo tolerance by word length.** One fuzzy clause and one word-start clause per word. Short words get no typo allowance. The first version allowed 2 typos everywhere, and `star` matched every Steam product.
*Trade-off:* it ties the app to Atlas. A local Mongo can't run search.

**Auth: short access tokens and rotating refresh tokens.** Access tokens last 15 minutes. Each refresh token works once; if a used one comes back, the API treats it as stolen and logs that user out everywhere. Logout revokes the access token at once. The client runs one refresh for many failing calls, even across tabs.
*Trade-off:* tokens sit in `localStorage`, which page scripts can read. httpOnly cookies are safer but need cookie setup across two domains (Vercel and Render). The app loads no third-party scripts, which limits the risk.

**Guest cart merges by taking the larger quantity.** When a guest signs in, a line in both carts keeps the larger quantity, not the sum. The same person adding the same game twice shouldn't get two, and a retried merge gives the same result.

**Next.js as a pure frontend.** All data is fetched in the browser. It keeps one clear boundary (the Go API) and one auth path.
*Trade-off:* no server-rendered content for search engines, and a missing product shows the 404 page with an HTTP 200 status.

**Simulated payment.** Format checks only, and one test number declines. A real test-mode provider would add webhooks and keys without teaching anything new about this domain.

**Own artwork only.** Covers are SVGs generated from each product's name and colors; icons come from lucide. No game art, no brand logos.

Every other decision, with its reason: [PLAN.md section 11](PLAN.md#11-decisions-made).

## Known limits

- **Cold start:** the free Render server sleeps after about 15 minutes idle; the first request can take up to a minute.
- **Rate limits reset** when the server restarts or sleeps, since the counts live in memory. bcrypt is the real defence against password guessing.
- **Store listings can undercount stock** for a short time after an abandoned checkout (see above). Buying always uses the live count.
- **A stolen refresh token** ends every refresh session, but access tokens already issued stay valid for up to 15 minutes.
- **Key codes are stored as plain text.** They're fake. A real store would encrypt them.
- **Search needs the Atlas index.** If `$search` errors, store search and autocomplete return a 500 while the rest of the store works. `/healthz` reports a missing index.
- **Signing in to a different account in another tab** without logging out first isn't noticed until reload.
- **Filter counts are for the whole store,** not the current selection.
- **The header heart opens the 404 page.** Wishlist isn't built.
- **Not built:** seller accounts, admin, wishlist, reviews, refunds, emailed keys, social login, other currencies or languages.
- **Vercel preview deploys can't reach the API:** CORS allows only the production URL.

## Future improvements

The full list with fixes is in [PLAN.md section 12](PLAN.md#12-future-improvements). The main ones:

1. Keep the hold after a declined card, so the buyer can retry without losing the keys.
2. End live access tokens at once when a stolen refresh token is detected.
3. httpOnly cookies instead of `localStorage`.
4. Encrypt key codes at rest.
5. Live stock on store listings.
6. Seller accounts that upload their own keys and set prices.
7. Emailed keys and receipts.
8. A plain-text search fallback when Atlas Search fails.
9. Rate limit counts in a shared store, plus a per-email limit on failed logins.
10. A "Cheap games" promo page built on the same block system (planned as phase 15b, parked).

## How I used AI

I built this with **Claude Code in VS Code** (Claude Sonnet 4.6), phase by phase. The logs show I switched to Opus partway through phase 1 (`/model opus`). I used a **separate Claude chat** for planning, stack choices and learning Go: [planning chat](TODO-add-link).

**How the work went.** I wrote the brief and planned with the agent before any code ([PLAN.md](PLAN.md), [API.md](API.md)). Then I ran one phase per session. Each session started the same way: read [CLAUDE.md](CLAUDE.md), PLAN.md and API.md, build only that phase, run a real check (tests, curl, or headless Chrome), and stop. Between phases I asked the agent to walk me through what it built (how a request flows, what happens under a race, what breaks first), so I could explain every part. CLAUDE.md holds the rules it had to follow: the folder layout, money in cents, no git commands, no secrets in chat.

**Where I pushed back or changed course:**

- **The IP behind the rate limit.** I asked for the riskiest part of phase 1. The answer was chi's `RealIP` middleware, which trusts headers the client sets, so a fake header per request would dodge the login rate limit. I had it removed and replaced with our own middleware that uses only the last `X-Forwarded-For` entry, plus a test that sends fake headers. ([phase 1 log](chat-logs/02-phase1-server-skeleton-claude-code.md))
- **Don't overengineer.** After phase 1 I added a rule to CLAUDE.md: build the simplest version that meets the assignment, test only auth, key claiming and checkout, and ask me before adding anything extra. ([phase 1 log](chat-logs/02-phase1-server-skeleton-claude-code.md))
- **Rate limit on `/me`.** The agent put `/me` behind the login limit of 10 per minute. The client calls `/me` on every protected page, so normal use would have hit 429. I moved it out and asked for a separate, higher limit for refresh. ([phase 5 log](chat-logs/06-phase5-auth-claude-code.md))
- **Two pending orders for one user.** Walking through checkout, we saw that a user who went back and checked out again would hold two sets of keys, and on low stock their own first order would block the second. I asked for a new order to expire the user's older pending orders in the same transaction, with a test. ([phase 9 log](chat-logs/10-phase9-orders-claude-code.md))
- **A paid order whose answer got lost.** I asked what happens if the server crashes mid-checkout. That led to a rule for phase 15: when Pay returns 409, reload the order and show the keys if it's already paid. Later I added the `sessionStorage` fix so a refresh mid-payment finishes the same order. ([phase 9](chat-logs/10-phase9-orders-claude-code.md), [phase 15](chat-logs/16-phase15-checkout-claude-code.md))
- **Checking a "fix".** I asked the agent to fix a race when two writes create a user's first cart. It found that Mongo already reports that case as a write conflict and retries it, so the fix wasn't needed. I asked it to confirm that plainly, and it added an integration test that forces the worst order instead of new code. ([phase 8 log](chat-logs/09-phase8-cart-claude-code.md))
- **No fake controls.** Eneba shows a wishlist heart on product cards. Wishlist isn't built, so I told the agent to leave it off the cards and list it as future work, rather than add a button that does nothing. ([phase 11 log](chat-logs/12-phase11-homepage-ui-claude-code.md))
- **Keep it a known limit.** I asked what happens if two people run the seed at once. The agent offered a transaction; I had it noted as a known limit instead, since only I run the seed. ([phase 2 log](chat-logs/03-phase2-seed-data-claude-code.md))

**Chat logs** (full prompts and responses):

| Log | Phase |
|---|---|
| [01-planning](chat-logs/01-planning-claude-code.md) | Planning with Claude Code |
| [02-phase1](chat-logs/02-phase1-server-skeleton-claude-code.md) | Server skeleton |
| [03-phase2](chat-logs/03-phase2-seed-data-claude-code.md) | Indexes and seed data |
| [04-phase3](chat-logs/04-phase3-catalog-claude-code.md) | Catalog API |
| [05-phase4](chat-logs/05-phase4-search-claude-code.md) | Atlas Search |
| [06-phase5](chat-logs/06-phase5-auth-claude-code.md) | Auth part 1 |
| [07-phase6](chat-logs/07-phase6-auth-refresh-claude-code.md) | Auth part 2: refresh and logout |
| [08-phase7](chat-logs/08-phase7-homepage-claude-code.md) | Homepage API |
| [09-phase8](chat-logs/09-phase8-cart-claude-code.md) | Cart API |
| [10-phase9](chat-logs/10-phase9-orders-claude-code.md) | Orders and key claiming |
| [11-phase10](chat-logs/11-phase10-client-skeleton-claude-code.md) | Client skeleton |
| [12-phase11](chat-logs/12-phase11-homepage-ui-claude-code.md) | Homepage UI |
| [13-phase12](chat-logs/13-phase12-store-search-claude-code.md) | Store and search UI |
| [14-phase13](chat-logs/14-phase13-product-cart-claude-code.md) | Product page and cart |
| [15-phase14](chat-logs/15-phase14-auth-ui-claude-code.md) | Login, register, cart merge |
| [16-phase15](chat-logs/16-phase15-checkout-claude-code.md) | Checkout, payment, orders |
| [17-phase16](chat-logs/17-phase16-deploy-claude-code.md) | Deploy |
| 18-phase17 _(TODO: export this session)_ | Polish, README, assignment check |
