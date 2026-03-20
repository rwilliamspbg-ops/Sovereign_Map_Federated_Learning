// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"fmt"
	"log"
	"time"
)

// BlockProposer coordinates block production in a federated learning round
type BlockProposer struct {
	nodeID       string
	blockchain   *BlockChain
	validator    *BlockValidator
	miners       []string // Active miners/validators
	currentIndex int
}

// NewBlockProposer creates a new block proposer
func NewBlockProposer(nodeID string, bc *BlockChain) *BlockProposer {
	return &BlockProposer{
		nodeID:     nodeID,
		blockchain: bc,
		validator:  NewBlockValidator(bc, bc.StateDB),
	}
}

// ProposeBlock creates and proposes a new block
func (bp *BlockProposer) ProposeBlock(validatorID string, roundData map[string]interface{}) (*Block, error) {
	bp.blockchain.RefreshVerificationPolicyFromState()
	if err := bp.blockchain.EvaluateFLRoundVerification(roundData); err != nil {
		return nil, fmt.Errorf("FL verification gate failed: %w", err)
	}

	// Get pending transactions from mempool
	selectedTxns := bp.blockchain.Mempool.GetTransactionsForBlock(1000)

	// Add FL round transaction
	flTxn := Transaction{
		ID:        fmt.Sprintf("fl_%d_%s", bp.blockchain.Height()+1, validatorID),
		Type:      TxTypeFlRound,
		From:      validatorID,
		Nonce:     uint64(bp.blockchain.Height()),
		Amount:    0,
		Timestamp: time.Now().Unix(),
		Signature: []byte("system-fl-round"),
		Data:      roundData,
	}
	selectedTxns = append([]Transaction{flTxn}, selectedTxns...)

	// Compute Merkle root
	merkleRoot := ComputeMerkleRoot(selectedTxns)

	// Compute projected state root after applying selected transactions.
	stateRoot := bp.computeProjectedStateRoot(selectedTxns)

	// Create block
	block := &Block{
		Header: BlockHeader{
			Index:        bp.blockchain.Height() + 1,
			Timestamp:    time.Now().Unix(),
			PreviousHash: bp.blockchain.Tip.Header.Hash,
			ValidatorID:  validatorID,
			Difficulty:   1,
			Version:      1,
		},
		Transactions: selectedTxns,
		MerkleRoot:   merkleRoot,
		StateRoot:    stateRoot,
		ProofData:    make(map[string][]byte),
	}

	// Compute and set block hash
	block.Header.Hash = block.ComputeHash()

	// Validate block before accepting
	if err := bp.validator.ValidateBlock(block); err != nil {
		return nil, fmt.Errorf("block validation failed: %w", err)
	}

	return block, nil
}

func (bp *BlockProposer) computeProjectedStateRoot(txns []Transaction) string {
	tempState := bp.blockchain.StateDB.Clone()
	ledger := NewWalletLedger(tempState)

	for _, txn := range txns {
		switch txn.Type {
		case TxTypeFlRound:
			flKey := fmt.Sprintf("fl_round:%s", txn.Data["round_id"])
			tempState.Set(flKey, txn.Data)
			verification := BuildFLVerificationMetadata(txn.Data, bp.blockchain.Height()+1, txn.Timestamp)
			verificationKey := fmt.Sprintf("fl_verification:%s", txn.Data["round_id"])
			tempState.Set(verificationKey, verification)

		case TxTypeStake:
			stakeKey := fmt.Sprintf("stake:%s", txn.From)
			currentStake := uint64(0)
			if val, err := tempState.Get(stakeKey); err == nil {
				currentStake = val.(uint64)
			}
			tempState.Set(stakeKey, currentStake+txn.Amount)

		case TxTypeUnstake:
			stakeKey := fmt.Sprintf("stake:%s", txn.From)
			if val, err := tempState.Get(stakeKey); err == nil {
				currentStake := val.(uint64)
				if currentStake >= txn.Amount {
					tempState.Set(stakeKey, currentStake-txn.Amount)
				}
			}

		case TxTypeReward:
			_ = ledger.Credit(txn.To, txn.Amount)
			rewardKey := fmt.Sprintf("reward:%s", txn.To)
			currentReward := uint64(0)
			if val, err := tempState.Get(rewardKey); err == nil {
				if parsed, ok := asUint64Value(val); ok {
					currentReward = parsed
				}
			}
			tempState.Set(rewardKey, currentReward+txn.Amount)

		case TxTypeTransfer:
			if err := applyTransferToState(tempState, ledger, &txn); err != nil {
				// Invalid transfer does not mutate projected state; block validator will reject on replay.
				continue
			}

		case TxTypeSmartContract:
			contractKey := fmt.Sprintf("contract:%s", txn.ID)
			tempState.Set(contractKey, txn.Data)

		case TxTypeCheckpoint:
			checkpointKey := fmt.Sprintf("checkpoint:%s", txn.Data["checkpoint_hash"])
			tempState.Set(checkpointKey, txn.Data)
		}
	}

	tempState.IncrementVersion()
	return tempState.ComputeRoot()
}

