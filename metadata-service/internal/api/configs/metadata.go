package configs

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/content_types"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	db "github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/focadecombate/incus-metadata-service/metadata-service/pkg/types"
	"github.com/gin-gonic/gin"
)

func (h *Handler) AllMetadataHandler(c *gin.Context) {
	clientIP := c.ClientIP()

	row, err := h.Database.GetInstanceMetadataByIP(c, &clientIP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// A 404 here makes cloud-init skip the whole datasource, which is
			// the correct signal for an unknown IP. Caveat: during a sync race
			// a freshly-created instance may briefly 404 before its metadata
			// row is written; that is tolerated as a transient miss.
			c.JSON(http.StatusNotFound, gin.H{"error": "no metadata found for this instance"})
			return
		}
		logs.Logger.Error().Err(err).Str("ip", clientIP).Msg("failed to get instance metadata")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var metadata types.Metadata
	if err := db.ToJSONB(row.Metadata, &metadata); err != nil {
		logs.Logger.Error().Err(err).Msg("failed to unmarshal metadata")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// cloud-init parses meta-data with yaml.safe_load and sends "Accept: */*".
	// The documented contract is a YAML file, so YAML is the default; JSON is
	// only served when explicitly requested. Missing/compound Accept never 406s.
	if content_types.WantsJSON(c.GetHeader("Accept")) {
		c.JSON(http.StatusOK, metadata)
		return
	}

	c.YAML(http.StatusOK, metadata)
}

func (h *Handler) MetadataByKeyHandler(c *gin.Context) {
	key := c.Param("key")
	clientIP := c.ClientIP()

	row, err := h.Database.GetInstanceMetadataByIP(c, &clientIP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no metadata found for this instance"})
			return
		}
		logs.Logger.Error().Err(err).Str("ip", clientIP).Msg("failed to get instance metadata")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var metadata types.Metadata
	if err := db.ToJSONB(row.Metadata, &metadata); err != nil {
		logs.Logger.Error().Err(err).Msg("failed to unmarshal metadata")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Convert metadata to map for key-based lookup
	raw, err := json.Marshal(metadata)
	if err != nil {
		logs.Logger.Error().Err(err).Msg("failed to marshal metadata to JSON")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var metadataMap map[string]any
	if err := json.Unmarshal(raw, &metadataMap); err != nil {
		logs.Logger.Error().Err(err).Msg("failed to unmarshal metadata to map")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	value, ok := metadataMap[key]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}

	if content_types.WantsJSON(c.GetHeader("Accept")) {
		c.JSON(http.StatusOK, value)
		return
	}

	c.YAML(http.StatusOK, value)
}
