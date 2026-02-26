// Package handler contains the HTTP delivery layer for the application.
package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthResponse is the JSON response body for the health-check endpoint.
type HealthResponse struct {
	Message string `json:"message"`
}

// HelloWorld handles GET / and returns a simple health-check JSON response.
func HelloWorld(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Message: "Hello, World!",
	})
}
