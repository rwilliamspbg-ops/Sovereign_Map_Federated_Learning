package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
)

type ModelUpdate struct {
	NodeID      string    `json:"node_id"`
	Round       int       `json:"round"`
	Weights     []float64 `json:"weights"`
	IsByzantine bool      `json:"is_byzantine"`
	AttackType  string    `json:"attack_type,omitempty"`
}

func main() {
	if err := os.MkdirAll("test-data", 0750); err != nil {
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
			Weights:     generateWeights(isByzantine),
			IsByzantine: isByzantine,
		}
		if isByzantine {
			idx, err := secureInt(len(attackTypes))
			if err != nil {
				fmt.Printf("failed to generate secure attack-type index: %v\n", err)
				os.Exit(1)
			}
			update.AttackType = attackTypes[idx]
		}
		updates[i] = update
	}

	data, err := json.MarshalIndent(updates, "", "  ")
	if err != nil {
		fmt.Printf("failed to encode updates: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("test-data/200-nodes-model-updates.json", data, 0600); err != nil {
		fmt.Printf("failed to write output file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated test-data/200-nodes-model-updates.json")
}

func generateWeights(byzantine bool) []float64 {
	weights := make([]float64, 1000)
	for i := range weights {
		r, err := secureFloat64()
		if err != nil {
			fmt.Printf("failed to generate secure random weight: %v\n", err)
			os.Exit(1)
		}
		if byzantine {
			weights[i] = -(r * 0.1)
		} else {
			weights[i] = (r - 0.5) * 0.02
		}
	}
	return weights
}

func secureFloat64() (float64, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return 0, err
	}
	u := binary.LittleEndian.Uint64(bytes[:])
	return float64(u>>11) * (1.0 / (1 << 53)), nil
}

func secureInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive")
	}
	v, err := secureFloat64()
	if err != nil {
		return 0, err
	}
	idx := int(v * float64(max))
	if idx >= max {
		idx = max - 1
	}
	return idx, nil
}
