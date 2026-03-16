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

// Package consensus implements distributed consensus for model aggregation
// Reference: Byzantine fault-tolerant consensus with quorum-based voting

package consensus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
)

// ModelProposal represents a proposed model update for consensus
type ModelProposal struct {
	Round      int
	Weights    []byte
	ProposerID string
	Proof      []byte
	Timestamp  time.Time
}

// Vote represents a node's vote on a proposal
type Vote struct {
	NodeID     string
	ProposalID string
	Approve    bool
	Signature  []byte
	Timestamp  time.Time
}

// ConsensusState tracks the current state of consensus
type ConsensusState int

const (
	Proposing ConsensusState = iota
	Voting
	Committed
	Aborted
)

// ConsensusRound represents a single round of model consensus
type ConsensusRound struct {
	RoundNumber    int
	ProposerID     string
	ProposalID     string
	ModelWeights   []byte
	Metrics        map[string]float64
	CommitTime     time.Time
	ValidatorVotes []*Vote
}

// Coordinator manages distributed consensus for model aggregation
type Coordinator struct {
	mu                   sync.RWMutex
	nodeID               string
	proposals            map[string]*ModelProposal
	votes                map[string][]*Vote
	state                ConsensusState
	quorumSize           int
	totalNodes           int
	timeout              time.Duration
	convergenceThreshold float64

	// Blockchain integration (NEW)
	blockchain      *blockchain.BlockChain
	blockchainState *blockchain.StateDatabase
	blockProposer   *blockchain.BlockProposer
	contractExec    *blockchain.SmartContractExecutor
	validators      *blockchain.ValidatorSet
	mempool         *blockchain.Mempool
	roundNumber     int
}

// NewCoordinator creates a new consensus coordinator
func NewCoordinator(nodeID string, totalNodes int, timeout time.Duration) *Coordinator {
	// Byzantine fault tolerance: quorum = 2f + 1 where f is max faulty nodes
	// For n nodes, f < n/3, so quorum = ⌈(2n/3)⌉
	quorumSize := (2 * totalNodes / 3) + 1

	return &Coordinator{
		nodeID:               nodeID,
		proposals:            make(map[string]*ModelProposal),
		votes:                make(map[string][]*Vote),
		state:                Proposing,
		quorumSize:           quorumSize,
		totalNodes:           totalNodes,
		timeout:              timeout,
		convergenceThreshold: 0.01,

		// Initialize blockchain components (NEW)
		blockchain:      &blockchain.BlockChain{},
		blockchainState: blockchain.NewStateDatabase(),
		validators:      blockchain.NewValidatorSet(),
		mempool:         blockchain.NewMempool(),
		roundNumber:     0,
	}
}

// SetupBlockchainIntegration configures blockchain components (NEW)
func (c *Coordinator) SetupBlockchainIntegration(proposer *blockchain.BlockProposer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if proposer == nil {
		return fmt.Errorf("block proposer cannot be nil")
	}

	c.blockProposer = proposer
	return nil
}

// SetupContractExecutor configures a smart contract executor for governance actions.
func (c *Coordinator) SetupContractExecutor(executor *blockchain.SmartContractExecutor) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if executor == nil {
		return fmt.Errorf("contract executor cannot be nil")
	}

	c.contractExec = executor
	return nil
}

// RegisterValidator adds a node as a validator in the blockchain (NEW)
func (c *Coordinator) RegisterValidator(nodeID string, stake uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.validators.Stake(nodeID, stake)
}

// ProposeModel submits a new model update for consensus
func (c *Coordinator) ProposeModel(ctx context.Context, proposal *ModelProposal) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != Proposing {
		return "", fmt.Errorf("cannot propose: current state is %v", c.state)
	}

	proposalID := fmt.Sprintf("%s-%d-%d", proposal.ProposerID, proposal.Round, proposal.Timestamp.Unix())
	c.proposals[proposalID] = proposal
	c.votes[proposalID] = make([]*Vote, 0)

	// Transition to voting state
	c.state = Voting

	return proposalID, nil
}

