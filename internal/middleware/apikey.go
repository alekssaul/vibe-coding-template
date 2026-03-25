package middleware

import (
	"log/slog"
	"net/http"

	"github.com/alekssaul/template/internal/model"
	"github.com/alekssaul/template/internal/response"
	"github.com/alekssaul/template/internal/store"
)

// APIKeyMiddleware validates X-API-Key headers and enforces permissions.
type APIKeyMiddleware struct {
	store  *store.Store
	logger *slog.Logger
}

// NewAPIKey creates a new APIKeyMiddleware.
func NewAPIKey(s *store.Store, logger *slog.Logger) *APIKeyMiddleware {
	return &APIKeyMiddleware{store: s, logger: logger}
}

// RequireRead allows requests authenticated with either a read or write key.
func (m *APIKeyMiddleware) RequireRead(next http.Handler) http.Handler {
	return m.require(next, model.PermissionRead)
}

// RequireWrite only allows requests authenticated with a write key.
func (m *APIKeyMiddleware) RequireWrite(next http.Handler) http.Handler {
	return m.require(next, model.PermissionWrite)
}

func (m *APIKeyMiddleware) require(next http.Handler, minPerm model.Permission) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			response.WriteError(w, http.StatusUnauthorized, "missing X-API-Key header", "UNAUTHORIZED")
			return
		}

		ak, err := m.store.ValidateAPIKey(r.Context(), key)
		if err != nil {
			response.WriteError(w, http.StatusUnauthorized, "invalid API key", "UNAUTHORIZED")
			return
		}

		// A write key satisfies a read requirement; a read key cannot satisfy write.
		if minPerm == model.PermissionWrite && ak.Permission != model.PermissionWrite {
			response.WriteError(w, http.StatusForbidden, "write permission required", "FORBIDDEN")
			return
		}

		m.logger.InfoContext(r.Context(), "request authenticated",
			"key_name", ak.Name,
			"permission", ak.Permission,
			"request_id", GetRequestID(r.Context()),
		)
		next.ServeHTTP(w, r)
	})
}
