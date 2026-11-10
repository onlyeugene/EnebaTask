package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// maxBody caps request bodies. Our largest body is a guest cart merge, far under 1 MB.
const maxBody = 1 << 20

// DecodeJSON reads one JSON object from the body into dst. Malformed JSON, a body that is
// too big, or anything after the object is a 400 BAD_REQUEST.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody))
	if err := dec.Decode(dst); err != nil {
		return NewError(http.StatusBadRequest, CodeBadRequest, "Request body must be valid JSON")
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return NewError(http.StatusBadRequest, CodeBadRequest, "Request body must hold a single JSON object")
	}
	return nil
}