// CastVote records a vote for a proposal
func (c *Coordinator) CastVote(ctx context.Context, vote *Vote) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != Voting {
		return fmt.Errorf("cannot vote: current state is %v", c.state)
	}

	// Verify proposal exists
	if _, exists := c.proposals[vote.ProposalID]; !exists {
		return fmt.Errorf("proposal %s not found", vote.ProposalID)
	}

	// Record vote
	c.votes[vote.ProposalID] = append(c.votes[vote.ProposalID], vote)

	return nil
}

// CheckConsensus determines if consensus has been reached
func (c *Coordinator) CheckConsensus(proposalID string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	votes, exists := c.votes[proposalID]
	if !exists {
		return false, fmt.Errorf("proposal %s not found", proposalID)
	}

	// Count affirmative votes
	approvalCount := 0
	for _, vote := range votes {
		if vote.Approve {
			approvalCount++
		}
	}

	// Check if quorum reached
	return approvalCount >= c.quorumSize, nil
}

// CommitModel finalizes the consensus and commits the model
func (c *Coordinator) CommitModel(ctx context.Context, proposalID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	votes, exists := c.votes[proposalID]
	if !exists {
		return fmt.Errorf("proposal %s not found", proposalID)
	}

	approvalCount := 0
	for _, vote := range votes {
		if vote.Approve {
			approvalCount++
		}
	}

	if approvalCount < c.quorumSize {
		c.state = Aborted
		return fmt.Errorf("consensus not reached: insufficient votes")
	}

	c.state = Committed

	// NEW: Create blockchain block for this consensus round
	if c.blockProposer != nil && c.proposals[proposalID] != nil {
		proposal := c.proposals[proposalID]
		c.roundNumber++

		// Prepare FL round data for blockchain
		roundData := map[string]interface{}{
			"round":               c.roundNumber,
			"proposer":            proposal.ProposerID,
			"weights_hash":        fmt.Sprintf("%x", proposal.Weights),
			"approval_vote_count": approvalCount,
			"total_votes":         len(votes),
			"timestamp":           time.Now().Unix(),
		}

		// Propose and commit block asynchronously
		go func() {
			// ProposeBlock includes FL round transaction and other mempool transactions
			if block, err := c.blockProposer.ProposeBlock(c.nodeID, roundData); err == nil && block != nil {
				// Commit block to blockchain
				if err := c.blockProposer.CommitBlock(block); err != nil {
					fmt.Printf("Error committing block: %v\n", err)
					return
				}

				// Get participating node IDs from votes
				participatingNodes := make([]string, 0, len(votes))
				for _, vote := range votes {
					participatingNodes = append(participatingNodes, vote.NodeID)
				}

				// Distribute rewards to participating validators
				blockHeight := uint64(c.roundNumber)
				baseReward := uint64(10000) // 10K tokens per round
				if err := c.blockProposer.DistributeFlRewards(blockHeight, participatingNodes, baseReward); err != nil {
					fmt.Printf("Error distributing rewards: %v\n", err)
				}
			} else if err != nil {
				fmt.Printf("Error proposing block: %v\n", err)
			}
		}()
	}

	return nil
}

