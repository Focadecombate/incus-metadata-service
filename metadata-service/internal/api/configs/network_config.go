package configs

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/content_types"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	db "github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/focadecombate/incus-metadata-service/metadata-service/pkg/types"
	"github.com/gin-gonic/gin"
)

func (h *Handler) NetworkConfigHandler(c *gin.Context) {
	clientIP := c.ClientIP()

	row, err := h.Database.GetInstanceNetworkConfigByIP(c, &clientIP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no network config found for this instance"})
			return
		}
		logs.Logger.Error().Err(err).Str("ip", clientIP).Msg("failed to get instance network config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var networkConfig types.NetworkConfig
	if err := db.ToJSONB(row.NetworkConfig, &networkConfig); err != nil {
		logs.Logger.Error().Err(err).Msg("failed to unmarshal network config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	requestedContentType := c.GetHeader("Accept")

	if !content_types.ValidateContentType(c, requestedContentType, [][]string{content_types.YamlContentTypes}) {
		return
	}

	c.YAML(http.StatusOK, networkConfig)
}
