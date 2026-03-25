package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard error payload.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// SuccessResponse wraps a single data payload.
type SuccessResponse struct {
	Data any `json:"data"`
}

// ListResponse wraps a paginated list payload.
type ListResponse struct {
	Data   any `json:"data"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// WriteJSON encodes v as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a structured error response.
func WriteError(w http.ResponseWriter, status int, message, code string) {
	WriteJSON(w, status, ErrorResponse{Error: message, Code: code})
}

// WriteSuccess writes a single-item success response.
func WriteSuccess(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, SuccessResponse{Data: data})
}

// WriteList writes a paginated list response.
func WriteList(w http.ResponseWriter, data any, total, limit, offset int) {
	WriteJSON(w, http.StatusOK, ListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
