# CLAUDE.md

"Eugene" — an Eneba-style game key marketplace (homework). Go API in `/server`, Next.js frontend in `/client`, MongoDB Atlas.

Read first: [ASSIGNMENT.md](ASSIGNMENT.md), [PLAN.md](PLAN.md), [API.md](API.md), [design-reference/notes.md](design-reference/notes.md). Screenshots are in `design-reference/screenshots/` (gitignored); open the matching ones before building a page.

## Rules

1. **Check against ASSIGNMENT.md before calling a feature done.** Name the assignment item it covers and what is still missing.
2. **One phase at a time.** Each phase ends with a check that was actually run. Mark the phase `✅` in PLAN.md when done.
3. **Explain any decision that isn't obvious**, in a short code comment or in PLAN.md. The owner must be able to explain every line. Prefer plain code over clever code.

## Folder rules

### Server (`/server`)
- `cmd/api/main.go` wires only: config → db → routes → cleanup job → ListenAndServe. No logic.
- `cmd/seed/main.go` for the seed script. Data in `cmd/seed/data.go`.
- `internal/` is grouped by feature: `auth`, `products`, `search`, `cart`, `orders`, `home`, `keys`.
- Each feature has exactly: `model.go` (types), `repository.go` (interface), `repository_mongo.go` (Mongo implementation), `service.go` (business logic), `handler.go` (HTTP only), `routes.go` (its chi routes). Tests live next to the file they test: `service_test.go`, not `auth_test.go`.
- **Each feature owns its routes.** `internal/<feature>/routes.go` has `func (h *Handler) Routes() chi.Router` (exported so `cmd/api` can call it), with paths relative to the mount point. `cmd/api/routes.go` holds only: global middleware, NotFound/MethodNotAllowed, `/healthz`, and an `/api` router that mounts each feature: `api.Mount("/auth", authHandler.Routes())`. No top-level `routes/` folder.
- **Handlers never call Mongo directly.** They call service methods only.
- **Services never import chi or net/http.** They call repository interfaces only.
- Shared infrastructure goes in `internal/platform/`: `config/`, `db/`, `httpx/`, `middleware/`. Platform packages never import feature packages.
- Short lowercase file names. No `_util`, no `_helper`.

### Client (`/client`)
- `app/` is routes and layouts only. Each `page.tsx` renders one view component from `features/`. No query calls, no business logic in `app/`.
- `features/` is grouped by feature: `home`, `store`, `search`, `product`, `cart`, `checkout`, `orders`, `auth`. Each feature owns its `components/`, `hooks/`, and `api.ts`.
- `components/ui/` is the shared design system: Button, Badge, Input, Price, ProductCard, Skeleton, Modal, ErrorBoundary, ErrorState. No imports from `features/` or `lib/client.ts`.
- `components/layout/` holds the Header, NavBar, Footer (server) and CartButton (client). No feature imports.
- `lib/` holds `client.ts` (fetch + refresh), `tokens.ts`, `query-client.ts`, `providers.tsx`, `format.ts`. No feature imports.
- `types/api.ts` holds all API types from API.md. No runtime code.
- kebab-case filenames for all client files.
- One job per file. Split before a file passes ~150 lines.

## Working notes

- Run the API locally with live reload: `cd server && air` (config in `server/.air.toml`, builds to `server/tmp/`). Install once with `go install github.com/air-verse/air@latest`. Without air: `go run ./cmd/api`.
- API.md is the contract. Update it in the same phase as any endpoint change.
- Money is integer cents everywhere. Never floats.
- Client: no Next API routes, no rewrites, no data fetching in server components. Data views are client components using TanStack Query and `apiFetch`. Layouts, header shell, footer and static content stay server components.
- Styling only through the tokens in `client/src/app/globals.css` (from notes.md). Square corners.
- No copyrighted images or brand logos. Footer says the app is a demo, not affiliated with Eneba.
- Never run git commands (no init, add, commit, or push). Git is handled by the owner.
- Never write `.env` files. Never seed `eneba_prod` without `--allow-prod`. Never point `MONGODB_TEST_URI` at `eneba_dev` or `eneba_prod`.
- Integration tests are in `*_integration_test.go` with `//go:build integration`. Normal `go test ./...` never runs them. Run them with:
  ```
  MONGODB_TEST_URI="mongodb+srv://..." go test -tags integration ./internal/orders/... -v -count=1
  ```
  They always use the `eneba_test` database, which they drop and recreate before each test.
- Keep the full agent chat log; it is a required deliverable.
