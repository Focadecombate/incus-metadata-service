package configs

import (
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
)

// RegisterConfigRoutes registers the public API routes for the metadata service.
func RegisterConfigRoutes(router *gin.Engine, cfg *config.Config, db db.Querier) {
	publicGroup := router.Group("/configs")

	handlers := &Handler{
		Config:   cfg,
		Database: db,
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
