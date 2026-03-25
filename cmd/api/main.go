// Package main is the entrypoint for the API server.
//
//	@title					Template API
//	@version				1.0
//	@description			CRUD API template with API key authentication and SQLite.
//	@host					localhost:8080
//	@BasePath				/
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in						header
//	@name					X-API-Key
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/alekssaul/template/internal/config"
	"github.com/alekssaul/template/internal/handler"
	"github.com/alekssaul/template/internal/middleware"
	"github.com/alekssaul/template/internal/model"
	"github.com/alekssaul/template/internal/store"
)

// Build-time variables injected via -ldflags.
var (
	gitSHA    = "dev"
	buildTime = "unknown"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Load()

	s, err := store.New(cfg.DBPath)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer s.Close()

	// Auto-create a default write key on first run.
	count, err := s.CountAPIKeys(context.Background())
	if err != nil {
		logger.Error("failed to count api keys", "error", err)
		os.Exit(1)
	}
	if count == 0 {
		resp, err := s.CreateAPIKey(context.Background(), &model.CreateAPIKeyRequest{
			Name:       "default",
			Permission: model.PermissionWrite,
		})
		if err != nil {
			logger.Error("failed to create default api key", "error", err)
			os.Exit(1)
		}
		logger.Info("⚠️  default API key created — save this, it will not be shown again",
			"key", resp.Key,
			"permission", resp.APIKey.Permission,
		)
	}

	h := handler.New(s, logger, gitSHA, buildTime)
	apiKeyMW := middleware.NewAPIKey(s, logger)

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("/docs/", h.Docs())

	// Items — read
	mux.Handle("GET /v1/items", apiKeyMW.RequireRead(http.HandlerFunc(h.ListItems)))
	mux.Handle("GET /v1/items/{id}", apiKeyMW.RequireRead(http.HandlerFunc(h.GetItem)))

	// Items — write
	mux.Handle("POST /v1/items", apiKeyMW.RequireWrite(http.HandlerFunc(h.CreateItem)))
	mux.Handle("PUT /v1/items/{id}", apiKeyMW.RequireWrite(http.HandlerFunc(h.UpdateItem)))
	mux.Handle("DELETE /v1/items/{id}", apiKeyMW.RequireWrite(http.HandlerFunc(h.DeleteItem)))

	// API key management — write only
	mux.Handle("GET /v1/keys", apiKeyMW.RequireWrite(http.HandlerFunc(h.ListAPIKeys)))
	mux.Handle("POST /v1/keys", apiKeyMW.RequireWrite(http.HandlerFunc(h.CreateAPIKey)))
	mux.Handle("DELETE /v1/keys/{id}", apiKeyMW.RequireWrite(http.HandlerFunc(h.DeleteAPIKey)))

	// Global middleware chain: CORS → RequestID → mux
	var root http.Handler = mux
	root = middleware.RequestID(root)
	root = middleware.CORS(cfg.CORSOrigins)(root)

	srv := &http.Server{
		Addr:         ":" + cfg.APIPort,
		Handler:      root,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("🚀 server starting",
		"port", cfg.APIPort,
		"git_sha", gitSHA,
		"build_time", buildTime,
		"go_version", runtime.Version(),
		"env", cfg.Env,
		"db", cfg.DBPath,
	)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}
