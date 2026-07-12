package configs

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
)

func (h *Handler) UserDataHandler(c *gin.Context) {
	clientIP := c.ClientIP()

	row, err := h.Database.GetInstanceUserDataByIP(c, &clientIP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// GetInstanceUserDataByIP inner-joins instances, so ErrNoRows means
			// either the IP is unknown OR the instance exists but has no
			// user-data row. A non-2xx on user-data makes cloud-init skip the
			// whole datasource, so a KNOWN instance with no user-data must get
			// 200 (empty body). Only a genuinely unknown IP returns 404.
			if _, instErr := h.Database.GetInstanceByIP(c, &clientIP); instErr != nil {
				if errors.Is(instErr, sql.ErrNoRows) {
					c.JSON(http.StatusNotFound, gin.H{"error": "no instance found for this IP"})
					return
				}
				logs.Logger.Error().Err(instErr).Str("ip", clientIP).Msg("failed to look up instance")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user data"})
				return
			}

			c.Data(http.StatusOK, "text/plain", []byte{})
			return
		}

		logs.Logger.Error().Err(err).Str("ip", clientIP).Msg("failed to retrieve user data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user data"})
		return
	}

	// User-data is stored as raw cloud-init YAML or shell script text.
	// Serve it as-is with the appropriate content type.
	var rawData string
	switch v := row.UserData.(type) {
	case string:
		rawData = v
	case []byte:
		rawData = string(v)
	default:
		rawData = fmt.Sprintf("%v", v)
	}

	c.Data(http.StatusOK, "text/yaml", []byte(rawData))
}
