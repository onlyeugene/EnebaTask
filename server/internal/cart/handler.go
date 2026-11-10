package cart

import (
	"errors"
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

type itemsBody struct {
	Items []ItemInput `json:"items"`
}

func (h *Handler) quote(w http.ResponseWriter, r *http.Request) {
	var body itemsBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	v, err := h.svc.Quote(r.Context(), body.Items)
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, v)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Get(r.Context(), auth.ClaimsFrom(r.Context()).UserID)
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, v)
}

func (h *Handler) put(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Quantity int `json:"quantity"`
	}
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	v, err := h.svc.SetQuantity(r.Context(), auth.ClaimsFrom(r.Context()).UserID, chi.URLParam(r, "offerId"), body.Quantity)
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, v)
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Remove(r.Context(), auth.ClaimsFrom(r.Context()).UserID, chi.URLParam(r, "offerId"))
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, v)
}

func (h *Handler) merge(w http.ResponseWriter, r *http.Request) {
	var body itemsBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	v, dropped, err := h.svc.Merge(r.Context(), auth.ClaimsFrom(r.Context()).UserID, body.Items)
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, map[string]any{"cart": v, "dropped": dropped})
}

// toHTTP maps service errors to API.md codes. Anything else becomes a logged 500.
func toHTTP(err error) error {
	var v *ValidationError
	switch {
	case errors.As(err, &v):
		e := httpx.NewError(http.StatusUnprocessableEntity, httpx.CodeValidation, "Some fields are invalid")
		e.Fields = v.Fields
		return e
	case errors.Is(err, ErrOfferNotFound):
		return httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Offer not found")
	}
	return err
}
