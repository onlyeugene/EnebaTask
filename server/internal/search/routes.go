package search

import "github.com/go-chi/chi/v5"

// Routes is mounted at /api/search.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/autocomplete", h.autocomplete)
	return r
}
