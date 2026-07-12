package api

import (
	"github.com/Raezil/GoEventBus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api/configs"
	internal_routes "github.com/focadecombate/incus-metadata-service/metadata-service/internal/api/internal"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/consensus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	incus "github.com/lxc/incus/client"
)

type App struct {
	Config        *config.Config
	Router        *gin.Engine
	Database      *db.Queries
	Incus         incus.InstanceServer
	EventStore    *GoEventBus.EventStore
	CronScheduler *gocron.Scheduler
	RaftNode      *consensus.RaftNode
}

// SetupRouter initializes the Gin router with the necessary routes for the metadata service.
func SetupRouter(app *App) *gin.Engine {
	// Define a simple health check endpoint
	app.Router.GET("/health", HealthCheck)

	// RAFT status endpoint (when consensus is enabled)
	if app.RaftNode != nil {
		app.Router.GET("/raft/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"is_leader":   app.RaftNode.IsLeader(),
				"leader_addr": app.RaftNode.LeaderAddr(),
				"state":       app.RaftNode.Raft.State().String(),
			})
		})
	}

	// Register config API routes (public surface only). Internal mutation routes
	// are mounted on a separate host-only listener via SetupInternalRouter.
	configs.RegisterConfigRoutes(app.Router, app.Config, app.Database)

	return app.Router
}

// SetupInternalRouter builds a standalone Gin engine that serves only the
// internal mutation routes. It is intended to run on a host-only listener so
// guests cannot reach it. Trusted proxies are disabled so ClientIP relies on
// the direct socket peer address.
func SetupInternalRouter(app *App) (*gin.Engine, error) {
	router := gin.New()
	router.Use(gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	internal_routes.RegisterInternalRoutes(router, app.Config, app.Database)

	return router, nil
}
