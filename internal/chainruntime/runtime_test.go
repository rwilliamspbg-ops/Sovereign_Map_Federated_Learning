package chainruntime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
)

func TestIngestGossipTransaction(t *testing.T) {
	r := &Runtime{
		chain:       blockchain.NewBlockChain(),
		knownTxns:   make(map[string]struct{}),
		knownBlocks: make(map[string]struct{}),
	}

	txn := blockchain.Transaction{
		ID:        "tx-1",
		Type:      blockchain.TxTypeTransfer,
		From:      "0xabc",
		To:        "0xdef",
		Nonce:     0,
		Amount:    100,
		Gas:       21000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig"),
	}
	rawTxn, err := json.Marshal(txn)
	if err != nil {
		t.Fatalf("marshal txn: %v", err)
	}

	rawEnvelope, err := json.Marshal(GossipEnvelope{Type: gossipTypeTransaction, Payload: rawTxn})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	if err := r.IngestGossip(rawEnvelope); err != nil {
		t.Fatalf("ingest gossip txn failed: %v", err)
	}

	if got := r.chain.Mempool.Size(); got != 1 {
		t.Fatalf("expected mempool size 1, got %d", got)
	}
}

func TestIngestGossipBlock(t *testing.T) {
	bc := blockchain.NewBlockChain()
	if err := bc.ValidatorSet.AddValidator("node-1", 1_000_000); err != nil {
		t.Fatalf("add validator: %v", err)
	}

	r := &Runtime{
		chain:       bc,
		knownTxns:   make(map[string]struct{}),
		knownBlocks: make(map[string]struct{}),
	}

	block := blockchain.Block{
		Header: blockchain.BlockHeader{
			Index:        1,
			Timestamp:    time.Now().Unix(),
			PreviousHash: bc.Tip.Header.Hash,
			ValidatorID:  "node-1",
			Difficulty:   1,
			Version:      1,
		},
		Transactions: []blockchain.Transaction{},
		MerkleRoot:   blockchain.ComputeMerkleRoot([]blockchain.Transaction{}),
		StateRoot:    bc.StateDB.ComputeRoot(),
	}
	block.Header.Hash = block.ComputeHash()

	rawBlock, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("marshal block: %v", err)
	}

	rawEnvelope, err := json.Marshal(GossipEnvelope{Type: gossipTypeBlock, Payload: rawBlock})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	if err := r.IngestGossip(rawEnvelope); err != nil {
		t.Fatalf("ingest gossip block failed: %v", err)
	}

	if got := r.chain.Height(); got != 1 {
		t.Fatalf("expected chain height 1, got %d", got)
	}
}
