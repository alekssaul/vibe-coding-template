package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/alekssaul/template/internal/model"
	"github.com/alekssaul/template/internal/request"
	"github.com/alekssaul/template/internal/response"
)

// ListItems godoc
//
//	@Summary		List items
//	@Description	Returns paginated list of items
//	@Tags			items
//	@Produce		json
//	@Param			limit	query		int		false	"Max items (default 20, max 100)"
//	@Param			offset	query		int		false	"Pagination offset (default 0)"
//	@Param			search	query		string	false	"Search name/description"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	response.ListResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/v1/items [get]
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20, 100)
	offset := queryInt(r, "offset", 0, -1)
	search := r.URL.Query().Get("search")

	items, total, err := h.store.ListItems(r.Context(), limit, offset, search)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list items", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "failed to list items", "INTERNAL_ERROR")
		return
	}
	if items == nil {
		items = []*model.Item{}
	}
	response.WriteList(w, items, total, limit, offset)
}

// GetItem godoc
//
//	@Summary		Get item
//	@Description	Returns a single item by ID
//	@Tags			items
//	@Produce		json
//	@Param			id	path		int	true	"Item ID"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	response.SuccessResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/v1/items/{id} [get]
func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid item id", "BAD_REQUEST")
		return
	}
	item, err := h.store.GetItem(r.Context(), id)
	if err == sql.ErrNoRows {
		response.WriteError(w, http.StatusNotFound, "item not found", "NOT_FOUND")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "get item", "error", err, "id", id)
		response.WriteError(w, http.StatusInternalServerError, "failed to get item", "INTERNAL_ERROR")
		return
	}
	response.WriteSuccess(w, http.StatusOK, item)
}

// CreateItem godoc
//
//	@Summary		Create item
//	@Description	Creates a new item
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.CreateItemRequest	true	"Item payload"
//	@Security		ApiKeyAuth
//	@Success		201	{object}	response.SuccessResponse
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/v1/items [post]
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req model.CreateItemRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}
	item, err := h.store.CreateItem(r.Context(), &req)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "create item", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "failed to create item", "INTERNAL_ERROR")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, item)
}

// UpdateItem godoc
//
//	@Summary		Update item
//	@Description	Updates an existing item
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Item ID"
//	@Param			body	body		model.UpdateItemRequest	true	"Item payload"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	response.SuccessResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/v1/items/{id} [put]
func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid item id", "BAD_REQUEST")
		return
	}
	var req model.UpdateItemRequest
	if err := request.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}
	item, err := h.store.UpdateItem(r.Context(), id, &req)
	if err == sql.ErrNoRows {
		response.WriteError(w, http.StatusNotFound, "item not found", "NOT_FOUND")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "update item", "error", err, "id", id)
		response.WriteError(w, http.StatusInternalServerError, "failed to update item", "INTERNAL_ERROR")
		return
	}
	response.WriteSuccess(w, http.StatusOK, item)
}

// DeleteItem godoc
//
//	@Summary		Delete item
//	@Description	Deletes an item by ID
//	@Tags			items
//	@Param			id	path	int	true	"Item ID"
//	@Security		ApiKeyAuth
//	@Success		204	"No Content"
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/v1/items/{id} [delete]
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid item id", "BAD_REQUEST")
		return
	}
	err = h.store.DeleteItem(r.Context(), id)
	if err == sql.ErrNoRows {
		response.WriteError(w, http.StatusNotFound, "item not found", "NOT_FOUND")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "delete item", "error", err, "id", id)
		response.WriteError(w, http.StatusInternalServerError, "failed to delete item", "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// queryInt reads an integer query param with a default and optional max.
func queryInt(r *http.Request, key string, defaultVal, maxVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	if maxVal > 0 && v > maxVal {
		return maxVal
	}
	return v
}
