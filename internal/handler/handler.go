package handler

import (
	"log/slog"
	"net/http"
	"runtime"
	"time"

	_ "github.com/alekssaul/template/docs"
	"github.com/alekssaul/template/internal/response"
	"github.com/alekssaul/template/internal/store"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// Handler holds dependencies shared across all HTTP handlers.
type Handler struct {
	store     *store.Store
	logger    *slog.Logger
	gitSHA    string
	buildTime string
	startTime time.Time
}

// New creates a new Handler.
func New(s *store.Store, logger *slog.Logger, gitSHA, buildTime string) *Handler {
	return &Handler{
		store:     s,
		logger:    logger,
		gitSHA:    gitSHA,
		buildTime: buildTime,
		startTime: time.Now(),
	}
}

type healthResponse struct {
	Status    string `json:"status"`
	GitSHA    string `json:"git_sha"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	UptimeSec int64  `json:"uptime_sec"`
}

// Health godoc
//
//	@Summary		Health check
//	@Description	Returns service health, version, and uptime. No authentication required.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	healthResponse
//	@Router			/health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		GitSHA:    h.gitSHA,
		BuildTime: h.buildTime,
		GoVersion: runtime.Version(),
		UptimeSec: int64(time.Since(h.startTime).Seconds()),
	})
}

// Docs serves the Swagger UI.
func (h *Handler) Docs() http.Handler {
	return httpSwagger.WrapHandler
}
