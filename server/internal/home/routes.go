package home

import "github.com/go-chi/chi/v5"

// Routes is mounted at /api/homepage.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.layout)
	r.Get("/blocks/{id}", h.block)
	return r
}
