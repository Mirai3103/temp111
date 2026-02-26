// Package handler contains the HTTP delivery layer for the application.
package handler

import (
	"encoding/json"
	"net/http"
)

// HealthResponse is the JSON response body for the health-check endpoint.
type HealthResponse struct {
	Message string `json:"message"`
}

// HelloWorld handles GET / and returns a simple health-check JSON response.
func HelloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{
		Message: "Hello, World!",
	})
}
