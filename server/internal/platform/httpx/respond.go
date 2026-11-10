package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write json", "err", err)
	}
}

func OK(w http.ResponseWriter, body any)      { JSON(w, http.StatusOK, body) }
func Created(w http.ResponseWriter, body any) { JSON(w, http.StatusCreated, body) }
func NoContent(w http.ResponseWriter)         { w.WriteHeader(http.StatusNoContent) }
