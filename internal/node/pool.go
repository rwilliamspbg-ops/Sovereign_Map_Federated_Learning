// Copyright 2026 Sovereign-Mohawk Core Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package node

import (
	"context"
	"fmt"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
)

// NodePool manages a group of FL nodes participating in the same network
// Coordinates FL rounds across multiple nodes with blockchain integration
type NodePool struct {
	nodes             map[string]*FLNode
	blockchain        *blockchain.BlockChain
	mempool           *blockchain.Mempool
	validators        *blockchain.ValidatorSet
	blockProposer     *blockchain.BlockProposer
	targetRoundTime   time.Duration
	rewardDistributor *RewardDistributor
	stakeManager      *StakeManager
}

// NewNodePool creates a new pool of FL nodes
func NewNodePool(bc *blockchain.BlockChain, mempool *blockchain.Mempool,
	validators *blockchain.ValidatorSet, blockProposer *blockchain.BlockProposer) *NodePool {
	return &NodePool{
		nodes:             make(map[string]*FLNode),
		blockchain:        bc,
		mempool:           mempool,
		validators:        validators,
		blockProposer:     blockProposer,
		targetRoundTime:   10 * time.Second,
		rewardDistributor: NewRewardDistributor(),
		stakeManager:      NewStakeManager(),
	}
}

// RegisterNode adds a new node to the pool with initial stake
func (p *NodePool) RegisterNode(nodeID string, region string, initialStake uint64) (*FLNode, error) {
	if _, exists := p.nodes[nodeID]; exists {
		return nil, fmt.Errorf("node %s already registered", nodeID)
	}

	node := NewFLNode(nodeID, region)
	if err := node.SetupBlockchain(p.blockchain, p.mempool, initialStake); err != nil {
		return nil, err
	}

	// Register node as validator in blockchain
	if err := p.validators.Stake(nodeID, initialStake); err != nil {
		return nil, err
	}

	// Track stake in manager
	p.stakeManager.AddStake(nodeID, initialStake)

	p.nodes[nodeID] = node
	return node, nil
}

// ExecuteRound runs a single FL training round across all nodes
func (p *NodePool) ExecuteRound(ctx context.Context) (*RoundResult, error) {
	roundStart := time.Now()
	roundData := &RoundResult{
		RoundStartTime:     roundStart,
		ParticipatingNodes: make([]string, 0, len(p.nodes)),
		NodeResults:        make(map[string]*TrainingResult),
		TransactionIDs:     make(map[string]string),
	}

	// 1. Execute training on all nodes
	for nodeID, node := range p.nodes {
		select {
		case <-ctx.Done():
			return roundData, ctx.Err()
		default:
		}

		// Train
		result, err := node.TrainRound(ctx)
		if err != nil {
			_ = p.validators.RecordParticipationQuality(nodeID, 0)
			continue
		}

		// Submit to blockchain
		txID, err := node.SubmitFlRound(ctx, result, []byte{})
		if err != nil {
			_ = p.validators.RecordParticipationQuality(nodeID, 0)
			continue
		}

		qualityScore := uint32(result.Accuracy * 10000)
		if qualityScore > 10000 {
			qualityScore = 10000
		}
		if qualityScore == 0 {
			qualityScore = 1
		}
		_ = p.validators.RecordParticipationQuality(nodeID, qualityScore)

		roundData.ParticipatingNodes = append(roundData.ParticipatingNodes, nodeID)
		roundData.NodeResults[nodeID] = result
		roundData.TransactionIDs[nodeID] = txID
	}

	if len(roundData.ParticipatingNodes) == 0 {
		return roundData, fmt.Errorf("no nodes participated in round")
	}

	// 2. Wait for block inclusion (simulate)
	// In production, would wait for actual consensus + block creation
	time.Sleep(1 * time.Second)

	// 3. Mark transactions as confirmed
	for nodeID := range roundData.TransactionIDs {
		txID := roundData.TransactionIDs[nodeID]
		node := p.nodes[nodeID]

		// Simulate block inclusion at height (round number)
		blockHeight := uint64(len(p.nodes))
		if err := node.ConfirmTransaction(txID, blockHeight); err != nil {
			continue
		}

		// Complete the round
		if err := node.CompleteRound(ctx, blockHeight); err != nil {
			continue
		}

		// Claim rewards
		reward, err := node.ClaimRewards(ctx)
		if err == nil {
			roundData.TotalRewardsDistributed += reward
		}
	}

	roundData.RoundEndTime = time.Now()
	roundData.RoundDuration = roundData.RoundEndTime.Sub(roundStart)

	return roundData, nil
}