// CommitBlock commits a block to the blockchain and distributes rewards
func (bp *BlockProposer) CommitBlock(block *Block) error {
	// Add block to chain
	if err := bp.blockchain.AppendBlock(block); err != nil {
		return fmt.Errorf("failed to append block: %w", err)
	}

	for i := range block.Transactions {
		txn := &block.Transactions[i]
		switch txn.Type {
		case TxTypeFlRound:
			if err := bp.ProcessFlRoundTransaction(txn); err != nil {
				return fmt.Errorf("failed to persist FL round transaction %s: %w", txn.ID, err)
			}
		case TxTypeTransfer:
			if err := bp.ProcessTransferTransaction(txn); err != nil {
				return fmt.Errorf("failed to persist transfer transaction %s: %w", txn.ID, err)
			}
		case TxTypeReward:
			if err := bp.ProcessRewardTransaction(txn); err != nil {
				return fmt.Errorf("failed to persist reward transaction %s: %w", txn.ID, err)
			}
		}
	}

	// Record state snapshot
	bp.blockchain.StateDB.RecordSnapshot(block.Header.Index)

	// Distribute block rewards to validator
	bp.blockchain.ValidatorSet.DistributeRewards(
		block.Header.Index,
		1000000, // 1M token base reward
		[]string{block.Header.ValidatorID},
	)

	log.Printf("Block #%d committed: %d transactions, validator: %s",
		block.Header.Index,
		len(block.Transactions),
		block.Header.ValidatorID,
	)

	return nil
}

// ProcessTransferTransaction executes transfer and bridge-transfer state updates.
func (bp *BlockProposer) ProcessTransferTransaction(txn *Transaction) error {
	if txn.Type != TxTypeTransfer {
		return fmt.Errorf("not a transfer transaction")
	}

	ledger := NewWalletLedger(bp.blockchain.StateDB)
	return applyTransferToState(bp.blockchain.StateDB, ledger, txn)
}

// ProcessRewardTransaction credits recipient balance and keeps reward accounting key in sync.
func (bp *BlockProposer) ProcessRewardTransaction(txn *Transaction) error {
	if txn.Type != TxTypeReward {
		return fmt.Errorf("not a reward transaction")
	}

	ledger := NewWalletLedger(bp.blockchain.StateDB)
	if err := ledger.Credit(txn.To, txn.Amount); err != nil {
		return err
	}

	rewardKey := fmt.Sprintf("reward:%s", txn.To)
	currentReward := uint64(0)
	if val, err := bp.blockchain.StateDB.Get(rewardKey); err == nil {
		if parsed, ok := asUint64Value(val); ok {
			currentReward = parsed
		}
	}
	return bp.blockchain.StateDB.Set(rewardKey, currentReward+txn.Amount)
}

func applyTransferToState(state *StateDatabase, ledger *WalletLedger, txn *Transaction) error {
	if isBridgeTransfer(txn) {
		sourceChain := asStringValue(txn.Data["source_chain"], "")
		targetChain := asStringValue(txn.Data["target_chain"], "")
		asset := asStringValue(txn.Data["asset"], "")
		if sourceChain == "" || targetChain == "" || asset == "" {
			return fmt.Errorf("bridge transfer metadata is incomplete")
		}

		escrowAddress := bridgeEscrowAddress(sourceChain, targetChain, asset)
		if err := ledger.Transfer(txn.From, escrowAddress, txn.Amount); err != nil {
			return err
		}

		bridgeRecord := map[string]interface{}{
			"tx_id":           txn.ID,
			"from":            txn.From,
			"to":              txn.To,
			"amount":          txn.Amount,
			"asset":           asset,
			"source_chain":    sourceChain,
			"target_chain":    targetChain,
			"finality_blocks": asUint64Default(txn.Data["finality_blocks"], 0),
			"status":          "escrowed",
			"timestamp":       txn.Timestamp,
		}
		if err := state.Set(fmt.Sprintf("bridge_transfer:%s", txn.ID), bridgeRecord); err != nil {
			return err
		}

		volumeKey := fmt.Sprintf("bridge_volume:%s:%s:%s", sourceChain, targetChain, asset)
		currentVolume := uint64(0)
		if val, err := state.Get(volumeKey); err == nil {
			if parsed, ok := asUint64Value(val); ok {
				currentVolume = parsed
			}
		}
		return state.Set(volumeKey, currentVolume+txn.Amount)
	}

	return ledger.ApplyTransaction(txn)
}

