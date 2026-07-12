package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Raezil/GoEventBus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/api"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/consensus"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/telemetry"
	"github.com/focadecombate/incus-metadata-service/metadata-service/pkg/types"
	incus "github.com/lxc/incus/shared/api"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"gopkg.in/yaml.v3"
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

	eventDuration metric.Float64Histogram
	eventTotal    metric.Int64Counter
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

	meter := telemetry.GetMeter()
	eventDuration, _ := meter.Float64Histogram(
		"event_processing_duration_seconds",
		metric.WithDescription("Duration of event processing in seconds"),
		metric.WithUnit("s"),
	)
	eventTotal, _ := meter.Int64Counter(
		"event_processing_total",
		metric.WithDescription("Total number of events processed"),
	)

	manager := &EventManager{
		app:           app,
		config:        config,
		logger:        logger,
		eventDuration: eventDuration,
		eventTotal:    eventTotal,
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

// instrumentedHandler wraps an event handler to record processing metrics.
func (em *EventManager) instrumentedHandler(projection string, handler EventHandler) EventHandler {
	return func(ctx context.Context, args map[string]any) (GoEventBus.Result, error) {
		start := time.Now()
		result, err := handler(ctx, args)
		duration := time.Since(start).Seconds()

		status := "success"
		if err != nil {
			status = "error"
		}

		attrs := metric.WithAttributes(
			attribute.String("projection", projection),
			attribute.String("status", status),
		)

		em.eventDuration.Record(ctx, duration, attrs)
		em.eventTotal.Add(ctx, 1, attrs)

		return result, err
	}
}

// createDispatcher creates the event dispatcher with all handlers
func (em *EventManager) createDispatcher() error {
	em.logger.Debug().Msg("Creating event dispatcher")

	handlers := em.getHandlerRegistry()
	dispatcher := make(GoEventBus.Dispatcher)

	for projection, handler := range handlers {
		dispatcher[projection] = em.instrumentedHandler(projection, handler)
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

	// Fetch the full instance from Incus and delegate to the same per-instance
	// sync path, closing the boot-race window where a freshly-created instance
	// requests metadata before the periodic full sync has run.
	full, _, err := em.app.Incus.GetInstanceFull(instanceName)
	if err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to fetch created instance from Incus")
		return GoEventBus.Result{
			Message: "Failed to fetch created instance",
		}, fmt.Errorf("failed to fetch created instance %q: %w", instanceName, err)
	}

	return em.handleInstanceSync(ctx, map[string]any{"instance": *full})
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
		createParams := db.CreateInstanceParams{
			Name:       instance.Name,
			Project:    instance.Project,
			SourceNode: em.nodeID(),
			IpAddress:  strPtrOrNil(primaryIPv4),
		}
		if err := em.applyWrite(consensus.CmdCreateInstance, createParams); err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to create instance in database")
			return GoEventBus.Result{
				Message: "Failed to create instance in database",
			}, fmt.Errorf("failed to create instance in database: %w", err)
		}
		// Re-fetch the instance to get the ID
		dbInstance, err = em.app.Database.GetInstance(ctx, db.GetInstanceParams{
			Name:    instance.Name,
			Project: instance.Project,
		})
		if err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to fetch created instance")
			return GoEventBus.Result{
				Message: "Failed to fetch created instance",
			}, fmt.Errorf("failed to fetch created instance: %w", err)
		}
	} else {
		// Instance exists, update its IP address
		updateParams := db.UpdateInstanceParams{
			ID:        dbInstance.ID,
			IpAddress: strPtrOrNil(primaryIPv4),
		}
		if err := em.applyWrite(consensus.CmdUpdateInstance, updateParams); err != nil {
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
		PublicKeys:    extractSSHKeys(instance.Config),
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

	if err := em.applyWrite(consensus.CmdCreateOrUpdateMetadata, db.CreateOrUpdateInstanceMetadataParams{
		InstanceID: dbInstance.ID,
		Metadata:   metadataBytes,
	}); err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to store instance metadata")
		return GoEventBus.Result{
			Message: "Failed to store instance metadata",
		}, fmt.Errorf("failed to store instance metadata: %w", err)
	}

	// Build and store user data from cloud-init config
	// User-data is stored as a raw string (YAML cloud-config or shell script)
	userData := instance.Config["cloud-init.user-data"]
	if userData == "" {
		userData = instance.Config["user.user-data"]
	}
	if userData != "" {
		if err := em.applyWrite(consensus.CmdCreateOrUpdateUserData, db.CreateOrUpdateInstanceUserDataParams{
			InstanceID: dbInstance.ID,
			UserData:   userData,
		}); err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to store instance user data")
			return GoEventBus.Result{
				Message: "Failed to store instance user data",
			}, fmt.Errorf("failed to store instance user data: %w", err)
		}
	}

	// Build and store vendor data from cloud-init config
	vendorData := instance.Config["cloud-init.vendor-data"]
	if vendorData == "" {
		vendorData = instance.Config["user.vendor-data"]
	}
	if vendorData != "" {
		if err := em.applyWrite(consensus.CmdCreateOrUpdateVendorData, db.CreateOrUpdateInstanceVendorDataParams{
			InstanceID: dbInstance.ID,
			VendorData: vendorData,
		}); err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to store instance vendor data")
			return GoEventBus.Result{
				Message: "Failed to store instance vendor data",
			}, fmt.Errorf("failed to store instance vendor data: %w", err)
		}
	}

	// Build and store network config
	// Prefer user-defined cloud-init.network-config from Incus config over auto-generated
	var networkConfigBytes []byte
	if cfgNetworkConfig := instance.Config["cloud-init.network-config"]; cfgNetworkConfig != "" {
		// Parse user-defined network config YAML into the NetworkConfig struct
		var userNetworkConfig types.NetworkConfig
		if err := yaml.Unmarshal([]byte(cfgNetworkConfig), &userNetworkConfig); err == nil && userNetworkConfig.Version > 0 {
			networkConfigBytes, err = json.Marshal(userNetworkConfig)
			if err != nil {
				handlerLogger.Error().Err(err).Msg("Failed to marshal user-defined network config")
			}
		}
		if networkConfigBytes == nil {
			// Could not parse as v2 struct, store as raw string
			handlerLogger.Warn().Msg("Could not parse cloud-init.network-config as v2 format, using auto-generated")
		}
	}

	if networkConfigBytes == nil {
		// Auto-generate network config from runtime state (netplan v2 format)
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

			eth := types.Ethernet{
				Addresses: addresses,
			}
			if iface.Hwaddr != "" {
				eth.Match = &types.Match{MacAddress: iface.Hwaddr}
			}
			ethernets[iface.Name] = eth
		}

		networkConfig := types.NetworkConfig{
			Version:   2,
			Ethernets: ethernets,
		}

		var err error
		networkConfigBytes, err = json.Marshal(networkConfig)
		if err != nil {
			handlerLogger.Error().Err(err).Msg("Failed to marshal network config")
			return GoEventBus.Result{
				Message: "Failed to marshal network config",
			}, fmt.Errorf("failed to marshal network config: %w", err)
		}
	}

	if err := em.applyWrite(consensus.CmdCreateOrUpdateNetworkConfig, db.CreateOrUpdateInstanceNetworkConfigParams{
		InstanceID:    dbInstance.ID,
		NetworkConfig: networkConfigBytes,
	}); err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to store instance network config")
		return GoEventBus.Result{
			Message: "Failed to store instance network config",
		}, fmt.Errorf("failed to store instance network config: %w", err)
	}

	// Update instance state
	if instance.State != nil {
		if err := em.applyWrite(consensus.CmdCreateOrUpdateState, db.CreateOrUpdateInstanceStateParams{
			InstanceID: dbInstance.ID,
			Status:     instance.State.Status,
			StatusCode: int64(instance.State.StatusCode),
		}); err != nil {
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

// applyWrite routes a write operation through RAFT when consensus is enabled,
// or executes it directly on the local database otherwise.
func (em *EventManager) applyWrite(cmdType consensus.CommandType, params any) error {
	if em.app.RaftNode != nil {
		_, err := em.app.RaftNode.Apply(cmdType, params)
		return err
	}

	// No RAFT — execute directly
	ctx := context.Background()
	switch cmdType {
	case consensus.CmdCreateInstance:
		_, err := em.app.Database.CreateInstance(ctx, params.(db.CreateInstanceParams))
		return err
	case consensus.CmdUpdateInstance:
		_, err := em.app.Database.UpdateInstance(ctx, params.(db.UpdateInstanceParams))
		return err
	case consensus.CmdCreateOrUpdateMetadata:
		_, err := em.app.Database.CreateOrUpdateInstanceMetadata(ctx, params.(db.CreateOrUpdateInstanceMetadataParams))
		return err
	case consensus.CmdCreateOrUpdateUserData:
		_, err := em.app.Database.CreateOrUpdateInstanceUserData(ctx, params.(db.CreateOrUpdateInstanceUserDataParams))
		return err
	case consensus.CmdCreateOrUpdateNetworkConfig:
		_, err := em.app.Database.CreateOrUpdateInstanceNetworkConfig(ctx, params.(db.CreateOrUpdateInstanceNetworkConfigParams))
		return err
	case consensus.CmdCreateOrUpdateState:
		_, err := em.app.Database.CreateOrUpdateInstanceState(ctx, params.(db.CreateOrUpdateInstanceStateParams))
		return err
	case consensus.CmdCreateOrUpdateVendorData:
		_, err := em.app.Database.CreateOrUpdateInstanceVendorData(ctx, params.(db.CreateOrUpdateInstanceVendorDataParams))
		return err
	case consensus.CmdDeleteInstance:
		return em.app.Database.DeleteInstance(ctx, params.(int64))
	default:
		return fmt.Errorf("unknown command type: %d", cmdType)
	}
}

// extractSSHKeys extracts SSH public keys from the Incus instance config.
// It checks user.meta-data (YAML with public-keys) and user.ssh-keys (newline-separated).
func extractSSHKeys(config map[string]string) []string {
	var keys []string

	// Check user.meta-data for public-keys field
	if metaData := config["user.meta-data"]; metaData != "" {
		var parsed struct {
			PublicKeys []string `yaml:"public-keys"`
		}
		if err := yaml.Unmarshal([]byte(metaData), &parsed); err == nil && len(parsed.PublicKeys) > 0 {
			keys = append(keys, parsed.PublicKeys...)
		}
	}

	// Check user.ssh-keys (newline-separated list of SSH public keys)
	if sshKeys := config["user.ssh-keys"]; sshKeys != "" {
		for _, key := range strings.Split(sshKeys, "\n") {
			key = strings.TrimSpace(key)
			if key != "" {
				keys = append(keys, key)
			}
		}
	}

	return keys
}

// strPtrOrNil returns a pointer to s if non-empty, otherwise nil.
func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// handleInstancesSync handles bulk instances synchronization events.
// When RAFT is enabled, only the leader node syncs from Incus.
func (em *EventManager) handleInstancesSync(ctx context.Context, args map[string]any) (GoEventBus.Result, error) {
	handlerLogger := em.logger.With().Str("handler", "instances_sync").Logger()

	// Skip sync if RAFT is enabled and this node is not the leader
	if em.app.RaftNode != nil && !em.app.RaftNode.IsLeader() {
		handlerLogger.Debug().Msg("Not the leader, skipping instances sync")
		return GoEventBus.Result{
			Message: "Skipped: not the leader",
		}, nil
	}

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

	// Reconcile: soft-delete DB instances that no longer exist in Incus so a
	// reused IP is never served a decommissioned instance's secrets.
	em.reconcileDeletedInstances(ctx, instances, handlerLogger)

	return GoEventBus.Result{
		Message: fmt.Sprintf("Successfully triggered sync for %d instances", len(instances)),
	}, nil
}

// nodeID returns the id of this node for stamping instance ownership.
// It is the RAFT node id when consensus is enabled, otherwise "standalone".
func (em *EventManager) nodeID() string {
	if em.app.Config.Raft != nil && em.app.Config.Raft.Enabled {
		return em.app.Config.Raft.NodeID
	}
	return "standalone"
}

// reconcileDeletedInstances soft-deletes DB instances absent from the current
// Incus listing. It runs on the leader only (handleInstancesSync is leader-gated).
//
// NOTE: reconciliation is intentionally GLOBAL (all DB instances), which is
// correct for the current shared-Incus HA model. A future per-node "Option B"
// reconcile would instead scope to instances stamped with this node's id via
// db.ListActiveInstancesBySourceNode(ctx, em.nodeID()).
//
// The soft-delete is routed through applyWrite so that, when RAFT is enabled,
// the deletion is replicated to all followers via the log rather than applied
// only to the leader's local DB (which would diverge the cluster). In non-RAFT
// mode applyWrite executes the delete directly on the local DB.
func (em *EventManager) reconcileDeletedInstances(ctx context.Context, instances []incus.InstanceFull, handlerLogger zerolog.Logger) {
	live := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		live[instance.Project+"/"+instance.Name] = struct{}{}
	}

	dbInstances, err := em.app.Database.ListInstances(ctx)
	if err != nil {
		handlerLogger.Error().Err(err).Msg("Failed to list DB instances for reconciliation")
		return
	}

	for _, dbInstance := range dbInstances {
		if _, ok := live[dbInstance.Project+"/"+dbInstance.Name]; ok {
			continue
		}
		if err := em.applyWrite(consensus.CmdDeleteInstance, dbInstance.ID); err != nil {
			handlerLogger.Error().Err(err).
				Str("instance_name", dbInstance.Name).
				Str("instance_project", dbInstance.Project).
				Msg("Failed to soft-delete stale instance")
			continue
		}
		handlerLogger.Info().
			Str("instance_name", dbInstance.Name).
			Str("instance_project", dbInstance.Project).
			Msg("Soft-deleted stale instance no longer present in Incus")
	}
}
