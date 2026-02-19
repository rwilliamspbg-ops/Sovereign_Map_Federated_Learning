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
	rand.Seed(time.Now().UnixNano())
	os.MkdirAll("test-data", 0755)

	updates := make([]ModelUpdate, 200)
	attackTypes := []string{"gradient_poisoning", "label_flipping", "sybil_attack", "free_rider"}

	for i := 0; i < 200; i++ {
		isByzantine := i < 111
		update := ModelUpdate{
			NodeID:      fmt.Sprintf("node-%03d", i+1),
			Round:       1,
			Weights:     generateWeights(isByzantine),
			IsByzantine: isByzantine,
		}
		if isByzantine {
			update.AttackType = attackTypes[rand.Intn(len(attackTypes))]
		}
		updates[i] = update
	}

	data, _ := json.MarshalIndent(updates, "", "  ")
	os.WriteFile("test-data/200-nodes-model-updates.json", data, 0644)
	fmt.Println("Generated test-data/200-nodes-model-updates.json")
}

func generateWeights(byzantine bool) []float64 {
	weights := make([]float64, 1000)
	for i := range weights {
		if byzantine {
			weights[i] = -rand.NormFloat64() * 0.1
		} else {
			weights[i] = rand.NormFloat64() * 0.01
		}
	}
	return weights
}
