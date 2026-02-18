package main

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "os"
    "time"
)

// Generate synthetic model updates for 200 nodes
func main() {
    rand.Seed(time.Now().UnixNano())
    
    updates := make([]map[string]interface{}, 200)
    for i := 0; i < 200; i++ {
        // Simulate model weights (simplified)
        weights := make([]float64, 1000)
        for j := range weights {
            weights[j] = rand.NormFloat64() * 0.01
        }
        
        updates[i] = map[string]interface{}{
            "node_id": fmt.Sprintf("node-%d", i+1),
            "round": 1,
            "weights": weights,
            "byzantine": i < 111, // First 111 are Byzantine
            "attack_type": "gradient_poisoning",
        }
    }
    
    file, _ := json.MarshalIndent(updates, "", "  ")
    os.WriteFile("test-data/200-nodes-model-updates.json", file, 0644)
    fmt.Println("Generated test data for 200 nodes (111 Byzantine)")
}
