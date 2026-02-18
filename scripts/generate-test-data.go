// generate-test-data.go
// Generates synthetic model updates for 200-node BFT testing
// Usage: go run scripts/generate-test-data.go

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

// ModelUpdate represents a federated learning model update
type ModelUpdate struct {
	NodeID        string    `json:"node_id"`
	Round         int       `json:"round"`
	Timestamp     int64     `json:"timestamp"`
	Weights       []float64 `json:"weights"`
	WeightShape   []int     `json:"weight_shape"`
	GradientNorm  float64   `json:"gradient_norm"`
	SampleCount   int       `json:"sample_count"`
	IsByzantine   bool      `json:"is_byzantine"`
	AttackType    string    `json:"attack_type,omitempty"`
	Signature     string    `json:"signature"`
	Proof         string    `json:"zk_proof,omitempty"`
}

// TestScenario defines a complete test scenario
type TestScenario struct {
	Metadata struct {
		ScenarioID    string    `json:"scenario_id"`
		Description   string    `json:"description"`
		TotalNodes    int       `json:"total_nodes"`
		ByzantineCount int      `json:"byzantine_count"`
		ByzantineRatio float64  `json:"byzantine_ratio"`
		CreatedAt     time.Time `json:"created_at"`
	} `json:"metadata"`
	Nodes      []NodeConfig `json:"nodes"`
	Updates    []ModelUpdate `json:"updates"`
}

type NodeConfig struct {
	ID          string   `json:"id"`
	Index       int      `json:"index"`
	IsByzantine bool     `json:"is_byzantine"`
	AttackType  string   `json:"attack_type"`
	CPUQuota    float64  `json:"cpu_quota"`
	MemoryLimit int64    `json:"memory_limit"`
}

const (
	TotalNodes     = 200
	ByzantineCount = 111 // 55.5% of 200
	ModelSize      = 1000 // Number of weights per model
	OutputDir      = "test-data"
)

var attackTypes = []string{
	"gradient_poisoning",
	"label_flipping", 
	"sybil_attack",
	"free_rider",
	"gradient_boosting", // Amplifies gradients
	"random_weights",    // Submit random noise
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	fmt.Println("🎲 Generating synthetic test data for 200-node BFT test...")

	// Create output directory
	if err := os.MkdirAll(OutputDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create output directory: %v", err))
	}

	// Generate test scenarios
	scenarios := []string{"baseline", "byzantine-55", "partition-3way", "churn-50"}

	for _, scenario := range scenarios {
		fmt.Printf("  Generating scenario: %s\n", scenario)
		data := generateScenario(scenario)
		filename := fmt.Sprintf("%s/%s-200-nodes.json", OutputDir, scenario)
		
		file, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
		}

		if err := os.WriteFile(filename, file, 0644); err != nil {
			panic(fmt.Sprintf("Failed to write file: %v", err))
		}
		
		fmt.Printf("    ✓ Saved %d nodes to %s\n", len(data.Updates), filename)
	}

	// Generate specialized files
	generateIndividualUpdates()
	generateNetworkPartitionConfig()
	generateChurnPattern()
	generateAttackSignatures()

	fmt.Println("\n✅ Test data generation complete!")
	fmt.Printf("📁 Output directory: %s/\n", OutputDir)
	fmt.Println("📊 Summary:")
	fmt.Printf("   - Total scenarios: %d\n", len(scenarios))
	fmt.Printf("   - Nodes per scenario: %d\n", TotalNodes)
	fmt.Printf("   - Byzantine nodes: %d (%.1f%%)\n", ByzantineCount, float64(ByzantineCount)/float64(TotalNodes)*100)
}

func generateScenario(scenarioType string) TestScenario {
	scenario := TestScenario{}
	
	// Metadata
	scenario.Metadata.ScenarioID = fmt.Sprintf("%s-%s", scenarioType, time.Now().Format("20060102"))
	scenario.Metadata.Description = getScenarioDescription(scenarioType)
	scenario.Metadata.TotalNodes = TotalNodes
	scenario.Metadata.ByzantineCount = ByzantineCount
	scenario.Metadata.ByzantineRatio = float64(ByzantineCount) / float64(TotalNodes)
	scenario.Metadata.CreatedAt = time.Now()

	// Generate node configs and updates
	scenario.Nodes = make([]NodeConfig, TotalNodes)
	scenario.Updates = make([]ModelUpdate, TotalNodes)

	// First 111 nodes are Byzantine (as per 55.5% ratio)
	byzantineIndices := rand.Perm(TotalNodes)[:ByzantineCount]
	byzantineMap := make(map[int]bool)
	for _, idx := range byzantineIndices {
		byzantineMap[idx] = true
	}

	for i := 0; i < TotalNodes; i++ {
		isByzantine := byzantineMap[i]
		
		// Node configuration
		node := NodeConfig{
			ID:          fmt.Sprintf("node-%03d", i+1),
			Index:       i,
			IsByzantine: isByzantine,
			CPUQuota:    0.5,
			MemoryLimit: 512 * 1024 * 1024, // 512MB
		}

		// Assign attack type if Byzantine
		if isByzantine {
			node.AttackType = attackTypes[rand.Intn(len(attackTypes))]
		}

		scenario.Nodes[i] = node

		// Generate model update
		update := generateModelUpdate(node, scenarioType)
		scenario.Updates[i] = update
	}

	return scenario
}

