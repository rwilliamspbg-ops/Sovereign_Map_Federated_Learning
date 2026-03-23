// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package api

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	internalproof "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/convergence"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/hybrid"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/island"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/monitoring"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/p2p"
)

type proofVerifyRequest struct {
	Proof       string `json:"proof"`
	Encoding    string `json:"encoding,omitempty"`
	PublicInput string `json:"public_input,omitempty"`
}

type hybridVerifyRequest struct {
	Mode         string `json:"mode,omitempty"`
	STARKBackend string `json:"stark_backend,omitempty"`
	SNARKProof   string `json:"snark_proof"`
	STARKProof   string `json:"stark_proof"`
	Encoding     string `json:"encoding,omitempty"`
}

type verificationPolicyRequest struct {
	RequireProof                bool   `json:"require_proof"`
	MinConfidenceBps            uint32 `json:"min_confidence_bps"`
	RejectOnVerificationFailure bool   `json:"reject_on_verification_failure"`
	AllowConsensusProof         bool   `json:"allow_consensus_proof"`
	AllowZKProof                bool   `json:"allow_zk_proof"`
	AllowTEEProof               bool   `json:"allow_tee_proof"`
}

// ConsensusStatusReader exposes consensus status snapshots for API responses.
type ConsensusStatusReader interface {
	GetRuntimeStatus() map[string]interface{}
}

// AggregationStatusReader exposes aggregation runtime snapshots for API responses.
type AggregationStatusReader interface {
	GetRuntimeStatus() map[string]interface{}
}

// Handler provides HTTP endpoints for the federated learning system
type Handler struct {
	convergence       *convergence.Detector
	island            *island.Manager
	metrics           *monitoring.Collector
	p2pNetwork        *p2p.Network
	ledger            *ProofLedger
	blockchain        *blockchain.BlockChain
	consensusReader   ConsensusStatusReader
	aggregationReader AggregationStatusReader
}

func writeJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-API-Version", "v1")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func ensureGetMethod(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return false
	}
	return true
}

func ensurePostMethod(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return false
	}
	return true
}

func decodePayload(payload string, encoding string) ([]byte, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return nil, fmt.Errorf("payload is required")
	}

	enc := strings.ToLower(strings.TrimSpace(encoding))
	if enc == "" {
		if strings.HasPrefix(trimmed, "0x") {
			enc = "hex"
		} else {
			enc = "base64"
		}
	}

	switch enc {
	case "hex":
		hexPayload := strings.TrimPrefix(trimmed, "0x")
		data, err := hex.DecodeString(hexPayload)
		if err != nil {
			return nil, fmt.Errorf("invalid hex payload")
		}
		return data, nil
	case "base64":
		data, err := base64.StdEncoding.DecodeString(trimmed)
		if err == nil {
			return data, nil
		}
		data, err = base64.URLEncoding.DecodeString(trimmed)
		if err == nil {
			return data, nil
		}
		return nil, fmt.Errorf("invalid base64 payload")
	case "raw":
		return []byte(trimmed), nil
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", enc)
	}
}

func boolEnv(name string, defaultValue bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if raw == "" {
		return defaultValue
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func parseRoleList(raw string) map[string]struct{} {
	roles := map[string]struct{}{}
	for _, part := range strings.Split(raw, ",") {
		role := strings.ToLower(strings.TrimSpace(part))
		if role == "" {
			continue
		}
		roles[role] = struct{}{}
	}
	return roles
}

func extractRequestToken(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[len("Bearer "):])
	}
	return strings.TrimSpace(r.Header.Get("X-API-Token"))
}

func loadExpectedToken() (string, error) {
	tokenFile := strings.TrimSpace(os.Getenv("MOHAWK_API_TOKEN_FILE"))
	if tokenFile == "" {
		tokenFile = "/run/secrets/mohawk_api_token"
	}

	cleanTokenFile, err := sanitizeReadablePath(tokenFile)
	if err != nil {
		return "", err
	}

	raw, err := os.ReadFile(cleanTokenFile) // #nosec G304,G703 -- path is sanitized before file read
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %w", err)
	}
	token := strings.TrimSpace(string(raw))
	if token == "" {
		return "", fmt.Errorf("token file is empty")
	}
	return token, nil
}

