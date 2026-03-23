// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package tpm

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	simulatorKeyOnce sync.Once
	simulatorPrivKey ed25519.PrivateKey
	simulatorPubKey  ed25519.PublicKey
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
	if strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("node id is required")
	}
	if len(nonce) == 0 {
		nonce = make([]byte, 16)
		if _, err := rand.Read(nonce); err != nil {
			return nil, fmt.Errorf("failed to generate fallback nonce: %w", err)
		}
	}

	// Generate TPM quote (this would interface with actual TPM hardware)
	quote, err := generateTPMQuoteForNode(nodeID)
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
	}

	ensureSimulatorKeypair()
	payload := buildAttestationPayload(report)
	report.Signature = ed25519.Sign(simulatorPrivKey, payload)
	report.PublicKey = append([]byte(nil), simulatorPubKey...)

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
	if report == nil {
		return false, fmt.Errorf("attestation report is nil")
	}
	if strings.TrimSpace(report.NodeID) == "" {
		return false, fmt.Errorf("attestation node id is missing")
	}
	if len(report.Quote) == 0 {
		return false, fmt.Errorf("attestation quote is missing")
	}
	if len(report.Nonce) == 0 {
		return false, fmt.Errorf("attestation nonce is missing")
	}
	if strings.TrimSpace(report.AttestationID) == "" {
		return false, fmt.Errorf("attestation id is missing")
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
	nowBucket := time.Now().UTC().Format("2006-01-02T15")
	pcr0 := sha256.Sum256([]byte("boot-sequence:" + nowBucket))
	pcr1 := sha256.Sum256([]byte("kernel-state:" + nowBucket))
	pcr7 := sha256.Sum256([]byte("secure-boot:" + nowBucket))

	return map[int][]byte{
		0: pcr0[:],
		1: pcr1[:],
		7: pcr7[:],
	}, nil
}

func verifyQuoteSignature(report *AttestationReport) error {
	if len(report.Signature) == 0 {
		return fmt.Errorf("missing attestation signature")
	}
	if len(report.PublicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: %d", len(report.PublicKey))
	}
	if err := Verify(report.NodeID, report.Quote); err != nil {
		return fmt.Errorf("invalid quote: %w", err)
	}

	payload := buildAttestationPayload(report)
	if !ed25519.Verify(ed25519.PublicKey(report.PublicKey), payload, report.Signature) {
		return fmt.Errorf("invalid attestation signature")
	}
	return nil
}

func verifyPCRValues(pcrValues map[int][]byte) error {
	if len(pcrValues) == 0 {
		return fmt.Errorf("empty PCR values")
	}

	requiredPCRs := []int{0, 1, 7}
	for _, pcr := range requiredPCRs {
		value, exists := pcrValues[pcr]
		if !exists {
			return fmt.Errorf("missing required PCR index: %d", pcr)
		}
		if len(value) == 0 {
			return fmt.Errorf("PCR %d value is empty", pcr)
		}
	}

	return nil
}

func ensureSimulatorKeypair() {
	simulatorKeyOnce.Do(func() {
		seed := sha256.Sum256([]byte("sovereign-map-attestation-simulator-key-v1"))
		simulatorPrivKey = ed25519.NewKeyFromSeed(seed[:])
		simulatorPubKey = simulatorPrivKey.Public().(ed25519.PublicKey)
	})
}

func buildAttestationPayload(report *AttestationReport) []byte {
	pcrDigest := hashPCRValues(report.PCRValues)
	payload := fmt.Sprintf(
		"%s|%d|%x|%x|%x",
		report.NodeID,
		report.Timestamp.UnixNano(),
		report.Quote,
		report.Nonce,
		pcrDigest,
	)
	return []byte(payload)
}

func hashPCRValues(pcrValues map[int][]byte) []byte {
	hasher := sha256.New()
	ordered := []int{0, 1, 7}
	for _, pcr := range ordered {
		_, _ = fmt.Fprintf(hasher, "%d", pcr)
		hasher.Write(pcrValues[pcr])
	}
	return hasher.Sum(nil)
}
