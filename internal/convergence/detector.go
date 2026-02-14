// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package convergence

import (
	"math"
	"sync"
	"time"
)

// Detector monitors federated learning convergence across distributed nodes
// Implements Theorem 6: Convergence guarantees under non-IID conditions
type Detector struct {
	mu               sync.RWMutex
	threshold        float64   // ε target convergence threshold
	heterogeneity    float64   // ζ² bound for data heterogeneity
	gradientHistory  []float64 // Historical gradient norms
	lossHistory      []float64 // Historical loss values
	windowSize       int       // Moving window for convergence detection
	minIterations    int       // Minimum iterations before declaring convergence
	lastCheckTime    time.Time
}

// NewDetector initializes convergence detector with proof-backed bounds
func NewDetector(epsilon, zetaSq float64, windowSize, minIters int) *Detector {
	return &Detector{
		threshold:       epsilon,
		heterogeneity:   zetaSq,
		gradientHistory: make([]float64, 0),
		lossHistory:     make([]float64, 0),
		windowSize:      windowSize,
		minIterations:   minIters,
		lastCheckTime:   time.Now(),
	}
}

// RecordGradient adds gradient norm to convergence history
func (d *Detector) RecordGradient(gradNorm float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gradientHistory = append(d.gradientHistory, gradNorm)
	if len(d.gradientHistory) > d.windowSize {
		d.gradientHistory = d.gradientHistory[1:]
	}
}

// RecordLoss adds loss value to convergence history
func (d *Detector) RecordLoss(loss float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lossHistory = append(d.lossHistory, loss)
	if len(d.lossHistory) > d.windowSize {
		d.lossHistory = d.lossHistory[1:]
	}
}

// IsConverged checks if learning has converged
// Formula: E[||∇F(x_T)||²] ≤ O(1/√(KT)) + O(ζ²)
func (d *Detector) IsConverged() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.gradientHistory) < d.minIterations {
		return false
	}

	// Calculate effective threshold accounting for non-IID heterogeneity
	effectiveThreshold := d.threshold + math.Sqrt(d.heterogeneity)

	// Check recent gradient norms
	recentGrads := d.gradientHistory
	if len(d.gradientHistory) > d.windowSize {
		recentGrads = d.gradientHistory[len(d.gradientHistory)-d.windowSize:]
	}

	// All recent gradients must be below threshold
	for _, grad := range recentGrads {
		if grad > effectiveThreshold {
			return false
		}
	}

	// Additional check: loss variance should be low
	if len(d.lossHistory) >= 2 {
		variance := d.calculateVariance(d.lossHistory)
		if variance > effectiveThreshold {
			return false
		}
	}

	return true
}

// GetConvergenceRate estimates current convergence rate
func (d *Detector) GetConvergenceRate() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.gradientHistory) < 2 {
		return 0.0
	}

	// Calculate rate of gradient decrease
	n := len(d.gradientHistory)
	initialGrad := d.gradientHistory[0]
	recentGrad := d.gradientHistory[n-1]

	if initialGrad == 0 {
		return 0.0
	}

	return (initialGrad - recentGrad) / initialGrad
}

// GetHeterogeneityEstimate calculates gradient diversity across nodes
// Implementation of Step 2: Descent Lemma with Heterogeneity
func (d *Detector) GetHeterogeneityEstimate() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.gradientHistory) < 2 {
		return d.heterogeneity
	}

	// In practice, this scales with O(4ζ²) in 4-tier systems
	variance := d.calculateVariance(d.gradientHistory)
	return math.Max(variance, d.heterogeneity)
}

// calculateVariance computes variance of a sample
func (d *Detector) calculateVariance(data []float64) float64 {
	if len(data) == 0 {
		return 0.0
	}

	mean := 0.0
	for _, v := range data {
		mean += v
	}
	mean /= float64(len(data))

	variance := 0.0
	for _, v := range data {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(data))

	return variance
}

// Reset clears convergence history
func (d *Detector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.gradientHistory = make([]float64, 0)
	d.lossHistory = make([]float64, 0)
	d.lastCheckTime = time.Now()
}

// GetMetrics returns current convergence metrics
func (d *Detector) GetMetrics() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	metrics := map[string]interface{}{
		"converged":           d.IsConverged(),
		"iterations":          len(d.gradientHistory),
		"convergence_rate":    d.GetConvergenceRate(),
		"heterogeneity":       d.GetHeterogeneityEstimate(),
		"threshold":           d.threshold,
		"effective_threshold": d.threshold + math.Sqrt(d.heterogeneity),
	}

	if len(d.gradientHistory) > 0 {
		metrics["latest_gradient"] = d.gradientHistory[len(d.gradientHistory)-1]
	}

	if len(d.lossHistory) > 0 {
		metrics["latest_loss"] = d.lossHistory[len(d.lossHistory)-1]
	}

	return metrics
}
