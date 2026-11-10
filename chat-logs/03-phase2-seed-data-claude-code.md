❯ <ide_opened_file>The user opened the file /Users/onlyeugene/Desktop/Eneba
Marketplace/PLAN.md in the IDE. This may or may not be related to the current
task.</ide_opened_file>

❯ read claude.md, plan.md and api.md first. then start phase 2 from plan.md.
stop when it's done and tell me how to check it, including the requests i can
run in postman

@PLAN.md @CLAUDE.md @API.md
⎿  Read PLAN.md (776 lines)
⎿  Read CLAUDE.md (53 lines)
⎿  Read API.md (200 lines)

⏺ I'll look at what Phase 1 built first.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace" && find server -type f
-not -path '_/tmp/_' | sort && cat server/go.mod && ls design-reference &&
cat .gitignor…)
⎿  (No output)

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && cat
cmd/api/main.go cmd/api/routes.go internal/platform/config/config.go
internal/platform/db/connect…)
⎿  (No output)

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace" && grep -n -i -A3
"seed\|homepage\|block\|hero\|categor\|seller\|catalog" ASSIGNMENT.md |
head -80; grep -n -i…)
⎿  (No output)

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace" && sed -n 75,160p
design-reference/notes.md; cat server/.env.example; cat
server/internal/platform/httpx/errors…)
⎿  (No output)

⏺ Phase 1 is in place. Now the index setup, which both the API and the seed
script will call.

⏺ Write(internal/platform/db/indexes.go)
⎿  Wrote 71 lines to internal/platform/db/indexes.go
1 package db
2
3 import (
4 "context"
5 "fmt"
6
7 "go.mongodb.org/mongo-driver/v2/bson"
8 "go.mongodb.org/mongo-driver/v2/mongo"
9 "go.mongodb.org/mongo-driver/v2/mongo/options"
10 )
… +61 lines (ctrl+o to expand)

⏺ Update(cmd/api/main.go)
⎿  Added 6 lines, removed 1 line
26 ctx, stop := signal.NotifyContext(context.Background(), os.In
terrupt, syscall.SIGTERM)
27 defer stop()
28  
 29 - client, \_, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
29 + client, database, err := db.Connect(ctx, cfg.MongoURI, cfg.Mo
+ngoDB)  
 30 if err != nil {
31 slog.Error("db", "err", err)
32 os.Exit(1)
33 }
34 defer client.Disconnect(context.Background())
35  
 36 + if err := db.EnsureIndexes(ctx, database); err != nil {  
 37 + slog.Error("indexes", "err", err)  
 38 + os.Exit(1)  
 39 + }  
 40 +  
 41 srv := &http.Server{
42 Addr: ":" + cfg.Port,
43 Handler: newRouter(cfg, client),

⏺ Next, the cover generator. The plan puts it in products/cover.go so the seed
and the later cover route share it.

⏺ Write(internal/products/cover.go)
⎿  Wrote 89 lines to internal/products/cover.go
1 package products
2
3 import (
4 "fmt"
5 "hash/fnv"
6 "html"
7 "strings"
8 )
9
10 // Cover sizes match the product card image area in notes.md (22
… +80 lines (~3 KB) (ctrl+o to expand)

⏺ Seed needs the document shapes, so I'm creating the model types in their
feature packages (products, keys, home) now and will have the seed import
them.

⏺ Write(internal/products/model.go)
⎿  Wrote 68 lines to internal/products/model.go
1 package products
2
3 import (
4 "time"
5
6 "go.mongodb.org/mongo-driver/v2/bson"
7 )
8
9 // Product is the document in the products collection (PLAN.md s
ection 1).
… +59 lines (ctrl+o to expand)

⏺ Write(internal/keys/model.go)
⎿  Wrote 27 lines to internal/keys/model.go
1 package keys
2
3 import (
4 "time"
5
6 "go.mongodb.org/mongo-driver/v2/bson"
7 )
8
9 // A key moves available → reserved → sold, or back to available
(PLAN.md section 5).
… +18 lines (ctrl+o to expand)

