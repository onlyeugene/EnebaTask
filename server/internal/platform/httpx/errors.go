package httpx

import (
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

// WriteError writes the API.md error body. Unknown errors are logged and hidden,
// so driver messages or stack details never reach the browser.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		slog.Error("unhandled error", "method", r.Method, "path", r.URL.Path, "err", err)
		appErr = NewError(http.StatusInternalServerError, CodeInternal, "Something went wrong")
	}
	JSON(w, appErr.Status, map[string]any{"error": appErr})
}
