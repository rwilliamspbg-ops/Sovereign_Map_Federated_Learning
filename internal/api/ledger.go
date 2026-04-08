// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultLedgerCap = 1000
const defaultCheckpointEvery = 100

// ProofLedgerStore defines append-only ledger operations shared by in-memory and SQL backends.
type ProofLedgerStore interface {
	Record(eventType string, proofBytes []byte, role string, accepted bool, latencyMs int64, verifyErr error)
	RecordWithOptions(eventType string, proofBytes []byte, role string, accepted bool, latencyMs int64, verifyErr error, opts LedgerRecordOptions) (LedgerEntry, bool)
	Entries() []LedgerEntry
	Len() int
	Checkpoints() []LedgerCheckpoint
	Reconcile() LedgerReconcileReport
	StorageMode() string
	Capacity() int
	Readiness() (bool, string)
}

// LedgerEntry records a single proof-verification event.
type LedgerEntry struct {
	ID        uint64 `json:"id"`
	EntryID   string `json:"entry_id"`
	StreamID  string `json:"stream_id"`
	SeqNo     uint64 `json:"seq_no"`
	Timestamp string `json:"timestamp"`
	// EventType is "snark_verify" or "hybrid_verify".
	EventType string `json:"event_type"`
	// ProofHash is the hex-encoded SHA-256 of the raw proof bytes.
	ProofHash      string `json:"proof_hash"`
	PrevHash       string `json:"prev_hash"`
	EntryHash      string `json:"entry_hash"`
	IdempotencyKey string `json:"idempotency_key"`
	Replay         bool   `json:"replay"`
	// Role is the X-API-Role value supplied by the caller.
	Role      string `json:"role"`
	Accepted  bool   `json:"accepted"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

// LedgerRecordOptions adds append semantics for stream and idempotency.
type LedgerRecordOptions struct {
	StreamID       string
	IdempotencyKey string
}

// LedgerCheckpoint anchors a stream hash at a deterministic sequence interval.
type LedgerCheckpoint struct {
	StreamID   string `json:"stream_id"`
	ThroughSeq uint64 `json:"through_seq"`
	RootHash   string `json:"root_hash"`
	Timestamp  string `json:"timestamp"`
}

// LedgerReconcileReport summarizes integrity checks over in-memory entries.
type LedgerReconcileReport struct {
	Healthy      bool                   `json:"healthy"`
	TotalEntries int                    `json:"total_entries"`
	TotalStreams int                    `json:"total_streams"`
	Streams      map[string]interface{} `json:"streams"`
	Issues       []string               `json:"issues"`
	VerifiedAt   string                 `json:"verified_at"`
}

// ProofLedger is a thread-safe, append-only ring buffer of verification events.
// Once capacity is reached the oldest entry is overwritten.
type ProofLedger struct {
	mu               sync.RWMutex
	entries          []LedgerEntry
	cap              int
	seq              uint64
	streamSeq        map[string]uint64
	streamHead       map[string]string
	idempotencyIndex map[string]LedgerEntry
	checkpointEvery  uint64
	checkpoints      []LedgerCheckpoint
}

// NewProofLedger creates a ledger holding at most capacity entries.
// If capacity ≤ 0 the default (1000) is used.
func NewProofLedger(capacity int) *ProofLedger {
	if capacity <= 0 {
		capacity = defaultLedgerCap
	}
	return &ProofLedger{
		entries:          make([]LedgerEntry, 0, capacity),
		cap:              capacity,
		streamSeq:        make(map[string]uint64),
		streamHead:       make(map[string]string),
		idempotencyIndex: make(map[string]LedgerEntry),
		checkpointEvery:  defaultCheckpointEvery,
		checkpoints:      make([]LedgerCheckpoint, 0, capacity/defaultCheckpointEvery+1),
	}
}

// Record appends a new verification event to the ledger.
// proofBytes is hashed with SHA-256 so raw proof data is never stored.
func (l *ProofLedger) Record(eventType string, proofBytes []byte, role string, accepted bool, latencyMs int64, verifyErr error) {
	l.RecordWithOptions(eventType, proofBytes, role, accepted, latencyMs, verifyErr, LedgerRecordOptions{})
}

// RecordWithOptions appends a verification event and supports idempotent replays.
// When an idempotency key reappears, the original entry is returned and replay=true.
func (l *ProofLedger) RecordWithOptions(eventType string, proofBytes []byte, role string, accepted bool, latencyMs int64, verifyErr error, opts LedgerRecordOptions) (LedgerEntry, bool) {
	h := sha256.Sum256(proofBytes)
	proofHash := hex.EncodeToString(h[:])
	errStr := ""
	if verifyErr != nil {
		errStr = verifyErr.Error()
	}
	streamID := strings.TrimSpace(opts.StreamID)
	if streamID == "" {
		streamID = strings.TrimSpace(eventType)
	}
	idempotencyKey := strings.TrimSpace(opts.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = computeDeterministicID(streamID, proofHash, role, "")
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if existing, ok := l.idempotencyIndex[idempotencyKey]; ok {
		replay := existing
		replay.Replay = true
		return replay, true
	}

	l.seq++
	streamSeq := l.streamSeq[streamID] + 1
	prevHash := l.streamHead[streamID]
	entryHash := computeEntryHash(streamID, streamSeq, proofHash, role, accepted, errStr, idempotencyKey, prevHash)
	entryID := computeDeterministicID(streamID, proofHash, role, idempotencyKey)
	entry := LedgerEntry{
		ID:             l.seq,
		EntryID:        entryID,
		StreamID:       streamID,
		SeqNo:          streamSeq,
		Timestamp:      time.Now().UTC().Format(time.RFC3339Nano),
		EventType:      eventType,
		ProofHash:      proofHash,
		PrevHash:       prevHash,
		EntryHash:      entryHash,
		IdempotencyKey: idempotencyKey,
		Replay:         false,
		Role:           role,
		Accepted:       accepted,
		LatencyMs:      latencyMs,
		Error:          errStr,
	}
	l.streamSeq[streamID] = streamSeq
	l.streamHead[streamID] = entryHash
	l.idempotencyIndex[idempotencyKey] = entry

	if len(l.entries) >= l.cap {
		// Ring-buffer: shift left, overwrite last slot while preserving append semantics.
		evicted := l.entries[0]
		if idxEntry, ok := l.idempotencyIndex[evicted.IdempotencyKey]; ok && idxEntry.ID == evicted.ID {
			delete(l.idempotencyIndex, evicted.IdempotencyKey)
		}
		copy(l.entries, l.entries[1:])
		l.entries[len(l.entries)-1] = entry
	} else {
		l.entries = append(l.entries, entry)
	}
	if l.checkpointEvery > 0 && streamSeq%l.checkpointEvery == 0 {
		l.checkpoints = append(l.checkpoints, LedgerCheckpoint{
			StreamID:   streamID,
			ThroughSeq: streamSeq,
			RootHash:   entryHash,
			Timestamp:  entry.Timestamp,
		})
	}

	return entry, false
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

// StorageMode reports the underlying ledger backend mode.
func (l *ProofLedger) StorageMode() string {
	return "cockroach-compatible-inmemory"
}

// Capacity reports the configured read-retention capacity.
func (l *ProofLedger) Capacity() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.cap
}

// Readiness reports whether the backend is writable and available for serving traffic.
func (l *ProofLedger) Readiness() (bool, string) {
	return true, ""
}

// Checkpoints returns a snapshot of generated stream checkpoints.
func (l *ProofLedger) Checkpoints() []LedgerCheckpoint {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]LedgerCheckpoint, len(l.checkpoints))
	copy(result, l.checkpoints)
	return result
}

// Reconcile verifies per-stream sequence continuity and hash chain integrity.
func (l *ProofLedger) Reconcile() LedgerReconcileReport {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return reconcileEntries(l.entries)
}

func reconcileEntries(entries []LedgerEntry) LedgerReconcileReport {
	issues := make([]string, 0)
	streams := make(map[string]interface{})
	byStream := make(map[string][]LedgerEntry)
	for _, entry := range entries {
		byStream[entry.StreamID] = append(byStream[entry.StreamID], entry)
	}
	for streamID, entries := range byStream {
		streamIssues := make([]string, 0)
		for i := 0; i < len(entries); i++ {
			if i > 0 {
				wantSeq := entries[i-1].SeqNo + 1
				if entries[i].SeqNo != wantSeq {
					streamIssues = append(streamIssues, fmt.Sprintf("sequence gap at stream=%s seq=%d expected=%d", streamID, entries[i].SeqNo, wantSeq))
				}
				expectedPrev := entries[i-1].EntryHash
				if entries[i].PrevHash != expectedPrev {
					streamIssues = append(streamIssues, fmt.Sprintf("prev_hash mismatch at stream=%s seq=%d", streamID, entries[i].SeqNo))
				}
			}
			recomputed := computeEntryHash(streamID, entries[i].SeqNo, entries[i].ProofHash, entries[i].Role, entries[i].Accepted, entries[i].Error, entries[i].IdempotencyKey, entries[i].PrevHash)
			if entries[i].EntryHash != recomputed {
				streamIssues = append(streamIssues, fmt.Sprintf("entry_hash mismatch at stream=%s seq=%d", streamID, entries[i].SeqNo))
			}
		}
		if len(streamIssues) > 0 {
			issues = append(issues, streamIssues...)
		}
		streams[streamID] = map[string]interface{}{
			"entries": len(entries),
			"issues":  streamIssues,
		}
	}

	return LedgerReconcileReport{
		Healthy:      len(issues) == 0,
		TotalEntries: len(entries),
		TotalStreams: len(byStream),
		Streams:      streams,
		Issues:       issues,
		VerifiedAt:   time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func computeEntryHash(streamID string, seqNo uint64, proofHash string, role string, accepted bool, errStr string, idempotencyKey string, prevHash string) string {
	base := strings.Join([]string{
		streamID,
		strconv.FormatUint(seqNo, 10),
		proofHash,
		role,
		strconv.FormatBool(accepted),
		errStr,
		idempotencyKey,
		prevHash,
	}, "|")
	h := sha256.Sum256([]byte(base))
	return hex.EncodeToString(h[:])
}

func computeDeterministicID(streamID string, proofHash string, role string, idempotencyKey string) string {
	base := strings.Join([]string{streamID, proofHash, role, idempotencyKey}, "|")
	h := sha256.Sum256([]byte(base))
	return hex.EncodeToString(h[:])
}
