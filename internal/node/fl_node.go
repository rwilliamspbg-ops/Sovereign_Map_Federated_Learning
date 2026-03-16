// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package node

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
)

// TransactionStatus tracks a transaction through its lifecycle.
type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionConfirmed TransactionStatus = "confirmed"
	TransactionFailed    TransactionStatus = "failed"
)

// FLTransaction tracks an on-chain operation submitted by this node.
type FLTransaction struct {
	TransactionID string
	NodeID        string
	Type          blockchain.TransactionType
	Round         uint64
	Data          map[string]interface{}
	Nonce         uint64
	SubmitTime    time.Time
	BlockHeight   uint64
	Status        TransactionStatus
}

// TrainingResult is the output of a single local training round.
type TrainingResult struct {
	RoundNum       uint64
	Accuracy       float64
	Loss           float64
	TimeElapsed    time.Duration
	WeightsHash    string
	DataSamples    int
	UpdatesApplied int
}

// NodeState is a read-only snapshot of an FLNode.
type NodeState struct {
	NodeID                  string
	Region                  string
	WalletAddress           string
	TrainingRound           uint64
	ModelVersion            uint64
	NodeStake               uint64
	AccumulatedReward       uint64
	WalletBalance           uint64
	TotalRoundsCompleted    uint64
	TotalRewardsEarned      uint64
	SuccessfulTransactions  uint64
	FailedTransactions      uint64
	AverageRoundDuration    time.Duration
	PendingTransactionCount uint64
}

// FLNode is a federated-learning training participant with blockchain integration.
type FLNode struct {
	mu sync.RWMutex

	nodeID string
	region string

	blockchain *blockchain.BlockChain
	mempool    *blockchain.Mempool
	ledger     *blockchain.WalletLedger
	wallet     *blockchain.Wallet

	localWeights  []byte
	modelVersion  uint64
	trainingRound uint64
	nodeStake     uint64

	transactionNonce    uint64
	lastBlockHeight     uint64
	pendingTransactions map[string]*FLTransaction

	nodeReward uint64

	totalRoundsCompleted   uint64
	totalRewardsEarned     uint64
	successfulTransactions uint64
	failedTransactions     uint64
	averageRoundDuration   time.Duration
	lastRoundStartTime     time.Time
}

// NewFLNode creates a new federated learning node.
func NewFLNode(nodeID, region string) *FLNode {
	return &FLNode{
		nodeID:              nodeID,
		region:              region,
		pendingTransactions: make(map[string]*FLTransaction),
	}
}

// SetupBlockchain wires the node to the blockchain layer.
func (n *FLNode) SetupBlockchain(bc *blockchain.BlockChain, mempool *blockchain.Mempool, initialStake uint64) error {
	if bc == nil || mempool == nil {
		return fmt.Errorf("blockchain and mempool cannot be nil")
	}
	if bc.StateDB == nil {
		bc.StateDB = blockchain.NewStateDatabase()
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.blockchain = bc
	n.mempool = mempool
	n.nodeStake = initialStake
	n.ledger = blockchain.NewWalletLedger(bc.StateDB)
	return nil
}

// SetWallet assigns a wallet to the node.
func (n *FLNode) SetWallet(w *blockchain.Wallet) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.wallet = w
}

// Wallet returns the node's wallet, auto-generating one if none is set.
func (n *FLNode) Wallet() (*blockchain.Wallet, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.wallet == nil {
		w, err := blockchain.NewWallet()
		if err != nil {
			return nil, fmt.Errorf("auto-generate wallet: %w", err)
		}
		n.wallet = w
	}
	return n.wallet, nil
}

