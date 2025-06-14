package api

import (
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api/public"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin router with the necessary routes for the metadata service.
func SetupRouter(router *gin.Engine) *gin.Engine {
	// Define a simple health check endpoint
	router.GET("/health", HealthCheck)

	// Register public API routes
	public.RegisterPublicRoutes(router)

	return router
}