func requireScopedAuth(w http.ResponseWriter, r *http.Request, envName string, defaultRoles string) bool {
	authMode := strings.ToLower(strings.TrimSpace(os.Getenv("MOHAWK_API_AUTH_MODE")))
	if authMode == "" {
		authMode = "file-only"
	}

	if authMode == "off" || authMode == "disabled" || authMode == "none" {
		return true
	}

	expectedToken, err := loadExpectedToken()
	if err != nil {
		http.Error(w, "auth configuration error", http.StatusInternalServerError)
		return false
	}

	receivedToken := extractRequestToken(r)
	if receivedToken == "" {
		http.Error(w, "missing api token", http.StatusUnauthorized)
		return false
	}

	if subtle.ConstantTimeCompare([]byte(expectedToken), []byte(receivedToken)) != 1 {
		http.Error(w, "invalid api token", http.StatusUnauthorized)
		return false
	}

	if !boolEnv("MOHAWK_API_ENFORCE_ROLES", true) {
		return true
	}

	allowedRolesRaw := strings.TrimSpace(os.Getenv(envName))
	if allowedRolesRaw == "" {
		allowedRolesRaw = defaultRoles
	}

	allowedRoles := parseRoleList(allowedRolesRaw)
	requestRole := strings.ToLower(strings.TrimSpace(r.Header.Get("X-API-Role")))
	if requestRole == "" {
		http.Error(w, "missing api role", http.StatusForbidden)
		return false
	}
	if _, ok := allowedRoles[requestRole]; !ok {
		http.Error(w, "role not allowed", http.StatusForbidden)
		return false
	}

	return true
}

func requireProofAuth(w http.ResponseWriter, r *http.Request) bool {
	allowedRolesRaw := strings.TrimSpace(os.Getenv("MOHAWK_API_PROOF_ALLOWED_ROLES"))
	if allowedRolesRaw == "" {
		allowedRolesRaw = strings.TrimSpace(os.Getenv("MOHAWK_API_HYBRID_ALLOWED_ROLES"))
	}
	if allowedRolesRaw == "" {
		allowedRolesRaw = "verifier,admin"
	}
	return requireScopedAuth(w, r, "MOHAWK_API_PROOF_ALLOWED_ROLES", allowedRolesRaw)
}

func requirePolicyAuth(w http.ResponseWriter, r *http.Request) bool {
	return requireScopedAuth(w, r, "MOHAWK_API_POLICY_ALLOWED_ROLES", "admin")
}

// NewHandler creates a new API handler with integrated backends
func NewHandler(detector *convergence.Detector, islandMgr *island.Manager, collector *monitoring.Collector, network *p2p.Network) *Handler {
	return &Handler{
		convergence: detector,
		island:      islandMgr,
		metrics:     collector,
		p2pNetwork:  network,
		ledger:      NewProofLedger(0),
	}
}

// SetBlockchain attaches an optional blockchain backend used by status endpoints.
func (h *Handler) SetBlockchain(chain *blockchain.BlockChain) {
	h.blockchain = chain
}

// SetConsensusReaders attaches optional consensus and aggregation status readers.
func (h *Handler) SetConsensusReaders(consensusReader ConsensusStatusReader, aggregationReader AggregationStatusReader) {
	h.consensusReader = consensusReader
	h.aggregationReader = aggregationReader
}

