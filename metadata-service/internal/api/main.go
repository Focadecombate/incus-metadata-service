package api

import (
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api/configs"
	internal_routes "github.com/focadecombate/incus-metadata-service/metadata-service/internal/api/internal"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin router with the necessary routes for the metadata service.
func SetupRouter(router *gin.Engine, cfg *config.Config) *gin.Engine {
	// Define a simple health check endpoint
	router.GET("/health", HealthCheck)

	// Register config API routes
	configs.RegisterConfigRoutes(router, cfg)

	// Register internal API routes
	internal_routes.RegisterInternalRoutes(router, cfg)

	return router
}