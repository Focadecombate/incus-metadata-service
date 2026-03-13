package configs

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
)

func (h *Handler) VendorDataHandler(c *gin.Context) {
	clientIP := c.ClientIP()

	// Try per-instance vendor-data first (from cloud-init.vendor-data in Incus config)
	row, err := h.Database.GetInstanceVendorDataByIP(c, &clientIP)
	if err == nil {
		// Per-instance vendor-data found, serve as raw text (same as user-data)
		var rawData string
		switch v := row.VendorData.(type) {
		case string:
			rawData = v
		case []byte:
			rawData = string(v)
		default:
			rawData = fmt.Sprintf("%v", v)
		}

		c.Data(http.StatusOK, "text/yaml", []byte(rawData))
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		logs.Logger.Error().Err(err).Str("ip", clientIP).Msg("failed to retrieve per-instance vendor data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve vendor data"})
		return
	}

	// Fall back to global vendor-data
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
