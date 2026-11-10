package home

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"eugene/internal/platform/httpx"
)

type Handler struct {
	svc *Service
	// allowFail turns on ?fail=1, which forces a block to fail so the client's
	// per-block error state can be shown. Development only.
	allowFail bool
}

func NewHandler(svc *Service, allowFail bool) *Handler {
	return &Handler{svc: svc, allowFail: allowFail}
}

func (h *Handler) layout(w http.ResponseWriter, r *http.Request) {
	blocks, err := h.svc.Layout(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.OK(w, map[string]any{"blocks": blocks})
}

func (h *Handler) block(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if h.allowFail && r.URL.Query().Get("fail") == "1" {
		writeUnavailable(w, r)
		return
	}

	res, err := h.svc.Resolve(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, r, httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Block not found"))
		return
	}
	if err != nil {
		// 503, not 500: only this block failed, and the client shows a retry for it alone.
		// Logged here because WriteError only logs errors it hides as 500s.
		slog.Error("block unavailable", "id", id, "err", err)
		writeUnavailable(w, r)
		return
	}
	httpx.OK(w, res)
}

func writeUnavailable(w http.ResponseWriter, r *http.Request) {
	httpx.WriteError(w, r, httpx.NewError(http.StatusServiceUnavailable, httpx.CodeBlockUnavailable, "This block is unavailable right now"))
}
