package internal_routes

import (
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
)

func RegisterInternalRoutes(router *gin.Engine, cfg *config.Config, db *db.Queries) {
	// Register internal routes here

	handler := Handler{
		Config:   cfg,
		Database: db,
	}

	router.PUT("/internal/vendor/:vendor_name/data", handler.UpdateVendorData)
	router.GET("/internal/vendor/:vendor_name/data", handler.GetVendorData)
	router.POST("/internal/vendor", handler.CreateVendorData)
}
