// Package api provides the localhost HTTP API server for orchestratr.
// It exposes health, registry, reload, and app lifecycle endpoints
// bound exclusively to 127.0.0.1.
package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the standard JSON error envelope returned by all
// API endpoints when an error occurs.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// HealthResponse is returned by the GET /health endpoint.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// writeJSON encodes v as JSON and writes it with the given HTTP status code.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a standard error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Error: message, Code: code})
}
