package auth

import (
	"context"
	"net/http"
	"strings"

	"eugene/internal/platform/httpx"
)

type ctxKey struct{}

// RequireAuth lets a request through only with a valid, unrevoked access token
// (Authorization: Bearer <token>), and puts the user id and jti in the context.
// Cart and orders take this as a dependency and apply it in their own Routes().
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			httpx.WriteError(w, r, httpx.NewError(http.StatusUnauthorized, httpx.CodeUnauthorized, "Missing access token"))
			return
		}
		claims, err := h.svc.Authenticate(r.Context(), token)
		if err != nil {
			httpx.WriteError(w, r, toHTTP(err))
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, claims)))
	})
}

// ClaimsFrom returns what RequireAuth stored. Only call it behind RequireAuth;
// elsewhere it returns zero Claims.
func ClaimsFrom(ctx context.Context) Claims {
	c, _ := ctx.Value(ctxKey{}).(Claims)
	return c
}
