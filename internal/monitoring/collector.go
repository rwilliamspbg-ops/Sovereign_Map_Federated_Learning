// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package monitoring

import (
	"sync"
	"time"
)

// MetricType represents the type of metric being collected
type MetricType string

const (
	MetricGradient    MetricType = "gradient"
	MetricLoss        MetricType = "loss"
	MetricAccuracy    MetricType = "accuracy"
	MetricRoundTime   MetricType = "round_time"
	MetricPeerCount   MetricType = "peer_count"
	MetricNetworkLag  MetricType = "network_lag"
	MetricTPMAttest   MetricType = "tpm_attestation"
	MetricConsensus   MetricType = "consensus_votes"
)

// Metric represents a single metric observation
type Metric struct {
	Type      MetricType
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
	NodeID    string
}

// Collector aggregates metrics from federated learning operations
type Collector struct {
	mu           sync.RWMutex
	metrics      []Metric
	maxHistory   int
	aggregations map[MetricType]*Aggregation
}

// Aggregation stores statistical aggregates for a metric type
type Aggregation struct {
	Count   int
	Sum     float64
	Min     float64
	Max     float64
	Mean    float64
	StdDev  float64
	Updated time.Time
}

// NewCollector creates a new metrics collector
func NewCollector(maxHistory int) *Collector {
	return &Collector{
		metrics:      make([]Metric, 0, maxHistory),
		maxHistory:   maxHistory,
		aggregations: make(map[MetricType]*Aggregation),
	}
}

// Record adds a new metric observation
func (c *Collector) Record(metricType MetricType, value float64, labels map[string]string, nodeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	metric := Metric{
		Type:      metricType,
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
		NodeID:    nodeID,
	}

	c.metrics = append(c.metrics, metric)

	// Maintain max history size
	if len(c.metrics) > c.maxHistory {
		c.metrics = c.metrics[1:]
	}

	// Update aggregations
	c.updateAggregation(metricType, value)
}

// updateAggregation recalculates statistics for a metric type
func (c *Collector) updateAggregation(metricType MetricType, newValue float64) {
	agg, exists := c.aggregations[metricType]
	if !exists {
		agg = &Aggregation{
			Min: newValue,
			Max: newValue,
		}
		c.aggregations[metricType] = agg
	}

	agg.Count++
	agg.Sum += newValue
	agg.Mean = agg.Sum / float64(agg.Count)
	agg.Updated = time.Now()

	if newValue < agg.Min {
		agg.Min = newValue
	}
	if newValue > agg.Max {
		agg.Max = newValue
	}
}

// GetMetrics returns all recorded metrics
func (c *Collector) GetMetrics() []Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Metric, len(c.metrics))
	copy(result, c.metrics)
	return result
}

// GetMetricsByType returns metrics filtered by type
func (c *Collector) GetMetricsByType(metricType MetricType) []Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]Metric, 0)
	for _, m := range c.metrics {
		if m.Type == metricType {
			result = append(result, m)
		}
	}
	return result
}

// GetAggregation returns statistical summary for a metric type
func (c *Collector) GetAggregation(metricType MetricType) *Aggregation {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if agg, exists := c.aggregations[metricType]; exists {
		// Return a copy
		copy := *agg
		return &copy
	}
	return nil
}

// GetAllAggregations returns all aggregated statistics
func (c *Collector) GetAllAggregations() map[MetricType]*Aggregation {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[MetricType]*Aggregation)
	for k, v := range c.aggregations {
		copy := *v
		result[k] = &copy
	}
	return result
}

// GetSummary returns a human-readable summary of all metrics
func (c *Collector) GetSummary() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := map[string]interface{}{
		"total_metrics": len(c.metrics),
		"max_history":   c.maxHistory,
		"metric_types":  len(c.aggregations),
	}

	aggSummary := make(map[string]interface{})
	for metricType, agg := range c.aggregations {
		aggSummary[string(metricType)] = map[string]interface{}{
			"count":   agg.Count,
			"mean":    agg.Mean,
			"min":     agg.Min,
			"max":     agg.Max,
			"updated": agg.Updated,
		}
	}
	summary["aggregations"] = aggSummary

	return summary
}

// Clear removes all collected metrics
func (c *Collector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics = make([]Metric, 0, c.maxHistory)
	c.aggregations = make(map[MetricType]*Aggregation)
}

// GetRecentMetrics returns metrics from the last N seconds
func (c *Collector) GetRecentMetrics(seconds int) []Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cutoff := time.Now().Add(-time.Duration(seconds) * time.Second)
	result := make([]Metric, 0)

	for _, m := range c.metrics {
		if m.Timestamp.After(cutoff) {
			result = append(result, m)
		}
	}
	return result
}

// GetMetricRate calculates the rate of change for a metric type
func (c *Collector) GetMetricRate(metricType MetricType, windowSeconds int) float64 {
	metrics := c.GetMetricsByType(metricType)
	if len(metrics) < 2 {
		return 0.0
	}

	cutoff := time.Now().Add(-time.Duration(windowSeconds) * time.Second)
	filtered := make([]Metric, 0)

	for _, m := range metrics {
		if m.Timestamp.After(cutoff) {
			filtered = append(filtered, m)
		}
	}

	if len(filtered) < 2 {
		return 0.0
	}

	first := filtered[0]
	last := filtered[len(filtered)-1]
	timeDiff := last.Timestamp.Sub(first.Timestamp).Seconds()

	if timeDiff == 0 {
		return 0.0
	}

	return (last.Value - first.Value) / timeDiff
}
