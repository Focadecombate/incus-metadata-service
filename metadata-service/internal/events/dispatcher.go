package events

import (
	"context"
	"fmt"

	"github.com/Raezil/GoEventBus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	incus "github.com/lxc/incus/shared/api"
	"github.com/rs/zerolog"
)

// Event projection constants
const (
	InstanceSyncProjection    = "instance.sync"
	InstancesSyncProjection   = "instances.sync"
	InstanceCreatedProjection = "instance.created"
)

// EventStoreConfig holds configuration for the event store
type EventStoreConfig struct {
	BufferSize   uint64
	DropPolicy   GoEventBus.OverrunPolicy
	AsyncEnabled bool
}

// DefaultEventStoreConfig returns default configuration for event store
func DefaultEventStoreConfig() *EventStoreConfig {
	return &EventStoreConfig{
		BufferSize:   1 << 16, // 65536
		DropPolicy:   GoEventBus.DropOldest,
		AsyncEnabled: true,
	}
}

// EventManager manages event dispatching and event store
type EventManager struct {
	app        *api.App
	dispatcher *GoEventBus.Dispatcher
	eventStore *GoEventBus.EventStore
	config     *EventStoreConfig
	logger     zerolog.Logger
}

// EventHandler represents a typed event handler function
type EventHandler = GoEventBus.HandlerFunc

// HandlerRegistry holds all event handlers
type HandlerRegistry map[string]EventHandler

// NewEventManager creates and initializes a new EventManager
func NewEventManager(app *api.App, config *EventStoreConfig) (*EventManager, error) {
	if config == nil {
		config = DefaultEventStoreConfig()
	}

	logger := logs.Logger.With().Str("component", "event_manager").Logger()
	logger.Info().Msg("Creating event manager")

	manager := &EventManager{
		app:    app,
		config: config,
		logger: logger,
	}

	// Create dispatcher with handlers
	if err := manager.createDispatcher(); err != nil {
		logger.Error().Err(err).Msg("Failed to create dispatcher")
		return nil, fmt.Errorf("failed to create dispatcher: %w", err)
	}

	// Create event store
	if err := manager.createEventStore(); err != nil {
		logger.Error().Err(err).Msg("Failed to create event store")
		return nil, fmt.Errorf("failed to create event store: %w", err)
	}

	logger.Info().Msg("Event manager created successfully")
	return manager, nil
}

// Close closes the event store and cleans up resources
func (em *EventManager) Close(ctx context.Context) error {
	em.logger.Info().Msg("Closing event manager")
	if em.eventStore != nil {
		return em.eventStore.Close(ctx)
	}
	return nil
}

// GetDispatcher returns the dispatcher (for backward compatibility)
func (em *EventManager) GetDispatcher() *GoEventBus.Dispatcher {
	return em.dispatcher
}

// GetEventStore returns the event store (for backward compatibility)
func (em *EventManager) GetEventStore() *GoEventBus.EventStore {
	return em.eventStore
}

// createDispatcher creates the event dispatcher with all handlers
func (em *EventManager) createDispatcher() error {
	em.logger.Debug().Msg("Creating event dispatcher")

	handlers := em.getHandlerRegistry()
	dispatcher := make(GoEventBus.Dispatcher)

	for projection, handler := range handlers {
		dispatcher[projection] = handler
		em.logger.Debug().
			Str("projection", projection).
			Msg("Registered event handler")
	}

	em.dispatcher = &dispatcher
	em.logger.Info().Int("handler_count", len(handlers)).Msg("Event dispatcher created")
	return nil
}

// createEventStore creates and configures the event store
func (em *EventManager) createEventStore() error {
	em.logger.Debug().Msg("Creating event store")

	store := GoEventBus.NewEventStore(
		em.dispatcher,
		em.config.BufferSize,
		em.config.DropPolicy,
	)
	store.Async = em.config.AsyncEnabled

	em.eventStore = store
	em.app.EventStore = store

	em.logger.Info().
		Uint64("buffer_size", em.config.BufferSize).
		Bool("async_enabled", em.config.AsyncEnabled).
		Msg("Event store created")

	return nil
}

