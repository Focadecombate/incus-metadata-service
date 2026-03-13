package configs

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/content_types"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	db "github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/focadecombate/incus-metadata-service/metadata-service/pkg/types"
)

func (h *Handler) UserDataHandler(c *gin.Context) {
	requested_content_type := c.GetHeader("Accept")

	if !content_types.ValidateContentType(c, requested_content_type, [][]string{content_types.ScriptContentTypes, content_types.YamlContentTypes}) {
		return
	}

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

	var userData types.UserData
	if err := db.ToJSONB(row.UserData, &userData); err != nil {
		logs.Logger.Error().Err(err).Str("ip", clientIP).Msg("failed to unmarshal user data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user data"})
		return
	}

	if content_types.IsYamlContentType(requested_content_type) {
		c.YAML(http.StatusOK, userData)
		return
	}

	// Need to implement the conversion to script format if requested.

	c.YAML(http.StatusOK, userData)
}
