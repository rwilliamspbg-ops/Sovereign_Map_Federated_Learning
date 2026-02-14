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
)

// ModelProposal represents a proposed model update for consensus
type ModelProposal struct {
	Round       int
	Weights     []byte
	ProposerID  string
	Proof       []byte
	Timestamp   time.Time
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

// Coordinator manages distributed consensus for model aggregation
type Coordinator struct {
	mu              sync.RWMutex
	nodeID          string
	proposals       map[string]*ModelProposal
	votes           map[string][]*Vote
	state           ConsensusState
	quorumSize      int
	totalNodes      int
	timeout         time.Duration
	convergenceThreshold float64
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
	}
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
	
	consensus, err := c.CheckConsensus(proposalID)
	if err != nil {
		return err
	}
	
	if !consensus {
		c.state = Aborted
		return fmt.Errorf("consensus not reached: insufficient votes")
	}
	
	c.state = Committed
	return nil
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
}
