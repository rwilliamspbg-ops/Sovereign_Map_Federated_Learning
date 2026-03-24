package tpm

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
)

func TestNewAttestationManager(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	if manager == nil {
		t.Fatal("expected non-nil attestation manager")
	}
}

func TestGenerateAttestation(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	nodeID := "test-node-001"

	attestation, err := manager.GenerateAttestation(nodeID, []byte("nonce-1"))
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}

	if attestation.NodeID != nodeID {
		t.Errorf("Expected nodeID %s, got %s", nodeID, attestation.NodeID)
	}

	if attestation.Quote == nil {
		t.Error("Expected non-nil quote")
	}

	if len(attestation.PCRValues) == 0 {
		t.Error("Expected PCR values")
	}
}

func TestVerifyAttestation(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	nodeID := "test-node-002"

	attestation, err := manager.GenerateAttestation(nodeID, []byte("nonce-2"))
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}

	valid, err := manager.VerifyAttestation(attestation)
	if err != nil {
		t.Fatalf("Failed to verify attestation: %v", err)
	}

	if !valid {
		t.Error("Expected valid attestation")
	}
}

func TestGetVerifiedQuoteCache(t *testing.T) {
	nodeID := "cache-node"

	q1, err := GetVerifiedQuote(nodeID)
	if err != nil {
		t.Fatalf("first GetVerifiedQuote failed: %v", err)
	}
	q2, err := GetVerifiedQuote(nodeID)
	if err != nil {
		t.Fatalf("second GetVerifiedQuote failed: %v", err)
	}
	if string(q1) != string(q2) {
		t.Fatal("expected cached quote to match initial quote")
	}
}

func TestByzantineVerification(t *testing.T) {
	tests := []struct {
		name          string
		totalNodes    int
		faultyNodes   int
		expectSuccess bool
	}{
		{"No faulty nodes", 10, 0, true},
		{"Within tolerance", 10, 3, true},
		{"At boundary", 10, 4, true},
		{"Exceeds tolerance", 10, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifyByzantineResilience(tt.totalNodes, tt.faultyNodes)
			if tt.expectSuccess && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expectSuccess {
				t.Errorf("Expected %v, got %v for %d total and %d faulty nodes",
					tt.expectSuccess, result, tt.totalNodes, tt.faultyNodes)
			}
		})
	}
}

func TestPCRIntegrityChecks(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	nodeID := "test-node-004"

	attestation, err := manager.GenerateAttestation(nodeID, []byte("nonce-4"))
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}

	if len(attestation.PCRValues) == 0 {
		t.Fatal("Expected PCR values")
	}

	expectedPCRs := []int{0, 1, 7}
	for _, pcr := range expectedPCRs {
		if _, exists := attestation.PCRValues[pcr]; !exists {
			t.Errorf("Expected PCR %d to be present", pcr)
		}
	}
}

