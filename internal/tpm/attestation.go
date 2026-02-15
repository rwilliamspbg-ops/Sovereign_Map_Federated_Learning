// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package tpm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// AttestationReport represents a complete TPM attestation
type AttestationReport struct {
	NodeID        string
	Timestamp     time.Time
	Quote         []byte
	PCRValues     map[int][]byte
	Nonce         []byte
	Signature     []byte
	PublicKey     []byte
	AttestationID string
}

// AttestationManager handles TPM attestation lifecycle
type AttestationManager struct {
	mu               sync.RWMutex
	reports          map[string]*AttestationReport
	maxReports       int
	attestationCache *AttestationCache
	enabled          bool
}

// AttestationCache stores recently verified attestations
type AttestationCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

type CacheEntry struct {
	Report    *AttestationReport
	ExpiresAt time.Time
	Verified  bool
}

// NewAttestationManager creates a new attestation manager
func NewAttestationManager(maxReports int, cacheTTL time.Duration, enabled bool) *AttestationManager {
	return &AttestationManager{
		reports:    make(map[string]*AttestationReport),
		maxReports: maxReports,
		attestationCache: &AttestationCache{
			entries: make(map[string]*CacheEntry),
			ttl:     cacheTTL,
		},
		enabled: enabled,
	}
}

// GenerateAttestation creates a new TPM attestation report
func (am *AttestationManager) GenerateAttestation(nodeID string, nonce []byte) (*AttestationReport, error) {
	if !am.enabled {
		return nil, fmt.Errorf("TPM attestation is disabled")
	}

	// Generate TPM quote (this would interface with actual TPM hardware)
	quote, err := GenerateTPMQuote()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TPM quote: %w", err)
	}

	// Read PCR values (Platform Configuration Registers)
	pcrValues, err := readPCRValues()
	if err != nil {
		return nil, fmt.Errorf("failed to read PCR values: %w", err)
	}

	// Create attestation report
	report := &AttestationReport{
		NodeID:    nodeID,
		Timestamp: time.Now(),
		Quote:     quote,
		PCRValues: pcrValues,
		Nonce:     nonce,
		Signature: []byte("signature-stub"), // Would be actual TPM signature
		PublicKey: []byte("public-key-stub"), // Would be actual TPM public key
	}

	// Generate attestation ID
	report.AttestationID = am.generateAttestationID(report)

	// Store report
	am.mu.Lock()
	if len(am.reports) >= am.maxReports {
		// Remove oldest report
		am.evictOldestReport()
	}
	am.reports[report.AttestationID] = report
	am.mu.Unlock()

	return report, nil
}

// VerifyAttestation verifies a TPM attestation report
func (am *AttestationManager) VerifyAttestation(report *AttestationReport) (bool, error) {
	if !am.enabled {
		return true, nil // Skip verification if TPM is disabled
	}

	// Check cache first
	if cached := am.attestationCache.Get(report.AttestationID); cached != nil {
		if cached.Verified && time.Now().Before(cached.ExpiresAt) {
			return true, nil
		}
	}

	// Verify timestamp is recent (within 5 minutes)
	if time.Since(report.Timestamp) > 5*time.Minute {
		return false, fmt.Errorf("attestation timestamp too old")
	}

	// Verify quote signature (would use actual TPM verification)
	if err := verifyQuoteSignature(report); err != nil {
		return false, fmt.Errorf("quote signature verification failed: %w", err)
	}

	// Verify PCR values match expected state
	if err := verifyPCRValues(report.PCRValues); err != nil {
		return false, fmt.Errorf("PCR verification failed: %w", err)
	}

	// Cache the verified attestation
	am.attestationCache.Set(report.AttestationID, report, true)

	return true, nil
}

// GetAttestationReport retrieves a stored attestation report
func (am *AttestationManager) GetAttestationReport(attestationID string) (*AttestationReport, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	report, exists := am.reports[attestationID]
	if !exists {
		return nil, fmt.Errorf("attestation report not found: %s", attestationID)
	}

	return report, nil
}

// generateAttestationID creates a unique ID for an attestation
func (am *AttestationManager) generateAttestationID(report *AttestationReport) string {
	data := fmt.Sprintf("%s:%d:%x", report.NodeID, report.Timestamp.Unix(), report.Quote)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// evictOldestReport removes the oldest attestation report
func (am *AttestationManager) evictOldestReport() {
	var oldestID string
	var oldestTime time.Time

	for id, report := range am.reports {
		if oldestID == "" || report.Timestamp.Before(oldestTime) {
			oldestID = id
			oldestTime = report.Timestamp
		}
	}

	if oldestID != "" {
		delete(am.reports, oldestID)
	}
}

// Get retrieves a cached attestation entry
func (ac *AttestationCache) Get(attestationID string) *CacheEntry {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	entry, exists := ac.entries[attestationID]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil
	}

	return entry
}

// Set stores an attestation in the cache
func (ac *AttestationCache) Set(attestationID string, report *AttestationReport, verified bool) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.entries[attestationID] = &CacheEntry{
		Report:    report,
		ExpiresAt: time.Now().Add(ac.ttl),
		Verified:  verified,
	}
}

// Helper functions (stubs for actual TPM operations)

func readPCRValues() (map[int][]byte, error) {
	// Would read actual PCR values from TPM
	return map[int][]byte{
		0: []byte("pcr0-value"),
		1: []byte("pcr1-value"),
		7: []byte("pcr7-value"),
	}, nil
}

func verifyQuoteSignature(report *AttestationReport) error {
	// Would verify actual TPM signature
	return nil
}

func verifyPCRValues(pcrValues map[int][]byte) error {
	// Would verify PCR values match expected state
	return nil
}