// SetNodeAttestationScore updates a node's attestation score in validator set state.
func (p *NodePool) SetNodeAttestationScore(nodeID string, score uint32) error {
	return p.validators.SetAttestationScore(nodeID, score)
}

// GetNodeState returns the current state of a specific node
func (p *NodePool) GetNodeState(nodeID string) (*NodeState, error) {
	node, exists := p.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	return node.GetState(), nil
}

// GetPoolStats returns aggregate statistics for the pool
func (p *NodePool) GetPoolStats() *PoolStats {
	totalRounds := uint64(0)
	totalRewards := uint64(0)
	totalStake := uint64(0)
	activeNodes := 0

	for _, node := range p.nodes {
		state := node.GetState()
		totalRounds += state.TotalRoundsCompleted
		totalRewards += state.TotalRewardsEarned
		totalStake += state.NodeStake
		if state.TotalRoundsCompleted > 0 {
			activeNodes++
		}
	}

	return &PoolStats{
		TotalNodes:              uint64(len(p.nodes)),
		ActiveNodes:             uint64(activeNodes),
		TotalRoundsCompleted:    totalRounds,
		TotalRewardsDistributed: totalRewards,
		TotalStaked:             totalStake,
		AverageRewardsPerNode:   totalRewards / uint64(len(p.nodes)),
		BlockchainHeight:        p.blockchain.Height(),
	}
}

// PoolStats represents aggregate pool statistics
type PoolStats struct {
	TotalNodes              uint64
	ActiveNodes             uint64
	TotalRoundsCompleted    uint64
	TotalRewardsDistributed uint64
	TotalStaked             uint64
	AverageRewardsPerNode   uint64
	BlockchainHeight        uint64
}

// RoundResult contains the results of a single FL round
type RoundResult struct {
	RoundStartTime          time.Time
	RoundEndTime            time.Time
	RoundDuration           time.Duration
	ParticipatingNodes      []string
	NodeResults             map[string]*TrainingResult
	TransactionIDs          map[string]string
	TotalRewardsDistributed uint64
}

// RewardDistributor handles reward distribution logic
type RewardDistributor struct {
	baseRewardPerRound uint64
	accuracyBonusRate  float64 // Bonus as % of base reward per 1% accuracy above 50%
}

// NewRewardDistributor creates a new reward distributor
func NewRewardDistributor() *RewardDistributor {
	return &RewardDistributor{
		baseRewardPerRound: 10000,
		accuracyBonusRate:  100.0,
	}
}

// CalculateReward computes reward for a node based on its performance
func (rd *RewardDistributor) CalculateReward(accuracy float64, stake uint64) uint64 {
	reward := rd.baseRewardPerRound

	// Add accuracy bonus (1% accuracy over 50% = 100 additional tokens)
	if accuracy > 50.0 {
		bonus := uint64((accuracy - 50.0) * rd.accuracyBonusRate)
		reward += bonus
	}

	// Stake multiplier (higher stake = higher rewards)
	// Max multiplier is 2x at 100,000 tokens stake
	stakeMultiplier := float64(stake) / 100000.0
	if stakeMultiplier > 2.0 {
		stakeMultiplier = 2.0
	} else if stakeMultiplier < 1.0 {
		stakeMultiplier = 1.0
	}

	return uint64(float64(reward) * stakeMultiplier)
}

// StakeManager tracks node stake in the system
type StakeManager struct {
	stakes map[string]uint64
}

// NewStakeManager creates a new stake manager
func NewStakeManager() *StakeManager {
	return &StakeManager{
		stakes: make(map[string]uint64),
	}
}

// AddStake records initial stake for a node
func (sm *StakeManager) AddStake(nodeID string, amount uint64) {
	sm.stakes[nodeID] = amount
}

// GetTotalStake returns the total staked across all nodes
func (sm *StakeManager) GetTotalStake() uint64 {
	total := uint64(0)
	for _, stake := range sm.stakes {
		total += stake
	}
	return total
}

// GetNodeStake returns the stake for a specific node
func (sm *StakeManager) GetNodeStake(nodeID string) uint64 {
	return sm.stakes[nodeID]
}