func TestConcurrentAttestations(t *testing.T) {
	manager := NewAttestationManager(100, time.Minute, true)
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			nodeID := fmt.Sprintf("node-%d", id)
			_, err := manager.GenerateAttestation(nodeID, []byte("nonce"))
			if err != nil {
				t.Errorf("Failed concurrent attestation: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func TestAttestationReportStructure(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	attestation, err := manager.GenerateAttestation("test-node-006", []byte("nonce-6"))
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}

	if attestation.NodeID == "" {
		t.Error("Expected non-empty NodeID")
	}

	if attestation.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if attestation.Signature == nil {
		t.Error("Expected non-nil signature")
	}
}

func TestVerifyRejectsReplay(t *testing.T) {
	nodeID := "replay-node"
	quote, err := generateTPMQuoteForNode(nodeID)
	if err != nil {
		t.Fatalf("failed to generate quote: %v", err)
	}

	if err := Verify(nodeID, quote); err != nil {
		t.Fatalf("first verify failed unexpectedly: %v", err)
	}

	if err := Verify(nodeID, quote); err == nil {
		t.Fatal("expected replay verify to fail")
	}
}

func TestGenerateAttestationRejectsEmptyNodeID(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	if _, err := manager.GenerateAttestation("", []byte("nonce")); err == nil {
		t.Fatal("expected empty node ID to fail")
	}
}

func TestVerifyAttestationRejectsNil(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	if _, err := manager.VerifyAttestation(nil); err == nil {
		t.Fatal("expected nil attestation to fail")
	}
}

// --- TPMProofVerifier tests ---

// TestTPMProofVerifierTEEProofSucceeds verifies that a "tee" proof type with an
// enabled manager returns verified=true and full confidence (9500 bps).
func TestTPMProofVerifierTEEProofSucceeds(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	verifier := NewTPMProofVerifier(manager)

	req := blockchain.ProofVerificationRequest{
		RoundID:   "round-tee-1",
		ProofType: "tee",
		ProofHash: "abc123",
		Payload:   map[string]interface{}{"node_id": "node-tee-1"},
	}

	ok, confidence, err := verifier.VerifyProof(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected TEE proof to pass with enabled manager")
	}
	if confidence != 9500 {
		t.Errorf("expected confidence 9500, got %d", confidence)
	}
}

// TestTPMProofVerifierTPMProofAlias verifies the "tpm" alias behaves identically to "tee".
func TestTPMProofVerifierTPMProofAlias(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	verifier := NewTPMProofVerifier(manager)

	req := blockchain.ProofVerificationRequest{
		RoundID:   "round-tpm-1",
		ProofType: "tpm",
		Payload:   map[string]interface{}{"node_id": "node-tpm-1"},
	}

	ok, confidence, err := verifier.VerifyProof(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected 'tpm' proof type to pass")
	}
	if confidence != 9500 {
		t.Errorf("expected confidence 9500, got %d", confidence)
	}
}

// TestTPMProofVerifierDisabledManagerSoftFail verifies that a disabled AttestationManager
// causes the verifier to return (false, 2500, nil) since GenerateAttestation errors.
func TestTPMProofVerifierDisabledManagerSoftFail(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, false) // disabled
	verifier := NewTPMProofVerifier(manager)

	req := blockchain.ProofVerificationRequest{
		RoundID:   "round-disabled-1",
		ProofType: "tee",
		Payload:   map[string]interface{}{"node_id": "node-disabled"},
	}

	ok, confidence, err := verifier.VerifyProof(req)
	if err != nil {
		t.Fatalf("unexpected hard error from disabled manager: %v", err)
	}
	if ok {
		t.Fatal("expected disabled manager to produce a failed verification")
	}
	if confidence != 2500 {
		t.Errorf("expected soft-fail confidence 2500, got %d", confidence)
	}
}

// TestTPMProofVerifierConsensusPassthrough verifies non-TEE proof types get permissive defaults.
func TestTPMProofVerifierConsensusPassthrough(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	verifier := NewTPMProofVerifier(manager)

	// With no proof data → base permissive confidence
	req := blockchain.ProofVerificationRequest{
		RoundID:   "round-consensus-1",
		ProofType: "consensus",
	}
	ok, confidence, err := verifier.VerifyProof(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected consensus proof to pass through")
	}
	if confidence != 6500 {
		t.Errorf("expected permissive confidence 6500, got %d", confidence)
	}

	// With proof hash → slightly higher permissive confidence
	req.ProofHash = "somehash"
	ok2, confidence2, err2 := verifier.VerifyProof(req)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if !ok2 {
		t.Fatal("expected consensus proof with hash to pass through")
	}
	if confidence2 != 9000 {
		t.Errorf("expected permissive confidence 9000, got %d", confidence2)
	}
}

// TestNewTPMProofVerifierPanicsOnNil verifies the constructor panics on nil manager.
func TestNewTPMProofVerifierPanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when passing nil manager to NewTPMProofVerifier")
		}
	}()
	NewTPMProofVerifier(nil)
}

func TestTPMProofVerifierRoundScopedNoncePreferred(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	verifier := &TPMProofVerifier{manager: manager, roundScopedNonce: true}

	req := blockchain.ProofVerificationRequest{
		RoundID:   "round-42",
		ProofHash: "per-request-proof-hash",
	}

	nonce := verifier.nonceForRequest(req)
	if string(nonce) != "round:round-42" {
		t.Fatalf("expected round-scoped nonce, got %q", string(nonce))
	}
}

func TestTPMProofVerifierLegacyNonceUsesProofHash(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	verifier := &TPMProofVerifier{manager: manager, roundScopedNonce: false}

	req := blockchain.ProofVerificationRequest{
		RoundID:   "round-legacy",
		ProofHash: "proof-hash-legacy",
	}

	nonce := verifier.nonceForRequest(req)
	if string(nonce) != "proof-hash-legacy" {
		t.Fatalf("expected proof-hash nonce in legacy mode, got %q", string(nonce))
	}
}

func TestVerifyAttestationLatencySpikeEvictsCache(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	manager.SetLatencySpikeThreshold(1 * time.Nanosecond)

	report, err := manager.GenerateAttestation("spike-node", []byte("nonce-spike"))
	if err != nil {
		t.Fatalf("failed to generate attestation: %v", err)
	}

	valid, err := manager.VerifyAttestation(report)
	if err != nil {
		t.Fatalf("failed to verify attestation: %v", err)
	}
	if !valid {
		t.Fatal("expected valid attestation")
	}
	if manager.LatencySpikeCount() == 0 {
		t.Fatal("expected at least one latency spike to be recorded")
	}
	if cached := manager.attestationCache.Get(report.AttestationID); cached != nil {
		t.Fatal("expected spike circuit breaker to evict cache entry")
	}
}

