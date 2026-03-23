// Copyright 2026 Sovereign-Mohawk Core Team
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

// DistributedAggregator coordinates model aggregation across nodes with consensus.
type DistributedAggregator struct {
	mu          sync.RWMutex
	coordinator *Coordinator
	nodeID      string
	peerNodes   []string
	models      map[string]modelSubmission
	roundNumber int
	aggregated  []byte
	metrics     *AggregationMetrics
	asyncMode   bool
	maxStaleAge time.Duration
}

type modelSubmission struct {
	weights   []byte
	submitted time.Time
}

// AggregationMetrics tracks aggregation performance.
type AggregationMetrics struct {
	TotalRounds      int
	SuccessfulRounds int
	FailedRounds     int
	AverageLatency   time.Duration
	LastRoundTime    time.Time
	StaleDrops       int
	AsyncRounds      int
}

// NewDistributedAggregator creates a new distributed aggregator.
func NewDistributedAggregator(nodeID string, peerNodes []string, timeout time.Duration) *DistributedAggregator {
	totalNodes := len(peerNodes) + 1 // +1 for current node
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &DistributedAggregator{
		coordinator: NewCoordinator(nodeID, totalNodes, timeout),
		nodeID:      nodeID,
		peerNodes:   peerNodes,
		models:      make(map[string]modelSubmission),
		roundNumber: 0,
		metrics:     &AggregationMetrics{},
		asyncMode:   false,
		maxStaleAge: timeout,
	}
}

// EnableAsyncMode allows commits to progress with lower vote requirements while
// filtering stale model submissions.
func (da *DistributedAggregator) EnableAsyncMode(minVotes int, maxStaleAge time.Duration) {
	da.mu.Lock()
	defer da.mu.Unlock()

	if minVotes <= 0 {
		minVotes = 1
	}
	da.asyncMode = true
	if maxStaleAge > 0 {
		da.maxStaleAge = maxStaleAge
	}
	da.coordinator.SetAsyncMode(true, minVotes, da.maxStaleAge)
}

// SubmitModel submits a local model update for aggregation.
func (da *DistributedAggregator) SubmitModel(ctx context.Context, nodeID string, modelWeights []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	da.mu.Lock()
	defer da.mu.Unlock()

	da.models[nodeID] = modelSubmission{
		weights:   append([]byte(nil), modelWeights...),
		submitted: time.Now(),
	}
	return nil
}