// RegisterRoutes sets up HTTP routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Legacy + current endpoints
	mux.HandleFunc("/health", h.HealthCheck)
	mux.HandleFunc("/api/status", h.GetStatus)
	mux.HandleFunc("/api/metrics", h.GetMetrics)
	mux.HandleFunc("/api/convergence", h.GetConvergence)
	mux.HandleFunc("/api/convergence_status", h.GetConvergence)
	mux.HandleFunc("/api/island/status", h.GetIslandStatus)
	mux.HandleFunc("/api/peers", h.GetPeers)
	mux.HandleFunc("/api/network_status", h.GetNetworkStatus)
	mux.HandleFunc("/api/trust_status", h.GetTrustStatus)
	mux.HandleFunc("/api/trust_snapshot", h.GetTrustSnapshot)
	mux.HandleFunc("/api/consensus/status", h.GetConsensusStatus)

	// Versioned aliases
	mux.HandleFunc("/api/v1/status", h.GetStatus)
	mux.HandleFunc("/api/v1/metrics", h.GetMetrics)
	mux.HandleFunc("/api/v1/convergence", h.GetConvergence)
	mux.HandleFunc("/api/v1/island/status", h.GetIslandStatus)
	mux.HandleFunc("/api/v1/peers", h.GetPeers)
	mux.HandleFunc("/api/v1/network_status", h.GetNetworkStatus)
	mux.HandleFunc("/api/v1/trust_status", h.GetTrustStatus)
	mux.HandleFunc("/api/v1/trust_snapshot", h.GetTrustSnapshot)
	mux.HandleFunc("/api/v1/consensus/status", h.GetConsensusStatus)
	mux.HandleFunc("/api/proof/verify", h.VerifyProof)
	mux.HandleFunc("/api/v1/proof/verify", h.VerifyProof)
	mux.HandleFunc("/api/proof/hybrid/verify", h.VerifyHybridProof)
	mux.HandleFunc("/api/v1/proof/hybrid/verify", h.VerifyHybridProof)
	mux.HandleFunc("/api/capabilities", h.GetCapabilities)
	mux.HandleFunc("/api/v1/capabilities", h.GetCapabilities)
	mux.HandleFunc("/api/ledger", h.GetLedger)
	mux.HandleFunc("/api/v1/ledger", h.GetLedger)
	mux.HandleFunc("/api/verification_policy", h.HandleVerificationPolicy)
	mux.HandleFunc("/api/v1/verification_policy", h.HandleVerificationPolicy)
}

// HealthCheck returns basic health status
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]string{
		"status":  "healthy",
		"service": "sovereign-map-fl",
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, response)
}

// GetStatus returns overall system status
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]interface{}{
		"status":  "operational",
		"version": "0.1.0",
		"components": map[string]string{
			"aggregator":  "ready",
			"convergence": "monitoring",
			"island_mode": "online",
			"tpm":         "initialized",
			"consensus":   "active",
		},
	}

	if h.consensusReader != nil {
		response["consensus"] = h.consensusReader.GetRuntimeStatus()
	}
	if h.aggregationReader != nil {
		response["aggregation"] = h.aggregationReader.GetRuntimeStatus()
	}

	writeJSON(w, response)
}

// GetMetrics returns collected metrics
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]interface{}{
		"total_rounds":                0,
		"active_nodes":                0,
		"convergence_rate":            0.0,
		"network_lag_ms":              0,
		"churn_joins_total":           0,
		"churn_leaves_total":          0,
		"async_staleness_avg_seconds": 0.0,
	}

	if h.metrics != nil {
		summary := h.metrics.GetSummary()
		response["total_metrics"] = summary["total_metrics"]
		response["aggregations"] = summary["aggregations"]

		// Get specific metric aggregations
		if agg := h.metrics.GetAggregation(monitoring.MetricPeerCount); agg != nil {
			response["active_nodes"] = int(agg.Mean)
		}
		if agg := h.metrics.GetAggregation(monitoring.MetricNetworkLag); agg != nil {
			response["network_lag_ms"] = agg.Mean
		}
		if agg := h.metrics.GetAggregation(monitoring.MetricNodeJoin); agg != nil {
			response["churn_joins_total"] = agg.Sum
		}
		if agg := h.metrics.GetAggregation(monitoring.MetricNodeLeave); agg != nil {
			response["churn_leaves_total"] = agg.Sum
		}
		if agg := h.metrics.GetAggregation(monitoring.MetricStaleness); agg != nil {
			response["async_staleness_avg_seconds"] = agg.Mean
		}
	}

	writeJSON(w, response)
}

