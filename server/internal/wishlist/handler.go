package wishlist

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"eugene/internal/auth"
	"eugene/internal/platform/httpx"
)

type Handler struct {
	svc         *Service
	requireAuth func(http.Handler) http.Handler
}

// NewHandler takes auth's RequireAuth so this feature can guard its own routes.
func NewHandler(svc *Service, requireAuth func(http.Handler) http.Handler) *Handler {
	return &Handler{svc: svc, requireAuth: requireAuth}
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Get(r.Context(), auth.ClaimsFrom(r.Context()).UserID)
	respond(w, r, v, err)
}

func (h *Handler) put(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Add(r.Context(), auth.ClaimsFrom(r.Context()).UserID, chi.URLParam(r, "productId"))
	respond(w, r, v, err)
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Remove(r.Context(), auth.ClaimsFrom(r.Context()).UserID, chi.URLParam(r, "productId"))
	respond(w, r, v, err)
}

func (h *Handler) merge(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Items []MergeInput `json:"items"`
	}
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	v, err := h.svc.Merge(r.Context(), auth.ClaimsFrom(r.Context()).UserID, body.Items)
	respond(w, r, v, err)
}

func respond(w http.ResponseWriter, r *http.Request, v View, err error) {
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, v)
}

// toHTTP maps service errors to API.md codes. Anything else becomes a logged 500.
func toHTTP(err error) error {
	var v *ValidationError
	switch {
	case errors.As(err, &v):
		e := httpx.NewError(http.StatusUnprocessableEntity, httpx.CodeValidation, "Some fields are invalid")
		e.Fields = v.Fields
		return e
	case errors.Is(err, ErrFull):
		e := httpx.NewError(http.StatusUnprocessableEntity, httpx.CodeValidation, "Your wishlist is full")
		e.Fields = map[string]string{"productId": fmt.Sprintf("wishlist holds at most %d products", MaxItems)}
		return e
	case errors.Is(err, ErrProductNotFound):
		return httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Product not found")
	}
	return err
}
