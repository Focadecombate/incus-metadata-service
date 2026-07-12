package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/consensus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/events"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/incus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/telemetry"
	"github.com/gin-gonic/gin"
)

// startServer initializes and starts the metadata service server.
func startServer() {
	// Load configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize the logger with Info level
	logs.InitLogger(cfg.LogLevel)
	logs.Logger.Info().Msg("Starting metadata service server...")

	// Initialize OpenTelemetry metrics with Prometheus exporter
	if cfg.Otel == nil || cfg.Otel.MetricsEnabled {
		meterProvider, err := telemetry.InitMetrics()
		if err != nil {
			logs.Logger.Fatal().Err(err).Msg("Failed to initialize metrics")
		}
		defer func() {
			if err := meterProvider.Shutdown(context.Background()); err != nil {
				logs.Logger.Error().Err(err).Msg("Failed to shutdown meter provider")
			}
		}()
		logs.Logger.Info().Msg("OpenTelemetry metrics initialized with Prometheus exporter")
	}

	// Connect to the database
	db, err := db.ConnectDB(cfg)
	if err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to connect to the database")
	}

	incusClient, err := incus.ConnectToIncus(cfg)

	if err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to connect to Incus")
	}

	router := gin.Default()
	// Trust only the direct socket peer address; ignore X-Forwarded-For / X-Real-IP
	// so guests cannot spoof their identity via forwarded headers.
	if err := router.SetTrustedProxies(nil); err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to configure trusted proxies")
	}
	router.Use(telemetry.MetricsMiddleware())
	router.GET("/metrics", telemetry.MetricsHandler())

	app := &api.App{
		Config:   cfg,
		Router:   router,
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

	// Register public API routes (guest-facing surface only)
	api.SetupRouter(app)

	// Build the host-only internal listener for mutation routes so guests on the
	// public listener cannot reach them. The internal routes live in an internal/
	// package that only the api package may import, so it is wired there.
	internalRouter, err := api.SetupInternalRouter(app)
	if err != nil {
		logs.Logger.Fatal().Err(err).Msg("Failed to configure internal router")
	}

	// Listen for termination signals so deferred cleanups actually run.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publicSrv := &http.Server{Addr: ":" + cfg.Port, Handler: app.Router}
	internalSrv := &http.Server{Addr: cfg.InternalAddr, Handler: internalRouter}

	go func() {
		logs.Logger.Info().Str("addr", cfg.InternalAddr).Msg("Internal listener started")
		if err := internalSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Logger.Error().Err(err).Msg("Internal server error")
		}
	}()

	go func() {
		logs.Logger.Info().Msg("Metadata service server started on port " + cfg.Port)
		if err := publicSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Logger.Error().Err(err).Msg("Failed to start server")
		}
	}()

	// Block until a termination signal is received.
	<-ctx.Done()
	stop()
	logs.Logger.Info().Msg("Shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := publicSrv.Shutdown(shutdownCtx); err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to gracefully shut down public server")
	}
	if err := internalSrv.Shutdown(shutdownCtx); err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to gracefully shut down internal server")
	}
}

// main function to run the server
func main() {
	// Start the metadata service server
	startServer()
}