// GetConsensusStatus returns consensus and aggregation runtime status.
func (h *Handler) GetConsensusStatus(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]interface{}{
		"consensus": map[string]interface{}{
			"available": false,
		},
		"aggregation": map[string]interface{}{
			"available": false,
		},
	}

	if h.consensusReader != nil {
		status := h.consensusReader.GetRuntimeStatus()
		status["available"] = true
		response["consensus"] = status
	}
	if h.aggregationReader != nil {
		status := h.aggregationReader.GetRuntimeStatus()
		status["available"] = true
		response["aggregation"] = status
	}

	writeJSON(w, response)
}

// GetConvergence returns convergence status
func (h *Handler) GetConvergence(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]interface{}{
		"converged":           false,
		"iterations":          0,
		"convergence_rate":    0.0,
		"heterogeneity":       0.0,
		"threshold":           0.001,
		"effective_threshold": 0.001,
	}

	if h.convergence != nil {
		metrics := h.convergence.GetMetrics()
		converged := h.convergence.IsConverged()

		response["converged"] = converged
		response["iterations"] = metrics["iterations"]
		response["convergence_rate"] = metrics["convergence_rate"]
		response["heterogeneity"] = metrics["heterogeneity"]
		response["threshold"] = metrics["threshold"]
		response["effective_threshold"] = metrics["effective_threshold"]
	}

	writeJSON(w, response)
}

// GetIslandStatus returns Island Mode status
func (h *Handler) GetIslandStatus(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]interface{}{
		"mode":                 "online",
		"cached_updates":       0,
		"max_cached_updates":   100,
		"time_since_last_sync": "0s",
	}

	if h.island != nil {
		mode := h.island.CurrentMode()
		modeStr := "online"
		switch mode {
		case island.ModeIsland:
			modeStr = "island"
		case island.ModeTransition:
			modeStr = "transition"
		}

		cachedCount, maxCached := h.island.GetCachedUpdateStats()
		lastSync := h.island.GetLastSyncTime()
		timeSinceSync := time.Since(lastSync)

		response["mode"] = modeStr
		response["cached_updates"] = cachedCount
		response["max_cached_updates"] = maxCached
		response["time_since_last_sync"] = timeSinceSync.String()
	}

	writeJSON(w, response)
}

// GetPeers returns information about connected peers
func (h *Handler) GetPeers(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]interface{}{
		"total_peers":  0,
		"active_peers": 0,
		"peers":        []map[string]interface{}{},
	}

	if h.p2pNetwork != nil {
		peers := h.p2pNetwork.GetPeers()
		activePeers := h.p2pNetwork.GetActivePeerCount()

		response["total_peers"] = len(peers)
		response["active_peers"] = activePeers
		response["peers"] = peers
	}

	writeJSON(w, response)
}

// GetNetworkStatus returns concise network health details
func (h *Handler) GetNetworkStatus(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]interface{}{
		"network":      "ready",
		"node_id":      "unknown",
		"total_peers":  0,
		"active_nodes": 0,
	}

	if h.p2pNetwork != nil {
		response["node_id"] = h.p2pNetwork.GetNodeID()
		peers := h.p2pNetwork.GetPeers()
		response["total_peers"] = len(peers)
		response["active_nodes"] = h.p2pNetwork.GetActivePeerCount()
	}

	writeJSON(w, response)
}

// GetTrustStatus returns trust-related network summary metrics
func (h *Handler) GetTrustStatus(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}
	writeJSON(w, h.buildTrustStatusPayload())
}

func (h *Handler) GetTrustSnapshot(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]interface{}{
		"trust_status":   h.buildTrustStatusPayload(),
		"policy_history": h.getVerificationPolicyHistory(),
	}
	writeJSON(w, response)
}

