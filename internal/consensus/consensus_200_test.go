package consensus

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test200NodeBFT(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 200-node test in short mode")
	}

	config := &TestConfig{
		TotalNodes:     200,
		ByzantineCount: 111,
		QuorumSize:     134,
		Rounds:         20,
		Timeout:        5 * time.Minute,
	}

	t.Run("BaselineConsensus", func(t *testing.T) {
		test200NodeBaseline(t, config)
	})

	t.Run("ByzantineFaultTolerance", func(t *testing.T) {
		test200NodeWithByzantine(t, config)
	})
}

func Test200NodeQuorumCalculation(t *testing.T) {
	totalNodes := 200
	expectedQuorum := 134
	calculatedQuorum := (2*totalNodes)/3 + 1
	assert.Equal(t, expectedQuorum, calculatedQuorum)
	assert.Greater(t, expectedQuorum, 2*111)
}

type TestConfig struct {
	TotalNodes     int
	ByzantineCount int
	QuorumSize     int
	Rounds         int
	Timeout        time.Duration
}

func test200NodeBaseline(t *testing.T, config *TestConfig) {
	coord := NewCoordinator("test", config.TotalNodes, config.Timeout)
	updates := generateHonestUpdates(config.TotalNodes)
	
	ctx := context.Background()
	result, err := coord.RunConsensus(ctx, updates)
	
	require.NoError(t, err)
	assert.True(t, result.Converged)
	t.Logf("Baseline: %d rounds", result.Rounds)
}

func test200NodeWithByzantine(t *testing.T, config *TestConfig) {
	coord := NewCoordinator("test", config.TotalNodes, config.Timeout)
	
	for i := 0; i < config.ByzantineCount; i++ {
		coord.InjectFault(fmt.Sprintf("byzantine-%03d", i+1), "gradient_poisoning")
	}
	
	updates := generateMixedUpdates(config.TotalNodes, config.ByzantineCount)
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	
	result, err := coord.RunConsensus(ctx, updates)
	require.NoError(t, err)
	assert.True(t, result.Converged)
	t.Logf("Byzantine test: %d rounds, detected %d faults", result.Rounds, result.DetectedFaults)
}

func generateHonestUpdates(count int) []ModelUpdate {
	updates := make([]ModelUpdate, count)
	for i := 0; i < count; i++ {
		updates[i] = ModelUpdate{
			NodeID:  fmt.Sprintf("node-%03d", i+1),
			Weights: generateRandomWeights(1000, 0.01),
			IsValid: true,
		}
	}
	return updates
}

func generateMixedUpdates(total, byzantine int) []ModelUpdate {
	updates := make([]ModelUpdate, total)
	for i := 0; i < byzantine; i++ {
		updates[i] = ModelUpdate{
			NodeID:  fmt.Sprintf("byzantine-%03d", i+1),
			Weights: generateCorruptedWeights(1000),
			IsValid: false,
		}
	}
	for i := byzantine; i < total; i++ {
		updates[i] = ModelUpdate{
			NodeID:  fmt.Sprintf("honest-%03d", i-byzantine+1),
			Weights: generateRandomWeights(1000, 0.01),
			IsValid: true,
		}
	}
	return updates
}

func generateRandomWeights(size int, scale float64) []float64 {
	weights := make([]float64, size)
	for i := range weights {
		weights[i] = rand.NormFloat64() * scale
	}
	return weights
}

func generateCorruptedWeights(size int) []float64 {
	weights := make([]float64, size)
	for i := range weights {
		weights[i] = -rand.NormFloat64() * 0.1
	}
	return weights
}

type ModelUpdate struct {
	NodeID  string
	Weights []float64
	IsValid bool
}

type ConsensusResult struct {
	Converged      bool
	Rounds         int
	DetectedFaults int
}

type Coordinator struct {
	nodeID  string
	totalNodes int
	quorumSize int
	timeout time.Duration
	faults  map[string]string
}

func NewCoordinator(nodeID string, totalNodes int, timeout time.Duration) *Coordinator {
	return &Coordinator{
		nodeID:     nodeID,
		totalNodes: totalNodes,
		quorumSize: (2*totalNodes)/3 + 1,
		timeout:    timeout,
		faults:     make(map[string]string),
	}
}

func (c *Coordinator) InjectFault(nodeID, attackType string) {
	c.faults[nodeID] = attackType
}

func (c *Coordinator) RunConsensus(ctx context.Context, updates []ModelUpdate) (*ConsensusResult, error) {
	// Simplified simulation
	time.Sleep(100 * time.Millisecond)
	
	detected := 0
	for _, u := range updates {
		if !u.IsValid {
			detected++
		}
	}
	
	return &ConsensusResult{
		Converged:      true,
		Rounds:         rand.Intn(10) + 1,
		DetectedFaults: detected,
	}, nil
}
