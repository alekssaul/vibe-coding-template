package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alekssaul/template/internal/handler"
	"github.com/alekssaul/template/internal/middleware"
	"github.com/alekssaul/template/internal/model"
	"github.com/alekssaul/template/internal/store"
)

// testServer wires a full handler+middleware stack against an in-memory SQLite DB.
type testServer struct {
	srv    *httptest.Server
	store  *store.Store
	apiKey string // plaintext write key
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()

	// Use a temp file so each test gets an isolated, clean database.
	f, err := os.CreateTemp(t.TempDir(), "test-*.db")
	if err != nil {
		t.Fatalf("create temp db: %v", err)
	}
	f.Close()

	s, err := store.New(f.Name())
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	// Seed a write key for use in all test requests.
	resp, err := s.CreateAPIKey(context.Background(), &model.CreateAPIKeyRequest{
		Name:       "test-key",
		Permission: model.PermissionWrite,
	})
	if err != nil {
		t.Fatalf("seed api key: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.New(s, logger, "test", "test")
	apiKeyMW := middleware.NewAPIKey(s, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("GET /v1/items", apiKeyMW.RequireRead(http.HandlerFunc(h.ListItems)))
	mux.Handle("GET /v1/items/{id}", apiKeyMW.RequireRead(http.HandlerFunc(h.GetItem)))
	mux.Handle("POST /v1/items", apiKeyMW.RequireWrite(http.HandlerFunc(h.CreateItem)))
	mux.Handle("PUT /v1/items/{id}", apiKeyMW.RequireWrite(http.HandlerFunc(h.UpdateItem)))
	mux.Handle("DELETE /v1/items/{id}", apiKeyMW.RequireWrite(http.HandlerFunc(h.DeleteItem)))
	mux.Handle("GET /v1/keys", apiKeyMW.RequireWrite(http.HandlerFunc(h.ListAPIKeys)))
	mux.Handle("POST /v1/keys", apiKeyMW.RequireWrite(http.HandlerFunc(h.CreateAPIKey)))
	mux.Handle("DELETE /v1/keys/{id}", apiKeyMW.RequireWrite(http.HandlerFunc(h.DeleteAPIKey)))

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	return &testServer{
		srv:    ts,
		store:  s,
		apiKey: resp.Key,
	}
}

// do sends an authenticated HTTP request and returns the decoded JSON body.
func (ts *testServer) do(t *testing.T, method, path string, body any) (int, map[string]any) {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, ts.srv.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("X-API-Key", ts.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer res.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil && res.StatusCode != http.StatusNoContent {
		t.Fatalf("decode response: %v", err)
	}
	return res.StatusCode, result
}

// ── Health ─────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	ts := newTestServer(t)
	code, body := ts.do(t, http.MethodGet, "/health", nil)
	if code != http.StatusOK {
		t.Fatalf("want 200, got %d — %v", code, body)
	}
	if body["status"] != "ok" {
		t.Errorf("want status=ok, got %v", body["status"])
	}
}

// ── Items CRUD ─────────────────────────────────────────────────────────────

func TestItems_CRUD(t *testing.T) {
	ts := newTestServer(t)

	t.Run("list empty", func(t *testing.T) {
		code, body := ts.do(t, http.MethodGet, "/v1/items", nil)
		if code != http.StatusOK {
			t.Fatalf("want 200, got %d — %v", code, body)
		}
		data := body["data"].([]any)
		if len(data) != 0 {
			t.Errorf("want empty list, got %v", data)
		}
	})

	t.Run("create valid", func(t *testing.T) {
		code, body := ts.do(t, http.MethodPost, "/v1/items", map[string]any{
			"name":        "Test Item",
			"description": "A test",
		})
		if code != http.StatusCreated {
			t.Fatalf("want 201, got %d — %v", code, body)
		}
		data := body["data"].(map[string]any)
		if data["name"] != "Test Item" {
			t.Errorf("want name=Test Item, got %v", data["name"])
		}
	})

	t.Run("create invalid — name too short", func(t *testing.T) {
		code, body := ts.do(t, http.MethodPost, "/v1/items", map[string]any{
			"name": "X",
		})
		if code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d — %v", code, body)
		}
	})

	t.Run("create invalid — name missing", func(t *testing.T) {
		code, body := ts.do(t, http.MethodPost, "/v1/items", map[string]any{
			"description": "No name",
		})
		if code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d — %v", code, body)
		}
	})

	// Create an item and capture its ID for update/get/delete
	code, body := ts.do(t, http.MethodPost, "/v1/items", map[string]any{
		"name":        "Persist Me",
		"description": "for further tests",
	})
	if code != http.StatusCreated {
		t.Fatalf("seed item: want 201, got %d — %v", code, body)
	}
	itemID := int64(body["data"].(map[string]any)["id"].(float64))

	t.Run("get by id", func(t *testing.T) {
		code, body := ts.do(t, http.MethodGet, fmt.Sprintf("/v1/items/%d", itemID), nil)
		if code != http.StatusOK {
			t.Fatalf("want 200, got %d — %v", code, body)
		}
		data := body["data"].(map[string]any)
		if data["id"].(float64) != float64(itemID) {
			t.Errorf("want id=%d, got %v", itemID, data["id"])
		}
	})

	t.Run("get not found", func(t *testing.T) {
		code, body := ts.do(t, http.MethodGet, "/v1/items/99999", nil)
		if code != http.StatusNotFound {
			t.Fatalf("want 404, got %d — %v", code, body)
		}
	})

	t.Run("update", func(t *testing.T) {
		code, body := ts.do(t, http.MethodPut, fmt.Sprintf("/v1/items/%d", itemID), map[string]any{
			"name":        "Updated Item",
			"description": "updated",
		})
		if code != http.StatusOK {
			t.Fatalf("want 200, got %d — %v", code, body)
		}
		data := body["data"].(map[string]any)
		if data["name"] != "Updated Item" {
			t.Errorf("want name=Updated Item, got %v", data["name"])
		}
	})

	t.Run("list has item", func(t *testing.T) {
		code, body := ts.do(t, http.MethodGet, "/v1/items", nil)
		if code != http.StatusOK {
			t.Fatalf("want 200, got %d — %v", code, body)
		}
		data := body["data"].([]any)
		if len(data) < 2 { // "Persist Me" + "Test Item"
			t.Errorf("want at least 2 items, got %d", len(data))
		}
	})

	t.Run("delete", func(t *testing.T) {
		code, _ := ts.do(t, http.MethodDelete, fmt.Sprintf("/v1/items/%d", itemID), nil)
		if code != http.StatusNoContent {
			t.Fatalf("want 204, got %d", code)
		}
	})

	t.Run("get deleted returns 404", func(t *testing.T) {
		code, body := ts.do(t, http.MethodGet, fmt.Sprintf("/v1/items/%d", itemID), nil)
		if code != http.StatusNotFound {
			t.Fatalf("want 404, got %d — %v", code, body)
		}
	})
}

// ── Auth ───────────────────────────────────────────────────────────────────

func TestItems_Unauthorized(t *testing.T) {
	ts := newTestServer(t)

	// Request without API key.
	req, _ := http.NewRequest(http.MethodGet, ts.srv.URL+"/v1/items", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", res.StatusCode)
	}
}

func TestItems_ReadKeyCannotWrite(t *testing.T) {
	ts := newTestServer(t)

	// Create a read-only key.
	resp, err := ts.store.CreateAPIKey(context.Background(), &model.CreateAPIKeyRequest{
		Name:       "read-only",
		Permission: model.PermissionRead,
	})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodPost, ts.srv.URL+"/v1/items",
		bytes.NewBufferString(`{"name":"Should Fail","description":""}`))
	req.Header.Set("X-API-Key", resp.Key)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("want 403, got %d", res.StatusCode)
	}
}