func BenchmarkTPMProofVerifierNonceModes(b *testing.B) {
	manager := NewAttestationManager(256, 30*time.Second, true)

	request := blockchain.ProofVerificationRequest{
		RoundID:   "benchmark-round",
		ProofType: "tee",
		ProofHash: "benchmark-proof-hash",
		Payload:   map[string]interface{}{"node_id": "benchmark-node"},
	}

	b.Run("round_scoped", func(b *testing.B) {
		verifier := &TPMProofVerifier{manager: manager, roundScopedNonce: true}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ok, _, err := verifier.VerifyProof(request)
			if err != nil || !ok {
				b.Fatalf("unexpected verify failure: ok=%v err=%v", ok, err)
			}
		}
	})

	b.Run("legacy_proofhash", func(b *testing.B) {
		verifier := &TPMProofVerifier{manager: manager, roundScopedNonce: false}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := request
			req.ProofHash = fmt.Sprintf("benchmark-proof-hash-%d", i)
			ok, _, err := verifier.VerifyProof(req)
			if err != nil || !ok {
				b.Fatalf("unexpected verify failure: ok=%v err=%v", ok, err)
			}
		}
	})
}

func BenchmarkTPMProofVerifierRoundScopedParallel(b *testing.B) {
	manager := NewAttestationManager(1024, 30*time.Second, true)
	verifier := &TPMProofVerifier{manager: manager, roundScopedNonce: true}

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			request := blockchain.ProofVerificationRequest{
				RoundID:   "parallel-round",
				ProofType: "tee",
				ProofHash: "parallel-proof-hash",
				Payload:   map[string]interface{}{"node_id": "parallel-node"},
			}
			ok, _, err := verifier.VerifyProof(request)
			if err != nil || !ok {
				b.Fatalf("unexpected verify failure: ok=%v err=%v", ok, err)
			}
		}
	})
}

func TestTPMProofVerifierDefaultRoundScopedSetting(t *testing.T) {
	t.Setenv("TPM_ROUND_SCOPED_NONCE", "true")
	manager := NewAttestationManager(10, time.Minute, true)
	verifier := NewTPMProofVerifier(manager)
	if !verifier.roundScopedNonce {
		t.Fatal("expected round-scoped nonce to be enabled by default")
	}

	t.Setenv("TPM_ROUND_SCOPED_NONCE", "false")
	legacyVerifier := NewTPMProofVerifier(manager)
	if legacyVerifier.roundScopedNonce {
		t.Fatal("expected round-scoped nonce to be disabled when env is false")
	}

	t.Setenv("TPM_ROUND_SCOPED_NONCE", strings.ToUpper("true"))
	upperVerifier := NewTPMProofVerifier(manager)
	if !upperVerifier.roundScopedNonce {
		t.Fatal("expected case-insensitive true handling for round-scoped nonce env")
	}
}

func TestPrewarmVerifiedQuotesCachesEntries(t *testing.T) {
	nodes := []string{"prewarm-a", "prewarm-b", "prewarm-a", " "}
	warmed := PrewarmVerifiedQuotes(nodes)
	if warmed < 2 {
		t.Fatalf("expected at least 2 nodes to be prewarmed, got %d", warmed)
	}
	if !isQuoteCached("prewarm-a") {
		t.Fatal("expected prewarm-a to be present in quote cache")
	}
	if !isQuoteCached("prewarm-b") {
		t.Fatal("expected prewarm-b to be present in quote cache")
	}
}

func TestTPMProofVerifierRoundStartPrewarm(t *testing.T) {
	t.Setenv("TPM_ROUND_PREWARM", "true")
	t.Setenv("TPM_PREWARM_NODE_IDS", "node-x,node-y")

	manager := NewAttestationManager(10, time.Minute, true)
	verifier := NewTPMProofVerifier(manager)

	request := blockchain.ProofVerificationRequest{
		RoundID:   "prewarm-round-1",
		ProofType: "tee",
		ProofHash: "proofhash",
		Payload:   map[string]interface{}{"node_id": "node-z"},
	}

	ok, _, err := verifier.VerifyProof(request)
	if err != nil || !ok {
		t.Fatalf("unexpected verify failure: ok=%v err=%v", ok, err)
	}

	if !isQuoteCached("node-z") {
		t.Fatal("expected current node quote to be prewarmed")
	}
	if !isQuoteCached("node-x") || !isQuoteCached("node-y") {
		t.Fatal("expected configured prewarm nodes to be cached")
	}
}

