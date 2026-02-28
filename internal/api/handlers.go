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
	mux.HandleFunc("/health", h.HealthCheck)
	mux.HandleFunc("/api/status", h.GetStatus)
	mux.HandleFunc("/api/metrics", h.GetMetrics)
	mux.HandleFunc("/api/convergence", h.GetConvergence)
	mux.HandleFunc("/api/island/status", h.GetIslandStatus)
	mux.HandleFunc("/api/peers", h.GetPeers)
}

// HealthCheck returns basic health status
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"status":  "healthy",
		"service": "sovereign-map-fl",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStatus returns overall system status
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMetrics returns collected metrics
func (h *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetConvergence returns convergence status
func (h *Handler) GetConvergence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetIslandStatus returns Island Mode status
func (h *Handler) GetIslandStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
		if mode == island.ModeIsland {
			modeStr = "island"
		} else if mode == island.ModeTransition {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPeers returns information about connected peers
func (h *Handler) GetPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