func generateModelUpdate(node NodeConfig, scenarioType string) ModelUpdate {
	update := ModelUpdate{
		NodeID:    node.ID,
		Round:     1,
		Timestamp: time.Now().Unix(),
		WeightShape: []int{ModelSize},
		SampleCount: rand.Intn(100) + 50, // 50-150 samples
		IsByzantine: node.IsByzantine,
		Signature:   generateSignature(node.ID),
	}

	// Generate weights based on node type
	if node.IsByzantine {
		update.AttackType = node.AttackType
		update.Weights = generateByzantineWeights(node.AttackType)
		update.Proof = "INVALID_PROOF_" + node.ID // Invalid ZK proof
	} else {
		update.Weights = generateHonestWeights()
		update.Proof = generateValidProof(node.ID)
	}

	// Calculate gradient norm
	update.GradientNorm = calculateGradientNorm(update.Weights)

	return update
}

func generateHonestWeights() []float64 {
	weights := make([]float64, ModelSize)
	
	// Simulate honest training with small, reasonable gradients
	for i := range weights {
		// Normal distribution around 0, small variance
		weights[i] = rand.NormFloat64() * 0.01
	}
	
	return weights
}

func generateByzantineWeights(attackType string) []float64 {
	weights := make([]float64, ModelSize)
	
	switch attackType {
	case "gradient_poisoning":
		// Invert gradients to harm model
		for i := range weights {
			weights[i] = -rand.NormFloat64() * 0.1 // Large negative values
		}
		
	case "label_flipping":
		// Random sign flips
		for i := range weights {
			val := rand.NormFloat64() * 0.01
			if rand.Float64() < 0.5 {
				val = -val
			}
			weights[i] = val
		}
		
	case "sybil_attack":
		// Coordinated attack - all same values
		coordVal := rand.NormFloat64() * 0.05
		for i := range weights {
			weights[i] = coordVal + rand.NormFloat64()*0.001
		}
		
	case "free_rider":
		// Submit zeros or near-zeros
		for i := range weights {
			weights[i] = rand.NormFloat64() * 0.0001
		}
		
	case "gradient_boosting":
		// Amplify gradients significantly
		for i := range weights {
			weights[i] = rand.NormFloat64() * 0.5 // 50x larger
		}
		
	case "random_weights":
		// Complete noise
		for i := range weights {
			weights[i] = rand.Float64()*2 - 1 // Random -1 to 1
		}
		
	default:
		// Default to honest-looking but slightly off
		for i := range weights {
			weights[i] = rand.NormFloat64() * 0.02
		}
	}
	
	return weights
}

func calculateGradientNorm(weights []float64) float64 {
	sum := 0.0
	for _, w := range weights {
		sum += w * w
	}
	return math.Sqrt(sum)
}

func generateSignature(nodeID string) string {
	// Simulate ECDSA signature
	return fmt.Sprintf("SIG_%s_%d", nodeID, rand.Int63())
}

func generateValidProof(nodeID string) string {
	// Simulate valid ZK-SNARK proof
	return fmt.Sprintf("ZKPROOF_VALID_%s_%x", nodeID, rand.Int63())
}

func getScenarioDescription(scenarioType string) string {
	descriptions := map[string]string{
		"baseline":      "Baseline test with 0% Byzantine nodes for comparison",
		"byzantine-55":  "Full BFT test with 55.5% Byzantine nodes (111 of 200)",
		"partition-3way": "Network partitioned into 3 groups during consensus",
		"churn-50":      "50% node churn (join/leave) during test",
	}
	
	if desc, ok := descriptions[scenarioType]; ok {
		return desc
	}
	return "Custom test scenario"
}

