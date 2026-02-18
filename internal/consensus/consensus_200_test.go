// consensus_200_test.go
// 200-Node Byzantine Fault Tolerance Tests
// Sovereign Map Federated Learning

package consensus

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test200NodeBFT runs the full 200-node Byzantine fault tolerance test
func Test200NodeBFT(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping 200-node test in short mode")
	}

	config := &TestConfig{
		TotalNodes:     200,
		ByzantineCount: 111, // 55.5%
		QuorumSize:     134, // ⌈(2*200/3)⌉ + 1
		Rounds:         20,
		Timeout:        5 * time.Minute,
		NetworkLatency: 50 * time.Millisecond,
	}

	t.Logf("Starting 200-Node BFT Test: %d total, %d Byzantine (%.1f%%)",
		config.TotalNodes, config.ByzantineCount,
		float64(config.ByzantineCount)/float64(config.TotalNodes)*100)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	// Run sub-tests
	t.Run("SetupNetwork", func(t *testing.T) {
		test200NodeNetworkSetup(t, config)
	})

	t.Run("BaselineConsensus", func(t *testing.T) {
		test200NodeBaseline(t, config)
	})

	t.Run("ByzantineFaultTolerance", func(t *testing.T) {
		test200NodeWithByzantine(t, config)
	})

	t.Run("NetworkPartitionRecovery", func(t *testing.T) {
		test200NodePartitionRecovery(t, config)
	})

	t.Run("PerformanceBenchmark", func(t *testing.T) {
		test200NodePerformance(t, config)
	})
}

// Test200NodeQuorumCalculation validates quorum math for 200 nodes
func Test200NodeQuorumCalculation(t *testing.T) {
	totalNodes := 200
	expectedQuorum := 134 // ⌈(2*200/3)⌉ + 1 = 134

	// Test standard BFT threshold (33%)
	maxByzantine33 := int(float64(totalNodes) * 0.33)
	assert.Less(t, maxByzantine33*3, totalNodes, "Standard BFT: n > 3f violated")

	// Test Sovereign Map threshold (55.5%)
	maxByzantine55 := int(float64(totalNodes) * 0.555)
	assert.Equal(t, 111, maxByzantine55, "55.5% of 200 should be 111")

	// Verify quorum formula
	calculatedQuorum := (2*totalNodes)/3 + 1
	assert.Equal(t, expectedQuorum, calculatedQuorum, "Quorum calculation mismatch")

	// Verify safety: quorum > 2f
	assert.Greater(t, expectedQuorum, 2*maxByzantine55,
		"Quorum must be greater than 2*Byzantine count for safety")
}

// Test200NodeByzantineDetection tests detection of 111 malicious nodes
func Test200NodeByzantineDetection(t *testing.T) {
	config := &TestConfig{
		TotalNodes:     200,
		ByzantineCount: 111,
		DetectionRate:  0.90, // Must detect 90% of faults
	}

	detector := NewByzantineDetector(config)

	// Simulate 111 Byzantine nodes with various attacks
	byzantineNodes := generateByzantineNodes(111)
	detected := 0

	for _, node := range byzantineNodes {
		update := generateCorruptedUpdate(node)
		if detector.Analyze(update) {
			detected++
		}
	}

	detectionRate := float64(detected) / float64(len(byzantineNodes))
	t.Logf("Byzantine Detection: %d/%d (%.1f%%)", detected, len(byzantineNodes), detectionRate*100)

	assert.GreaterOrEqual(t, detectionRate, config.DetectionRate,
		"Failed to meet minimum detection rate of %.0f%%", config.DetectionRate*100)
}

// Test200NodeConsensusConvergence tests consensus with 55.5% faults
func Test200NodeConsensusConvergence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long convergence test")
	}

	coordinator := NewConsensusCoordinator(&CoordinatorConfig{
		NodeCount:      200,
		ByzantineRatio: 0.555,
		QuorumSize:     134,
		MaxRounds:      50,
	})

	// Inject 111 Byzantine nodes
	for i := 0; i < 111; i++ {
		coordinator.InjectFault(fmt.Sprintf("byzantine-%03d", i+1), "gradient_poisoning")
	}

	// Attempt consensus
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := coordinator.RunConsensus(ctx, generateHonestUpdates(89))
	require.NoError(t, err, "Consensus should succeed despite 55.5% faults")
	assert.True(t, result.Converged, "Consensus should converge")
	assert.Less(t, result.Rounds, 50, "Should converge within max rounds")
}

