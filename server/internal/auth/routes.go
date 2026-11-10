package auth

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"eugene/internal/platform/httpx"
)

// Routes is mounted at /api/auth.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// One limiter shared by /register and /login only: 10 requests per minute per IP across both,
	// to slow down password guessing. /me and /logout are not limited: they already need a valid
	// token, and /me runs on every protected page and in the header.
	login := perIP(10)
	// /refresh has its own limiter, so failed logins from the same IP (a shared office or home
	// network) never block a signed-in tab from refreshing. 60 is generous: each tab refreshes
	// about once per 15 minutes. It is still limited because a request costs db writes.
	refresh := perIP(60)

	r.With(login).Post("/register", h.register)
	r.With(login).Post("/login", h.login)
	r.With(refresh).Post("/refresh", h.refresh)
	r.With(h.RequireAuth).Post("/logout", h.logout)
	r.With(h.RequireAuth).Get("/me", h.me)
	return r
}

// perIP allows n requests per minute per client IP. Each call makes its own counter.
// The key is r.RemoteAddr, which middleware.ClientIP has already set to the real client IP.
// Not httprate.KeyByRealIP: it trusts headers the client can fake.
func perIP(n int) func(http.Handler) http.Handler {
	return httprate.Limit(n, time.Minute,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			return httprate.CanonicalizeIP(r.RemoteAddr), nil
		}),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			httpx.WriteError(w, r, httpx.NewError(http.StatusTooManyRequests, httpx.CodeRateLimited, "Too many attempts, try again in a minute"))
		}),
	)
}
