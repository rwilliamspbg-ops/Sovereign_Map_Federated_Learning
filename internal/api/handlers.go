// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/convergence"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/island"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/monitoring"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/p2p"
)

// Handler provides HTTP endpoints for the federated learning system
type Handler struct {
	convergence *convergence.Detector
	island      *island.Manager
	metrics     *monitoring.Collector
	p2pNetwork  *p2p.Network
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

// NewHandler creates a new API handler with integrated backends
func NewHandler(detector *convergence.Detector, islandMgr *island.Manager, collector *monitoring.Collector, network *p2p.Network) *Handler {
	return &Handler{
		convergence: detector,
		island:      islandMgr,
		metrics:     collector,
		p2pNetwork:  network,
	}
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

	// Versioned aliases
	mux.HandleFunc("/api/v1/status", h.GetStatus)
	mux.HandleFunc("/api/v1/metrics", h.GetMetrics)
	mux.HandleFunc("/api/v1/convergence", h.GetConvergence)
	mux.HandleFunc("/api/v1/island/status", h.GetIslandStatus)
	mux.HandleFunc("/api/v1/peers", h.GetPeers)
	mux.HandleFunc("/api/v1/network_status", h.GetNetworkStatus)
	mux.HandleFunc("/api/v1/trust_status", h.GetTrustStatus)
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

	writeJSON(w, response)
}

// GetMetrics returns collected metrics
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if !ensureGetMethod(w, r) {
		return
	}

	response := map[string]interface{}{
		"total_rounds":     0,
		"active_nodes":     0,
		"convergence_rate": 0.0,
		"network_lag_ms":   0,
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

	writeJSON(w, response)
}