func TestNonceCleanupRemovesExpiredEntries(t *testing.T) {
	key := "cleanup-node:nonce"
	shard := nonceShard(key)
	shard.mu.Lock()
	shard.entries[key] = time.Now().Add(-1 * time.Second)
	shard.mu.Unlock()

	cleanupExpiredNonceEntries()

	shard.mu.Lock()
	_, exists := shard.entries[key]
	shard.mu.Unlock()
	if exists {
		t.Fatal("expected expired nonce entry to be removed by cleanup")
	}
}

func TestTPMMetricsTrackCacheAndCleanup(t *testing.T) {
	before := tpmMetricsSnapshot()

	nodeID := fmt.Sprintf("metrics-node-%d", time.Now().UnixNano())
	q1, err := GetVerifiedQuote(nodeID)
	if err != nil {
		t.Fatalf("first quote retrieval failed: %v", err)
	}
	if _, err := GetVerifiedQuote(nodeID); err != nil {
		t.Fatalf("second quote retrieval failed: %v", err)
	}

	if err := Verify(nodeID, q1); err != nil {
		t.Fatalf("first verify failed unexpectedly: %v", err)
	}
	if err := Verify(nodeID, q1); err == nil {
		t.Fatal("expected replay verify to fail")
	}

	manager := NewAttestationManager(10, time.Minute, true)
	manager.SetLatencySpikeThreshold(time.Hour)
	report, err := manager.GenerateAttestation(nodeID, []byte("metrics-nonce"))
	if err != nil {
		t.Fatalf("failed to generate attestation: %v", err)
	}
	if ok, err := manager.VerifyAttestation(report); err != nil || !ok {
		t.Fatalf("first attestation verify failed: ok=%v err=%v", ok, err)
	}
	if ok, err := manager.VerifyAttestation(report); err != nil || !ok {
		t.Fatalf("second attestation verify failed: ok=%v err=%v", ok, err)
	}

	cleanupKey := fmt.Sprintf("metrics-cleanup:%d", time.Now().UnixNano())
	cleanupShard := nonceShard(cleanupKey)
	cleanupShard.mu.Lock()
	cleanupShard.entries[cleanupKey] = time.Now().Add(-1 * time.Second)
	cleanupShard.mu.Unlock()
	cleanupExpiredNonceEntries()

	after := tpmMetricsSnapshot()

	if after.quoteCacheMisses <= before.quoteCacheMisses {
		t.Fatalf("expected quote cache miss metric to increase: before=%d after=%d", before.quoteCacheMisses, after.quoteCacheMisses)
	}
	if after.quoteCacheHits <= before.quoteCacheHits {
		t.Fatalf("expected quote cache hit metric to increase: before=%d after=%d", before.quoteCacheHits, after.quoteCacheHits)
	}
	if after.attestationCacheMisses <= before.attestationCacheMisses {
		t.Fatalf("expected attestation cache miss metric to increase: before=%d after=%d", before.attestationCacheMisses, after.attestationCacheMisses)
	}
	if after.attestationCacheHits <= before.attestationCacheHits {
		t.Fatalf("expected attestation cache hit metric to increase: before=%d after=%d", before.attestationCacheHits, after.attestationCacheHits)
	}
	if after.nonceReplayRejects <= before.nonceReplayRejects {
		t.Fatalf("expected nonce replay rejection metric to increase: before=%d after=%d", before.nonceReplayRejects, after.nonceReplayRejects)
	}
	if after.nonceCleanupRuns <= before.nonceCleanupRuns {
		t.Fatalf("expected nonce cleanup run metric to increase: before=%d after=%d", before.nonceCleanupRuns, after.nonceCleanupRuns)
	}
	if after.nonceCleanupRemoved <= before.nonceCleanupRemoved {
		t.Fatalf("expected nonce cleanup removed metric to increase: before=%d after=%d", before.nonceCleanupRemoved, after.nonceCleanupRemoved)
	}
}

func TestTPMMetricsRegistered(t *testing.T) {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	expected := map[string]bool{
		"mohawk_tpm_cache_events_total":            false,
		"mohawk_tpm_lock_wait_seconds":             false,
		"mohawk_tpm_nonce_cleanup_runs_total":      false,
		"mohawk_tpm_nonce_cleanup_removed_total":   false,
		"mohawk_tpm_prewarm_requests_total":        false,
		"mohawk_tpm_prewarm_warmed_nodes_total":    false,
		"mohawk_tpm_nonce_replay_rejections_total": false,
	}

	for _, family := range families {
		if _, ok := expected[family.GetName()]; ok {
			expected[family.GetName()] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Fatalf("expected metric %s to be registered", name)
		}
	}
}
