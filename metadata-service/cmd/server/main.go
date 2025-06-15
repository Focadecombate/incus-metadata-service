package main

import (
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/gin-gonic/gin"
)

// StartServer initializes and starts the metadata service server.
func startServer() {
	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize the logger with Info level
	logs.InitLogger(cfg.LogLevel)
	logs.Logger.Info().Msg("Starting metadata service server...")
	// Create a new Gin router
	router := gin.Default()

	// Register public API routes
	api.SetupRouter(router, cfg)

	logs.Logger.Info().Msg("Metadata service server started on port " + cfg.Port)

	// Start the server on the configured port
	if err := router.Run(":" + cfg.Port); err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to start server")
		panic("Failed to start server: " + err.Error())
	}
}

// main function to run the server
func main() {
	// Start the metadata service server
	startServer()
}