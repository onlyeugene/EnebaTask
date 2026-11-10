package cart

import "github.com/go-chi/chi/v5"

// Routes is mounted at /api/cart. /quote is public (guest carts); the rest need a signed-in user.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/quote", h.quote)
	r.Group(func(r chi.Router) {
		r.Use(h.requireAuth)
		r.Get("/", h.get)
		r.Put("/items/{offerId}", h.put)
		r.Delete("/items/{offerId}", h.remove)
		r.Post("/merge", h.merge)
	})
	return r
}
