package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alekssaul/template/internal/model"
	"github.com/alekssaul/template/internal/response"
)

// ListAPIKeys godoc
//
//	@Summary		List API keys
//	@Description	Returns all API keys (plaintext keys are never returned after creation)
//	@Tags			keys
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	response.ListResponse
//	@Router			/v1/keys [get]
func (h *Handler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := h.store.ListAPIKeys(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list api keys", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "failed to list API keys", "INTERNAL_ERROR")
		return
	}
	if keys == nil {
		keys = []*model.APIKey{}
	}
	response.WriteList(w, keys, len(keys), len(keys), 0)
}

// CreateAPIKey godoc
//
//	@Summary		Create API key
//	@Description	Creates a new API key. The plaintext key is returned ONCE and never stored.
//	@Tags			keys
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.CreateAPIKeyRequest	true	"API key payload"
//	@Security		ApiKeyAuth
//	@Success		201	{object}	response.SuccessResponse
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/v1/keys [post]
func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "name is required", "VALIDATION_ERROR")
		return
	}
	if req.Permission != model.PermissionRead && req.Permission != model.PermissionWrite {
		response.WriteError(w, http.StatusBadRequest, "permission must be 'read' or 'write'", "VALIDATION_ERROR")
		return
	}
	resp, err := h.store.CreateAPIKey(r.Context(), &req)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "create api key", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "failed to create API key", "INTERNAL_ERROR")
		return
	}
	h.logger.InfoContext(r.Context(), "api key created", "name", req.Name, "permission", req.Permission)
	response.WriteSuccess(w, http.StatusCreated, resp)
}

// DeleteAPIKey godoc
//
//	@Summary		Delete API key
//	@Description	Deletes an API key by ID
//	@Tags			keys
//	@Param			id	path	int	true	"API Key ID"
//	@Security		ApiKeyAuth
//	@Success		204	"No Content"
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/v1/keys/{id} [delete]
func (h *Handler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid key id", "BAD_REQUEST")
		return
	}
	if err := h.store.DeleteAPIKey(r.Context(), id); err == sql.ErrNoRows {
		response.WriteError(w, http.StatusNotFound, "API key not found", "NOT_FOUND")
		return
	} else if err != nil {
		h.logger.ErrorContext(r.Context(), "delete api key", "error", err, "id", id)
		response.WriteError(w, http.StatusInternalServerError, "failed to delete API key", "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