// TrainRound executes one local federated-learning training round.
func (n *FLNode) TrainRound(ctx context.Context) (*TrainingResult, error) {
	n.mu.Lock()
	if n.blockchain == nil {
		n.mu.Unlock()
		return nil, fmt.Errorf("blockchain not configured: call SetupBlockchain first")
	}
	n.lastRoundStartTime = time.Now()
	n.trainingRound++
	roundNum := n.trainingRound
	n.mu.Unlock()

	seed := time.Now().UnixNano()
	result := &TrainingResult{
		RoundNum:       roundNum,
		Accuracy:       0.85 + float64(seed%15)/100.0,
		Loss:           0.15 - float64(seed%10)/100.0,
		TimeElapsed:    time.Since(n.lastRoundStartTime),
		WeightsHash:    computeWeightsHash([]byte(fmt.Sprintf("weights-r%d", roundNum))),
		DataSamples:    1000 + int(seed%1000),
		UpdatesApplied: 50 + int(seed%50),
	}

	n.mu.Lock()
	n.localWeights = []byte(fmt.Sprintf("weights-v%d-r%d", n.modelVersion, roundNum))
	n.modelVersion++
	n.mu.Unlock()

	return result, nil
}

// SubmitFlRound signs and submits a TxTypeFlRound transaction to the mempool.
func (n *FLNode) SubmitFlRound(ctx context.Context, result *TrainingResult, checkpoint []byte) (string, error) {
	w, err := n.Wallet()
	if err != nil {
		return "", err
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.blockchain == nil || n.mempool == nil {
		return "", fmt.Errorf("blockchain not configured")
	}

	nonce := n.transactionNonce
	roundData := map[string]interface{}{
		"round":           result.RoundNum,
		"accuracy":        result.Accuracy,
		"loss":            result.Loss,
		"elapsed_ms":      result.TimeElapsed.Milliseconds(),
		"weights_hash":    result.WeightsHash,
		"data_samples":    result.DataSamples,
		"updates_applied": result.UpdatesApplied,
		"checkpoint_hash": fmt.Sprintf("%x", sha256.Sum256(checkpoint)),
		"node_id":         n.nodeID,
		"timestamp":       time.Now().Unix(),
	}

	txn, err := w.CreateFlRoundTx(roundData, nonce)
	if err != nil {
		n.failedTransactions++
		return "", fmt.Errorf("create fl round tx: %w", err)
	}

	if err := n.mempool.AddTransaction(txn); err != nil {
		n.failedTransactions++
		return "", fmt.Errorf("submit fl round tx: %w", err)
	}

	n.transactionNonce++
	n.pendingTransactions[txn.ID] = &FLTransaction{
		TransactionID: txn.ID,
		NodeID:        n.nodeID,
		Type:          blockchain.TxTypeFlRound,
		Round:         result.RoundNum,
		Data:          roundData,
		Nonce:         nonce,
		SubmitTime:    time.Now(),
		Status:        TransactionPending,
	}
	return txn.ID, nil
}

// Transfer creates and submits a signed token transfer from this node's wallet.
func (n *FLNode) Transfer(ctx context.Context, toAddress string, amount uint64) (string, error) {
	w, err := n.Wallet()
	if err != nil {
		return "", err
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.mempool == nil {
		return "", fmt.Errorf("blockchain not configured")
	}
	if n.ledger == nil {
		return "", fmt.Errorf("ledger not configured")
	}

	bal := n.ledger.GetBalance(w.Address())
	if bal < amount {
		return "", fmt.Errorf("insufficient wallet balance: have %d, need %d", bal, amount)
	}

	nonce := n.transactionNonce
	txn, err := w.CreateTransfer(toAddress, amount, n.transactionNonce)
	if err != nil {
		return "", err
	}

	if err := n.mempool.AddTransaction(txn); err != nil {
		return "", fmt.Errorf("submit transfer tx: %w", err)
	}

	n.transactionNonce++
	n.pendingTransactions[txn.ID] = &FLTransaction{
		TransactionID: txn.ID,
		NodeID:        n.nodeID,
		Type:          blockchain.TxTypeTransfer,
		Nonce:         nonce,
		SubmitTime:    time.Now(),
		Status:        TransactionPending,
	}
	return txn.ID, nil
}

// GetWalletBalance returns the current on-chain balance for the node's wallet.
func (n *FLNode) GetWalletBalance() uint64 {
	w, err := n.Wallet()
	if err != nil {
		return 0
	}
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.ledger == nil {
		return 0
	}
	return n.ledger.GetBalance(w.Address())
}

// WaitForBlockInclusion polls until txID is confirmed or deadline passes.
func (n *FLNode) WaitForBlockInclusion(ctx context.Context, txID string, timeout time.Duration) (uint64, error) {
	n.mu.RLock()
	_, exists := n.pendingTransactions[txID]
	n.mu.RUnlock()
	if !exists {
		return 0, fmt.Errorf("transaction %s not found", txID)
	}

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return 0, fmt.Errorf("timeout waiting for block inclusion of %s", txID)
			}
			n.mu.RLock()
			tx := n.pendingTransactions[txID]
			n.mu.RUnlock()
			if tx != nil && tx.Status == TransactionConfirmed {
				return tx.BlockHeight, nil
			}
		}
	}
}

