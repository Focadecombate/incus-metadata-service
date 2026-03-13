package configs

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"

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

	requested_content_type := c.GetHeader("Accept")

	if !content_types.ValidateContentType(c, requested_content_type, [][]string{content_types.JsonContentTypes, content_types.YamlContentTypes}) {
		return
	}

	if slices.Contains(content_types.JsonContentTypes, requested_content_type) {
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

	requested_content_type := c.GetHeader("Accept")

	if !content_types.ValidateContentType(c, requested_content_type, [][]string{content_types.JsonContentTypes, content_types.YamlContentTypes}) {
		return
	}

	if slices.Contains(content_types.JsonContentTypes, requested_content_type) {
		c.JSON(http.StatusOK, value)
		return
	}

	c.YAML(http.StatusOK, value)
}
