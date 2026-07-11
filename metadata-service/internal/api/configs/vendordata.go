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
	if errors.Is(err, sql.ErrNoRows) {
		// No per-instance and no global vendor-data. cloud-init tolerates a
		// vendor-data 404 cleanly, whereas a bare "{}" body triggers warnings.
		logs.Logger.Info().Msg("No vendor data found, returning 404")
		c.JSON(http.StatusNotFound, gin.H{"error": "no vendor data found"})
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
