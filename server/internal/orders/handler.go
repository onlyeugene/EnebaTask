package orders

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

// create takes no body fields: the order is made from the user's server cart.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Create(r.Context(), auth.ClaimsFrom(r.Context()).UserID)
	var oos *OutOfStockError
	if errors.As(err, &oos) {
		// API.md: the error plus "details", so the client can point at the line and fix it.
		httpx.JSON(w, http.StatusConflict, map[string]any{
			"error":   httpx.NewError(http.StatusConflict, httpx.CodeOutOfStock, "Not enough keys left for an item in your cart"),
			"details": map[string]any{"offerId": oos.OfferID.Hex(), "available": oos.Available},
		})
		return
	}
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.Created(w, map[string]any{"order": v})
}

func (h *Handler) pay(w http.ResponseWriter, r *http.Request) {
	var card Card
	if err := httpx.DecodeJSON(w, r, &card); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	v, err := h.svc.Pay(r.Context(), auth.ClaimsFrom(r.Context()).UserID, chi.URLParam(r, "id"), card)
	if errors.Is(err, ErrDeclined) {
		// API.md: 402 with the failed order, so the client can show it and send the user back to the cart.
		httpx.JSON(w, http.StatusPaymentRequired, map[string]any{
			"error": httpx.NewError(http.StatusPaymentRequired, httpx.CodePaymentDeclined, "Your card was declined"),
			"order": v,
		})
		return
	}
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, map[string]any{"order": v})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context(), auth.ClaimsFrom(r.Context()).UserID)
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, map[string]any{"items": list})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Get(r.Context(), auth.ClaimsFrom(r.Context()).UserID, chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, map[string]any{"order": v})
}

// toHTTP maps service errors to API.md codes. Anything else becomes a logged 500.
func toHTTP(err error) error {
	var v *ValidationError
	switch {
	case errors.As(err, &v):
		e := httpx.NewError(http.StatusUnprocessableEntity, httpx.CodeValidation, "Some fields are invalid")
		e.Fields = v.Fields
		return e
	case errors.Is(err, ErrCartEmpty):
		e := httpx.NewError(http.StatusUnprocessableEntity, httpx.CodeValidation, "Your cart is empty")
		e.Fields = map[string]string{"cart": "is empty"}
		return e
	case errors.Is(err, ErrNotFound):
		return httpx.NewError(http.StatusNotFound, httpx.CodeNotFound, "Order not found")
	case errors.Is(err, ErrNotPayable):
		return httpx.NewError(http.StatusConflict, httpx.CodeOrderNotPayable, "This order can no longer be paid")
	}
	return err
}