// generateIndividualUpdates creates separate files for each node type
func generateIndividualUpdates() {
	fmt.Println("  Generating individual update files...")
	
	// Honest updates sample (10 nodes)
	honest := make([]ModelUpdate, 10)
	for i := 0; i < 10; i++ {
		node := NodeConfig{
			ID:          fmt.Sprintf("honest-%03d", i+1),
			IsByzantine: false,
		}
		honest[i] = generateModelUpdate(node, "baseline")
	}
	
	file, _ := json.MarshalIndent(honest, "", "  ")
	os.WriteFile(fmt.Sprintf("%s/sample-honest-updates.json", OutputDir), file, 0644)
	
	// Byzantine updates sample (10 nodes)
	byzantine := make([]ModelUpdate, 10)
	for i := 0; i < 10; i++ {
		node := NodeConfig{
			ID:          fmt.Sprintf("byzantine-%03d", i+1),
			IsByzantine: true,
			AttackType:  attackTypes[i%len(attackTypes)],
		}
		byzantine[i] = generateModelUpdate(node, "byzantine-55")
	}
	
	file, _ = json.MarshalIndent(byzantine, "", "  ")
	os.WriteFile(fmt.Sprintf("%s/sample-byzantine-updates.json", OutputDir), file, 0644)
	
	fmt.Println("    ✓ Sample update files created")
}

// generateNetworkPartitionConfig creates partition scenarios
func generateNetworkPartitionConfig() {
	fmt.Println("  Generating network partition configurations...")
	
	partitions := []map[string]interface{}{
		{
			"name":        "3-way-split",
			"description": "Split 200 nodes into 3 partitions (67, 67, 66)",
			"partitions": []map[string]interface{}{
				{"id": "A", "nodes": 67, "leader": "node-001"},
				{"id": "B", "nodes": 67, "leader": "node-068"},
				{"id": "C", "nodes": 66, "leader": "node-135"},
			},
			"duration_seconds": 60,
		},
		{
			"name":        "isolation-attack",
			"description": "Isolate 55 honest nodes from 145 others (including 111 Byzantine)",
			"partitions": []map[string]interface{}{
				{"id": "honest-island", "nodes": 55, "all_honest": true},
				{"id": "byzantine-land", "nodes": 145, "byzantine_count": 111},
			},
			"duration_seconds": 120,
		},
	}
	
	file, _ := json.MarshalIndent(partitions, "", "  ")
	os.WriteFile(fmt.Sprintf("%s/partition-scenarios.json", OutputDir), file, 0644)
	
	fmt.Println("    ✓ Partition scenarios created")
}

// generateChurnPattern simulates node join/leave patterns
func generateChurnPattern() {
	fmt.Println("  Generating node churn patterns...")
	
	churn := map[string]interface{}{
		"scenario": "50-percent-churn",
		"total_nodes": 200,
		"rounds": 20,
		"churn_events": []map[string]interface{}{},
	}
	
	// Generate churn events
	for round := 1; round <= 20; round++ {
		if round%4 == 0 { // Every 4th round
			event := map[string]interface{}{
				"round": round,
				"leaving_nodes": rand.Intn(50) + 25, // 25-75 nodes leave
				"joining_nodes": rand.Intn(50) + 25, // 25-75 new nodes join
				"duration_seconds": 30,
			}
			churn["churn_events"] = append(churn["churn_events"].([]map[string]interface{}), event)
		}
	}
	
	file, _ := json.MarshalIndent(churn, "", "  ")
	os.WriteFile(fmt.Sprintf("%s/churn-pattern.json", OutputDir), file, 0644)
	
	fmt.Println("    ✓ Churn pattern created")
}

// generateAttackSignatures creates known-bad signatures for detection testing
func generateAttackSignatures() {
	fmt.Println("  Generating attack signature database...")
	
	signatures := map[string]interface{}{
		"version": "1.0",
		"attack_patterns": map[string]interface{}{
			"gradient_poisoning": map[string]interface{}{
				"indicators": []string{
					"negative_gradient_sum",
					"high_variance",
					"opposite_direction",
				},
				"threshold": 0.8,
			},
			"sybil_attack": map[string]interface{}{
				"indicators": []string{
					"identical_weights",
					"coordinated_timing",
					"similar_signatures",
				},
				"threshold": 0.9,
			},
			"free_rider": map[string]interface{}{
				"indicators": []string{
					"near_zero_gradients",
					"no_training_progress",
					"minimal_sample_count",
				},
				"threshold": 0.95,
			},
		},
		"known_bad_nodes": []string{}, // To be filled during test
	}
	
	file, _ := json.MarshalIndent(signatures, "", "  ")
	os.WriteFile(fmt.Sprintf("%s/attack-signatures.json", OutputDir), file, 0644)
	
	fmt.Println("    ✓ Attack signature database created")
}
