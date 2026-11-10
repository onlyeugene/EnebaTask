package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"eugene/internal/auth"
	"eugene/internal/platform/config"
	"eugene/internal/platform/db"
	"eugene/internal/platform/httpx"
	"eugene/internal/platform/middleware"
	"eugene/internal/products"
	"eugene/internal/search"
)

// handlers holds one HTTP handler per feature. main.go builds them; newRouter only mounts them.
type handlers struct {
	auth     *auth.Handler
	products *products.Handler
	search   *search.Handler
}

func newRouter(cfg config.Config, database *mongo.Database, h handlers) http.Handler {
	r := chi.NewRouter()

	// Order matters: request id and client IP first so the logger can print them;
	// recoverer inside the logger so a panic is still logged as a 500.
	// Not chimw.RealIP: it trusts X-Real-IP and the first X-Forwarded-For entry,
	// which any client can set, so per-IP rate limits could be dodged.
	r.Use(chimw.RequestID)
	r.Use(middleware.ClientIP(!cfg.IsDev()))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS(cfg.CORSOrigins))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Route not found"))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, r, httpx.NewError(http.StatusMethodNotAllowed, httpx.CodeBadRequest, "Method not allowed"))
	})

	r.Get("/healthz", healthz(database))

	r.Route("/api", func(api chi.Router) {
		api.Mount("/auth", h.auth.Routes())
		api.Mount("/products", h.products.Routes())
		api.Mount("/search", h.search.Routes())
	})

	return r
}

// healthz pings Mongo so Render's health check fails when the db is unreachable,
// not just when the process is down.
//
// It also reports the Atlas Search index, because $search on a missing index returns no
// results and no error, so a broken search looks like "nothing matches". A missing index
// still returns 200: Render restarts a service that fails its health check, and a restart
// can't create an index, while the rest of the API works fine without it. Only db down is 503.
func healthz(database *mongo.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx, database.Client()); err != nil {
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "db": "down"})
			return
		}
		// "missing" also covers an index that exists but can't be queried yet (still building).
		if err := db.CheckSearchIndex(ctx, database); err != nil {
			httpx.OK(w, map[string]string{"status": "degraded", "db": "ok", "search": "missing"})
			return
		}
		httpx.OK(w, map[string]string{"status": "ok", "db": "ok", "search": "ok"})
	}
}