func (h *Handler) buildTrustStatusPayload() map[string]interface{} {
	response := map[string]interface{}{
		"trust_mode":         "p2p-reputation",
		"total_peers":        0,
		"active_peers":       0,
		"average_reputation": 0.0,
	}

	if h.p2pNetwork != nil {
		peers := h.p2pNetwork.GetPeers()
		activePeers := h.p2pNetwork.GetActivePeerCount()

		totalReputation := 0.0
		for _, peer := range peers {
			if rep, ok := peer["reputation"].(float64); ok {
				totalReputation += rep
			}
		}

		avgReputation := 0.0
		if len(peers) > 0 {
			avgReputation = totalReputation / float64(len(peers))
		}

		response["total_peers"] = len(peers)
		response["active_peers"] = activePeers
		response["average_reputation"] = avgReputation
	}

	if h.blockchain != nil {
		policy := h.blockchain.GetVerificationPolicy()
		verification := h.blockchain.GetFLVerificationMetrics()

		response["trust_mode"] = "p2p-reputation+governed-proof-verification"
		response["verification_policy"] = verificationPolicyPayload(policy)
		response["fl_verification"] = flVerificationPayload(verification)
	}

	return response
}

func (h *Handler) getVerificationPolicyHistory() []map[string]interface{} {
	if h.blockchain == nil {
		return []map[string]interface{}{}
	}

	entries := h.blockchain.StateDB.GetAll()
	history := make([]map[string]interface{}, 0)
	for key, value := range entries {
		if !strings.HasPrefix(key, "governance_verification_audit:") && !strings.HasPrefix(key, "api_verification_policy_audit:") {
			continue
		}
		entry, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		copied := make(map[string]interface{}, len(entry)+1)
		for k, v := range entry {
			copied[k] = v
		}
		copied["audit_key"] = key
		history = append(history, copied)
	}

	sort.Slice(history, func(i, j int) bool {
		return auditTimestamp(history[i]) > auditTimestamp(history[j])
	})

	return history
}

func auditTimestamp(entry map[string]interface{}) int64 {
	switch ts := entry["timestamp"].(type) {
	case int64:
		return ts
	case int:
		return int64(ts)
	case uint64:
		if ts <= math.MaxInt64 {
			return int64(ts)
		}
	case float64:
		if ts >= math.MinInt64 && ts <= math.MaxInt64 {
			return int64(ts)
		}
	default:
		return 0
	}
	return 0
}

func verificationPolicyPayload(policy blockchain.VerificationPolicy) map[string]interface{} {
	return map[string]interface{}{
		"require_proof":                  policy.RequireProof,
		"min_confidence_bps":             policy.MinConfidenceBps,
		"reject_on_verification_failure": policy.RejectOnVerificationFailure,
		"allow_consensus_proof":          policy.AllowConsensusProof,
		"allow_zk_proof":                 policy.AllowZKProof,
		"allow_tee_proof":                policy.AllowTEEProof,
	}
}

func flVerificationPayload(verification blockchain.FLVerificationMetrics) map[string]interface{} {
	return map[string]interface{}{
		"total_rounds":           verification.TotalRounds,
		"verified_rounds":        verification.VerifiedRounds,
		"failed_rounds":          verification.FailedRounds,
		"verified_ratio":         verification.VerifiedRatio,
		"average_confidence_bps": verification.AverageConfidenceBps,
		"last_round_id":          verification.LastRoundID,
		"last_proof_type":        verification.LastProofType,
	}
}