func isBridgeTransfer(txn *Transaction) bool {
	if txn == nil || txn.Data == nil {
		return false
	}
	return asBoolDefault(txn.Data["bridge"], false)
}

func bridgeEscrowAddress(sourceChain, targetChain, asset string) string {
	return fmt.Sprintf("bridge_escrow:%s:%s:%s", sourceChain, targetChain, asset)
}

func asStringValue(v interface{}, fallback string) string {
	if s, ok := v.(string); ok {
		if s != "" {
			return s
		}
	}
	return fallback
}

func asUint64Default(v interface{}, fallback uint64) uint64 {
	if parsed, ok := asUint64Value(v); ok {
		return parsed
	}
	return fallback
}

func asUint64Value(v interface{}) (uint64, bool) {
	switch n := v.(type) {
	case uint64:
		return n, true
	case uint32:
		return uint64(n), true
	case int:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case float64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	default:
		return 0, false
	}
}

// HandleFLRound processes a federated learning round and creates block
func (bp *BlockProposer) HandleFLRound(
	roundID string,
	aggregateModel []byte,
	modelHash string,
	accuracy float64,
	consensusProof []byte,
) (*Block, error) {
	// Create FL round data
	flRoundData := map[string]interface{}{
		"round_id":        roundID,
		"model_hash":      modelHash,
		"accuracy":        accuracy,
		"timestamp":       time.Now().Unix(),
		"proof_type":      "consensus",
		"proof_hash":      hashData(string(consensusProof)),
		"consensus_proof": consensusProof,
	}

	// Propose block containing FL round
	block, err := bp.ProposeBlock(bp.nodeID, flRoundData)
	if err != nil {
		return nil, err
	}

	// Commit block
	if err := bp.CommitBlock(block); err != nil {
		return nil, err
	}

	return block, nil
}

// ProcessFlRoundTransaction executes FL round transaction and updates state
func (bp *BlockProposer) ProcessFlRoundTransaction(txn *Transaction) error {
	if txn.Type != TxTypeFlRound {
		return fmt.Errorf("not an FL round transaction")
	}

	// Record FL round in state
	flKey := fmt.Sprintf("fl_round:%s", txn.Data["round_id"])
	if err := bp.blockchain.StateDB.Set(flKey, txn.Data); err != nil {
		return err
	}

	verification := BuildFLVerificationMetadata(txn.Data, bp.blockchain.Height(), txn.Timestamp)
	verificationKey := fmt.Sprintf("fl_verification:%s", txn.Data["round_id"])
	if err := bp.blockchain.StateDB.Set(verificationKey, verification); err != nil {
		return err
	}

	// Update accuracy metrics if available
	if accuracy, ok := txn.Data["accuracy"]; ok {
		metricsKey := "accuracy:latest"
		bp.blockchain.StateDB.Set(metricsKey, accuracy)
	}

	return nil
}

// DistributeFlRewards distributes rewards to participating nodes based on FL contribution.
// Each node receives an equal share of baseReward.
func (bp *BlockProposer) DistributeFlRewards(blockHeight uint64, participatingNodes []string, baseReward uint64) error {
	rewardPerNode := baseReward / uint64(len(participatingNodes))

	for _, nodeID := range participatingNodes {
		rewardKey := fmt.Sprintf("fl_reward:%s", nodeID)

		// Get existing reward
		currentReward := uint64(0)
		if val, err := bp.blockchain.StateDB.Get(rewardKey); err == nil {
			currentReward = val.(uint64)
		}

		// Add new reward
		if err := bp.blockchain.StateDB.Set(rewardKey, currentReward+rewardPerNode); err != nil {
			return err
		}
	}

	// Also distribute to validators
	validators := bp.blockchain.ValidatorSet.SelectValidators(1)
	if len(validators) > 0 {
		bp.blockchain.ValidatorSet.DistributeRewards(blockHeight, baseReward/10, []string{validators[0].NodeID})
	}

	return nil
}