// CommitGovernanceProposal finalizes consensus and executes a governance proposal via contract executor.
func (c *Coordinator) CommitGovernanceProposal(ctx context.Context, consensusProposalID string, contractAddress string, governanceProposalID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	votes, exists := c.votes[consensusProposalID]
	if !exists {
		return fmt.Errorf("proposal %s not found", consensusProposalID)
	}

	approvalCount := 0
	for _, vote := range votes {
		if vote.Approve {
			approvalCount++
		}
	}

	if approvalCount < c.quorumSize {
		c.state = Aborted
		return fmt.Errorf("consensus not reached: insufficient votes")
	}
	if c.contractExec == nil {
		return fmt.Errorf("contract executor not configured")
	}

	txn := &blockchain.Transaction{
		ID:        fmt.Sprintf("gov_exec_%s_%d", governanceProposalID, time.Now().UnixNano()),
		Type:      blockchain.TxTypeSmartContract,
		From:      c.nodeID,
		Nonce:     0,
		Gas:       300000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("consensus-governance"),
		Data: map[string]interface{}{
			"call":             true,
			"contract_address": contractAddress,
			"function":         "executeProposal",
			"params": map[string]interface{}{
				"proposal_id": governanceProposalID,
			},
		},
	}

	if err := c.contractExec.ExecuteContractTransaction(txn); err != nil {
		return fmt.Errorf("execute governance proposal: %w", err)
	}

	c.state = Committed
	return nil
}

// SubmitGovernancePolicyProposal creates a governance proposal transaction for policy updates.
func (c *Coordinator) SubmitGovernancePolicyProposal(
	ctx context.Context,
	contractAddress string,
	title string,
	description string,
	policy blockchain.ReputationPolicy,
	minVotes uint64,
) (string, error) {
	c.mu.RLock()
	exec := c.contractExec
	nodeID := c.nodeID
	c.mu.RUnlock()

	if exec == nil {
		return "", fmt.Errorf("contract executor not configured")
	}
	if contractAddress == "" {
		return "", fmt.Errorf("contract address is required")
	}
	if minVotes == 0 {
		minVotes = 1
	}

	proposalID := fmt.Sprintf("gov_policy_%d", time.Now().UnixNano())
	txn := &blockchain.Transaction{
		ID:        proposalID,
		Type:      blockchain.TxTypeSmartContract,
		From:      nodeID,
		Nonce:     0,
		Gas:       300000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("consensus-governance"),
		Data: map[string]interface{}{
			"call":             true,
			"contract_address": contractAddress,
			"function":         "createProposal",
			"params": map[string]interface{}{
				"title":       title,
				"description": description,
				"action":      "set_reputation_policy",
				"min_votes":   minVotes,
				"params": map[string]interface{}{
					"slash_penalty":                        policy.SlashPenalty,
					"reward_gain":                          policy.RewardGain,
					"epoch_recovery":                       policy.EpochRecovery,
					"reputation_weight":                    policy.ReputationWeight,
					"attestation_weight":                   policy.AttestationWeight,
					"quality_weight":                       policy.QualityWeight,
					"max_attestation_age_blocks":           policy.MaxAttestationAgeBlocks,
					"invalid_attestation_penalty":          policy.InvalidAttestationPenalty,
					"invalid_attestation_slash_bps":        policy.InvalidAttestationSlashBps,
					"stale_attestation_slash_bps":          policy.StaleAttestationSlashBps,
					"max_consecutive_attestation_failures": policy.MaxConsecutiveAttestationFailures,
					"min_quality_score":                    policy.MinQualityScore,
					"quality_penalty":                      policy.QualityPenalty,
					"max_consecutive_low_quality":          policy.MaxConsecutiveLowQuality,
				},
			},
		},
	}

	if err := exec.ExecuteContractTransaction(txn); err != nil {
		return "", fmt.Errorf("submit governance policy proposal: %w", err)
	}

	return proposalID, nil
}