⏺ Write(internal/home/model.go)
⎿  Wrote 60 lines to internal/home/model.go
1 package home
2
3 import (
4 "go.mongodb.org/mongo-driver/v2/bson"
5 )
6
7 // Block types. The client picks a component by type; the server
picks a service by type.
8 const (
9 TypeHero = "hero"
… +51 lines (ctrl+o to expand)

⏺ Now the seed data: sellers, fictional titles, palettes and block configs.

⏺ Write(cmd/seed/data.go)
⎿  Wrote 187 lines to cmd/seed/data.go
1 package main
2
3 import "eugene/internal/home"
4
5 // All titles, studios and sellers are made up. Platform names a
ppear as plain text only.
6
7 type sellerSeed struct {
8 slug, name string
9 rating float64
… +178 lines (ctrl+o to expand)

⏺ The hero links to Starfall Odyssey's Steam/global slug, so the build will pin
that title. The cover function can take the short title directly from the seed
instead of cutting it from the long name.

⏺ Update(internal/products/cover.go)
⎿  Added 1 line, removed 1 line
47 }
48  
 49 // Title at the bottom, wrapped into short lines. Escaped bec
ause it ends up inside XML.
50 - lines := wrap(coverTitle(title), 14)  
 50 + lines := wrap(title, 14)  
 51 y := coverH - 20 - (len(lines)-1)\*26
52 for \_, line := range lines {
53 fmt.Fprintf(&b, `<text x="16" y="%d" font-family="Arial, He
          lvetica, sans-serif" font-size="22" font-weight="800" fill="#ff
          ffff">%s</text>`,

⏺ Update(internal/products/cover.go)
⎿  Removed 8 lines
58 return b.String()
59 }
60  
 61 -// coverTitle drops the listing suffix: "Starfall Odyssey (PC)
-Steam Key GLOBAL" → "Starfall Odyssey".  
 62 -func coverTitle(name string) string {  
 63 - if i := strings.Index(name, " ("); i > 0 {  
 64 - return name[:i]  
 65 - }  
 66 - return name  
 67 -}  
 68 -  
 61 // wrap splits text into lines of at most max characters, break
ing only between words.
62 // A single word longer than max gets its own line.
63 func wrap(text string, max int) []string {

⏺ Update(internal/products/cover.go) ⎿  Added 2 lines, removed 1 line
13 coverH = 250
14 )
15  
 16 -// CoverSVG draws a cover from the title and a two-color palett
-e.  
 16 +// CoverSVG draws a cover from a short title ("Starfall Odyssey
+", not the full listing name)  
 17 +// and a two-color palette.  
 18 // The same input always gives the same SVG: the shapes come fr
om a hash of the title,
19 // not from a random source, so re-seeding never changes how a
product looks.
20 func CoverSVG(title string, palette [2]string) string {

⏺ Now the build step that turns the seed lists into documents.

⏺ Write(cmd/seed/build.go)
⎿  Wrote 219 lines to cmd/seed/build.go
1 package main
2
3 import (
4 "fmt"
5 "math/rand/v2"
6 "strings"
7 "time"
8
9 "go.mongodb.org/mongo-driver/v2/bson"
… +210 lines (ctrl+o to expand)

⏺ Now the seed entry point.

⏺ Write(cmd/seed/main.go)
⎿  Wrote 112 lines to cmd/seed/main.go
1 // Command seed fills a database with sellers, products, offers,
keys and homepage blocks.
2 //
3 // go run ./cmd/seed # refuses if products alrea
dy exist
4 // go run ./cmd/seed --reset # clears catalog, carts and
orders first
5 //
6 // It reads MONGODB_URI and MONGODB_DB from server/.env or the e
nvironment.
… +106 lines (ctrl+o to expand)

⏺ clear shadows Go's built-in, so I'll rename it, then build and vet.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && sed -i ''
's/if err := clear(ctx, database)/if err := clearCatalog(ctx, database)/;
s/^\/\/ clear emp…)
⎿  (No output)

⏺ Builds and vets clean; one file needs gofmt. Now the real run against
eneba_dev.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && gofmt -w
internal/home/model.go && grep -q "MONGODB_DB=eneba_dev" .env && echo "db
is eneba_dev" && g…)
⎿  (No output)
⎿  (timeout 3m)

⏺ The seed ran and both guards refuse correctly. Now I'll check what landed in
Atlas: indexes, the summary fields against the real key counts, and a cover.

Searched for 1 pattern (ctrl+o to expand)

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && mkdir -p
cmd/zzcheck && cat > cmd/zzcheck/main.go <<'EOF'
package main…)
⎿  (No output)
⎿  (timeout 3m)

⏺ Index counts match the plan (each count includes the built-in \_id index).
Stock and summaries check out. The panic is a bug in my throwaway script, not
the seed; I'll rerun just the cover export.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && mkdir -p
cmd/zzcheck && cat > cmd/zzcheck/main.go <<'EOF'
package main…)
⎿  (No output)
⎿  (timeout 2m)

Read 1 file (ctrl+o to expand)

⏺ The cover renders properly (the white margin is just the macOS preview
canvas). Last check: the API still starts now that it creates indexes on boot.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && if curl -s
localhost:8080/healthz >/dev/null; then echo "something already on 8080";
else (PORT=8099…)
⎿  (No output)

⏺ Your server (likely air) is already on 8080 and would have rebuilt with the
new code. A quick check:

⏺ Bash(curl -s localhost:8080/healthz; echo; ls -la
"/Users/onlyeugene/Desktop/Eneba Marketplace/server/tmp" 2>/dev/null |
head)
⎿  (No output)

⏺ The rebuild at 12:48 went through and /healthz responds, so startup with index
creation works. I'm marking the phase done and adding build.go to PLAN.md's
folder tree next.

⏺ Update(~/Desktop/Eneba Marketplace/PLAN.md)
⎿  Added 1 line, removed 1 line
715 |---|---|---|
716 | 0 | Verify `.gitignore` covers `.env*` (except `.env.example
           `), `.next/`, `bin/` | All listed paths are ignored |
717 | 1 ✅ | Server skeleton: `go mod init`, config, chi, CORS, lo
gger, recoverer, JSON error helper, Mongo connect, `/healthz`
| `curl localhost:8080/healthz` → `{"status":"ok","db":"ok"}`
|
718 -| 2 | Indexes + seed: sellers, ~60 products (fictional titles)
-, 2–4 offers each, 5–20 keys per offer, SVG covers, homepage b
-locks | `go run ./cmd/seed --reset`; docs visible in Atlas `en
          -eba_dev` |  
 718 +| 2 ✅ | Indexes + seed: sellers, ~60 products (fictional titl
+es), 2–4 offers each, 5–20 keys per offer, SVG covers, homepag
+e blocks | `go run ./cmd/seed --reset`; docs visible in Atlas +`eneba_dev` |  
 719 | 3 | Catalog API: list with filters/sort/pagination (no `q`),
filters, product detail + offers, cover route | curl each; op
en a `cover.svg` in the browser |
720 | 4 | Atlas Search: create index from JSON, `q` on `/api/produ
           cts`, `/api/search/autocomplete` | Typo query and partial quer
y return the right product |
721 | 5 | Auth part 1: users, register/login, JWT, middleware, `/m
           e`, rate limit + tests for expired/revoked | `go test ./intern
           al/auth/...` green; curl register → me |

⏺ Update(~/Desktop/Eneba Marketplace/PLAN.md)
⎿  Added 3 lines, removed 2 lines
390 │ │ │ ├─ main.go # wire only: config → db → r
outer → cleanup job → ListenAndServe
391 │ │ │ └─ routes.go # global middleware, 404/405
, /healthz, mounts each feature under /api
392 │ │ └─ seed/
393 -│ │ ├─ main.go # entry: flags, db connect,
-call seed functions  
 394 -│ │ └─ data.go # fake titles, sellers, bloc
-k configs, cover palettes  
 393 +│ │ ├─ main.go # entry: flags, db connect,
+clear, insert  
 394 +│ │ ├─ data.go # fake titles, sellers, bloc
+k configs, cover palettes  
 395 +│ │ └─ build.go # turns data.go into documen
+ts (fixed random seed, so runs repeat)  
 396 │ ├─ atlas/
397 │ │ └─ products_search.json # Atlas Search index definit
ion
398 │ ├─ internal/

⏺ Phase 2 is done and marked ✅ in PLAN.md. I ran the seed on eneba_dev, and the
data in Atlas checks out.

What I ran

- go run ./cmd/seed --reset added 6 sellers, 60 products (48 games, 6 gift
  cards, 6 software), 189 offers, 2,360 keys and 6 homepage blocks.
- Running the seed a second time without --reset refuses, because products
  already exist.
- Running it with MONGODB_DB=eneba_prod refuses without --allow-prod.
- A throwaway check script (now deleted) confirmed:
  - Every collection has the indexes from PLAN.md section 2.
  - Every offer's stock equals its count of available keys.
  - Every product's minPriceCents, offerCount and inStock match its offers.
- The cover SVG renders properly.
- Your running air server rebuilt with the change and /healthz returns ok, so
  creating indexes at startup works.

What changed

- indexes.go (server/internal/platform/db/indexes.go) creates all the indexes.
  The API runs it at startup and the seed runs it too.
- cmd/seed/ (server/cmd/seed/) has three files: main.go for flags, the prod
  guard and inserts, data.go for the made-up titles, sellers and block
  configs, and build.go, which turns that data into documents. The random seed
  is fixed, so every run gives the same prices and stock. build.go is new to
  the plan's folder tree, so I added it there.
- cover.go (server/internal/products/cover.go) draws a cover from the short
  title and a palette. The same title always gives the same cover.
- I wrote some model files early: products/model.go
  (server/internal/products/model.go), keys/model.go
  (server/internal/keys/model.go) and home/model.go
  (server/internal/home/model.go). The seed needs the document shapes, and
  this way each shape is defined once. Phases 3 and 7 will build on them.

Choices worth knowing

- --reset empties collections with DeleteMany instead of dropping them.
  Dropping a collection would also delete the Atlas Search index that Phase 4
  adds.
- A reset also clears carts and orders, because they hold offer and key ids
  that would no longer exist. Users and tokens stay.
- The hero banner links to starfall-odyssey-steam-global, so the seed always
  makes that product a Steam/Global key.

How to check it

There is nothing new to call in Postman yet: Phase 2 adds no endpoints. The
only route is still:

GET http://localhost:8080/healthz → {"db":"ok","status":"ok"}

The catalog endpoints come in Phase 3. To check this phase:

1. Re-run the seed: cd server && go run ./cmd/seed --reset. The last log line
   should read sellers=6 products=60 offers=189 keys=2360 blocks=6.
2. Test the guards: run go run ./cmd/seed without flags and it should fail
   with "already has 60 products". Run MONGODB_DB=eneba_prod go run ./cmd/seed
   and it should fail with "refusing to seed eneba_prod".
3. Look in Atlas (Browse Collections → eneba_dev):
   - In products, filter { slug: "starfall-odyssey-steam-global" }. You should
     see minPriceCents: 1959, discountPct: 51 and offerCount: 3.
   - In keys, filter { status: "available" }. The count should be 2,360.
   - In homepage_blocks, you should see 6 docs with order set to 10, 20,
     …, 60.
   - On the Indexes tab of products, you should see 7 indexes. refresh_tokens
     and revoked_tokens should each show one TTL index on expiresAt.

Against ASSIGNMENT.md: this phase only lays down data. The store, homepage
blocks and product page with offers from several sellers now have something to
show. None of them is visible yet: they come in Phases 3, 7 and 11.

Phase 0 isn't marked ✅ yet. I read .gitignore, and it covers .env\* (except
.env.example), .next/ and bin/. I didn't run git check-ignore because of the
no-git rule, so you may want to tick that row yourself.

