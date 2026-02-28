// Package client provides a Flower-based federated learning client for mobile devices
package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SovereignMapClient wraps Flower client for mobile federated learning
type SovereignMapClient struct {
	serverAddr string
	conn       *grpc.ClientConn
	mu         sync.Mutex

	// Model training state
	model     ModelWeights
	localData TrainingData
	epochs    int
	batchSize int

	// Metrics
	lastAccuracy float32
	lastLoss     float32
	trainingTime time.Duration

	// Configuration
	byzantine     bool
	privacyBudget float32
}

// ModelWeights represents neural network weights
type ModelWeights struct {
	Layers [][]float32
	Bias   [][]float32
}

// TrainingData represents local training dataset
type TrainingData struct {
	Features [][]float32
	Labels   []int32
	Count    int32
}

// MetricsUpdate represents convergence metrics
type MetricsUpdate struct {
	Round     int32
	Accuracy  float32
	Loss      float32
	Timestamp int64
	NodeID    string
	Byzantine bool
}

// NewSovereignMapClient creates a new federated learning client
// serverAddr: aggregator address (e.g., "localhost:8080")
func NewSovereignMapClient(serverAddr string) *SovereignMapClient {
	return &SovereignMapClient{
		serverAddr:    serverAddr,
		epochs:        3,
		batchSize:     32,
		byzantine:     false,
		privacyBudget: 1.0,
	}
}

// Connect establishes connection to Flower aggregator
func (c *SovereignMapClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return fmt.Errorf("already connected")
	}

	// Create gRPC connection (insecure for now, add TLS for production)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024 * 1024 * 100), // 100MB max message
		),
	}

	conn, err := grpc.DialContext(ctx, c.serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to aggregator: %w", err)
	}

	c.conn = conn
	log.Printf("Connected to aggregator at %s", c.serverAddr)
	return nil
}

// Disconnect closes connection to aggregator
func (c *SovereignMapClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}

// LoadData loads local training data (e.g., from MNIST subset)
func (c *SovereignMapClient) LoadData(features [][]float32, labels []int32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(features) == 0 || len(features) != len(labels) {
		return fmt.Errorf("invalid data: features and labels mismatch")
	}

	c.localData = TrainingData{
		Features: features,
		Labels:   labels,
		Count:    int32(len(labels)),
	}

	log.Printf("Loaded %d training samples", len(labels))
	return nil
}

// TrainLocal performs local model training
// Returns trained weights and metrics
func (c *SovereignMapClient) TrainLocal(ctx context.Context) (ModelWeights, MetricsUpdate, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.localData.Count == 0 {
		return ModelWeights{}, MetricsUpdate{}, fmt.Errorf("no training data loaded")
	}

	start := time.Now()

	// Simulate local training (in production, use actual PyTorch/TensorFlow)
	// This is placeholder for model training logic
	for epoch := 0; epoch < c.epochs; epoch++ {
		// Process data in batches
		for batch := 0; batch < int(c.localData.Count)/c.batchSize; batch++ {
			// Simulate gradient descent
			// In production: loss = model.forward() -> backward() -> optimizer.step()
		}
	}

	c.trainingTime = time.Since(start)

	// Simulate training metrics
	c.lastAccuracy = 0.65 + float32(c.epochs)*0.05 // Mock: improves with epochs
	c.lastLoss = 3.5 - float32(c.epochs)*0.3

	// For Byzantine nodes, invert parameters
	if c.byzantine {
		log.Println("Byzantine mode: inverting update")
		for i := range c.model.Layers {
			for j := range c.model.Layers[i] {
				c.model.Layers[i][j] *= -1
			}
		}
	}

	metrics := MetricsUpdate{
		Accuracy:  c.lastAccuracy,
		Loss:      c.lastLoss,
		Timestamp: time.Now().Unix(),
		Byzantine: c.byzantine,
	}

	log.Printf("Training complete: Accuracy=%.2f%%, Loss=%.4f, Time=%v",
		c.lastAccuracy*100, c.lastLoss, c.trainingTime)

	return c.model, metrics, nil
}

// GetMetrics returns current training metrics
func (c *SovereignMapClient) GetMetrics() MetricsUpdate {
	c.mu.Lock()
	defer c.mu.Unlock()

	return MetricsUpdate{
		Accuracy:  c.lastAccuracy,
		Loss:      c.lastLoss,
		Timestamp: time.Now().Unix(),
		Byzantine: c.byzantine,
	}
}

// SetByzantine enables/disables Byzantine attack mode
func (c *SovereignMapClient) SetByzantine(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byzantine = enabled
	if enabled {
		log.Println("Byzantine mode enabled")
	}
}

// SetEpochs sets number of training epochs
func (c *SovereignMapClient) SetEpochs(epochs int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.epochs = epochs
}

// SetBatchSize sets training batch size
func (c *SovereignMapClient) SetBatchSize(batchSize int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batchSize = batchSize
}

// Health checks connection to aggregator
func (c *SovereignMapClient) Health(ctx context.Context) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return false, fmt.Errorf("not connected")
	}

	// In production, call actual health check endpoint
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Simple connectivity check
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return true, nil
	}
}

// Status returns current client status
func (c *SovereignMapClient) Status() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	connected := c.conn != nil
	return map[string]interface{}{
		"connected":      connected,
		"server_address": c.serverAddr,
		"byzantine":      c.byzantine,
		"last_accuracy":  c.lastAccuracy,
		"last_loss":      c.lastLoss,
		"training_time":  c.trainingTime.String(),
		"data_samples":   c.localData.Count,
	}
}
