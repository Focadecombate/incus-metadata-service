package api

import (
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api/configs"
	internal_routes "github.com/focadecombate/incus-metadata-service/metadata-service/internal/api/internal"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
)

type App struct {
	Config   *config.Config
	Router   *gin.Engine
	Database *db.Queries
}

// SetupRouter initializes the Gin router with the necessary routes for the metadata service.
func SetupRouter(app *App) *gin.Engine {
	// Define a simple health check endpoint
	app.Router.GET("/health", HealthCheck)

	// Register config API routes
	configs.RegisterConfigRoutes(app.Router, app.Config, app.Database)

	// Register internal API routes
	internal_routes.RegisterInternalRoutes(app.Router, app.Config, app.Database)

	return app.Router
}
