aggregator.go// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package consensus

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// DistributedAggregator coordinates model aggregation across nodes with consensus
type DistributedAggregator struct {
	mu          sync.RWMutex
	coordinator *Coordinator
	nodeID      string
	peerNodes   []string
	models      map[string][]byte
	roundNumber int
	aggregated  []byte
	metrics     *AggregationMetrics
}

// AggregationMetrics tracks aggregation performance
type AggregationMetrics struct {
	TotalRounds      int
	SuccessfulRounds int
	FailedRounds     int
	AverageLatency   time.Duration
	LastRoundTime    time.Time
}

// NewDistributedAggregator creates a new distributed aggregator
func NewDistributedAggregator(nodeID string, peerNodes []string, timeout time.Duration) *DistributedAggregator {
	totalNodes := len(peerNodes) + 1 // +1 for current node
	return &DistributedAggregator{
		coordinator: NewCoordinator(nodeID, totalNodes, timeout),
		nodeID:      nodeID,
		peerNodes:   peerNodes,
		models:      make(map[string][]byte),
		roundNumber: 0,
		metrics:     &AggregationMetrics{},
	}
}

// SubmitModel submits a local model update for aggregation
func (da *DistributedAggregator) SubmitModel(ctx context.Context, nodeID string, modelWeights []byte) error {
	da.mu.Lock()
	defer da.mu.Unlock()

	da.models[nodeID] = modelWeights
	return nil
}

// AggregateWithConsensus performs model aggregation with distributed consensus
func (da *DistributedAggregator) AggregateWithConsensus(ctx context.Context) ([]byte, error) {
	startTime := time.Now()
	da.mu.Lock()
	da.roundNumber++
	currentRound := da.roundNumber
	da.mu.Unlock()

	// Step 1: Aggregate local models
	aggregated, err := da.aggregateModels()
	if err != nil {
		da.metrics.FailedRounds++
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}

	// Step 2: Create proposal
	proposal := &ModelProposal{
		Round:      currentRound,
		Weights:    aggregated,
		ProposerID: da.nodeID,
		Proof:      da.generateProof(aggregated),
		Timestamp:  time.Now(),
	}

	// Step 3: Submit proposal to consensus
	proposalID, err := da.coordinator.ProposeModel(ctx, proposal)
	if err != nil {
		da.metrics.FailedRounds++
		return nil, fmt.Errorf("proposal failed: %w", err)
	}

	// Step 4: Collect votes from peers (simulated)
	if err := da.collectVotes(ctx, proposalID); err != nil {
		da.metrics.FailedRounds++
		return nil, fmt.Errorf("vote collection failed: %w", err)
	}

	// Step 5: Check consensus
	consensusReached, err := da.coordinator.CheckConsensus(proposalID)
	if err != nil {
		da.metrics.FailedRounds++
		return nil, fmt.Errorf("consensus check failed: %w", err)
	}

	if !consensusReached {
		da.metrics.FailedRounds++
		return nil, fmt.Errorf("consensus not reached for round %d", currentRound)
	}

	// Step 6: Commit the aggregated model
	if err := da.coordinator.CommitModel(ctx, proposalID); err != nil {
		da.metrics.FailedRounds++
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	// Update metrics
	da.mu.Lock()
	da.aggregated = aggregated
	da.metrics.SuccessfulRounds++
	da.metrics.TotalRounds++
	da.metrics.LastRoundTime = time.Now()
	da.metrics.AverageLatency = time.Since(startTime)
	da.mu.Unlock()

	// Reset for next round
	da.coordinator.Reset()

	return aggregated, nil
}

// aggregateModels performs weighted average aggregation
func (da *DistributedAggregator) aggregateModels() ([]byte, error) {
	if len(da.models) == 0 {
		return nil, fmt.Errorf("no models to aggregate")
	}

	// Simple averaging for demonstration
	// In production, this would be a proper weighted average
	aggregated := make([]byte, 0)
	for _, model := range da.models {
		if len(aggregated) == 0 {
			aggregated = make([]byte, len(model))
		}
		// Simplified aggregation
		for i := range model {
			if i < len(aggregated) {
				aggregated[i] = (aggregated[i] + model[i]) / 2
			}
		}
	}

	return aggregated, nil
}

// generateProof creates a cryptographic proof of the aggregation
func (da *DistributedAggregator) generateProof(aggregated []byte) []byte {
	hash := sha256.Sum256(aggregated)
	return []byte(hex.EncodeToString(hash[:]))
}

// collectVotes simulates collecting votes from peer nodes
func (da *DistributedAggregator) collectVotes(ctx context.Context, proposalID string) error {
	// In production, this would send requests to peer nodes
	// For now, simulate votes from peers
	for _, peerID := range da.peerNodes {
		vote := &Vote{
			NodeID:     peerID,
			ProposalID: proposalID,
			Approve:    true, // Simplified: assume all approve
			Signature:  []byte("signature-" + peerID),
			Timestamp:  time.Now(),
		}
		if err := da.coordinator.CastVote(ctx, vote); err != nil {
			return err
		}
	}
	return nil
}

// GetMetrics returns aggregation metrics
func (da *DistributedAggregator) GetMetrics() *AggregationMetrics {
	da.mu.RLock()
	defer da.mu.RUnlock()
	return da.metrics
}

// GetLastAggregated returns the most recently aggregated model
func (da *DistributedAggregator) GetLastAggregated() []byte {
	da.mu.RLock()
	defer da.mu.RUnlock()
	return da.aggregated
}
