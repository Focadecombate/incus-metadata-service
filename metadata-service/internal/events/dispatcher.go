package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Raezil/GoEventBus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/focadecombate/incus-metadata-service/metadata-service/pkg/types"
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

// extractNetworkAddresses extracts IPv4 and IPv6 addresses, MAC addresses, and interface
// info from the Incus instance state network data.
type ifaceInfo struct {
	Name    string
	Hwaddr  string
	IPv4    string
	IPv6    string
	Netmask string
}

func extractInterfaceInfo(state *incus.InstanceState) []ifaceInfo {
	if state == nil || state.Network == nil {
		return nil
	}

	var ifaces []ifaceInfo
	for name, net := range state.Network {
		if name == "lo" {
			continue
		}

		info := ifaceInfo{
			Name:   name,
			Hwaddr: net.Hwaddr,
		}

		for _, addr := range net.Addresses {
			if addr.Scope == "link" {
				continue
			}
			switch addr.Family {
			case "inet":
				if info.IPv4 == "" {
					info.IPv4 = addr.Address
					info.Netmask = addr.Netmask
				}
			case "inet6":
				if info.IPv6 == "" {
					info.IPv6 = addr.Address
				}
			}
		}

		ifaces = append(ifaces, info)
	}

	return ifaces
}

// handleInstanceSync handles single instance synchronization events
func (em *EventManager) handleInstanceSync(ctx context.Context, args map[string]any) (GoEventBus.Result, error) {
	handlerLogger := em.logger.With().Str("handler", "instance_sync").Logger()
	handlerLogger.Info().Msg("Starting instance synchronization")

	instance, ok := args["instance"].(incus.InstanceFull)
	if !ok {
		handlerLogger.Error().Msg("Missing instance argument")
		return GoEventBus.Result{
			Message: "Missing instance argument",
		}, fmt.Errorf("instance argument is required")
	}

	handlerLogger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("instance_name", instance.Name)
	})

	// Extract network interface information from instance state
	ifaces := extractInterfaceInfo(instance.State)

	// Determine primary IPv4/IPv6 from the first non-loopback interface
	var primaryIPv4, primaryIPv6 string
	if len(ifaces) > 0 {
		primaryIPv4 = ifaces[0].IPv4
		primaryIPv6 = ifaces[0].IPv6
	}

	// Create or update instance in database
	dbInstance, err := em.app.Database.GetInstance(ctx, db.GetInstanceParams{
		Name:    instance.Name,
		Project: instance.Project,
	})

	if err != nil {
		handlerLogger.Info().Err(err).Msg("Instance not found in database, creating new instance")
		dbInstance, err = em.app.Database.CreateInstance(ctx, db.CreateInstanceParams{
			Name:      instance.Name,
			Project:   instance.Project,
			IpAddress: strPtrOrNil(primaryIPv4),
		})
		if err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to create instance in database")
			return GoEventBus.Result{
				Message: "Failed to create instance in database",
			}, fmt.Errorf("failed to create instance in database: %w", err)
		}
	} else {
		// Instance exists, update its IP address
		dbInstance, err = em.app.Database.UpdateInstance(ctx, db.UpdateInstanceParams{
			ID:        dbInstance.ID,
			IpAddress: strPtrOrNil(primaryIPv4),
		})
		if err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to update instance in database")
			return GoEventBus.Result{
				Message: "Failed to update instance in database",
			}, fmt.Errorf("failed to update instance in database: %w", err)
		}
	}

	handlerLogger.Info().Int64("instance_id", dbInstance.ID).Msg("Instance synced in database")

	// Build and store metadata
	macs := make(map[string]types.Mac)
	for _, iface := range ifaces {
		if iface.Hwaddr != "" {
			macs[iface.Hwaddr] = types.Mac{
				DeviceNumber:  iface.Name,
				LocalHostname: instance.Name,
				LocalIPv4:     iface.IPv4,
				LocalIPv6:     iface.IPv6,
				Mac:           iface.Hwaddr,
			}
		}
	}

	metadata := types.Metadata{
		InstanceID:    fmt.Sprintf("%d", dbInstance.ID),
		Hostname:      instance.Name,
		LocalHostname: instance.Name,
		LocalIPv4:     primaryIPv4,
		LocalIPv6:     primaryIPv6,
		Placement: types.Placement{
			Project: instance.Project,
			HostID:  instance.Location,
		},
		Network: types.Network{
			Interfaces: types.Interfaces{
				Macs: macs,
			},
		},
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to marshal metadata")
		return GoEventBus.Result{
			Message: "Failed to marshal metadata",
		}, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = em.app.Database.CreateOrUpdateInstanceMetadata(ctx, db.CreateOrUpdateInstanceMetadataParams{
		InstanceID: dbInstance.ID,
		Metadata:   metadataBytes,
	})
	if err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to store instance metadata")
		return GoEventBus.Result{
			Message: "Failed to store instance metadata",
		}, fmt.Errorf("failed to store instance metadata: %w", err)
	}

	// Build and store user data from cloud-init config
	userData := instance.Config["cloud-init.user-data"]
	if userData == "" {
		userData = instance.Config["user.user-data"]
	}
	if userData != "" {
		userDataBytes, err := json.Marshal(userData)
		if err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to marshal user data")
			return GoEventBus.Result{
				Message: "Failed to marshal user data",
			}, fmt.Errorf("failed to marshal user data: %w", err)
		}

		_, err = em.app.Database.CreateOrUpdateInstanceUserData(ctx, db.CreateOrUpdateInstanceUserDataParams{
			InstanceID: dbInstance.ID,
			UserData:   userDataBytes,
		})
		if err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to store instance user data")
			return GoEventBus.Result{
				Message: "Failed to store instance user data",
			}, fmt.Errorf("failed to store instance user data: %w", err)
		}
	}

	// Build and store network config (netplan v2 format)
	ethernets := make(map[string]types.Ethernet)
	for _, iface := range ifaces {
		var addresses []string
		if iface.IPv4 != "" {
			addr := iface.IPv4
			if iface.Netmask != "" {
				addr = fmt.Sprintf("%s/%s", iface.IPv4, iface.Netmask)
			}
			addresses = append(addresses, addr)
		}
		if iface.IPv6 != "" {
			addresses = append(addresses, iface.IPv6)
		}

		ethernets[iface.Name] = types.Ethernet{
			Match: types.Match{
				MacAddress: iface.Hwaddr,
			},
			Addresses: addresses,
		}
	}

	networkConfig := types.NetworkConfig{
		Version:   2,
		Ethernets: ethernets,
	}

	networkConfigBytes, err := json.Marshal(networkConfig)
	if err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to marshal network config")
		return GoEventBus.Result{
			Message: "Failed to marshal network config",
		}, fmt.Errorf("failed to marshal network config: %w", err)
	}

	_, err = em.app.Database.CreateOrUpdateInstanceNetworkConfig(ctx, db.CreateOrUpdateInstanceNetworkConfigParams{
		InstanceID:    dbInstance.ID,
		NetworkConfig: networkConfigBytes,
	})
	if err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to store instance network config")
		return GoEventBus.Result{
			Message: "Failed to store instance network config",
		}, fmt.Errorf("failed to store instance network config: %w", err)
	}

	// Update instance state
	if instance.State != nil {
		_, err = em.app.Database.CreateOrUpdateInstanceState(ctx, db.CreateOrUpdateInstanceStateParams{
			InstanceID: dbInstance.ID,
			Status:     instance.State.Status,
			StatusCode: int64(instance.State.StatusCode),
		})
		if err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to store instance state")
			return GoEventBus.Result{
				Message: "Failed to store instance state",
			}, fmt.Errorf("failed to store instance state: %w", err)
		}
	}

	handlerLogger.Info().Msg("Instance sync completed")

	return GoEventBus.Result{
		Message: "Instance sync completed",
	}, nil
}

// strPtrOrNil returns a pointer to s if non-empty, otherwise nil.
func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// handleInstancesSync handles bulk instances synchronization events
func (em *EventManager) handleInstancesSync(ctx context.Context, args map[string]any) (GoEventBus.Result, error) {
	handlerLogger := em.logger.With().Str("handler", "instances_sync").Logger()
	handlerLogger.Info().Msg("Starting instances synchronization")

	// Get all instances from Incus
	instances, err := em.app.Incus.GetInstancesFullAllProjects(incus.InstanceTypeAny)
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