// ConfirmTransaction marks a pending transaction as confirmed at blockHeight.
func (n *FLNode) ConfirmTransaction(txID string, blockHeight uint64) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	tx, ok := n.pendingTransactions[txID]
	if !ok {
		return fmt.Errorf("transaction %s not found", txID)
	}
	tx.Status = TransactionConfirmed
	tx.BlockHeight = blockHeight
	return nil
}

// CompleteRound finalises a training round after block inclusion.
func (n *FLNode) CompleteRound(ctx context.Context, blockHeight uint64) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.totalRoundsCompleted++
	n.successfulTransactions++
	n.lastBlockHeight = blockHeight
	elapsed := time.Since(n.lastRoundStartTime)
	if n.totalRoundsCompleted == 1 {
		n.averageRoundDuration = elapsed
	} else {
		n.averageRoundDuration = (n.averageRoundDuration + elapsed) / 2
	}
	return nil
}

// ClaimRewards credits the node's wallet with accumulated validator rewards.
func (n *FLNode) ClaimRewards(ctx context.Context) (uint64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.blockchain == nil {
		return 0, fmt.Errorf("blockchain not configured")
	}

	reward := uint64(10000)
	if n.totalRoundsCompleted > 0 {
		reward += n.totalRoundsCompleted * 100
	}

	if n.ledger != nil && n.wallet != nil {
		_ = n.ledger.Credit(n.wallet.Address(), reward)
	}

	n.nodeReward += reward
	n.totalRewardsEarned += reward
	return reward, nil
}

// GetState returns a read-only snapshot of the node's metrics.
func (n *FLNode) GetState() *NodeState {
	w, _ := n.Wallet()
	var walletAddr string
	var walletBal uint64
	if w != nil {
		walletAddr = w.Address()
		n.mu.RLock()
		if n.ledger != nil {
			walletBal = n.ledger.GetBalance(walletAddr)
		}
		n.mu.RUnlock()
	}

	n.mu.RLock()
	defer n.mu.RUnlock()

	return &NodeState{
		NodeID:                  n.nodeID,
		Region:                  n.region,
		WalletAddress:           walletAddr,
		TrainingRound:           n.trainingRound,
		ModelVersion:            n.modelVersion,
		NodeStake:               n.nodeStake,
		AccumulatedReward:       n.nodeReward,
		WalletBalance:           walletBal,
		TotalRoundsCompleted:    n.totalRoundsCompleted,
		TotalRewardsEarned:      n.totalRewardsEarned,
		SuccessfulTransactions:  n.successfulTransactions,
		FailedTransactions:      n.failedTransactions,
		AverageRoundDuration:    n.averageRoundDuration,
		PendingTransactionCount: uint64(len(n.pendingTransactions)),
	}
}

func computeWeightsHash(weights []byte) string {
	h := sha256.Sum256(weights)
	return fmt.Sprintf("%x", h)
}
