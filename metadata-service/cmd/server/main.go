package main

import (
	"context"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/consensus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/events"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/incus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
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

	// Connect to the database
	db, err := db.ConnectDB(cfg)
	if err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to connect to the database")
	}

	incusClient, err := incus.ConnectToIncus(cfg)

	if err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to connect to Incus")
	}

	app := &api.App{
		Config:   cfg,
		Router:   gin.Default(),
		Database: db,
		Incus:    incusClient,
	}

	// Initialize RAFT consensus if enabled
	if cfg.Raft != nil && cfg.Raft.Enabled {
		raftCfg := &consensus.RaftConfig{
			NodeID:    cfg.Raft.NodeID,
			BindAddr:  cfg.Raft.BindAddr,
			DataDir:   cfg.Raft.DataDir,
			Peers:     cfg.Raft.Peers,
			Bootstrap: cfg.Raft.Bootstrap,
		}

		raftNode, err := consensus.NewRaftNode(raftCfg, db)
		if err != nil {
			logs.Logger.Fatal().Err(err).Msg("Failed to initialize RAFT node")
		}
		defer func() {
			if err := raftNode.Shutdown(); err != nil {
				logs.Logger.Error().Err(err).Msg("Failed to shutdown RAFT node")
			}
		}()

		app.RaftNode = raftNode
		logs.Logger.Info().
			Str("node_id", cfg.Raft.NodeID).
			Str("bind_addr", cfg.Raft.BindAddr).
			Msg("RAFT consensus enabled")
	}

	// Initialize event manager
	eventManager, err := events.NewEventManager(app, nil)
	if err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to create event manager")
	}
	defer func() {
		if err := eventManager.Close(context.Background()); err != nil {
			logs.Logger.Error().Err(err).Msg("Failed to close event manager")
		}
	}()

	// Initialize cron manager
	cronManager, err := events.NewCronManager(app, nil)
	if err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to create cron manager")
	}

	// Start cron scheduler
	if err := cronManager.Start(); err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to start cron scheduler")
	}
	defer func() {
		if err := cronManager.Stop(); err != nil {
			logs.Logger.Error().Err(err).Msg("Failed to stop cron scheduler")
		}
	}()

	// Add cron jobs
	if err := cronManager.AddJobs(); err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to add cron jobs")
	}

	// Register public API routes
	api.SetupRouter(app)

	logs.Logger.Info().Msg("Metadata service server started on port " + cfg.Port)

	// Start the server on the configured port
	if err := app.Router.Run(":" + cfg.Port); err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to start server")
		panic("Failed to start server: " + err.Error())
	}
}

// main function to run the server
func main() {
	// Start the metadata service server
	startServer()
}
