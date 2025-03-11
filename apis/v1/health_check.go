package v1

import (
	"fcm/common/response"

	"github.com/gin-gonic/gin"
)

type HealthCheckHandler struct {
}

// NewHealthCheck sets up the route for the health check endpoint.
// It initializes a HealthCheckHandler and registers the endpoint
// at /health-check.
func NewHealthCheck(engine *gin.Engine) {
	handler := &HealthCheckHandler{}

	// Register the health check endpoint
	engine.GET("health-check", handler.HealthCheck)
}

// HealthCheck is a health check handler
func (handler *HealthCheckHandler) HealthCheck(c *gin.Context) {
	// Return OK response
	c.JSON(response.OK(map[string]string{
		"ping": "pong",
	}))
}
