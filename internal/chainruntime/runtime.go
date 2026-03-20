package chainruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
	meshruntime "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/node/network"
)

const (
	gossipTypeTransaction = "tx"
	gossipTypeBlock       = "block"
)

// GossipEnvelope is the wire format for chain messages shared over mesh gossip.
type GossipEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Runtime wires blockchain execution with mesh gossip transport.
type Runtime struct {
	mu sync.Mutex

	mesh     *meshruntime.Mesh
	chain    *blockchain.BlockChain
	proposer *blockchain.BlockProposer
	policies *BridgePolicySet

	knownTxns   map[string]struct{}
	knownBlocks map[string]struct{}
}

func New(mesh *meshruntime.Mesh, chain *blockchain.BlockChain, policies *BridgePolicySet) (*Runtime, error) {
	if mesh == nil {
		return nil, fmt.Errorf("mesh is required")
	}
	if chain == nil {
		return nil, fmt.Errorf("blockchain is required")
	}

	r := &Runtime{
		mesh:        mesh,
		chain:       chain,
		proposer:    blockchain.NewBlockProposer(mesh.PeerID(), chain),
		policies:    policies,
		knownTxns:   make(map[string]struct{}),
		knownBlocks: make(map[string]struct{}),
	}

	localValidatorID := mesh.PeerID()
	if localValidatorID != "" {
		if _, err := chain.ValidatorSet.GetValidator(localValidatorID); err != nil {
			_ = chain.ValidatorSet.AddValidator(localValidatorID, chain.ValidatorSet.MinStakeAmount)
		}
	}

	return r, nil
}

// Start begins a receive loop consuming tx and block gossip.
func (r *Runtime) Start(ctx context.Context) error {
	sub, err := r.mesh.Subscribe()
	if err != nil {
		return err
	}

	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				return
			}
			_ = r.IngestGossip(msg.Data)
		}
	}()

	return nil
}

// SubmitTransaction validates and gossips a transaction to the mesh.
func (r *Runtime) SubmitTransaction(ctx context.Context, txn *blockchain.Transaction) error {
	if txn == nil {
		return fmt.Errorf("transaction is required")
	}
	if err := txn.Validate(); err != nil {
		return err
	}

	if err := r.chain.Mempool.AddTransaction(txn); err != nil {
		return err
	}

	r.mu.Lock()
	r.knownTxns[txn.ID] = struct{}{}
	r.mu.Unlock()

	payload, err := json.Marshal(txn)
	if err != nil {
		return err
	}

	envelope, err := json.Marshal(GossipEnvelope{Type: gossipTypeTransaction, Payload: payload})
	if err != nil {
		return err
	}

	return r.mesh.Publish(ctx, envelope)
}

// SubmitBridgeTransfer validates bridge route policy before gossiping the transfer.
func (r *Runtime) SubmitBridgeTransfer(ctx context.Context, txn *blockchain.Transaction, req BridgeTransferRequest) error {
	if txn == nil {
		return fmt.Errorf("transaction is required")
	}
	if err := r.policies.ValidateTransfer(req); err != nil {
		return err
	}
	if txn.Data == nil {
		txn.Data = map[string]interface{}{}
	}
	txn.Data["bridge"] = true
	txn.Data["source_chain"] = req.SourceChain
	txn.Data["target_chain"] = req.TargetChain
	txn.Data["asset"] = req.Asset
	txn.Data["finality_blocks"] = req.FinalityBlocks
	return r.SubmitTransaction(ctx, txn)
}

// ProposeAndBroadcastBlock creates a block from current mempool and gossips it.
func (r *Runtime) ProposeAndBroadcastBlock(ctx context.Context, validatorID string, roundData map[string]interface{}) (*blockchain.Block, error) {
	block, err := r.proposer.ProposeBlock(validatorID, roundData)
	if err != nil {
		return nil, err
	}
	if err := r.proposer.CommitBlock(block); err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.knownBlocks[block.Header.Hash] = struct{}{}
	r.mu.Unlock()

	payload, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	envelope, err := json.Marshal(GossipEnvelope{Type: gossipTypeBlock, Payload: payload})
	if err != nil {
		return nil, err
	}

	if err := r.mesh.Publish(ctx, envelope); err != nil {
		return nil, err
	}

	return block, nil
}

// IngestGossip accepts a serialized GossipEnvelope and applies it locally.
func (r *Runtime) IngestGossip(raw []byte) error {
	var envelope GossipEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("decode gossip envelope: %w", err)
	}

	switch envelope.Type {
	case gossipTypeTransaction:
		var txn blockchain.Transaction
		if err := json.Unmarshal(envelope.Payload, &txn); err != nil {
			return fmt.Errorf("decode transaction gossip: %w", err)
		}
		if txn.ID == "" {
			return fmt.Errorf("transaction gossip missing id")
		}

		r.mu.Lock()
		_, seen := r.knownTxns[txn.ID]
		if !seen {
			r.knownTxns[txn.ID] = struct{}{}
		}
		r.mu.Unlock()
		if seen {
			return nil
		}

		if err := r.chain.Mempool.AddTransaction(&txn); err != nil {
			return fmt.Errorf("ingest tx into mempool: %w", err)
		}
		return nil

	case gossipTypeBlock:
		var block blockchain.Block
		if err := json.Unmarshal(envelope.Payload, &block); err != nil {
			return fmt.Errorf("decode block gossip: %w", err)
		}
		if block.Header.Hash == "" {
			return fmt.Errorf("block gossip missing hash")
		}

		r.mu.Lock()
		_, seen := r.knownBlocks[block.Header.Hash]
		if !seen {
			r.knownBlocks[block.Header.Hash] = struct{}{}
		}
		r.mu.Unlock()
		if seen {
			return nil
		}

		if err := r.chain.AppendBlock(&block); err != nil {
			return fmt.Errorf("append gossiped block: %w", err)
		}

		for i := range block.Transactions {
			txn := &block.Transactions[i]
			switch txn.Type {
			case blockchain.TxTypeFlRound:
				if err := r.proposer.ProcessFlRoundTransaction(txn); err != nil {
					return fmt.Errorf("apply gossiped fl round transaction: %w", err)
				}
			case blockchain.TxTypeTransfer:
				if err := r.proposer.ProcessTransferTransaction(txn); err != nil {
					return fmt.Errorf("apply gossiped transfer transaction: %w", err)
				}
			case blockchain.TxTypeReward:
				if err := r.proposer.ProcessRewardTransaction(txn); err != nil {
					return fmt.Errorf("apply gossiped reward transaction: %w", err)
				}
			}
		}
		r.chain.StateDB.RecordSnapshot(block.Header.Index)
		return nil

	default:
		return fmt.Errorf("unsupported gossip type: %s", envelope.Type)
	}
}
