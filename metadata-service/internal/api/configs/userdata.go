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
			c.JSON(http.StatusNotFound, gin.H{"error": "no user data found for this instance"})
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
