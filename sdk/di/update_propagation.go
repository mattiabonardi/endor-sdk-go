package di

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// UpdateType represents the type of dependency update
type UpdateType int

const (
	UpdateTypeUnknown UpdateType = iota
	UpdateTypeConfiguration
	UpdateTypeHotReload
	UpdateTypeVersionChange
	UpdateTypeDependencyChange
	UpdateTypeForceRefresh
)

// String returns the string representation of UpdateType
func (ut UpdateType) String() string {
	switch ut {
	case UpdateTypeConfiguration:
		return "Configuration"
	case UpdateTypeHotReload:
		return "HotReload"
	case UpdateTypeVersionChange:
		return "VersionChange"
	case UpdateTypeDependencyChange:
		return "DependencyChange"
	case UpdateTypeForceRefresh:
		return "ForceRefresh"
	default:
		return "Unknown"
	}
}

// UpdatePropagationStrategy defines how updates should be propagated
type UpdatePropagationStrategy int

const (
	UpdatePropagationImmediate UpdatePropagationStrategy = iota
	UpdatePropagationBatched
	UpdatePropagationScheduled
	UpdatePropagationManual
)

// String returns the string representation of UpdatePropagationStrategy
func (ups UpdatePropagationStrategy) String() string {
	switch ups {
	case UpdatePropagationImmediate:
		return "Immediate"
	case UpdatePropagationBatched:
		return "Batched"
	case UpdatePropagationScheduled:
		return "Scheduled"
	case UpdatePropagationManual:
		return "Manual"
	default:
		return "Unknown"
	}
}

// DependencyUpdate represents an update to a dependency
type DependencyUpdate struct {
	DependencyType  string
	UpdateType      UpdateType
	OldVersion      string
	NewVersion      string
	OldInstance     interface{}
	NewInstance     interface{}
	Configuration   map[string]interface{}
	PropagationPath []string
	Timestamp       time.Time
	Error           error
	RequiresRestart bool
}

// UpdateListener handles dependency update events
type UpdateListener interface {
	OnDependencyUpdated(update DependencyUpdate)
}

// Versionable represents a dependency that supports versioning
type Versionable interface {
	GetVersion() string
}

// HotReloadable represents a dependency that supports hot reloading
type HotReloadable interface {
	Reload(ctx context.Context, config map[string]interface{}) error
}

// ConfigUpdatable represents a dependency that can update its configuration
type ConfigUpdatable interface {
	UpdateConfig(config map[string]interface{}) error
}