// ── API Keys CRUD ──────────────────────────────────────────────────────────

func TestAPIKeys_CRUD(t *testing.T) {
	ts := newTestServer(t)

	t.Run("list returns seeded key", func(t *testing.T) {
		code, body := ts.do(t, http.MethodGet, "/v1/keys", nil)
		if code != http.StatusOK {
			t.Fatalf("want 200, got %d — %v", code, body)
		}
		data := body["data"].([]any)
		if len(data) == 0 {
			t.Error("want at least one key (the test seed key)")
		}
	})

	t.Run("create valid key", func(t *testing.T) {
		code, body := ts.do(t, http.MethodPost, "/v1/keys", map[string]any{
			"name":       "ci-key",
			"permission": "read",
		})
		if code != http.StatusCreated {
			t.Fatalf("want 201, got %d — %v", code, body)
		}
		data := body["data"].(map[string]any)
		if data["key"] == "" || data["key"] == nil {
			t.Error("want plaintext key in response")
		}
	})

	t.Run("create invalid — bad permission", func(t *testing.T) {
		code, body := ts.do(t, http.MethodPost, "/v1/keys", map[string]any{
			"name":       "bad-perm",
			"permission": "superuser",
		})
		if code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d — %v", code, body)
		}
	})

	t.Run("create invalid — name missing", func(t *testing.T) {
		code, body := ts.do(t, http.MethodPost, "/v1/keys", map[string]any{
			"permission": "read",
		})
		if code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d — %v", code, body)
		}
	})

	// Create then delete
	code, body := ts.do(t, http.MethodPost, "/v1/keys", map[string]any{
		"name":       "disposable",
		"permission": "read",
	})
	if code != http.StatusCreated {
		t.Fatalf("create for delete: want 201, got %d — %v", code, body)
	}
	keyID := int64(body["data"].(map[string]any)["id"].(float64))

	t.Run("delete key", func(t *testing.T) {
		code, _ := ts.do(t, http.MethodDelete, fmt.Sprintf("/v1/keys/%d", keyID), nil)
		if code != http.StatusNoContent {
			t.Fatalf("want 204, got %d", code)
		}
	})
}
