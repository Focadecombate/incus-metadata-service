package public

import (
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers the public API routes for the metadata service.
func RegisterPublicRoutes(router *gin.Engine) {
	publicGroup := router.Group("/public")
	// Metadata endpoints
	publicGroup.GET("/meta-data", AllMetadataHandler)
	publicGroup.GET("/meta-data/:key", MetadataByKeyHandler)

	// User data endpoint
	publicGroup.GET("/user-data", UserDataHandler)

	// Vendor data endpoint
	publicGroup.GET("/vendor-data", VendorDataHandler)

	// Network configuration endpoint
	publicGroup.GET("/network-config", NetworkConfigHandler)
}