// AggregateWithConsensus performs model aggregation with distributed consensus.
func (da *DistributedAggregator) AggregateWithConsensus(ctx context.Context) ([]byte, error) {
	startTime := time.Now()

	da.mu.Lock()
	da.roundNumber++
	currentRound := da.roundNumber
	da.mu.Unlock()

	// Step 1: Aggregate local models.
	aggregated, err := da.aggregateModels()
	if err != nil {
		da.recordFailedRound()
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}

	// Step 2: Create proposal.
	proposal := &ModelProposal{
		Round:      currentRound,
		Weights:    aggregated,
		ProposerID: da.nodeID,
		Proof:      da.generateProof(aggregated),
		Timestamp:  time.Now(),
	}

	// Step 3: Submit proposal to consensus.
	proposalID, err := da.coordinator.ProposeModel(ctx, proposal)
	if err != nil {
		da.recordFailedRound()
		return nil, fmt.Errorf("proposal failed: %w", err)
	}

	if err := da.castSelfVote(ctx, proposalID); err != nil {
		da.recordFailedRound()
		return nil, fmt.Errorf("self vote failed: %w", err)
	}

	// Step 4: Collect votes from peers unless async mode is enabled.
	if !da.isAsyncMode() {
		if err := da.collectVotes(ctx, proposalID); err != nil {
			da.recordFailedRound()
			return nil, fmt.Errorf("vote collection failed: %w", err)
		}
	} else {
		da.mu.Lock()
		da.metrics.AsyncRounds++
		da.mu.Unlock()
	}

	// Step 5: Check consensus.
	consensusReached, err := da.coordinator.CheckConsensus(proposalID)
	if err != nil {
		da.recordFailedRound()
		return nil, fmt.Errorf("consensus check failed: %w", err)
	}

	if !consensusReached {
		da.recordFailedRound()
		return nil, fmt.Errorf("consensus not reached for round %d", currentRound)
	}

	// Step 6: Commit the aggregated model.
	if err := da.coordinator.CommitModel(ctx, proposalID); err != nil {
		da.recordFailedRound()
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	// Update metrics.
	da.mu.Lock()
	latency := time.Since(startTime)
	da.aggregated = append([]byte(nil), aggregated...)
	da.metrics.SuccessfulRounds++
	da.metrics.TotalRounds++
	da.metrics.LastRoundTime = time.Now()
	if da.metrics.SuccessfulRounds == 1 {
		da.metrics.AverageLatency = latency
	} else {
		prev := da.metrics.AverageLatency * time.Duration(da.metrics.SuccessfulRounds-1)
		da.metrics.AverageLatency = (prev + latency) / time.Duration(da.metrics.SuccessfulRounds)
	}
	da.mu.Unlock()

	// Reset for next round.
	da.coordinator.Reset()

	return append([]byte(nil), aggregated...), nil
}

// aggregateModels performs weighted average aggregation.
func (da *DistributedAggregator) aggregateModels() ([]byte, error) {
	da.mu.RLock()
	maxStaleAge := da.maxStaleAge
	models := make(map[string]modelSubmission, len(da.models))
	for nodeID, model := range da.models {
		models[nodeID] = model
	}
	da.mu.RUnlock()

	if len(models) == 0 {
		return nil, fmt.Errorf("no models to aggregate")
	}

	var aggregated []byte
	now := time.Now()
	validModels := 0

	for _, model := range models {
		if maxStaleAge > 0 && now.Sub(model.submitted) > maxStaleAge {
			da.mu.Lock()
			da.metrics.StaleDrops++
			da.mu.Unlock()
			continue
		}

		if len(aggregated) == 0 {
			aggregated = make([]byte, len(model.weights))
		}
		if len(model.weights) != len(aggregated) {
			return nil, fmt.Errorf("inconsistent model size: expected %d, got %d", len(aggregated), len(model.weights))
		}

		for i := range model.weights {
			aggregated[i] += model.weights[i]
		}
		validModels++
	}

	if validModels == 0 {
		return nil, fmt.Errorf("all candidate models were stale")
	}

	for i := range aggregated {
		aggregated[i] /= byte(validModels)
	}

	return aggregated, nil
}

// generateProof creates a cryptographic proof of the aggregation.
func (da *DistributedAggregator) generateProof(aggregated []byte) []byte {
	hash := sha256.Sum256(aggregated)
	return []byte(hex.EncodeToString(hash[:]))
}

// collectVotes simulates collecting votes from peer nodes.
func (da *DistributedAggregator) collectVotes(ctx context.Context, proposalID string) error {
	for _, peerID := range da.peerNodes {
		vote := &Vote{
			NodeID:     peerID,
			ProposalID: proposalID,
			Approve:    true,
			Signature:  []byte("signature-" + peerID),
			Timestamp:  time.Now(),
		}
		if err := da.coordinator.CastVote(ctx, vote); err != nil {
			return err
		}
	}
	return nil
}

func (da *DistributedAggregator) castSelfVote(ctx context.Context, proposalID string) error {
	vote := &Vote{
		NodeID:     da.nodeID,
		ProposalID: proposalID,
		Approve:    true,
		Signature:  []byte("signature-" + da.nodeID),
		Timestamp:  time.Now(),
	}
	return da.coordinator.CastVote(ctx, vote)
}

func (da *DistributedAggregator) isAsyncMode() bool {
	da.mu.RLock()
	defer da.mu.RUnlock()
	return da.asyncMode
}

func (da *DistributedAggregator) recordFailedRound() {
	da.mu.Lock()
	defer da.mu.Unlock()
	da.metrics.FailedRounds++
	da.metrics.TotalRounds++
	da.metrics.LastRoundTime = time.Now()
}

// GetMetrics returns aggregation metrics.
func (da *DistributedAggregator) GetMetrics() *AggregationMetrics {
	da.mu.RLock()
	defer da.mu.RUnlock()
	copyMetrics := *da.metrics
	return &copyMetrics
}

// GetLastAggregated returns the most recently aggregated model.
func (da *DistributedAggregator) GetLastAggregated() []byte {
	da.mu.RLock()
	defer da.mu.RUnlock()
	return append([]byte(nil), da.aggregated...)
}