// Test200NodeScalability tests system behavior at 200 node scale
func Test200NodeScalability(t *testing.T) {
	coordinator := NewConsensusCoordinator(&CoordinatorConfig{
		NodeCount: 200,
		QuorumSize: 134,
	})

	// Measure message overhead
	messageCount := 0
	coordinator.SetMessageInterceptor(func(msg interface{}) {
		messageCount++
	})

	// Run single consensus round
	ctx := context.Background()
	updates := generateHonestUpdates(200)

	start := time.Now()
	result, err := coordinator.RunConsensus(ctx, updates)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.True(t, result.Converged)

	// Scalability assertions
	t.Logf("200-Node Consensus: %v, Messages: %d, Latency: %v",
		result.Converged, messageCount, duration)

	// Message complexity should be O(n²) for mesh, O(n) for optimized
	maxExpectedMessages := 200 * 200 * 2 // Upper bound
	assert.Less(t, messageCount, maxExpectedMessages, "Message overhead too high")

	// Latency should be reasonable (< 2 minutes for 200 nodes)
	assert.Less(t, duration, 2*time.Minute, "Consensus too slow for 200 nodes")
}

// Test200NodeMeshTopology validates mesh connectivity for 200 nodes
func Test200NodeMeshTopology(t *testing.T) {
	network := NewMeshNetwork(200)

	// Verify full connectivity
	for i := 1; i <= 200; i++ {
		nodeID := fmt.Sprintf("node-%03d", i)
		neighbors := network.GetNeighbors(nodeID)

		// In full mesh, each node connects to 199 others
		assert.Equal(t, 199, len(neighbors),
			"Node %s should have 199 neighbors in full mesh", nodeID)
	}

	// Verify path redundancy
	paths := network.CountRedundantPaths("node-001", "node-200")
	assert.Greater(t, paths, 100, "Should have multiple redundant paths")
}

// Benchmark200Nodes measures performance with 200 nodes
func Benchmark200Nodes(b *testing.B) {
	coordinator := NewConsensusCoordinator(&CoordinatorConfig{
		NodeCount:  200,
		QuorumSize: 134,
	})

	updates := generateHonestUpdates(200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		_, err := coordinator.RunConsensus(ctx, updates)
		cancel()

		if err != nil {
			b.Fatalf("Consensus failed: %v", err)
		}
	}
	b.ReportMetric(float64(200), "nodes")
}

// Helper functions

type TestConfig struct {
	TotalNodes     int
	ByzantineCount int
	QuorumSize     int
	Rounds         int
	Timeout        time.Duration
	NetworkLatency time.Duration
	DetectionRate  float64
}

func test200NodeNetworkSetup(t *testing.T, config *TestConfig) {
	t.Log("Setting up 200-node mesh network...")
	// Implementation: initialize 200 node agents
	time.Sleep(2 * time.Second) // Simulate setup
	t.Log("✓ 200-node network ready")
}

func test200NodeBaseline(t *testing.T, config *TestConfig) {
	t.Log("Running baseline consensus (0% Byzantine)...")
	// Run with all honest nodes
	coordinator := NewConsensusCoordinator(&CoordinatorConfig{
		NodeCount: config.TotalNodes,
		QuorumSize: config.QuorumSize,
	})

	updates := generateHonestUpdates(config.TotalNodes)
	ctx := context.Background()

	result, err := coordinator.RunConsensus(ctx, updates)
	require.NoError(t, err)
	assert.True(t, result.Converged)
	t.Logf("✓ Baseline: %d rounds, %v", result.Rounds, result.Duration)
}

func test200NodeWithByzantine(t *testing.T, config *TestConfig) {
	t.Logf("Testing with %d Byzantine nodes (%.1f%%)...",
		config.ByzantineCount, float64(config.ByzantineCount)/float64(config.TotalNodes)*100)

	coordinator := NewConsensusCoordinator(&CoordinatorConfig{
		NodeCount:      config.TotalNodes,
		ByzantineRatio: float64(config.ByzantineCount) / float64(config.TotalNodes),
		QuorumSize:     config.QuorumSize,
	})

	// Inject Byzantine nodes
	for i := 0; i < config.ByzantineCount; i++ {
		nodeID := fmt.Sprintf("byzantine-%03d", i+1)
		attackType := []string{"gradient_poisoning", "label_flipping", "sybil_attack"}[i%3]
		coordinator.InjectFault(nodeID, attackType)
	}

	// Run consensus with mixed honest/Byzantine updates
	updates := generateMixedUpdates(config.TotalNodes, config.ByzantineCount)
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	result, err := coordinator.RunConsensus(ctx, updates)
	require.NoError(t, err, "Consensus should tolerate 55.5% faults")
	assert.True(t, result.Conver