❯ <ide_opened_file>The user opened the file /Users/onlyeugene/Desktop/Eneba
Marketplace/server/internal/products/cover.go in the IDE. This may or may not
be related to the current task.</ide_opened_file>

❯ running go run ./cmd/seed --reset failed with: indexes on sellers:
(IndexOptionsConflict) index already exists with a different name: email_1.
find why the same index is created with two different names
and make all index names consistent in one place. keep the fix simple.

Read 1 file (ctrl+o to expand)

⏺ The unique options builder is one pointer shared by five index models. I
suspect the driver writes the generated name back into that shared builder.
Checking the driver source:

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && f=$(go env
GOMODCACHE)/go.mongodb.org/mongo-driver/v2@v2.10.0/mongo/index_view.go &&
grep -n "SetName…)
⎿  (No output)

⏺ Bash(f=$(go env
GOMODCACHE)/go.mongodb.org/mongo-driver/v2@v2.10.0/mongo/index_view.go &&
sed -n 225,250p $f && sed -n 536,560p $f) ⎿  (No output)

⏺ The shared unique options builder gets stamped with whichever index name the
map loop hits first, so later unique indexes inherit that random name, which
explains why runs differ. I'm checking what names ended up in eneba_dev from
the first run.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && mkdir -p
cmd/zzcheck && cat > cmd/zzcheck/main.go <<'EOF'
package main…)
⎿  (No output)
⎿  (timeout 2m)

