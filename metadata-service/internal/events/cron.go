package events

import (
	"context"
	"fmt"
	"time"

	"github.com/Raezil/GoEventBus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog"
)

// CronConfig holds configuration for cron jobs
type CronConfig struct {
	InstancesSyncInterval time.Duration
}

// DefaultCronConfig returns default configuration for cron jobs
func DefaultCronConfig() *CronConfig {
	return &CronConfig{
		InstancesSyncInterval: 10 * time.Second,
	}
}

// CronManager manages cron scheduler and jobs
type CronManager struct {
	scheduler gocron.Scheduler
	app       *api.App
	config    *CronConfig
	logger    zerolog.Logger
}

// JobDefinition represents a cron job configuration
type JobDefinition struct {
	Name        string
	Interval    time.Duration
	Task        func()
	Description string
}

// NewCronManager creates and initializes a new CronManager
func NewCronManager(app *api.App, config *CronConfig) (*CronManager, error) {
	if config == nil {
		config = DefaultCronConfig()
	}

	logger := logs.Logger.With().Str("component", "cron_manager").Logger()
	logger.Info().Msg("Creating cron manager")

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create scheduler")
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	manager := &CronManager{
		scheduler: scheduler,
		app:       app,
		config:    config,
		logger:    logger,
	}

	// Set the scheduler in the app
	app.CronScheduler = &scheduler

	logger.Info().Msg("Cron manager created successfully")
	return manager, nil
}

// Start starts the cron scheduler
func (cm *CronManager) Start() error {
	cm.logger.Info().Msg("Starting cron scheduler")
	cm.scheduler.Start()
	return nil
}

// Stop stops the cron scheduler
func (cm *CronManager) Stop() error {
	cm.logger.Info().Msg("Stopping cron scheduler")
	return cm.scheduler.Shutdown()
}

// AddJobs registers all predefined cron jobs
func (cm *CronManager) AddJobs() error {
	cm.logger.Info().Msg("Adding cron jobs")

	jobs := cm.getJobDefinitions()
	
	for _, job := range jobs {
		if err := cm.addJob(job); err != nil {
			cm.logger.Error().
				Err(err).
				Str("job_name", job.Name).
				Msg("Failed to add cron job")
			return fmt.Errorf("failed to add job %s: %w", job.Name, err)
		}
		
		cm.logger.Info().
			Str("job_name", job.Name).
			Dur("interval", job.Interval).
			Str("description", job.Description).
			Msg("Cron job added successfully")
	}

	cm.logger.Info().Int("job_count", len(jobs)).Msg("All cron jobs added successfully")
	return nil
}

// getJobDefinitions returns all job definitions
func (cm *CronManager) getJobDefinitions() []JobDefinition {
	return []JobDefinition{
		{
			Name:        "instances_sync",
			Interval:    cm.config.InstancesSyncInterval,
			Task:        cm.createInstancesSyncTask(),
			Description: "Triggers synchronization of all instances",
		},
	}
}

// addJob adds a single job to the scheduler
func (cm *CronManager) addJob(job JobDefinition) error {
	_, err := cm.scheduler.NewJob(
		gocron.DurationJob(job.Interval),
		gocron.NewTask(job.Task),
	)
	return err
}

// createInstancesSyncTask creates the instances sync task function
func (cm *CronManager) createInstancesSyncTask() func() {
	return func() {
		taskLogger := cm.logger.With().Str("task", "instances_sync").Logger()
		taskLogger.Info().Msg("Triggering instances sync event")

		// Create and publish the sync event
		event := GoEventBus.Event{
			Projection: InstancesSyncProjection,
			ID:         time.Now().Format(time.RFC3339),
		}

		cm.app.EventStore.Subscribe(context.Background(), event)
		cm.app.EventStore.Publish()

		taskLogger.Debug().Str("event_id", event.ID).Msg("Instances sync event published")
	}
}
