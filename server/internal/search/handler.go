package search

import (
	"net/http"

	"eugene/internal/platform/httpx"
	"eugene/internal/products"
)

// Handler serves autocomplete only. Store search (?q= on /api/products) goes through
// products.Handler, which calls Service.Search: chi can't mount two routers on /api/products.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) autocomplete(w http.ResponseWriter, r *http.Request) {
	text, err := products.SearchText(r.URL.Query().Get("q"))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	suggestions, err := h.svc.Autocomplete(r.Context(), text)
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.OK(w, map[string]any{"suggestions": suggestions})
}
