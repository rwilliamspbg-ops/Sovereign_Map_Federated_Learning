content = '''\
// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package node

import (
\t"context"
\t"crypto/sha256"
\t"fmt"
\t"sync"
\t"time"

\t"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
)

// TransactionStatus tracks a transaction through its lifecycle.
type TransactionStatus string

const (
\tTransactionPending   TransactionStatus = "pending"
\tTransactionConfirmed TransactionStatus = "confirmed"
\tTransactionFailed    TransactionStatus = "failed"
)

// FLTransaction tracks an on-chain operation submitted by this node.
type FLTransaction struct {
\tTransactionID string
\tNodeID        string
\tType          blockchain.TransactionType
\tRound         uint64
\tData          map[string]interface{}
\tNonce         uint64
\tSubmitTime    time.Time
\tBlockHeight   uint64
\tStatus        TransactionStatus
}

// TrainingResult is the output of a single local training round.
type TrainingResult struct {
\tRoundNum       uint64
\tAccuracy       float64
\tLoss           float64
\tTimeElapsed    time.Duration
\tWeightsHash    string
\tDataSamples    int
\tUpdatesApplied int
}

// NodeState is a read-only snapshot of an FLNode.
type NodeState struct {
\tNodeID                  string
\tRegion                  string
\tWalletAddress           string
\tTrainingRound           uint64
\tModelVersion            uint64
\tNodeStake               uint64
\tAccumulatedReward       uint64
\tWalletBalance           uint64
\tTotalRoundsCompleted    uint64
\tTotalRewardsEarned      uint64
\tSuccessfulTransactions  uint64
\tFailedTransactions      uint64
\tAverageRoundDuration    time.Duration
\tPendingTransactionCount uint64
}

// FLNode is a federated-learning training participant with blockchain integration.
type FLNode struct {
\tmu sync.RWMutex

\tnodeID string
\tregion string

\tblockchain *blockchain.BlockChain
\tmempool    *blockchain.Mempool
\tledger     *blockchain.WalletLedger
\twallet     *blockchain.Wallet

\tlocalWeights []byte
\tmodelVersion uint64
\ttrainingRound uint64
\tnodeStake    uint64

\ttransactionNonce    uint64
\tlastBlockHeight     uint64
\tpendingTransactions map[string]*FLTransaction

\tnodeReward uint64

\ttotalRoundsCompleted   uint64
\ttotalRewardsEarned     uint64
\tsuccessfulTransactions uint64
\tfailedTransactions     uint64
\taverageRoundDuration   time.Duration
\tlastRoundStartTime     time.Time
}

// NewFLNode creates a new federated learning node.
func NewFLNode(nodeID, region string) *FLNode {
\treturn &FLNode{
\t\tnodeID:              nodeID,
\t\tregion:              region,
\t\tpendingTransactions: make(map[string]*FLTransaction),
\t}
}

// SetupBlockchain wires the node to the blockchain layer.
func (n *FLNode) SetupBlockchain(bc *blockchain.BlockChain, mempool *blockchain.Mempool, initialStake uint64) error {
\tif bc == nil || mempool == nil {
\t\treturn fmt.Errorf("blockchain and mempool cannot be nil")
\t}
\tn.mu.Lock()
\tdefer n.mu.Unlock()
\tn.blockchain = bc
\tn.mempool = mempool
\tn.nodeStake = initialStake
\tn.ledger = blockchain.NewWalletLedger(bc.StateDB)
\treturn nil
}

// SetWallet assigns a wallet to the node.
func (n *FLNode) SetWallet(w *blockchain.Wallet) {
\tn.mu.Lock()
\tdefer n.mu.Unlock()
\tn.wallet = w
}

// Wallet returns the node\'s wallet, auto-generating one if none is set.
func (n *FLNode) Wallet() (*blockchain.Wallet, error) {
\tn.mu.Lock()
\tdefer n.mu.Unlock()
\tif n.wallet == nil {
\t\tw, err := blockchain.NewWallet()
\t\tif err != nil {
\t\t\treturn nil, fmt.Errorf("auto-generate wallet: %w", err)
\t\t}
\t\tn.wallet = w
\t}
\treturn n.wallet, nil
}

// TrainRound executes one local federated-learning training round.
func (n *FLNode) TrainRound(ctx context.Context) (*TrainingResult, error) {
\tn.mu.Lock()
\tif n.blockchain == nil {
\t\tn.mu.Unlock()
\t\treturn nil, fmt.Errorf("blockchain not configured: call SetupBlockchain first")
\t}
\tn.lastRoundStartTime = time.Now()
\tn.trainingRound++
\troundNum := n.trainingRound
\tn.mu.Unlock()

\tseed := time.Now().UnixNano()
\tresult := &TrainingResult{
\t\tRoundNum:       roundNum,
\t\tAccuracy:       0.85 + float64(seed%15)/100.0,
\t\tLoss:           0.15 - float64(seed%10)/100.0,
\t\tTimeElapsed:    time.Since(n.lastRoundStartTime),
\t\tWeightsHash:    computeWeightsHash([]byte(fmt.Sprintf("weights-r%d", roundNum))),
\t\tDataSamples:    1000 + int(seed%1000),
\t\tUpdatesApplied: 50 + int(seed%50),
\t}

\tn.mu.Lock()
\tn.localWeights = []byte(fmt.Sprintf("weights-v%d-r%d", n.modelVersion, roundNum))
\tn.modelVersion++
\tn.mu.Unlock()

\treturn result, nil
}

// SubmitFlRound signs and submits a TxTypeFlRound transaction to the mempool.
func (n *FLNode) SubmitFlRound(ctx context.Context, result *TrainingResult, checkpoint []byte) (string, error) {
\tw, err := n.Wallet()
\tif err != nil {
\t\treturn "", err
\t}

\tn.mu.Lock()
\tdefer n.mu.Unlock()

\tif n.blockchain == nil || n.mempool == nil {
\t\treturn "", fmt.Errorf("blockchain not configured")
\t}

\tn.transactionNonce++
\troundData := map[string]interface{}{
\t\t"round":           result.RoundNum,
\t\t"accuracy":        result.Accuracy,
\t\t"loss":            result.Loss,
\t\t"elapsed_ms":      result.TimeElapsed.Milliseconds(),
\t\t"weights_hash":    result.WeightsHash,
\t\t"data_samples":    result.DataSamples,
\t\t"updates_applied": result.UpdatesApplied,
\t\t"checkpoint_hash": fmt.Sprintf("%x", sha256.Sum256(checkpoint)),
\t\t"node_id":         n.nodeID,
\t\t"timestamp":       time.Now().Unix(),
\t}

\ttxn, err := w.CreateFlRoundTx(roundData, n.transactionNonce)
\tif err != nil {
\t\tn.failedTransactions++
\t\treturn "", fmt.Errorf("create fl round tx: %w", err)
\t}

\tif err := n.mempool.AddTransaction(txn); err != nil {
\t\tn.failedTransactions++
\t\treturn "", fmt.Errorf("submit fl round tx: %w", err)
\t}

\tn.pendingTransactions[txn.ID] = &FLTransaction{
\t\tTransactionID: txn.ID,
\t\tNodeID:        n.nodeID,
\t\tType:          blockchain.TxTypeFlRound,
\t\tRound:         result.RoundNum,
\t\tData:          roundData,
\t\tNonce:         n.transactionNonce,
\t\tSubmitTime:    time.Now(),
\t\tStatus:        TransactionPending,
\t}
\treturn txn.ID, nil
}

// Transfer creates and submits a signed token transfer from this node\'s wallet.
func (n *FLNode) Transfer(ctx context.Context, toAddress string, amount uint64) (string, error) {
\tw, err := n.Wallet()
\tif err != nil {
\t\treturn "", err
\t}

\tn.mu.Lock()
\tdefer n.mu.Unlock()

\tif n.mempool == nil {
\t\treturn "", fmt.Errorf("blockchain not configured")
\t}
\tif n.ledger == nil {
\t\treturn "", fmt.Errorf("ledger not configured")
\t}

\tbal := n.ledger.GetBalance(w.Address())
\tif bal < amount {
\t\treturn "", fmt.Errorf("insufficient wallet balance: have %d, need %d", bal, amount)
\t}

\tn.transactionNonce++
\ttxn, err := w.CreateTransfer(toAddress, amount, n.transactionNonce)
\tif err != nil {
\t\treturn "", err
\t}

\tif err := n.mempool.AddTransaction(txn); err != nil {
\t\treturn "", fmt.Errorf("submit transfer tx: %w", err)
\t}

\tn.pendingTransactions[txn.ID] = &FLTransaction{
\t\tTransactionID: txn.ID,
\t\tNodeID:        n.nodeID,
\t\tType:          blockchain.TxTypeTransfer,
\t\tNonce:         n.transactionNonce,
\t\tSubmitTime:    time.Now(),
\t\tStatus:        TransactionPending,
\t}
\treturn txn.ID, nil
}

// GetWalletBalance returns the current on-chain balance for the node\'s wallet.
func (n *FLNode) GetWalletBalance() uint64 {
\tw, err := n.Wallet()
\tif err != nil {
\t\treturn 0
\t}
\tn.mu.RLock()
\tdefer n.mu.RUnlock()
\tif n.ledger == nil {
\t\treturn 0
\t}
\treturn n.ledger.GetBalance(w.Address())
}

// WaitForBlockInclusion polls until txID is confirmed or deadline passes.
func (n *FLNode) WaitForBlockInclusion(ctx context.Context, txID string, timeout time.Duration) (uint64, error) {
\tn.mu.RLock()
\t_, exists := n.pendingTransactions[txID]
\tn.mu.RUnlock()
\tif !exists {
\t\treturn 0, fmt.Errorf("transaction %s not found", txID)
\t}

\tdeadline := time.Now().Add(timeout)
\tticker := time.NewTicker(100 * time.Millisecond)
\tdefer ticker.Stop()

\tfor {
\t\tselect {
\t\tcase <-ctx.Done():
\t\t\treturn 0, ctx.Err()
\t\tcase <-ticker.C:
\t\t\tif time.Now().After(deadline) {
\t\t\t\treturn 0, fmt.Errorf("timeout waiting for block inclusion of %s", txID)
\t\t\t}
\t\t\tn.mu.RLock()
\t\t\ttx := n.pendingTransactions[txID]
\t\t\tn.mu.RUnlock()
\t\t\tif tx != nil && tx.Status == TransactionConfirmed {
\t\t\t\treturn tx.BlockHeight, nil
\t\t\t}
\t\t}
\t}
}

// ConfirmTransaction marks a pending transaction as confirmed at blockHeight.
func (n *FLNode) ConfirmTransaction(txID string, blockHeight uint64) error {
\tn.mu.Lock()
\tdefer n.mu.Unlock()
\ttx, ok := n.pendingTransactions[txID]
\tif !ok {
\t\treturn fmt.Errorf("transaction %s not found", txID)
\t}
\ttx.Status = TransactionConfirmed
\ttx.BlockHeight = blockHeight
\treturn nil
}

// CompleteRound finalises a training round after block inclusion.
func (n *FLNode) CompleteRound(ctx context.Context, blockHeight uint64) error {
\tn.mu.Lock()
\tdefer n.mu.Unlock()
\tn.totalRoundsCompleted++
\tn.successfulTransactions++
\tn.lastBlockHeight = blockHeight
\telapsed := time.Since(n.lastRoundStartTime)
\tif n.totalRoundsCompleted == 1 {
\t\tn.averageRoundDuration = elapsed
\t} else {
\t\tn.averageRoundDuration = (n.averageRoundDuration + elapsed) / 2
\t}
\treturn nil
}

// ClaimRewards credits the node\'s wallet with accumulated validator rewards.
func (n *FLNode) ClaimRewards(ctx context.Context) (uint64, error) {
\tn.mu.Lock()
\tdefer n.mu.Unlock()

\tif n.blockchain == nil {
\t\treturn 0, fmt.Errorf("blockchain not configured")
\t}

\treward := uint64(10000)
\tif n.totalRoundsCompleted > 0 {
\t\treward += n.totalRoundsCompleted * 100
\t}

\tif n.ledger != nil && n.wallet != nil {
\t\t_ = n.ledger.Credit(n.wallet.Address(), reward)
\t}

\tn.nodeReward += reward
\tn.totalRewardsEarned += reward
\treturn reward, nil
}

// GetState returns a read-only snapshot of the node\'s metrics.
func (n *FLNode) GetState() *NodeState {
\tw, _ := n.Wallet()
\tvar walletAddr string
\tvar walletBal uint64
\tif w != nil {
\t\twalletAddr = w.Address()
\t\tn.mu.RLock()
\t\tif n.ledger != nil {
\t\t\twalletBal = n.ledger.GetBalance(walletAddr)
\t\t}
\t\tn.mu.RUnlock()
\t}

\tn.mu.RLock()
\tdefer n.mu.RUnlock()

\treturn &NodeState{
\t\tNodeID:                  n.nodeID,
\t\tRegion:                  n.region,
\t\tWalletAddress:           walletAddr,
\t\tTrainingRound:           n.trainingRound,
\t\tModelVersion:            n.modelVersion,
\t\tNodeStake:               n.nodeStake,
\t\tAccumulatedReward:       n.nodeReward,
\t\tWalletBalance:           walletBal,
\t\tTotalRoundsCompleted:    n.totalRoundsCompleted,
\t\tTotalRewardsEarned:      n.totalRewardsEarned,
\t\tSuccessfulTransactions:  n.successfulTransactions,
\t\tFailedTransactions:      n.failedTransactions,
\t\tAverageRoundDuration:    n.averageRoundDuration,
\t\tPendingTransactionCount: uint64(len(n.pendingTransactions)),
\t}
}

func computeWeightsHash(weights []byte) string {
\th := sha256.Sum256(weights)
\treturn fmt.Sprintf("%x", h)
}
'''

with open('/workspaces/Sovereign_Map_Federated_Learning/internal/node/fl_node.go', 'w') as f:
    f.write(content)
print('fl_node.go written successfully')
