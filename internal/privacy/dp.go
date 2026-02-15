// Package privacy implements SGP-001 differential privacy standard
// Epsilon (ε) = 1.0, Delta (δ) = 1e-5
package privacy

import (
	"crypto/rand"
	"fmt"
	"math"
	"sync"
  	"encoding/binary"
)

// SGP001Config defines the privacy budget parameters for SGP-001 standard
type SGP001Config struct {
	Epsilon float64 // Privacy loss parameter (ε = 1.0)
	Delta   float64 // Privacy failure probability (δ = 1e-5)
	L2Sensitivity float64 // L2 sensitivity of the query
	mu sync.Mutex
}

// NewSGP001Config creates a new SGP-001 configuration with standard parameters
func NewSGP001Config() *SGP001Config {
	return &SGP001Config{
		Epsilon: 1.0,
		Delta:   1e-5,
		L2Sensitivity: 1.0,
	}
}

// DifferentialPrivacy handles privacy-preserving operations
type DifferentialPrivacy struct {
	config *SGP001Config
	budgetUsed float64
	mu sync.RWMutex
}

// NewDifferentialPrivacy creates a new differential privacy manager
func NewDifferentialPrivacy(config *SGP001Config) *DifferentialPrivacy {
	return &DifferentialPrivacy{
		config: config,
		budgetUsed: 0.0,
	}
}

// AddGaussianNoise adds calibrated Gaussian noise for differential privacy
// Implements the Gaussian mechanism for (ε,δ)-differential privacy
func (dp *DifferentialPrivacy) AddGaussianNoise(value float64) (float64, error) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Check if privacy budget is exhausted
	if dp.budgetUsed >= dp.config.Epsilon {
		return 0, fmt.Errorf("privacy budget exhausted: used %.2f/%.2f", dp.budgetUsed, dp.config.Epsilon)
	}

	// Calculate noise scale using Gaussian mechanism
	// σ² = 2 * ln(1.25/δ) * Δ²/ε²
	sigma := dp.calculateNoiseScale()

	// Generate Gaussian noise
	noise, err := dp.gaussianNoise(0, sigma)
	if err != nil {
		return 0, fmt.Errorf("failed to generate noise: %w", err)
	}

	// Update privacy budget
	dp.budgetUsed += dp.config.Epsilon / 10.0 // Incremental budget consumption

	return value + noise, nil
}

// AddLaplaceNoise adds Laplace noise for differential privacy
// Implements the Laplace mechanism for ε-differential privacy
func (dp *DifferentialPrivacy) AddLaplaceNoise(value float64, sensitivity float64) (float64, error) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Calculate Laplace parameter: b = Δf/ε
	scale := sensitivity / dp.config.Epsilon

	// Generate Laplace noise
	noise, err := dp.laplaceNoise(scale)
	if err != nil {
		return 0, fmt.Errorf("failed to generate Laplace noise: %w", err)
	}

	return value + noise, nil
}

// calculateNoiseScale computes the noise scale for Gaussian mechanism
func (dp *DifferentialPrivacy) calculateNoiseScale() float64 {
	// σ = Δ * sqrt(2 * ln(1.25/δ)) / ε
	dp.config.mu.Lock()
	defer dp.config.mu.Unlock()
	
	term := 2.0 * math.Log(1.25/dp.config.Delta)
	sigma := dp.config.L2Sensitivity * math.Sqrt(term) / dp.config.Epsilon
	return sigma
}

// gaussianNoise generates noise from a Gaussian distribution N(mean, variance)
func (dp *DifferentialPrivacy) gaussianNoise(mean, stddev float64) (float64, error) {
	// Box-Muller transform for generating Gaussian random variables
	var u1, u2 float64
	
	// Generate uniform random numbers
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return 0, err
	}
	u1 = float64(binary.BigEndian.Uint64(buf)) / float64(math.MaxUint64)
	
	if _, err := rand.Read(buf); err != nil {
		return 0, err
	}
	u2 = float64(binary.BigEndian.Uint64(buf)) / float64(math.MaxUint64)
	
	// Box-Muller transform
	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	return mean + z0*stddev, nil
}

// laplaceNoise generates noise from a Laplace distribution
func (dp *DifferentialPrivacy) laplaceNoise(scale float64) (float64, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return 0, err
	}
	
	u := float64(binary.BigEndian.Uint64(buf))/float64(math.MaxUint64) - 0.5
	return -scale * math.Copysign(1.0, u) * math.Log(1.0-2.0*math.Abs(u)), nil
}

// GetPrivacyBudget returns the current privacy budget usage
func (dp *DifferentialPrivacy) GetPrivacyBudget() (used, total float64) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.budgetUsed, dp.config.Epsilon
}

// ResetPrivacyBudget resets the privacy budget counter
func (dp *DifferentialPrivacy) ResetPrivacyBudget() {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.budgetUsed = 0.0
}

// VerifyPrivacyCompliance checks if privacy parameters meet SGP-001 standard
func (dp *DifferentialPrivacy) VerifyPrivacyCompliance() error {
	dp.config.mu.Lock()
	defer dp.config.mu.Unlock()
	
	if dp.config.Epsilon != 1.0 {
		return fmt.Errorf("SGP-001 violation: epsilon must be 1.0, got %.2f", dp.config.Epsilon)
	}
	if dp.config.Delta != 1e-5 {
		return fmt.Errorf("SGP-001 violation: delta must be 1e-5, got %.2e", dp.config.Delta)
	}
	return nil
}

// AddNoiseToGradients adds differential privacy noise to model gradients
func (dp *DifferentialPrivacy) AddNoiseToGradients(gradients []float64, clipNorm float64) ([]float64, error) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	noisyGradients := make([]float64, len(gradients))
	
	// Clip gradients to bound sensitivity
	clippedGradients := dp.clipGradients(gradients, clipNorm)
	
	// Add Gaussian noise to each gradient component
	sigma := dp.calculateNoiseScale()
	for i, grad := range clippedGradients {
		noise, err := dp.gaussianNoise(0, sigma)
		if err != nil {
			return nil, fmt.Errorf("failed to add noise to gradient %d: %w", i, err)
		}
		noisyGradients[i] = grad + noise
	}
	
	return noisyGradients, nil
}

// clipGradients clips gradient norms to bound sensitivity
func (dp *DifferentialPrivacy) clipGradients(gradients []float64, clipNorm float64) []float64 {
	// Calculate L2 norm
	norm := 0.0
	for _, g := range gradients {
		norm += g * g
	}
	norm = math.Sqrt(norm)
	
	// Clip if necessary
	if norm > clipNorm {
		clipFactor := clipNorm / norm
		clipped := make([]float64, len(gradients))
		for i, g := range gradients {
			clipped[i] = g * clipFactor
		}
		return clipped
	}
	
	return gradients
}