// DependencyUpdateManager manages dependency updates and propagation
type DependencyUpdateManager struct {
	container       *containerImpl
	listeners       []UpdateListener
	updateQueue     chan DependencyUpdate
	batchQueue      []DependencyUpdate
	batchTimeout    time.Duration
	strategy        UpdatePropagationStrategy
	versions        map[string]string
	dependencyGraph map[string][]string // dependency -> dependents
	mu              sync.RWMutex
	updateHistory   []DependencyUpdate
	maxHistorySize  int
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// NewDependencyUpdateManager creates a new dependency update manager
func NewDependencyUpdateManager(container *containerImpl) *DependencyUpdateManager {
	ctx, cancel := context.WithCancel(context.Background())

	dum := &DependencyUpdateManager{
		container:       container,
		listeners:       make([]UpdateListener, 0),
		updateQueue:     make(chan DependencyUpdate, 1000),
		batchQueue:      make([]DependencyUpdate, 0),
		batchTimeout:    time.Second * 5,
		strategy:        UpdatePropagationImmediate,
		versions:        make(map[string]string),
		dependencyGraph: make(map[string][]string),
		updateHistory:   make([]DependencyUpdate, 0),
		maxHistorySize:  1000,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start the update processor
	dum.wg.Add(1)
	go dum.processUpdates()

	return dum
}

// AddUpdateListener adds a listener for dependency updates
func (dum *DependencyUpdateManager) AddUpdateListener(listener UpdateListener) {
	dum.mu.Lock()
	defer dum.mu.Unlock()
	dum.listeners = append(dum.listeners, listener)
}

// SetUpdatePropagationStrategy sets the strategy for update propagation
func (dum *DependencyUpdateManager) SetUpdatePropagationStrategy(strategy UpdatePropagationStrategy) {
	dum.mu.Lock()
	defer dum.mu.Unlock()
	dum.strategy = strategy
}

// SetBatchTimeout sets the timeout for batched updates
func (dum *DependencyUpdateManager) SetBatchTimeout(timeout time.Duration) {
	dum.mu.Lock()
	defer dum.mu.Unlock()
	dum.batchTimeout = timeout
}

// RegisterDependencyVersion registers a version for a dependency
func (dum *DependencyUpdateManager) RegisterDependencyVersion(dependencyType string, version string) {
	dum.mu.Lock()
	defer dum.mu.Unlock()
	dum.versions[dependencyType] = version
}

// GetDependencyVersion returns the current version of a dependency
func (dum *DependencyUpdateManager) GetDependencyVersion(dependencyType string) (string, bool) {
	dum.mu.RLock()
	defer dum.mu.RUnlock()
	version, exists := dum.versions[dependencyType]
	return version, exists
}

// AddDependencyRelationship adds a dependency relationship for update propagation
func (dum *DependencyUpdateManager) AddDependencyRelationship(dependency, dependent string) {
	dum.mu.Lock()
	defer dum.mu.Unlock()

	if dependents, exists := dum.dependencyGraph[dependency]; exists {
		// Check if already exists
		for _, existing := range dependents {
			if existing == dependent {
				return
			}
		}
		dum.dependencyGraph[dependency] = append(dependents, dependent)
	} else {
		dum.dependencyGraph[dependency] = []string{dependent}
	}
}

// TriggerDependencyUpdate triggers an update for a specific dependency
func (dum *DependencyUpdateManager) TriggerDependencyUpdate(dependencyType string, updateType UpdateType, newInstance interface{}, config map[string]interface{}) error {
	dum.mu.RLock()
	// For now, we'll use the new instance as old instance since we don't have instance tracking yet
	// In a full implementation, this would track resolved instances
	var oldInstance interface{} = nil
	oldVersion := dum.versions[dependencyType]
	dum.mu.RUnlock()

	var newVersion string
	if versionable, ok := newInstance.(Versionable); ok {
		newVersion = versionable.GetVersion()
	}

	update := DependencyUpdate{
		DependencyType:  dependencyType,
		UpdateType:      updateType,
		OldVersion:      oldVersion,
		NewVersion:      newVersion,
		OldInstance:     oldInstance,
		NewInstance:     newInstance,
		Configuration:   config,
		PropagationPath: []string{dependencyType},
		Timestamp:       time.Now(),
		RequiresRestart: dum.requiresRestart(updateType, oldInstance, newInstance),
	}

	return dum.queueUpdate(update)
}

// TriggerConfigurationUpdate triggers a configuration update for a dependency
func (dum *DependencyUpdateManager) TriggerConfigurationUpdate(dependencyType string, config map[string]interface{}) error {
	// For this initial implementation, we'll assume the caller provides the instance
	// In a full implementation, this would resolve the instance from the container

	update := DependencyUpdate{
		DependencyType:  dependencyType,
		UpdateType:      UpdateTypeConfiguration,
		Configuration:   config,
		PropagationPath: []string{dependencyType},
		Timestamp:       time.Now(),
		RequiresRestart: false,
	}

	return dum.queueUpdate(update)
}

// ForceRefresh forces a refresh of all instances of a dependency type
func (dum *DependencyUpdateManager) ForceRefresh(dependencyType string) error {
	// For this initial implementation, we'll create an update without resolving instances
	// In a full implementation, this would resolve the current instance from the container

	update := DependencyUpdate{
		DependencyType:  dependencyType,
		UpdateType:      UpdateTypeForceRefresh,
		PropagationPath: []string{dependencyType},
		Timestamp:       time.Now(),
		RequiresRestart: true,
	}

	return dum.queueUpdate(update)
}

// GetUpdateHistory returns the update history
func (dum *DependencyUpdateManager) GetUpdateHistory() []DependencyUpdate {
	dum.mu.RLock()
	defer dum.mu.RUnlock()

	history := make([]DependencyUpdate, len(dum.updateHistory))
	copy(history, dum.updateHistory)
	return history
}

// GetDependencyGraph returns the current dependency graph for update propagation
func (dum *DependencyUpdateManager) GetDependencyGraph() map[string][]string {
	dum.mu.RLock()
	defer dum.mu.RUnlock()

	graph := make(map[string][]string)
	for dep, dependents := range dum.dependencyGraph {
		graph[dep] = make([]string, len(dependents))
		copy(graph[dep], dependents)
	}
	return graph
}

// Close shuts down the update manager
func (dum *DependencyUpdateManager) Close() {
	dum.cancel()
	close(dum.updateQueue)
	dum.wg.Wait()
}

// queueUpdate queues an update for processing
func (dum *DependencyUpdateManager) queueUpdate(update DependencyUpdate) error {
	select {
	case dum.updateQueue <- update:
		return nil
	case <-dum.ctx.Done():
		return fmt.Errorf("update manager is shutting down")
	default:
		return fmt.Errorf("update queue is full")
	}
}

// processUpdates processes queued updates based on the propagation strategy
func (dum *DependencyUpdateManager) processUpdates() {
	defer dum.wg.Done()

	var batchTimer *time.Timer

	for {
		select {
		case update, ok := <-dum.updateQueue:
			if !ok {
				// Channel closed, process any remaining batched updates
				if len(dum.batchQueue) > 0 {
					dum.processBatchedUpdates()
				}
				return
			}

			dum.mu.RLock()
			strategy := dum.strategy
			dum.mu.RUnlock()

			switch strategy {
			case UpdatePropagationImmediate:
				dum.processUpdate(update)
			case UpdatePropagationBatched:
				dum.mu.Lock()
				dum.batchQueue = append(dum.batchQueue, update)
				dum.mu.Unlock()

				if batchTimer == nil {
					batchTimer = time.AfterFunc(dum.batchTimeout, func() {
						dum.processBatchedUpdates()
						batchTimer = nil
					})
				}
			case UpdatePropagationManual:
				dum.addToHistory(update)
			}

		case <-dum.ctx.Done():
			return
		}
	}
}

// processUpdate processes a single update and propagates it
func (dum *DependencyUpdateManager) processUpdate(update DependencyUpdate) {
	// Add to history
	dum.addToHistory(update)

	// Update version if applicable
	if update.NewVersion != "" {
		dum.mu.Lock()
		dum.versions[update.DependencyType] = update.NewVersion
		dum.mu.Unlock()
	}

	// Notify listeners
	dum.notifyListeners(update)

	// Propagate to dependents
	dum.propagateUpdate(update)
}

// processBatchedUpdates processes all batched updates
func (dum *DependencyUpdateManager) processBatchedUpdates() {
	dum.mu.Lock()
	updates := make([]DependencyUpdate, len(dum.batchQueue))
	copy(updates, dum.batchQueue)
	dum.batchQueue = dum.batchQueue[:0]
	dum.mu.Unlock()

	for _, update := range updates {
		dum.processUpdate(update)
	}
}

// propagateUpdate propagates an update to dependent services
func (dum *DependencyUpdateManager) propagateUpdate(update DependencyUpdate) {
	dum.mu.RLock()
	dependents, exists := dum.dependencyGraph[update.DependencyType]
	dum.mu.RUnlock()

	if !exists || len(dependents) == 0 {
		return
	}

	for _, dependent := range dependents {
		// Check if already in propagation path to avoid cycles
		inPath := false
		for _, pathItem := range update.PropagationPath {
			if pathItem == dependent {
				inPath = true
				break
			}
		}

		if inPath {
			continue
		}

		// Create propagated update
		propagatedUpdate := DependencyUpdate{
			DependencyType:  dependent,
			UpdateType:      UpdateTypeDependencyChange,
			OldInstance:     update.OldInstance,
			NewInstance:     update.NewInstance,
			Configuration:   update.Configuration,
			PropagationPath: append(update.PropagationPath, dependent),
			Timestamp:       time.Now(),
			RequiresRestart: update.RequiresRestart,
		}

		// Queue the propagated update
		select {
		case dum.updateQueue <- propagatedUpdate:
		case <-dum.ctx.Done():
			return
		default:
			// Queue full, skip this propagation
		}
	}
}

// notifyListeners notifies all registered listeners about an update
func (dum *DependencyUpdateManager) notifyListeners(update DependencyUpdate) {
	dum.mu.RLock()
	listeners := make([]UpdateListener, len(dum.listeners))
	copy(listeners, dum.listeners)
	dum.mu.RUnlock()

	for _, listener := range listeners {
		// Call listener with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Log panic but don't let it crash the update manager
					fmt.Printf("Update listener panicked: %v\n", r)
				}
			}()
			listener.OnDependencyUpdated(update)
		}()
	}
}

// addToHistory adds an update to the history
func (dum *DependencyUpdateManager) addToHistory(update DependencyUpdate) {
	dum.mu.Lock()
	defer dum.mu.Unlock()

	dum.updateHistory = append(dum.updateHistory, update)

	// Trim history if too large
	if len(dum.updateHistory) > dum.maxHistorySize {
		copy(dum.updateHistory, dum.updateHistory[1:])
		dum.updateHistory = dum.updateHistory[:dum.maxHistorySize]
	}
}

// requiresRestart determines if an update requires a service restart
func (dum *DependencyUpdateManager) requiresRestart(updateType UpdateType, oldInstance, newInstance interface{}) bool {
	switch updateType {
	case UpdateTypeHotReload, UpdateTypeConfiguration:
		return false
	case UpdateTypeVersionChange, UpdateTypeForceRefresh:
		return true
	case UpdateTypeDependencyChange:
		// Check if instances are the same
		return oldInstance != newInstance
	default:
		return true
	}
}