func (h *Handler) HandleVerificationPolicy(w http.ResponseWriter, r *http.Request) {
	if h.blockchain == nil {
		http.Error(w, "blockchain backend unavailable", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, map[string]interface{}{
			"verification_policy": verificationPolicyPayload(h.blockchain.GetVerificationPolicy()),
			"fl_verification":     flVerificationPayload(h.blockchain.GetFLVerificationMetrics()),
		})
		return
	case http.MethodPost:
		if !requirePolicyAuth(w, r) {
			return
		}
	default:
		methodNotAllowed(w)
		return
	}

	var req verificationPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	policy := blockchain.VerificationPolicy{
		RequireProof:                req.RequireProof,
		MinConfidenceBps:            req.MinConfidenceBps,
		RejectOnVerificationFailure: req.RejectOnVerificationFailure,
		AllowConsensusProof:         req.AllowConsensusProof,
		AllowZKProof:                req.AllowZKProof,
		AllowTEEProof:               req.AllowTEEProof,
	}
	if err := h.blockchain.SetVerificationPolicy(policy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.blockchain.StateDB.Set("verification_policy:active", verificationPolicyPayload(policy)); err != nil {
		http.Error(w, "failed to persist verification policy", http.StatusInternalServerError)
		return
	}

	ts := time.Now().Unix()
	auditKey := fmt.Sprintf("api_verification_policy_audit:%d", ts)
	_ = h.blockchain.StateDB.Set(auditKey, map[string]interface{}{
		"proposal_id": "api-direct",
		"action":      "set_verification_policy",
		"source":      "api",
		"actor_role":  strings.ToLower(strings.TrimSpace(r.Header.Get("X-API-Role"))),
		"timestamp":   ts,
		"new_policy":  verificationPolicyPayload(policy),
	})

	writeJSON(w, map[string]interface{}{
		"status":              "updated",
		"verification_policy": verificationPolicyPayload(policy),
		"fl_verification":     flVerificationPayload(h.blockchain.GetFLVerificationMetrics()),
		"policy_history":      h.getVerificationPolicyHistory(),
	})
}

func (h *Handler) VerifyProof(w http.ResponseWriter, r *http.Request) {
	if !ensurePostMethod(w, r) {
		return
	}
	if !requireProofAuth(w, r) {
		return
	}

	var req proofVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	proofBytes, err := decodePayload(req.Proof, req.Encoding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	inputBytes := []byte(req.PublicInput)
	started := time.Now()
	ok, verifyErr := internalproof.VerifyProof(proofBytes, inputBytes)
	latencyDur := time.Since(started)
	latency := latencyDur.Milliseconds()
	observeProofVerification("snark", "groth16_bn254", ok, latencyDur)

	role := strings.ToLower(strings.TrimSpace(r.Header.Get("X-API-Role")))
	h.ledger.Record("snark_verify", proofBytes, role, ok, latency, verifyErr)
	observeLedgerEvent("snark_verify", h.ledger.Len())

	response := map[string]interface{}{
		"valid":      ok,
		"latency_ms": latency,
	}
	if verifyErr != nil {
		response["error"] = verifyErr.Error()
	}

	writeJSON(w, response)
}

func (h *Handler) VerifyHybridProof(w http.ResponseWriter, r *http.Request) {
	if !ensurePostMethod(w, r) {
		return
	}
	if !requireProofAuth(w, r) {
		return
	}

	var req hybridVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	snarkBytes, err := decodePayload(req.SNARKProof, req.Encoding)
	if err != nil {
		http.Error(w, "invalid snark_proof: "+err.Error(), http.StatusBadRequest)
		return
	}

	starkBytes, err := decodePayload(req.STARKProof, req.Encoding)
	if err != nil {
		http.Error(w, "invalid stark_proof: "+err.Error(), http.StatusBadRequest)
		return
	}

	mode := hybrid.HybridMode(req.Mode)
	if mode == "" {
		mode = hybrid.ModePreferSNARK
	}

	started := time.Now()
	result, verifyErr := hybrid.VerifyHybrid(hybrid.VerifyRequest{
		Mode:         mode,
		SNARKProof:   snarkBytes,
		STARKProof:   starkBytes,
		STARKBackend: req.STARKBackend,
	})
	latencyDur := time.Since(started)

	combinedProof := append(snarkBytes, starkBytes...)
	hyRole := strings.ToLower(strings.TrimSpace(r.Header.Get("X-API-Role")))
	h.ledger.Record("hybrid_verify", combinedProof, hyRole, result.Accepted, latencyDur.Milliseconds(), verifyErr)
	observeProofVerification("hybrid", result.STARKBackend, result.Accepted, latencyDur)
	observeLedgerEvent("hybrid_verify", h.ledger.Len())

	response := map[string]interface{}{
		"accepted": result.Accepted,
		"policy":   result.Policy,
		"snark":    result.SNARKValid,
		"stark":    result.STARKValid,
		"backend":  result.STARKBackend,
	}
	if verifyErr != nil {
		response["error"] = verifyErr.Error()
	}

	writeJSON(w, response)
}

// GetLedger returns the proof verification ledger (auth-gated).
func (h *Handler) GetLedger(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}
	if !requireProofAuth(w, r) {
		return
	}

	entries := h.ledger.Entries()
	writeJSON(w, map[string]interface{}{
		"count":   len(entries),
		"entries": entries,
	})
}

func (h *Handler) GetCapabilities(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	capPath := os.Getenv("MOHAWK_CAPABILITIES_PATH")
	if strings.TrimSpace(capPath) == "" {
		capPath = "capabilities.json"
	}
	bridgePath := os.Getenv("MOHAWK_BRIDGE_POLICIES_PATH")
	if strings.TrimSpace(bridgePath) == "" {
		bridgePath = "bridge-policies.json"
	}

	capabilities := map[string]interface{}{}
	bridgePolicies := map[string]interface{}{}

	cleanCapPath, capErr := sanitizeReadablePath(capPath)
	cleanBridgePath, bridgeErr := sanitizeReadablePath(bridgePath)

	if capErr == nil {
		if raw, err := os.ReadFile(cleanCapPath); err == nil { // #nosec G304,G703 -- path is sanitized before file read
			_ = json.Unmarshal(raw, &capabilities)
		}
	}
	if bridgeErr == nil {
		if raw, err := os.ReadFile(cleanBridgePath); err == nil { // #nosec G304,G703 -- path is sanitized before file read
			_ = json.Unmarshal(raw, &bridgePolicies)
		}
	}

	response := map[string]interface{}{
		"protocol_version":  "1.0.0",
		"stark_backends":    hybrid.AvailableSTARKBackends(),
		"capabilities":      capabilities,
		"bridge_policies":   bridgePolicies,
		"capabilities_path": capPath,
		"bridge_path":       bridgePath,
		"api": map[string]interface{}{
			"base_paths": []string{"/api", "/api/v1"},
			"open_endpoints": []string{
				"GET /health",
				"GET /api/v1/status",
				"GET /api/v1/consensus/status",
				"GET /api/v1/capabilities",
				"GET /api/v1/verification_policy",
				"GET /api/v1/trust_snapshot",
			},
			"auth_protected_endpoints": []string{
				"POST /api/v1/proof/verify",
				"POST /api/v1/proof/hybrid/verify",
				"GET /api/v1/ledger",
				"POST /api/v1/verification_policy",
			},
			"proof_payload": map[string]interface{}{
				"fields":              []string{"proof", "encoding", "public_input"},
				"supported_encodings": []string{"base64", "hex", "raw"},
			},
			"hybrid_payload": map[string]interface{}{
				"fields":              []string{"mode", "stark_backend", "snark_proof", "stark_proof", "encoding"},
				"supported_modes":     []string{string(hybrid.ModeAny), string(hybrid.ModeBoth), string(hybrid.ModePreferSNARK)},
				"supported_encodings": []string{"base64", "hex", "raw"},
			},
			"auth": map[string]interface{}{
				"default_mode":              "file-only",
				"disable_values":            []string{"off", "disabled", "none"},
				"token_headers":             []string{"Authorization: Bearer <token>", "X-API-Token: <token>"},
				"role_header":               "X-API-Role",
				"default_token_file":        "/run/secrets/mohawk_api_token",
				"roles_enforced_by_default": true,
				"default_allowed_roles":     []string{"verifier", "admin"},
				"policy_allowed_roles":      []string{"admin"},
			},
		},
		"observability": map[string]interface{}{
			"proof_metrics": []string{
				"mohawk_proof_verifications_total",
				"mohawk_proof_verification_latency_seconds",
			},
			"ledger_metrics": []string{
				"mohawk_ledger_events_total",
				"mohawk_ledger_entries",
			},
			"ledger_state": map[string]interface{}{
				"entries":  h.ledger.Len(),
				"capacity": h.ledger.cap,
			},
		},
	}

	writeJSON(w, response)
}

func sanitizeReadablePath(pathValue string) (string, error) {
	trimmed := strings.TrimSpace(pathValue)
	if trimmed == "" {
		return "", fmt.Errorf("path is empty")
	}
	if strings.ContainsRune(trimmed, 0) {
		return "", fmt.Errorf("path contains invalid null byte")
	}
	clean := filepath.Clean(trimmed)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("relative parent paths are not allowed")
	}
	return clean, nil
}