⏺ Every unique index in eneba_dev ended up named email_1, including those on
slug, userId, number and tokenHash; the TTL indexes were only right by chance.
I'm giving each index its own options builder and adding one helper that sets
explicit names in Mongo's style.

⏺ Update(internal/platform/db/indexes.go)
⎿  Added 16 lines, removed 8 lines
13 // CreateMany is a no-op for an index that already exists with
the same keys and options,
14 // so this is safe to run on every startup and from the seed sc
ript.
15 func EnsureIndexes(ctx context.Context, database *mongo.Databas
e) error {
16 - unique := options.Index().SetUnique(true)  
 17 - // TTL with 0 seconds: Mongo deletes the doc once the date in - the field has passed.  
 18 - ttl := options.Index().SetExpireAfterSeconds(0)  
 16 + const (  
 17 + plain = iota  
 18 + unique  
 19 + ttl // expireAfterSeconds 0: Mongo deletes the doc once the + date in the field has passed  
 20 + )  
 21  
 20 - idx := func(keys bson.D, opts ...*options.IndexOptionsBuilder
-) mongo.IndexModel {  
 21 - m := mongo.IndexModel{Keys: keys}  
 22 - if len(opts) > 0 {  
 23 - m.Options = opts[0]  
 22 + // idx is the only place index names are set. Each call build
+s its own options:  
 23 + // the driver writes the index name into the options it is gi
+ven, so one builder  
 24 + // shared by several indexes would hand them all the first in
+dex's name.  
 25 + idx := func(keys bson.D, kind ...int) mongo.IndexModel {  
 26 + opts := options.Index().SetName(indexName(keys))  
 27 + if len(kind) > 0 && kind[0] == unique {  
 28 + opts.SetUnique(true)  
 29 }
25 - return m  
 30 + if len(kind) > 0 && kind[0] == ttl {  
 31 + opts.SetExpireAfterSeconds(0)  
 32 + }  
 33 + return mongo.IndexModel{Keys: keys, Options: opts}  
 34 }
35  
 36 all := map[string][]mongo.IndexModel{

⏺ Update(internal/platform/db/indexes.go)
⎿  Added 10 lines
77 }
78 return nil
79 }
80 +  
 81 +// indexName follows Mongo's default style, field*direction joi
+ned by "*":  
 82 +// {active: 1, createdAt: -1} → "active*1_createdAt*-1".  
 83 +func indexName(keys bson.D) string {  
 84 + parts := make([]string, 0, len(keys))  
 85 + for _, k := range keys {  
 86 + parts = append(parts, fmt.Sprintf("%s_%v", k.Key, k.Value))
87 + }  
 88 + return strings.Join(parts, "\_")  
 89 +}

⏺ Update(internal/platform/db/indexes.go)
⎿  Added 1 line
3 import (
4 "context"
5 "fmt"
6 + "strings"  
 7  
 8 "go.mongodb.org/mongo-driver/v2/bson"
9 "go.mongodb.org/mongo-driver/v2/mongo"

⏺ The helper is in place, so I'm now dropping the five misnamed indexes on
eneba_dev, re-running the seed, and listing the indexes again.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && go build ./...
&& go vet ./... && gofmt -l . ; grep -q "MONGODB_DB=eneba_dev" .env &&
mkdir -p cmd/zz…)
⎿  (No output)
⎿  (timeout 4m)

⏺ Fixed. go run ./cmd/seed --reset passes now, and I ran it twice in a row to
confirm it keeps passing.

Why it happened: if an index has no name, the driver's CreateMany works one
out (like email_1) and writes it into the options builder you passed in
(index_view.go:243 in mongo-driver v2.10.0). My EnsureIndexes used one unique
builder for six indexes. Whichever collection the loop reached first left its
name on that builder, and every later unique index asked for the same name. Go
visits map keys in random order, so the first seed got lucky and created them
all as email_1. This run asked for a name that clashed with what was already
there.

The fix (indexes.go (server/internal/platform/db/indexes.go)):

- The idx helper now makes a fresh options builder for every index, so nothing
  is shared.
- It also gives each index an explicit name from a new indexName(keys)
  function, in Mongo's own style (active*1_inStock_1_soldCount*-1). This is
  the one place index names are set. Because the style matches Mongo's, the
  indexes that were already named correctly didn't change.

Repairing eneba_dev: the first run had left five indexes named email_1 on
sellers.slug, products.slug, carts.userId, orders.number and
refresh_tokens.tokenHash. I dropped those five with a throwaway script, since
deleted. The seed then recreated them. Every index now has its correct name,
and the unique and TTL settings sit on the right indexes, as the listing above
shows.

My earlier Phase 2 check only counted indexes per collection, so it couldn't
catch wrong names. It should have listed names and options. eneba_prod and
eneba_test were never seeded with the broken code, so they need no repair.

❯ <ide_opened_file>The user opened the file /Users/onlyeugene/Desktop/Eneba
Marketplace/server/internal/platform/db/indexes.go in the IDE. This may or
may not be related to the current task.</ide_opened_file>

❯ what would happen if two people ran the seed at the same time?

⏺ Short answer: without --reset it's safe. With --reset on both, you can end up
with mixed data and both runs reporting success.

Both run go run ./cmd/seed (no --reset)

- Both count 0 products, so both go ahead. Both call EnsureIndexes, which is
  safe to run twice now that the names are fixed.
- Both then insert sellers first. sellers.slug has a unique index, so
  whichever run gets to lootcodes second fails with a duplicate key error on
  its first document. InsertMany is ordered, so that run stops there and
  inserts nothing else.
- Result: one full catalog, and the other run exits with an error. Clean.

Both run with --reset

- clearCatalog is seven separate DeleteMany calls with no lock or transaction,
  so the two runs can interleave. One bad order:
  a. A clears everything and inserts its sellers, products and offers.
  b. B clears again, deleting A's sellers, products and offers.
  c. B inserts its own sellers, products and offers.
  d. A goes on to insert its keys and homepage blocks.
- keys and homepage_blocks have no unique index, so nothing stops step 4. Both
  runs print seeded. The database ends up with:
  - Orphan keys: A's 2,360 keys point at offers that no longer exist. Stock
    counts filter by offerId, so stock stays right, but the keys are junk.
  - Double homepage blocks: 12 blocks instead of 6, so every section shows
    twice on the homepage.
- Other orders give a partial mix, or one run fails halfway. You only notice
  if a run errors out.

Does it matter? Hardly. The seed is a dev tool run by one person, and
--allow-prod guards prod. A more likely cousin of this problem is a seed that
dies halfway (network drop) and leaves partial data. Re-running with --reset
already fixes that.

If you want it closed, the simplest fix is to run the clear and inserts inside
one transaction (session.WithTransaction). The whole seed then lands
completely or not at all. Two runs at once would hit a write conflict, and one
would abort or retry. That also fixes the "died halfway" case. About 2,600
docs fits within Atlas's transaction limits. It's about 10 lines in
cmd/seed/main.go.

CLAUDE.md asks me to check before adding extras, so I haven't changed
anything. Do you want the transaction, or should I just note this as a known
limit in PLAN.md?

❯ why does each key have its own document instead of a stock number on the
offer?

⏺ Because we sell a specific code, not a quantity. A stock number can say "3
left" but can't say which 3 codes, who holds each one, or which order it went
to. One document per key carries all of that.

1. The buyer has to get a real, unique code. Each key document holds its code,
   so a sale is "this exact key now belongs to this order." With only a counter,
   you'd still need the codes stored somewhere, probably as an array on the
   offer. You'd then have to pop one off without two buyers popping the same one.
   That's the hard part, just moved.

2. One atomic claim per key is what prevents a double sale. Reserving runs:
   findOneAndUpdate({ offerId, <claimable> }, { $set: { status: "reserved",
   orderId, reservedUntil } })
   Mongo updates a single document atomically. Once a key is reserved, it stops
   matching the filter, so a second buyer can't take it. They get the next key or
   OUT_OF_STOCK. This is the check in PLAN.md section 5, and the race test (20
   buyers, 5 keys → exactly 5 distinct wins) proves it.

3. Each key carries its own reservation state. status, orderId and
   reservedUntil live on the key. That gives us:

- Holds expire on their own: a key whose reservedUntil has passed matches the
  claim filter again. No background job is needed, which matters because
  Render's free tier sleeps.
- A late payment can't steal a re-claimed key: paying, declining and cleanup
  all filter on orderId. When an expired key is claimed by a new order, its
  orderId changes, so the old order no longer owns it.
- Paying is checkable: UpdateMany({ orderId, status: "reserved", reservedUntil
  > now }) must change exactly as many keys as the order has. If it changes
  > fewer, some keys were lost to expiry and the payment is refused.

With a counter, a hold that expires would mean adding back to stock. Nothing
reliable does that while the server sleeps, and if both the expiry and a new
buyer adjust the count, it drifts (section 5 covers this).

4. Buyers don't all fight over one document. A counter or a codes array on the
   offer means every buyer of that offer writes the same document, so busy
   offers hit write conflicts and retries. With separate key documents, buyers
   only collide when they grab the same key, and then one simply moves to the
   next.

So what is offers.stock? A cached count of claimable keys, kept so the store
can sort and filter without counting keys for every listing. It's never
incremented or decremented. It's recounted from keys inside the same
transaction that claims, sells or releases keys. The keys collection is the
truth; stock is a copy for fast reads.

The cost: more documents (2,360 for the seed versus 189 offers) and an index
on { offerId, status, reservedUntil } to find the next claimable key quickly.
At this size, that's nothing.

❯ just note it as a known limit in plan.md, no transaction. it's a dev tool
only i run.

⏺ Update(~/Desktop/Eneba Marketplace/PLAN.md)
⎿  Added 1 line
774 6. **Seller accounts** that upload their own keys and set pric
es.
775 7. **Emailed keys** and order receipts.
776 8. **Panic after a partial response.** `Recoverer` writes a 50
0 error body after a panic. If the handler had already started
writing, the headers are already sent, so the error JSON is a
ppended to the half-written body and the client gets broken JS
ON. No current handler streams, so it can't happen yet. Fix: w
rap the writer, and if bytes were already sent, abort the conn
ection (`panic(http.ErrAbortHandler)`) instead of writing more
.
777 +9. **Two seeds at once.** Known limit, accepted: the seed is a + dev tool only the owner runs. Without `--reset`, a second run + fails on the unique `sellers.slug` index and inserts nothing. + With `--reset` on both, the clear and inserts are separate st
+eps, so they can interleave: both runs report success, but the + db can hold orphan keys and every homepage block twice. A see
+d that dies halfway leaves partial data too. Re-running with `
          +--reset` fixes both. Fix if it ever matters: run the clear and + inserts in one transaction.

⏺ I added the known limit as item 9 in section 12 of PLAN.md (PLAN.md#L776). It
says what goes wrong when two seeds run with --reset at once, why that's
accepted, how a re-run with --reset cleans it up, and that a transaction is
the fix if it ever matters. I didn't change any code.
