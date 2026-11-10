package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"eugene/internal/platform/config"
	"eugene/internal/platform/db"
	"eugene/internal/platform/httpx"
	"eugene/internal/platform/middleware"
)

func newRouter(cfg config.Config, client *mongo.Client) http.Handler {
	r := chi.NewRouter()

	// Order matters: request id first so the logger can print it; recoverer
	// inside the logger so a panic is still logged as a 500.
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS(cfg.CORSOrigins))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Route not found"))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, r, httpx.NewError(http.StatusMethodNotAllowed, httpx.CodeBadRequest, "Method not allowed"))
	})

	r.Get("/healthz", healthz(client))

	return r
}

// healthz pings Mongo so Render's health check fails when the db is unreachable,
// not just when the process is down.
func healthz(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx, client); err != nil {
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "db": "down"})
			return
		}
		httpx.OK(w, map[string]string{"status": "ok", "db": "ok"})
	}
}