// getHandlerRegistry returns all event handlers
func (em *EventManager) getHandlerRegistry() HandlerRegistry {
	return HandlerRegistry{
		InstanceCreatedProjection: em.handleInstanceCreated,
		InstanceSyncProjection:    em.handleInstanceSync,
		InstancesSyncProjection:   em.handleInstancesSync,
	}
}

// handleInstanceCreated handles instance creation events
func (em *EventManager) handleInstanceCreated(ctx context.Context, args map[string]any) (GoEventBus.Result, error) {
	handlerLogger := em.logger.With().Str("handler", "instance_created").Logger()
	
	instanceName, ok := args["instance"].(string)
	if !ok {
		handlerLogger.Error().Msg("Invalid instance argument type")
		return GoEventBus.Result{
			Message: "Invalid instance argument",
		}, fmt.Errorf("expected string for instance argument, got %T", args["instance"])
	}

	handlerLogger.Debug().
		Str("instance", instanceName).
		Msg("Processing instance created event")

	// TODO: Add actual instance creation logic here
	
	return GoEventBus.Result{
		Message: "Instance created successfully",
	}, nil
}

// handleInstanceSync handles single instance synchronization events
func (em *EventManager) handleInstanceSync(ctx context.Context, args map[string]any) (GoEventBus.Result, error) {
	handlerLogger := em.logger.With().Str("handler", "instance_sync").Logger()
	handlerLogger.Info().Msg("Starting instance synchronization")

	instance, ok := args["instance"].(incus.Instance)
	if !ok {
		handlerLogger.Error().Msg("Missing instance argument")
		return GoEventBus.Result{
			Message: "Missing instance argument",
		}, fmt.Errorf("instance argument is required")
	}

	handlerLogger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("instance_name", instance.Name)
	})

	// TODO: Add actual instance sync logic here

	handlerLogger.Info().Msg("Instance sync completed")
	
	return GoEventBus.Result{
		Message: "Instance sync completed",
	}, nil
}

// handleInstancesSync handles bulk instances synchronization events
func (em *EventManager) handleInstancesSync(ctx context.Context, args map[string]any) (GoEventBus.Result, error) {
	handlerLogger := em.logger.With().Str("handler", "instances_sync").Logger()
	handlerLogger.Info().Msg("Starting instances synchronization")

	// Get all instances from Incus
	instances, err := em.app.Incus.GetInstancesAllProjects(incus.InstanceTypeAny)
	if err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to get instances from Incus")
		return GoEventBus.Result{
			Message: "Failed to get instances",
		}, fmt.Errorf("failed to get instances from Incus: %w", err)
	}

	handlerLogger.Info().
		Int("instances_count", len(instances)).
		Msg("Retrieved instances from Incus")

	// Begin transaction to publish individual instance sync events
	tx := em.app.EventStore.BeginTransaction()

	for _, instance := range instances {
		event := GoEventBus.Event{
			ID: instance.Name,
			Args: map[string]any{
				"instance": instance,
			},
			Projection: InstanceSyncProjection,
		}
		
		tx.Publish(event)
		
		handlerLogger.Debug().
			Str("instance_name", instance.Name).
			Str("instance_type", string(instance.Type)).
			Msg("Queued instance sync event")
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to commit event transaction")
		return GoEventBus.Result{
			Message: "Failed to commit transaction",
		}, fmt.Errorf("failed to commit event transaction: %w", err)
	}

	handlerLogger.Info().
		Int("events_published", len(instances)).
		Msg("Successfully published all instance sync events")

	return GoEventBus.Result{
		Message: fmt.Sprintf("Successfully triggered sync for %d instances", len(instances)),
	}, nil
}
