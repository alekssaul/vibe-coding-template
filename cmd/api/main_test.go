package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"log/slog"

	"github.com/alekssaul/template/internal/handler"
	"github.com/alekssaul/template/internal/store"
)

func newTestHandler(t *testing.T) *handler.Handler {
	t.Helper()
	tmpDB := t.TempDir() + "/test.db"
	s, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return handler.New(s, logger, "test-sha", "2026-01-01T00:00:00Z")
}

func TestHealthEndpoint(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestStoreBootsClean(t *testing.T) {
	tmpDB := t.TempDir() + "/boot.db"
	s, err := store.New(tmpDB)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	count, err := s.CountAPIKeys(t.Context())
	if err != nil {
		t.Fatalf("CountAPIKeys: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 keys on fresh DB, got %d", count)
	}
}
