package consensus

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test200NodeQuorumCalculation(t *testing.T) {
	totalNodes := 200
	expectedQuorum := 134
	calculatedQuorum := (2*totalNodes)/3 + 1
	if calculatedQuorum != expectedQuorum {
		t.Fatalf("expected quorum %d, got %d", expectedQuorum, calculatedQuorum)
	}
	if expectedQuorum <= totalNodes/2 {
		t.Fatalf("expected quorum to exceed 50%% threshold, got %d", expectedQuorum)
	}
}

func Test200NodeConsensusCommit(t *testing.T) {
	ctx := context.Background()
	coord := NewCoordinator("node-main", 200, 5*time.Second)

	proposalID, err := coord.ProposeModel(ctx, &ModelProposal{
		Round:      1,
		Weights:    []byte("weights"),
		ProposerID: "node-main",
		Proof:      []byte("proof"),
		Timestamp:  time.Now(),
	})
	if err != nil {
		t.Fatalf("propose model failed: %v", err)
	}

	for i := 1; i <= 134; i++ {
		err := coord.CastVote(ctx, &Vote{
			NodeID:     fmt.Sprintf("node-%03d", i),
			ProposalID: proposalID,
			Approve:    true,
			Signature:  []byte("sig"),
			Timestamp:  time.Now(),
		})
		if err != nil {
			t.Fatalf("cast vote failed at vote %d: %v", i, err)
		}
	}

	consensus, err := coord.CheckConsensus(proposalID)
	if err != nil {
		t.Fatalf("check consensus failed: %v", err)
	}
	if !consensus {
		t.Fatal("expected consensus with 134 approvals")
	}

	if err := coord.CommitModel(ctx, proposalID); err != nil {
		t.Fatalf("commit model failed: %v", err)
	}
	if coord.GetState() != Committed {
		t.Fatalf("expected committed state, got %v", coord.GetState())
	}
}
