package wishlist

import "github.com/go-chi/chi/v5"

// Routes is mounted at /api/wishlist. Every route needs a signed-in user;
// guests keep their wishlist in the browser.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(h.requireAuth)
	r.Get("/", h.get)
	r.Put("/{productId}", h.put)
	r.Delete("/{productId}", h.remove)
	r.Post("/merge", h.merge)
	return r
}
