// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package api

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

const defaultLedgerCap = 1000

// LedgerEntry records a single proof-verification event.
type LedgerEntry struct {
	ID        uint64 `json:"id"`
	Timestamp string `json:"timestamp"`
	// EventType is "snark_verify" or "hybrid_verify".
	EventType string `json:"event_type"`
	// ProofHash is the hex-encoded SHA-256 of the raw proof bytes.
	ProofHash string `json:"proof_hash"`
	// Role is the X-API-Role value supplied by the caller.
	Role      string `json:"role"`
	Accepted  bool   `json:"accepted"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

// ProofLedger is a thread-safe, append-only ring buffer of verification events.
// Once capacity is reached the oldest entry is overwritten.
type ProofLedger struct {
	mu      sync.RWMutex
	entries []LedgerEntry
	cap     int
	seq     uint64
}

// NewProofLedger creates a ledger holding at most capacity entries.
// If capacity ≤ 0 the default (1000) is used.
func NewProofLedger(capacity int) *ProofLedger {
	if capacity <= 0 {
		capacity = defaultLedgerCap
	}
	return &ProofLedger{
		entries: make([]LedgerEntry, 0, capacity),
		cap:     capacity,
	}
}

// Record appends a new verification event to the ledger.
// proofBytes is hashed with SHA-256 so raw proof data is never stored.
func (l *ProofLedger) Record(eventType string, proofBytes []byte, role string, accepted bool, latencyMs int64, verifyErr error) {
	h := sha256.Sum256(proofBytes)
	errStr := ""
	if verifyErr != nil {
		errStr = verifyErr.Error()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.seq++
	entry := LedgerEntry{
		ID:        l.seq,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		EventType: eventType,
		ProofHash: hex.EncodeToString(h[:]),
		Role:      role,
		Accepted:  accepted,
		LatencyMs: latencyMs,
		Error:     errStr,
	}

	if len(l.entries) >= l.cap {
		// Ring-buffer: shift left, overwrite last slot
		copy(l.entries, l.entries[1:])
		l.entries[len(l.entries)-1] = entry
	} else {
		l.entries = append(l.entries, entry)
	}
}

// Entries returns a snapshot of all ledger entries (oldest first).
func (l *ProofLedger) Entries() []LedgerEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]LedgerEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

// Len returns the current number of entries in the ledger.
func (l *ProofLedger) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}
