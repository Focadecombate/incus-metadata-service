package main

import (
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// StartServer initializes and starts the metadata service server.
func startServer() {
	logs.InitLogger(zerolog.InfoLevel) // Initialize the logger with Info level
	logs.Logger.Info().Msg("Starting metadata service server...")
	// Create a new Gin router
	router := gin.Default()

	// Register public API routes
	api.SetupRouter(router)

	logs.Logger.Info().Msg("Metadata service server started on port 8080")

	// Start the server on port 8080
	if err := router.Run(":8080"); err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to start server")
		panic("Failed to start server: " + err.Error())
	}

}

// main function to run the server
func main() {
	// Start the metadata service server
	startServer()
}