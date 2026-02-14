// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package island

import (
	"context"
	"sync"
	"time"
)

// Mode represents the operational mode of a node
type Mode int

const (
	// ModeOnline - Node is connected to network and participates in FL
	ModeOnline Mode = iota
	// ModeIsland - Node operates autonomously with cached updates
	ModeIsland
	// ModeTransition - Node is transitioning between modes
	ModeTransition
)

// Manager handles Island Mode transitions for offline operation
type Manager struct {
	mu                sync.RWMutex
	mode              Mode
	connectivityCheck func() bool
	checkInterval     time.Duration
	cachedUpdates     []Update
	maxCachedUpdates  int
	lastSync          time.Time
	listeners         []ModeChangeListener
	ctx               context.Context
	cancel            context.CancelFunc
}

// Update represents a federated learning update
type Update struct {
	Timestamp   time.Time
	Round       int
	ModelDelta  []byte
	Metadata    map[string]interface{}
	PeerID      string
}

// ModeChangeListener is called when mode changes
type ModeChangeListener func(oldMode, newMode Mode)

// NewManager creates a new Island Mode manager
func NewManager(checkInterval time.Duration, maxCachedUpdates int, connectivityCheck func() bool) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		mode:              ModeOnline,
		connectivityCheck: connectivityCheck,
		checkInterval:     checkInterval,
		cachedUpdates:     make([]Update, 0, maxCachedUpdates),
		maxCachedUpdates:  maxCachedUpdates,
		lastSync:          time.Now(),
		listeners:         make([]ModeChangeListener, 0),
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Start begins monitoring connectivity and managing mode transitions
func (m *Manager) Start() {
	go m.monitorConnectivity()
}

// Stop halts the Island Mode manager
func (m *Manager) Stop() {
	m.cancel()
}

// monitorConnectivity periodically checks network connectivity
func (m *Manager) monitorConnectivity() {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			isOnline := m.connectivityCheck()
			m.updateMode(isOnline)
		}
	}
}

// updateMode transitions between online and island modes
func (m *Manager) updateMode(isOnline bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldMode := m.mode
	newMode := oldMode

	if isOnline && oldMode == ModeIsland {
		// Transition to online: sync cached updates
		newMode = ModeOnline
		go m.syncCachedUpdates()
	} else if !isOnline && oldMode == ModeOnline {
		// Transition to island: cache future updates
		newMode = ModeIsland
	}

	if newMode != oldMode {
		m.mode = newMode
		m.notifyListeners(oldMode, newMode)
	}
}

// GetMode returns the current operational mode
func (m *Manager) GetMode() Mode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

// IsOnline returns true if node is connected to network
func (m *Manager) IsOnline() bool {
	return m.GetMode() == ModeOnline
}

// CacheUpdate stores an update for later synchronization
func (m *Manager) CacheUpdate(update Update) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.cachedUpdates) >= m.maxCachedUpdates {
		// Remove oldest update if cache is full
		m.cachedUpdates = m.cachedUpdates[1:]
	}

	m.cachedUpdates = append(m.cachedUpdates, update)
	return nil
}

// GetCachedUpdates returns all cached updates
func (m *Manager) GetCachedUpdates() []Update {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	updates := make([]Update, len(m.cachedUpdates))
	copy(updates, m.cachedUpdates)
	return updates
}

// syncCachedUpdates sends cached updates when coming back online
func (m *Manager) syncCachedUpdates() {
	m.mu.Lock()
	updates := m.cachedUpdates
	m.cachedUpdates = make([]Update, 0, m.maxCachedUpdates)
	m.lastSync = time.Now()
	m.mu.Unlock()

	// TODO: Send updates to aggregation server
	// This would integrate with the batch aggregator
	_ = updates // Placeholder for actual sync logic
}

// AddModeChangeListener registers a callback for mode changes
func (m *Manager) AddModeChangeListener(listener ModeChangeListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// notifyListeners calls all registered mode change listeners
func (m *Manager) notifyListeners(oldMode, newMode Mode) {
	for _, listener := range m.listeners {
		go listener(oldMode, newMode)
	}
}

// GetStatus returns current Island Mode status
func (m *Manager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var modeStr string
	switch m.mode {
	case ModeOnline:
		modeStr = "online"
	case ModeIsland:
		modeStr = "island"
	case ModeTransition:
		modeStr = "transition"
	default:
		modeStr = "unknown"
	}

	return map[string]interface{}{
		"mode":                 modeStr,
		"cached_updates":       len(m.cachedUpdates),
		"max_cached_updates":   m.maxCachedUpdates,
		"last_sync":            m.lastSync,
		"time_since_last_sync": time.Since(m.lastSync),
	}
}

// ForceSync immediately attempts to sync cached updates
func (m *Manager) ForceSync() error {
	if !m.IsOnline() {
		return nil // Skip if offline
	}
	go m.syncCachedUpdates()
	return nil
}
