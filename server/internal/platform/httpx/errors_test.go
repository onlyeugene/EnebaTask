package httpx

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A request the client gave up on ends as 499 with no error log; the same
// error while the client is still waiting stays a logged 500.
func TestWriteErrorClientClosed(t *testing.T) {
	var logs bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	defer slog.SetDefault(old)

	// Wrapped, as a repository would return it.
	err := fmt.Errorf("find order: %w", context.Canceled)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	WriteError(w, httptest.NewRequest(http.MethodGet, "/api/orders", nil).WithContext(ctx), err)
	if w.Code != StatusClientClosed {
		t.Fatalf("client gone: status %d, want %d", w.Code, StatusClientClosed)
	}
	if strings.Contains(logs.String(), "level=ERROR") || !strings.Contains(logs.String(), "level=DEBUG") {
		t.Fatalf("client gone: want one debug line and no error, got:\n%s", logs.String())
	}

	logs.Reset()
	w = httptest.NewRecorder()
	WriteError(w, httptest.NewRequest(http.MethodGet, "/api/orders", nil), err)
	if w.Code != http.StatusInternalServerError || !strings.Contains(logs.String(), "level=ERROR") {
		t.Fatalf("client still waiting: status %d, want a logged 500; logs:\n%s", w.Code, logs.String())
	}
}
