# Deploy

The Go API runs on Render (Docker, built from `server/`), the Next.js client on Vercel (from `client/`), and the data on the same Atlas M0 cluster as dev, in the `eneba_prod` database.

Do the steps in order. Never paste a secret into a file, a commit or a chat. Each one goes straight from its source (Atlas, your `server/.env`) into the Render or Vercel form.

Files in the repo:

- [server/Dockerfile](server/Dockerfile): builds the API binary in `golang:1.25-alpine` and copies it into `distroless/static` (about 21 MB, no shell, runs as non-root).
- [server/.dockerignore](server/.dockerignore): keeps `.env`, `tmp/` and `bin/` out of the image.
- [render.yaml](render.yaml): the Render Blueprint. It holds the plain settings; it asks for the secrets once and never stores them.

---

## Environment variables

### Render (API)

| Var | Value | New for prod? | Where it's set |
|---|---|---|---|
| `APP_ENV` | `production` | **yes** (dev: `development`) | `render.yaml` |
| `MONGODB_DB` | `eneba_prod` | **yes** (dev: `eneba_dev`) | `render.yaml` |
| `JWT_SECRET` | 32 random bytes | **yes**, never the dev one | Render makes it (`generateValue`); you never see or type it |
| `CORS_ORIGINS` | first `http://localhost:3000`, then your Vercel URL | **yes** | Asked for when you create the Blueprint; changed in step 5 |
| `MONGODB_URI` | the cluster's `mongodb+srv://…` string | no, same cluster as dev | Asked for when you create the Blueprint; copy it from `server/.env` |
| `JWT_ISSUER` | `eugene-api` | no | `render.yaml` |
| `ACCESS_TOKEN_TTL` | `15m` | no | `render.yaml` |
| `REFRESH_TOKEN_TTL` | `720h` | no | `render.yaml` |
| `RESERVATION_TTL` | `10m` | no | `render.yaml` |
| `BCRYPT_COST` | `12` | no | `render.yaml` |
| `PORT` | — | — | Render sets it; don't add it |

