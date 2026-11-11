package httpx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// Codes from API.md. The client switches on these, so they never change once shipped.
const (
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeTokenExpired     = "TOKEN_EXPIRED"
	CodeTokenRevoked     = "TOKEN_REVOKED"
	CodeTokenReused      = "TOKEN_REUSED"
	CodePaymentDeclined  = "PAYMENT_DECLINED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeEmailTaken       = "EMAIL_TAKEN"
	CodeOutOfStock       = "OUT_OF_STOCK"
	CodeOrderNotPayable  = "ORDER_NOT_PAYABLE"
	CodeValidation       = "VALIDATION_FAILED"
	CodeRateLimited      = "RATE_LIMITED"
	CodeBlockUnavailable = "BLOCK_UNAVAILABLE"
	CodeInternal         = "INTERNAL"
)

// AppError is an error the client is allowed to see. Anything else becomes a plain 500.
type AppError struct {
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *AppError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func NewError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

// StatusClientClosed is nginx's 499 "client closed request": no standard code
// exists for a caller that hung up before the answer was ready.
const StatusClientClosed = 499

// WriteError writes the API.md error body. Unknown errors are logged and hidden,
// so driver messages or stack details never reach the browser.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	// The client went away (tab closed, page reloaded, fetch aborted), so the
	// request's context was cancelled and the db call stopped with it. Nothing
	// broke on our side, so it's not logged as an error or counted as a 500.
	// Both checks are needed: a context.Canceled while the request's own context
	// is still live came from somewhere else in our code, and is a real 500.
	// No body: nobody is left to read it.
	if errors.Is(err, context.Canceled) && r.Context().Err() != nil {
		slog.Debug("client closed request", "method", r.Method, "path", r.URL.Path, "err", err)
		w.WriteHeader(StatusClientClosed)
		return
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		slog.Error("unhandled error", "method", r.Method, "path", r.URL.Path, "err", err)
		appErr = NewError(http.StatusInternalServerError, CodeInternal, "Something went wrong")
	}
	JSON(w, appErr.Status, map[string]any{"error": appErr})
}
