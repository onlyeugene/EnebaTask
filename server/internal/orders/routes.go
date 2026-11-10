package orders

import "github.com/go-chi/chi/v5"

// Routes is mounted at /api/orders. Every route needs a signed-in user.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAuth)
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{id}", h.get)
	r.Post("/{id}/pay", h.pay)
	return r
}
