package sdk_entity

import (
	"sync"
	"time"
)

const (
	// ephemeralRegistryTTL is the time-to-live for a per-user debug overlay registry.
	ephemeralRegistryTTL = 30 * time.Second

	// ephemeralCleanupInterval is how often the background cleanup goroutine runs.
	ephemeralCleanupInterval = 1 * time.Minute
)

// EphemeralRegistryEntry holds a per-user debug overlay registry and its expiry timestamp.
type EphemeralRegistryEntry struct {
	dictionary map[string]EndorHandlerDictionary
	expiresAt  time.Time
}

// EphemeralCacheManager manages per-user ephemeral registry entries with TTL-based expiry.
// All operations are thread-safe.
type EphemeralCacheManager struct {
	mu      sync.RWMutex
	entries map[string]*EphemeralRegistryEntry
}

// newEphemeralCacheManager creates a new EphemeralCacheManager and starts the background
// cleanup goroutine that prevents memory leaks from stale entries.
func newEphemeralCacheManager() *EphemeralCacheManager {
	mgr := &EphemeralCacheManager{
		entries: make(map[string]*EphemeralRegistryEntry),
	}
	go mgr.cleanupLoop()
	return mgr
}

// Get returns the cached dictionary for the given userID if it exists and has not expired.
// Returns nil on a cache miss or if the TTL has elapsed.
func (m *EphemeralCacheManager) Get(userID string) map[string]EndorHandlerDictionary {
	m.mu.RLock()
	entry, ok := m.entries[userID]
	m.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return nil
	}
	return entry.dictionary
}

// Set stores a dictionary for the given userID, resetting the TTL to ephemeralRegistryTTL.
func (m *EphemeralCacheManager) Set(userID string, dict map[string]EndorHandlerDictionary) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[userID] = &EphemeralRegistryEntry{
		dictionary: dict,
		expiresAt:  time.Now().Add(ephemeralRegistryTTL),
	}
}

// InvalidateAll removes all ephemeral registry entries, forcing a full rebuild on next access.
// Call this whenever the underlying production registry changes.
func (m *EphemeralCacheManager) InvalidateAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = make(map[string]*EphemeralRegistryEntry)
}

// cleanupLoop runs on a fixed interval, deleting expired entries to prevent memory leaks.
func (m *EphemeralCacheManager) cleanupLoop() {
	ticker := time.NewTicker(ephemeralCleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for userID, entry := range m.entries {
			if now.After(entry.expiresAt) {
				delete(m.entries, userID)
			}
		}
		m.mu.Unlock()
	}
}
