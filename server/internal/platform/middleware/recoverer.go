package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"eugene/internal/platform/httpx"
)

// Recoverer turns a panic into a 500 with the normal error body, instead of chi's plain-text one.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			// http.ErrAbortHandler is how net/http aborts on purpose; let it through.
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			slog.Error("panic", "err", rec, "stack", string(debug.Stack()))
			httpx.WriteError(w, r, fmt.Errorf("panic: %v", rec))
		}()
		next.ServeHTTP(w, r)
	})
}
