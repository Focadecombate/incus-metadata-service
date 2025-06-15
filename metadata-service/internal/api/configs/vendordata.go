package configs

import (
	"database/sql"
	"net/http"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
)

func (h *Handler) VendorDataHandler(c *gin.Context) {
	// This is a placeholder for the vendor data handler logic.
	// In a real application, you would retrieve and return vendor data here.

	vendorData, err := h.Database.GetVendorData(c, "default")

	if err == sql.ErrNoRows {
		logs.Logger.Info().Msg("No vendor data found, returning empty response")
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve vendor data"})
		return
	}

	var data map[string]any

	err = db.ToJSONB(vendorData.Data, &data)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse vendor data"})
		return
	}

	c.JSON(http.StatusOK, data)
}
