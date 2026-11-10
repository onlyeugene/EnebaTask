package auth

import (
	"errors"
	"net/http"

	"eugene/internal/platform/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	session, err := h.svc.Register(r.Context(), body.Email, body.Password, body.Name)
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.Created(w, session)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	session, err := h.svc.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, session)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, err := h.svc.Me(r.Context(), ClaimsFrom(r.Context()).UserID)
	if err != nil {
		httpx.WriteError(w, r, toHTTP(err))
		return
	}
	httpx.OK(w, map[string]any{"user": user})
}

// toHTTP maps service errors to the API.md codes. Anything not listed stays a plain error,
// which WriteError logs and hides behind a 500.
func toHTTP(err error) error {
	var v *ValidationError
	switch {
	case errors.As(err, &v):
		e := httpx.NewError(http.StatusUnprocessableEntity, httpx.CodeValidation, "Some fields are invalid")
		e.Fields = v.Fields
		return e
	case errors.Is(err, ErrEmailTaken):
		return httpx.NewError(http.StatusConflict, httpx.CodeEmailTaken, "An account with this email already exists")
	case errors.Is(err, ErrInvalidCredentials):
		return httpx.NewError(http.StatusUnauthorized, httpx.CodeUnauthorized, "Invalid email or password")
	case errors.Is(err, ErrTokenExpired):
		return httpx.NewError(http.StatusUnauthorized, httpx.CodeTokenExpired, "Access token expired")
	case errors.Is(err, ErrTokenRevoked):
		return httpx.NewError(http.StatusUnauthorized, httpx.CodeTokenRevoked, "Access token revoked")
	case errors.Is(err, ErrTokenInvalid):
		return httpx.NewError(http.StatusUnauthorized, httpx.CodeUnauthorized, "Invalid access token")
	}
	return err
}
