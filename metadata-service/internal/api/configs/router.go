package configs

import (
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/gin-gonic/gin"
)

// RegisterConfigRoutes registers the public API routes for the metadata service.
func RegisterConfigRoutes(router *gin.Engine, cfg *config.Config) {
	publicGroup := router.Group("/configs")

	handlers := &Handler{
		config: cfg,
	}

	// Metadata endpoints
	publicGroup.GET("/meta-data", handlers.AllMetadataHandler)
	publicGroup.GET("/meta-data/:key", handlers.MetadataByKeyHandler)

	// User data endpoint
	publicGroup.GET("/user-data", handlers.UserDataHandler)

	// Vendor data endpoint
	publicGroup.GET("/vendor-data", handlers.VendorDataHandler)

	// Network configuration endpoint
	publicGroup.GET("/network-config", handlers.NetworkConfigHandler)
}