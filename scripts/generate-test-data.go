package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type ModelUpdate struct {
	NodeID      string    `json:"node_id"`
	Round       int       `json:"round"`
	Weights     []float64 `json:"weights"`
	IsByzantine bool      `json:"is_byzantine"`
	AttackType  string    `json:"attack_type,omitempty"`
}

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	if err := os.MkdirAll("test-data", 0755); err != nil {
		fmt.Printf("failed to create test-data directory: %v\n", err)
		os.Exit(1)
	}

	updates := make([]ModelUpdate, 200)
	attackTypes := []string{"gradient_poisoning", "label_flipping", "sybil_attack", "free_rider"}

	for i := 0; i < 200; i++ {
		isByzantine := i < 111
		update := ModelUpdate{
			NodeID:      fmt.Sprintf("node-%03d", i+1),
			Round:       1,
			Weights:     generateWeights(rng, isByzantine),
			IsByzantine: isByzantine,
		}
		if isByzantine {
			update.AttackType = attackTypes[rng.Intn(len(attackTypes))]
		}
		updates[i] = update
	}

	data, err := json.MarshalIndent(updates, "", "  ")
	if err != nil {
		fmt.Printf("failed to encode updates: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("test-data/200-nodes-model-updates.json", data, 0644); err != nil {
		fmt.Printf("failed to write output file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated test-data/200-nodes-model-updates.json")
}

func generateWeights(rng *rand.Rand, byzantine bool) []float64 {
	weights := make([]float64, 1000)
	for i := range weights {
		if byzantine {
			weights[i] = -rng.NormFloat64() * 0.1
		} else {
			weights[i] = rng.NormFloat64() * 0.01
		}
	}
	return weights
}
