package convergence

import (
	"testing"
)

func TestNewDetector(t *testing.T) {
	epsilon := 0.001
	zetaSq := 0.1
	windowSize := 10
	minIters := 5

	detector := NewDetector(epsilon, zetaSq, windowSize, minIters)
	if detector == nil {
		t.Fatal("Expected non-nil detector")
	}

	if detector.threshold != epsilon {
		t.Errorf("Expected threshold %.3f, got %.3f", epsilon, detector.threshold)
	}

	if detector.heterogeneity != zetaSq {
		t.Errorf("Expected heterogeneity %.1f, got %.1f", zetaSq, detector.heterogeneity)
	}
}

func TestRecordGradient(t *testing.T) {
	detector := NewDetector(0.001, 0.1, 10, 5)

	gradNorms := []float64{1.0, 0.8, 0.6, 0.4, 0.2}
	for _, norm := range gradNorms {
		detector.RecordGradient(norm)
	}

	detector.mu.RLock()
	count := len(detector.gradientHistory)
	detector.mu.RUnlock()

	if count != 5 {
		t.Errorf("Expected 5 gradient records, got %d", count)
	}
}

func TestRecordLoss(t *testing.T) {
	detector := NewDetector(0.001, 0.1, 10, 5)

	losses := []float64{2.0, 1.5, 1.0, 0.8, 0.6}
	for _, loss := range losses {
		detector.RecordLoss(loss)
	}

	detector.mu.RLock()
	count := len(detector.lossHistory)
	detector.mu.RUnlock()

	if count != 5 {
		t.Errorf("Expected 5 loss records, got %d", count)
	}
}

func TestIsConvergedInsufficientData(t *testing.T) {
	detector := NewDetector(0.001, 0.1, 10, 5)

	// Add fewer than minIterations
	detector.RecordGradient(0.5)
	detector.RecordGradient(0.4)

	if detector.IsConverged() {
		t.Error("Expected not converged with insufficient data")
	}
}

func TestIsConvergedSuccess(t *testing.T) {
	detector := NewDetector(0.01, 0.0001, 5, 3)

	// Add gradients below threshold
	for i := 0; i < 5; i++ {
		detector.RecordGradient(0.001)
		detector.RecordLoss(0.001)
	}

	if !detector.IsConverged() {
		t.Error("Expected convergence with low gradients")
	}
}

func TestIsConvergedHighGradients(t *testing.T) {
	detector := NewDetector(0.01, 0.0001, 5, 3)

	// Add gradients above threshold
	for i := 0; i < 5; i++ {
		detector.RecordGradient(1.0)
	}

	if detector.IsConverged() {
		t.Error("Expected not converged with high gradients")
	}
}

func TestGetConvergenceRate(t *testing.T) {
	detector := NewDetector(0.001, 0.1, 10, 5)

	// Add decreasing gradients
	detector.RecordGradient(1.0)
	detector.RecordGradient(0.8)
	detector.RecordGradient(0.6)
	detector.RecordGradient(0.4)
	detector.RecordGradient(0.2)

	rate := detector.GetConvergenceRate()
	if rate <= 0 {
		t.Error("Expected positive convergence rate")
	}
}

func TestGetHeterogeneityEstimate(t *testing.T) {
	detector := NewDetector(0.001, 0.1, 10, 5)

	// Add varying gradients to create heterogeneity
	gradients := []float64{0.1, 0.5, 0.2, 0.8, 0.3}
	for _, g := range gradients {
		detector.RecordGradient(g)
	}

	heterogeneity := detector.GetHeterogeneityEstimate()
	if heterogeneity < detector.heterogeneity {
		t.Errorf("Expected heterogeneity >= %.3f, got %.3f", detector.heterogeneity, heterogeneity)
	}
}

func TestReset(t *testing.T) {
	detector := NewDetector(0.001, 0.1, 10, 5)

	// Add some data
	for i := 0; i < 5; i++ {
		detector.RecordGradient(0.5)
		detector.RecordLoss(0.5)
	}

	// Reset
	detector.Reset()

	detector.mu.RLock()
	gradCount := len(detector.gradientHistory)
	lossCount := len(detector.lossHistory)
	detector.mu.RUnlock()

	if gradCount != 0 || lossCount != 0 {
		t.Error("Expected empty history after reset")
	}
}

func TestGetMetrics(t *testing.T) {
	detector := NewDetector(0.001, 0.1, 10, 5)

	// Add some data
	for i := 0; i < 5; i++ {
		detector.RecordGradient(0.001)
		detector.RecordLoss(0.001)
	}

	metrics := detector.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	if _, ok := metrics["iterations"]; !ok {
		t.Error("Expected iterations in metrics")
	}

	if _, ok := metrics["convergence_rate"]; !ok {
		t.Error("Expected convergence_rate in metrics")
	}

	if _, ok := metrics["threshold"]; !ok {
		t.Error("Expected threshold in metrics")
	}
}

func TestWindowSizeLimit(t *testing.T) {
	windowSize := 5
	detector := NewDetector(0.001, 0.1, windowSize, 3)

	// Add more than window size
	for i := 0; i < 10; i++ {
		detector.RecordGradient(float64(i))
	}

	detector.mu.RLock()
	count := len(detector.gradientHistory)
	detector.mu.RUnlock()

	if count != windowSize {
		t.Errorf("Expected window size %d, got %d", windowSize, count)
	}
}

func BenchmarkRecordGradient(b *testing.B) {
	detector := NewDetector(0.001, 0.1, 1000, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.RecordGradient(0.5)
	}
}

func BenchmarkIsConverged(b *testing.B) {
	detector := NewDetector(0.001, 0.1, 100, 10)

	// Pre-populate
	for i := 0; i < 100; i++ {
		detector.RecordGradient(0.001)
		detector.RecordLoss(0.001)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.IsConverged()
	}
}
