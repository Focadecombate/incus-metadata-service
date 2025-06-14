package api

import (
	"net/http"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	// Respond with a simple JSON message indicating the service is healthy
	logs.Logger.Info().Msg("Health check endpoint hit")
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Metadata service is running",
	})
}
