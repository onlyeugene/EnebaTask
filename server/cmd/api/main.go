package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eugene/internal/auth"
	"eugene/internal/cart"
	"eugene/internal/home"
	"eugene/internal/home/blocks"
	"eugene/internal/keys"
	"eugene/internal/platform/config"
	"eugene/internal/platform/db"
	"eugene/internal/products"
	"eugene/internal/search"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client, database, err := db.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		slog.Error("db", "err", err)
		os.Exit(1)
	}
	defer client.Disconnect(context.Background())

	if err := db.EnsureIndexes(ctx, database); err != nil {
		slog.Error("indexes", "err", err)
		os.Exit(1)
	}

	// Only a warning: the rest of the API works without search, and the seed creates the index.
	if err := db.CheckSearchIndex(ctx, database); err != nil {
		slog.Warn("search index", "err", err)
	}

	authSvc, err := auth.NewService(auth.NewMongoUsers(database), auth.NewMongoRefresh(database), auth.NewMongoRevoked(database),
		auth.Settings{
			Secret: cfg.JWTSecret, Issuer: cfg.JWTIssuer,
			AccessTTL: cfg.AccessTokenTTL, RefreshTTL: cfg.RefreshTokenTTL, BcryptCost: cfg.BcryptCost,
		}, time.Now)
	if err != nil {
		slog.Error("auth", "err", err)
		os.Exit(1)
	}

	keyRepo := keys.NewMongoRepo(database)
	searchSvc := search.NewService(search.NewMongoRepo(database))
	productsSvc := products.NewService(products.NewMongoRepo(database), keyRepo, time.Now)
	homeSvc := home.NewService(home.NewMongoRepo(database), blocks.Registry(productsSvc))
	cartSvc := cart.NewService(cart.NewMongoRepo(database), keyRepo, time.Now)
	authHandler := auth.NewHandler(authSvc)
	h := handlers{
		auth:     authHandler,
		products: products.NewHandler(productsSvc, searchSvc),
		search:   search.NewHandler(searchSvc),
		home:     home.NewHandler(homeSvc, cfg.IsDev()),
		cart:     cart.NewHandler(cartSvc, authHandler.RequireAuth),
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           newRouter(cfg, database, h),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("listening", "port", cfg.Port, "env", cfg.AppEnv, "db", cfg.MongoDB)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	// Render sends SIGTERM on deploy; give in-flight requests a few seconds to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	slog.Info("stopped")
}