// DistributeFlRewardsVerificationWeighted distributes FL rewards scaled by the
// verification confidence recorded for the given roundID.  Nodes whose round
// carries a high-confidence proof receive proportionally more; rounds with very
// low confidence (e.g. failed TEE attestation) receive a reduced payout while
// still honouring the minimum-reward guarantee.
//
// Reward formula: perNode = (baseReward * confidenceBps / 10000) / len(nodes)
// Minimum perNode is always 1 to prevent zero payouts.
func (bp *BlockProposer) DistributeFlRewardsVerificationWeighted(blockHeight uint64, roundID string, participatingNodes []string, baseReward uint64) error {
	if len(participatingNodes) == 0 {
		return nil
	}

	// Default to 80 % confidence if the state key is missing (e.g. genesis round).
	confidenceBps := uint64(8000)
	if val, err := bp.blockchain.StateDB.Get(fmt.Sprintf("fl_verification:%s", roundID)); err == nil {
		if data, ok := val.(map[string]interface{}); ok {
			switch cv := data["verification_confidence_bps"].(type) {
			case uint32:
				confidenceBps = uint64(cv)
			case float64:
				confidenceBps = uint64(cv)
			case int:
				confidenceBps = uint64(cv)
			}
		}
	}

	// Scale base reward by confidence and split across nodes.
	scaledReward := (baseReward * confidenceBps) / 10000
	if scaledReward == 0 {
		scaledReward = 1
	}
	rewardPerNode := scaledReward / uint64(len(participatingNodes))
	if rewardPerNode == 0 {
		rewardPerNode = 1
	}

	for _, nodeID := range participatingNodes {
		rewardKey := fmt.Sprintf("fl_reward:%s", nodeID)
		currentReward := uint64(0)
		if val, err := bp.blockchain.StateDB.Get(rewardKey); err == nil {
			currentReward = val.(uint64)
		}
		if err := bp.blockchain.StateDB.Set(rewardKey, currentReward+rewardPerNode); err != nil {
			return err
		}
	}

	// Validator commission: 10 % of the scaled round reward.
	validators := bp.blockchain.ValidatorSet.SelectValidators(1)
	if len(validators) > 0 {
		bp.blockchain.ValidatorSet.DistributeRewards(blockHeight, scaledReward/10, []string{validators[0].NodeID})
	}

	return nil
}

// GetBlockStats returns statistics about the blockchain
func (bp *BlockProposer) GetBlockStats() map[string]interface{} {
	verification := bp.blockchain.GetFLVerificationMetrics()
	return map[string]interface{}{
		"height":                             bp.blockchain.Height(),
		"total_blocks":                       bp.blockchain.Length(),
		"pending_txns":                       bp.blockchain.Mempool.Size(),
		"validators":                         bp.blockchain.ValidatorSet.Count(),
		"total_stake":                        bp.blockchain.ValidatorSet.TotalStake,
		"fl_total_rounds":                    verification.TotalRounds,
		"fl_verified_rounds":                 verification.VerifiedRounds,
		"fl_failed_rounds":                   verification.FailedRounds,
		"fl_verification_ratio":              verification.VerifiedRatio,
		"fl_avg_verification_confidence_bps": verification.AverageConfidenceBps,
		"fl_last_round_id":                   verification.LastRoundID,
		"fl_last_proof_type":                 verification.LastProofType,
		"state_root":                         bp.blockchain.StateDB.ComputeRoot(),
		"tip_hash":                           bp.blockchain.Tip.Header.Hash,
		"last_block_time":                    bp.blockchain.Tip.Header.Timestamp,
	}
}

// SyncWithPeer synchronizes blockchain state with a peer
func (bp *BlockProposer) SyncWithPeer(peerBlocks []*Block) error {
	// Validate block sequence
	if err := bp.validator.ValidateBlockSequence(peerBlocks); err != nil {
		return fmt.Errorf("peer blocks invalid: %w", err)
	}

	// Check if peer has longer chain
	if len(peerBlocks) > bp.blockchain.Length() {
		// Verify continuity from current tip
		currentTip := bp.blockchain.Tip
		firstPeerBlock := peerBlocks[0]

		if firstPeerBlock.Header.PreviousHash != currentTip.Header.Hash {
			// Peer has forked chain, need to handle fork resolution
			return fmt.Errorf("peer chain forked at block %d", firstPeerBlock.Header.Index)
		}

		// Sync blocks from peer
		for _, block := range peerBlocks {
			if err := bp.blockchain.AppendBlock(block); err != nil {
				return fmt.Errorf("failed to sync block %d: %w", block.Header.Index, err)
			}
		}
	}

	return nil
}

// RollbackToBlock rolls back blockchain to a specific block (for handling forks)
func (bp *BlockProposer) RollbackToBlock(targetIndex uint64) error {
	if targetIndex >= bp.blockchain.Height() {
		return fmt.Errorf("cannot rollback to future block %d", targetIndex)
	}

	// Truncate blockchain
	bp.blockchain.mu.Lock()
	defer bp.blockchain.mu.Unlock()

	if targetIndex < uint64(len(bp.blockchain.Blocks)) {
		bp.blockchain.Blocks = bp.blockchain.Blocks[:targetIndex+1]
		bp.blockchain.Tip = bp.blockchain.Blocks[len(bp.blockchain.Blocks)-1]
	}

	// Rebuild block index
	for _, block := range bp.blockchain.Blocks {
		bp.blockchain.blockIndex[block.Header.Hash] = block
	}

	return nil
}
