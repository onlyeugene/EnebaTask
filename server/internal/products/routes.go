package products

import "github.com/go-chi/chi/v5"

// Routes is mounted at /api/products. chi tries static segments before {slug},
// so /filters never gets read as a product slug.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Get("/filters", h.filters)
	r.Get("/{slug}", h.detail)
	r.Get("/{slug}/cover.svg", h.cover)
	return r
}
