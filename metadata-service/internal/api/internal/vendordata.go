package internal_routes

import (
	"database/sql"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
)

type AddVendorDataKeyRequest struct {
	Data map[string]any `json:"data" binding:"required"`
}

func (h Handler) UpdateVendorData(c *gin.Context) {
	vendorName := c.Param("vendor_name")
	if vendorName == "" {
		c.JSON(400, gin.H{"error": "Vendor name is required"})
		return
	}

	logs.Logger.Info().Str("vendor_name", vendorName).Msg("Updating vendor data")

	var req AddVendorDataKeyRequest
	if err := c.BindJSON(&req); err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to bind request payload")
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	vendorData, err := h.Database.GetVendorData(c, vendorName)
	if err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to retrieve vendor data")
		c.JSON(500, gin.H{"error": "Failed to retrieve vendor data"})
		return
	}

	data, err := db.ToBytes(req.Data)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid vendor data format"})
		return
	}

	update := db.UpdateVendorDataParams{
		ID:   vendorData.ID,
		Data: data,
	}

	// Update the vendor data in the database
	if _, err := h.Database.UpdateVendorData(c, update); err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to update vendor data")
		c.JSON(500, gin.H{"error": "Failed to update vendor data"})
		return
	}

	c.JSON(200, gin.H{"message": "Vendor data updated successfully"})
}

type CreateVendorDataRequest struct {
	VendorName  string         `json:"vendor_name" binding:"required" minLength:"1"`
	Description *string        `json:"description,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

func (h Handler) CreateVendorData(c *gin.Context) {
	var req CreateVendorDataRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	_, err := h.Database.GetVendorData(c, req.VendorName)
	if err != nil && err != sql.ErrNoRows {
		logs.Logger.Error().Err(err).Msg("Failed to check existing vendor data")
		c.JSON(500, gin.H{"error": "Failed to check existing vendor data"})
		return
	}

	if err == nil {
		logs.Logger.Warn().Str("vendor_name", req.VendorName).Msg("Vendor data already exists")
		c.JSON(400, gin.H{"error": "Vendor data already exists"})
		return
	}

	data, err := db.ToBytes(req.Data)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid vendor data format"})
		return
	}

	// Create new vendor data with an empty map
	newData := db.CreateVendorDataParams{
		Name:        req.VendorName,
		Data:        data,
		Description: req.Description,
	}

	// Insert the new vendor data into the database
	if _, err := h.Database.CreateVendorData(c, newData); err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to create vendor data")
		c.JSON(500, gin.H{"error": "Failed to create vendor data"})
		return
	}

	c.JSON(201, gin.H{"message": "Vendor data created successfully"})
}

func (h Handler) GetVendorData(c *gin.Context) {
	vendorName := c.Param("vendor_name")
	logs.Logger.Info().Str("vendor_name", vendorName).Msg("Retrieving vendor data")
	if vendorName == "" {
		c.JSON(400, gin.H{"error": "Vendor name is required"})
		return
	}

	vendorData, err := h.Database.GetVendorData(c, vendorName)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Vendor data not found"})
			return
		}
		logs.Logger.Error().Err(err).Msg("Failed to retrieve vendor data")
		c.JSON(500, gin.H{"error": "Failed to retrieve vendor data"})
		return
	}

	var data map[string]any
	if err := db.ToJSONB(vendorData.Data, &data); err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to parse vendor data")
		c.JSON(500, gin.H{"error": "Failed to parse vendor data"})
		return
	}

	c.JSON(200, gin.H{"data": data})
}
