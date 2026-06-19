// Package handlers implements the HTTP layer for auth and task endpoints.
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/hireft/task-manager/internal/auth"
	"github.com/hireft/task-manager/internal/store"
)

// Handler holds the dependencies shared by all HTTP handlers.
type Handler struct {
	Store      store.Store
	Auth       *auth.Manager
	BcryptCost int
	// Now is injected so tests can control token timestamps.
	Now func() time.Time
}

// New builds a Handler with a real wall-clock.
func New(s store.Store, a *auth.Manager, bcryptCost int) *Handler {
	return &Handler{Store: s, Auth: a, BcryptCost: bcryptCost, Now: time.Now}
}

// ErrorBody is the consistent error envelope returned by every endpoint.
type ErrorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "error", err)
	}
}

// writeError emits a consistent error envelope.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorBody{Error: errorDetail{Message: message}})
}

// writeValidationError emits field-level validation errors with a 422 status.
func writeValidationError(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, ErrorBody{
		Error: errorDetail{Message: "validation failed", Fields: fields},
	})
}

// decodeJSON strictly decodes the request body, rejecting unknown fields and
// trailing data.
func decodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