`APP_ENV=production` does two things: it turns off the `?fail=1` homepage switch, and it makes the API read the client IP from the last `X-Forwarded-For` entry (Render's proxy).

### Vercel (client)

| Var | Value | New for prod? |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | your Render URL, e.g. `https://eugene-api.onrender.com`, **no trailing slash** | **yes** (dev: `http://localhost:8080`) |

It isn't a secret: Next.js writes it into the browser code at build time. That also means a change needs a new deploy to take effect.

---

## 1. Atlas (before anything else)

- [ ] **Network Access** (Security → Network Access): `0.0.0.0/0` is listed and **Active**. Render's free tier has no fixed outbound IP. Without it, the API's startup ping times out, it exits, and Render restarts it in a loop. The log says only `mongo ping: ... server selection timeout`, not "IP not allowed".
- [ ] **Database Access**: open the DB user the URI uses and check it can read and write `eneba_prod`. If its privileges list each database by name, add `eneba_prod` with the same role `eneba_dev` has. If it has a cluster-wide role (e.g. `readWriteAnyDatabase`), there is nothing to do.
- [ ] **Search indexes** (cluster → Atlas Search): M0 allows 3. `eneba_dev` uses 1 and `eneba_prod` will use 1. If there are already 3, delete any you don't need first.

## 2. Seed `eneba_prod` and create its search index

From your machine. `MONGODB_URI` comes from `server/.env`; the inline `MONGODB_DB` wins over the file, because the loader never overrides a var that is already set.

```
cd server
MONGODB_DB=eneba_prod go run ./cmd/seed --allow-prod
```

- [ ] It ends without an error. The seed creates the `products_search` index from `atlas/products_search.json` and waits until Atlas reports it `READY`.
- [ ] If it says `eneba_prod already has N products`, re-run with `--reset` added. That wipes the catalog, carts and orders in `eneba_prod`, so only do it before go-live.
- [ ] In Atlas → Browse Collections, `eneba_prod` has `products`, `offers`, `keys`, `sellers` and `homepage_blocks`, and Atlas Search shows `products_search` on `eneba_prod.products` as **Active/Ready**.

Fallback if the index step fails: Atlas Search → Create Search Index → JSON editor → database `eneba_prod`, collection `products`, name `products_search`, and paste the `definition` from `server/atlas/products_search.json`.

## 3. Render (API)

Merge the phase 16 branch into `main` first; Render builds from `main`.

- [ ] Check `region` in `render.yaml` (now `frankfurt`) is the one nearest your Atlas cluster's region (Atlas → cluster overview). Render can't change it later. If it's wrong, change it before merging.

- [ ] Render dashboard → **New → Blueprint** → connect the GitHub repo → branch `main`. Render reads `render.yaml` and shows one web service, `eugene-api`.
- [ ] It asks for two values:
  - `MONGODB_URI`: paste the URI from `server/.env`.
  - `CORS_ORIGINS`: `http://localhost:3000` for now. It can't be empty, and you don't know the Vercel URL yet.
- [ ] **Apply**. Wait for the build (a few minutes) and the deploy.
- [ ] Logs show `"msg":"listening"` with `"env":"production"` and `"db":"eneba_prod"`, and no `search index` warning.
- [ ] Copy the service URL from the top of the page (`https://eugene-api…onrender.com`). Then:
  ```
  curl https://<render-url>/healthz
  ```
  → `{"db":"ok","search":"ok","status":"ok"}`. On the free plan the first call after a sleep can take up to a minute.

If it restarts in a loop with `server selection timeout`, go back to step 1's network check.

## 4. Vercel (client)

- [ ] Vercel → **Add New → Project** → import the same repo.
- [ ] **Root Directory**: `client`. Framework preset: Next.js (found on its own). Leave build and output settings as they are.
- [ ] **Environment Variables**: `NEXT_PUBLIC_API_URL` = the Render URL from step 3, no trailing slash.
- [ ] **Deploy**.
- [ ] Note the production domain under Project → Settings → Domains, e.g. `https://<project>.vercel.app`. Use that one, not the per-deploy URL with a hash in it.

## 5. CORS: let the Vercel site call the API

Until this step the site loads, but every API call fails with a CORS error in the browser console.

- [ ] Render → `eugene-api` → **Environment** → `CORS_ORIGINS` → set it to the exact Vercel domain: `https://<project>.vercel.app`. Scheme included, no trailing slash, no path. For more than one, separate them with commas.
- [ ] **Save, rebuild and deploy** (or Save, then Manual Deploy). Wait for the new deploy to be live.
- [ ] Check from your terminal:
  ```
  curl -s -o /dev/null -D - -X OPTIONS \
    -H "Origin: https://<project>.vercel.app" \
    -H "Access-Control-Request-Method: GET" \
    https://<render-url>/api/homepage | grep -i access-control-allow-origin
  ```
  → prints `Access-Control-Allow-Origin: https://<project>.vercel.app`. No line printed means the value doesn't match exactly.

Vercel preview deploys get random URLs that aren't in this list, so previews can't reach the API. Accepted (PLAN.md section 7). If you add a custom domain or rename the project later, update `CORS_ORIGINS` again.

## 6. Verify the full flow on the live URLs

Open `https://<project>.vercel.app` with DevTools → Network open.

- [ ] **Home**: every block loads (hero, rails, categories). Requests go to `onrender.com`. No red CORS errors in the console.
- [ ] **Demo switch off**: open `/?fail=<blockId>` with a block id from the `GET /api/homepage` response in the Network tab. The block still loads, since the API ignores `?fail=1` in production.
- [ ] **Search**: type `skyforje` in the header. Suggestions show Skyforge. Enter opens the store with results.
- [ ] **Store**: tick a platform, change the sort, go to page 2, press Back. The URL and grid follow.
- [ ] **Product page**: covers load (they come from the API), offers are listed.
- [ ] **Guest cart**: add 2 products as a guest, reload. Both are still in the cart.
- [ ] **Register** a new account. The guest cart merges into it.
- [ ] **Pay**: cart → Checkout now → card `4242 4242 4242 4242`, any future `MM/YY`, any 3-digit CVC, any name → Pay. The order page shows the keys.
- [ ] **Decline**: add something, check out with `4000 0000 0000 0002`. You land on the cart with "Your card was declined", the cart unchanged, and the product's stock the same as before.
- [ ] **Orders**: `/orders` lists the paid and the failed order. Each opens.
- [ ] **Session**: log out, log back in, the cart is there. Open `/orders` while logged out → redirected to login.

### Logged IP (PLAN.md section 7)

- [ ] Render → `eugene-api` → **Logs**. Find a `"msg":"request"` line from your own clicks and read its `ip` field.
- [ ] Compare it with your public IP (search "what is my IP" in your browser).
- [ ] **Same** → done. The login rate limit counts each user on their own.
- [ ] **Different** (e.g. a `10.x`, `100.x` or proxy address) → Render has put a second proxy in front, so every user shares one rate limit bucket. Don't fix it in the dashboard; report it so `middleware.ClientIP` can be changed to pick the right `X-Forwarded-For` entry.

When every box is ticked, mark phase 16 `✅` in PLAN.md.
