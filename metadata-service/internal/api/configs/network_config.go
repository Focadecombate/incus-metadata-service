package configs

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	db "github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
)

func (h *Handler) NetworkConfigHandler(c *gin.Context) {
	clientIP := c.ClientIP()

	row, err := h.Database.GetInstanceNetworkConfigByIP(c, &clientIP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// A 404 on network-config is tolerated by cloud-init (treated as
			// None), so it is the correct response when no config exists for
			// this instance or the IP is unknown.
			c.JSON(http.StatusNotFound, gin.H{"error": "no network config found for this instance"})
			return
		}
		logs.Logger.Error().Err(err).Str("ip", clientIP).Msg("failed to get instance network config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// cloud-init parses network-config with yaml.safe_load and ignores the
	// response Content-Type, sending "Accept: */*". Always serve the stored
	// config as YAML with 200 (no Accept gate). Serve it as an OPAQUE
	// passthrough: unmarshal the raw JSONB into a generic value and render it
	// as YAML rather than projecting through a fixed typed struct.
	var networkConfig any
	if err := db.ToJSONB(row.NetworkConfig, &networkConfig); err != nil {
		logs.Logger.Error().Err(err).Msg("failed to unmarshal network config")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.YAML(http.StatusOK, networkConfig)
}