// CastGovernanceVote submits a governance vote transaction.
func (c *Coordinator) CastGovernanceVote(
	ctx context.Context,
	contractAddress string,
	voterID string,
	governanceProposalID string,
	support bool,
) error {
	c.mu.RLock()
	exec := c.contractExec
	c.mu.RUnlock()

	if exec == nil {
		return fmt.Errorf("contract executor not configured")
	}
	if contractAddress == "" || voterID == "" || governanceProposalID == "" {
		return fmt.Errorf("contract address, voter id, and proposal id are required")
	}

	txn := &blockchain.Transaction{
		ID:        fmt.Sprintf("gov_vote_%s_%d", voterID, time.Now().UnixNano()),
		Type:      blockchain.TxTypeSmartContract,
		From:      voterID,
		Nonce:     0,
		Gas:       200000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("consensus-governance"),
		Data: map[string]interface{}{
			"call":             true,
			"contract_address": contractAddress,
			"function":         "vote",
			"params": map[string]interface{}{
				"proposal_id": governanceProposalID,
				"support":     support,
			},
		},
	}

	if err := exec.ExecuteContractTransaction(txn); err != nil {
		return fmt.Errorf("cast governance vote: %w", err)
	}

	return nil
}

// CommitGovernancePolicy commits consensus and executes policy proposal.
func (c *Coordinator) CommitGovernancePolicy(ctx context.Context, consensusProposalID string, contractAddress string, governanceProposalID string) error {
	return c.CommitGovernanceProposal(ctx, consensusProposalID, contractAddress, governanceProposalID)
}

// GetState returns the current consensus state
func (c *Coordinator) GetState() ConsensusState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// Reset resets the coordinator for a new round
func (c *Coordinator) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.proposals = make(map[string]*ModelProposal)
	c.votes = make(map[string][]*Vote)
	c.state = Proposing
	// Note: roundNumber is NOT reset - it increments monotonically
}

// GetConsensusRound constructs a ConsensusRound from current consensus state (NEW)
func (c *Coordinator) GetConsensusRound(proposalID string) (*ConsensusRound, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	proposal, exists := c.proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal %s not found", proposalID)
	}

	votes := c.votes[proposalID]

	return &ConsensusRound{
		RoundNumber:    c.roundNumber,
		ProposerID:     proposal.ProposerID,
		ProposalID:     proposalID,
		ModelWeights:   proposal.Weights,
		CommitTime:     time.Now(),
		ValidatorVotes: votes,
	}, nil
}

// SubmitVerificationPolicyProposal creates a governance proposal to update the
// FL proof verification policy (proof type gates, confidence threshold, enforcement mode).
func (c *Coordinator) SubmitVerificationPolicyProposal(
	ctx context.Context,
	contractAddress string,
	title string,
	description string,
	policy blockchain.VerificationPolicy,
	minVotes uint64,
) (string, error) {
	c.mu.RLock()
	exec := c.contractExec
	nodeID := c.nodeID
	c.mu.RUnlock()

	if exec == nil {
		return "", fmt.Errorf("contract executor not configured")
	}
	if contractAddress == "" {
		return "", fmt.Errorf("contract address is required")
	}
	if minVotes == 0 {
		minVotes = 1
	}

	proposalID := fmt.Sprintf("gov_verif_policy_%d", time.Now().UnixNano())
	txn := &blockchain.Transaction{
		ID:        proposalID,
		Type:      blockchain.TxTypeSmartContract,
		From:      nodeID,
		Nonce:     0,
		Gas:       300000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("consensus-governance"),
		Data: map[string]interface{}{
			"call":             true,
			"contract_address": contractAddress,
			"function":         "createProposal",
			"params": map[string]interface{}{
				"title":       title,
				"description": description,
				"action":      "set_verification_policy",
				"min_votes":   minVotes,
				"params": map[string]interface{}{
					"require_proof":                  policy.RequireProof,
					"min_confidence_bps":             policy.MinConfidenceBps,
					"reject_on_verification_failure": policy.RejectOnVerificationFailure,
					"allow_consensus_proof":          policy.AllowConsensusProof,
					"allow_zk_proof":                 policy.AllowZKProof,
					"allow_tee_proof":                policy.AllowTEEProof,
				},
			},
		},
	}

	if err := exec.ExecuteContractTransaction(txn); err != nil {
		return "", fmt.Errorf("submit verification policy proposal: %w", err)
	}

	return proposalID, nil
}
