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
	// to slow down password guessing. /me is not limited: the client calls it on every protected
	// page and in the header, and it already needs a valid token. /refresh (phase 6) gets its own,
	// higher limit, since every open tab refreshes every 15 minutes.
	// The key is r.RemoteAddr, which middleware.ClientIP has already set to the real client IP.
	// Not httprate.KeyByRealIP: it trusts headers the client can fake.
	limit := httprate.Limit(10, time.Minute,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			return httprate.CanonicalizeIP(r.RemoteAddr), nil
		}),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			httpx.WriteError(w, r, httpx.NewError(http.StatusTooManyRequests, httpx.CodeRateLimited, "Too many attempts, try again in a minute"))
		}),
	)

	r.With(limit).Post("/register", h.register)
	r.With(limit).Post("/login", h.login)
	r.With(h.RequireAuth).Get("/me", h.me)
	return r
}
