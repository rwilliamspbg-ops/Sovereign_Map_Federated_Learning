// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// FLVerificationMetrics summarizes on-chain FL verification outcomes.
type FLVerificationMetrics struct {
	TotalRounds          uint64  `json:"total_rounds"`
	VerifiedRounds       uint64  `json:"verified_rounds"`
	FailedRounds         uint64  `json:"failed_rounds"`
	AverageConfidenceBps uint32  `json:"average_confidence_bps"`
	LastRoundID          string  `json:"last_round_id"`
	LastProofType        string  `json:"last_proof_type"`
	VerifiedRatio        float64 `json:"verified_ratio"`
}

// VerificationPolicy defines governance-tunable gates for FL proof acceptance.
type VerificationPolicy struct {
	RequireProof                bool   `json:"require_proof"`
	MinConfidenceBps            uint32 `json:"min_confidence_bps"`
	RejectOnVerificationFailure bool   `json:"reject_on_verification_failure"`
	AllowConsensusProof         bool   `json:"allow_consensus_proof"`
	AllowZKProof                bool   `json:"allow_zk_proof"`
	AllowTEEProof               bool   `json:"allow_tee_proof"`
}

// ProofVerificationRequest captures FL round verification context for pluggable verifiers.
type ProofVerificationRequest struct {
	RoundID   string
	ModelHash string
	ProofType string
	ProofHash string
	ProofData []byte
	Payload   map[string]interface{}
}

// ProofVerifier allows custom ZK/TEE/consensus proof validation.
type ProofVerifier interface {
	VerifyProof(req ProofVerificationRequest) (bool, uint32, error)
}

// PermissiveProofVerifier is a default verifier used until a strict provider is configured.
type PermissiveProofVerifier struct{}

// VerifyProof returns a permissive result while still producing confidence metadata.
func (v *PermissiveProofVerifier) VerifyProof(req ProofVerificationRequest) (bool, uint32, error) {
	if req.ProofHash == "" && len(req.ProofData) == 0 {
		return true, 6500, nil
	}
	return true, 9000, nil
}

// TransactionType defines the type of transaction
type TransactionType string

const (
	TxTypeFlRound       TransactionType = "fl_round"
	TxTypeStake         TransactionType = "stake"
	TxTypeUnstake       TransactionType = "unstake"
	TxTypeReward        TransactionType = "reward"
	TxTypeSmartContract TransactionType = "smart_contract"
	TxTypeCheckpoint    TransactionType = "checkpoint"
	TxTypeTransfer      TransactionType = "transfer" // token transfer between wallets
)

// Transaction represents an on-chain transaction
type Transaction struct {
	ID        string                 `json:"id"`
	Type      TransactionType        `json:"type"`
	From      string                 `json:"from"`
	To        string                 `json:"to,omitempty"`
	Nonce     uint64                 `json:"nonce"`
	Amount    uint64                 `json:"amount"`
	Gas       uint64                 `json:"gas"`
	GasPrice  uint64                 `json:"gas_price"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
	Signature []byte                 `json:"signature"`
}

// Validate checks if transaction is valid
func (t *Transaction) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("transaction missing ID")
	}
	if t.From == "" {
		return fmt.Errorf("transaction missing From address")
	}
	if t.Timestamp == 0 {
		return fmt.Errorf("transaction missing timestamp")
	}
	if len(t.Signature) == 0 {
		return fmt.Errorf("transaction missing signature")
	}
	return nil
}

// Hash returns the transaction hash
func (t *Transaction) Hash() string {
	data := fmt.Sprintf("%s:%s:%s:%d:%d:%d",
		t.ID, t.Type, t.From, t.Nonce, t.Amount, t.Timestamp)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// BlockHeader contains block metadata
type BlockHeader struct {
	Index        uint64 `json:"index"`
	Timestamp    int64  `json:"timestamp"`
	PreviousHash string `json:"previous_hash"`
	Hash         string `json:"hash"`
	Nonce        uint64 `json:"nonce"`
	ValidatorID  string `json:"validator_id"`
	Difficulty   uint32 `json:"difficulty"`
	Version      uint32 `json:"version"`
}

// Block represents a blockchain block
type Block struct {
	Header       BlockHeader       `json:"header"`
	Transactions []Transaction     `json:"transactions"`
	MerkleRoot   string            `json:"merkle_root"`
	StateRoot    string            `json:"state_root"`
	ProofData    map[string][]byte `json:"proof_data"`
}

// Validate checks if block is valid
func (b *Block) Validate() error {
	if b.Header.Index == 0 && b.Header.PreviousHash != "" {
		return fmt.Errorf("genesis block must have empty previous hash")
	}
	if b.Header.Hash == "" {
		return fmt.Errorf("block missing hash")
	}
	if b.Header.ValidatorID == "" {
		return fmt.Errorf("block missing validator ID")
	}
	// Validate all transactions
	for i, txn := range b.Transactions {
		if err := txn.Validate(); err != nil {
			return fmt.Errorf("transaction %d invalid: %w", i, err)
		}
	}
	return nil
}

// ComputeHash computes the block hash
func (b *Block) ComputeHash() string {
	data := fmt.Sprintf("%d:%d:%s:%s:%s",
		b.Header.Index,
		b.Header.Timestamp,
		b.Header.PreviousHash,
		b.MerkleRoot,
		b.StateRoot,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// BlockChain represents the full blockchain
type BlockChain struct {
	mu                 sync.RWMutex
	Blocks             []*Block
	PendingTxns        []Transaction
	blockIndex         map[string]*Block // hash -> block
	StateDB            *StateDatabase
	ValidatorSet       *ValidatorSet
	Mempool            *Mempool
	VerificationPolicy VerificationPolicy
	proofVerifier      ProofVerifier
	GenesisBlock       *Block
	Tip                *Block
}

// NewBlockChain creates a new blockchain instance
func NewBlockChain() *BlockChain {
	bc := &BlockChain{
		Blocks:       make([]*Block, 0),
		PendingTxns:  make([]Transaction, 0),
		blockIndex:   make(map[string]*Block),
		StateDB:      NewStateDatabase(),
		ValidatorSet: NewValidatorSet(),
		Mempool:      NewMempool(),
		VerificationPolicy: VerificationPolicy{
			RequireProof:                false,
			MinConfidenceBps:            6000,
			RejectOnVerificationFailure: false,
			AllowConsensusProof:         true,
			AllowZKProof:                true,
			AllowTEEProof:               true,
		},
		proofVerifier: &PermissiveProofVerifier{},
	}
	bc.createGenesisBlock()
	return bc
}

// createGenesisBlock initializes the genesis block
func (bc *BlockChain) createGenesisBlock() {
	genesis := &Block{
		Header: BlockHeader{
			Index:        0,
			Timestamp:    time.Now().Unix(),
			PreviousHash: "",
			ValidatorID:  "genesis",
			Difficulty:   1,
			Version:      1,
		},
		Transactions: []Transaction{},
		MerkleRoot:   "genesis",
		StateRoot:    bc.StateDB.ComputeRoot(),
	}
	genesis.Header.Hash = genesis.ComputeHash()

	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.Blocks = append(bc.Blocks, genesis)
	bc.blockIndex[genesis.Header.Hash] = genesis
	bc.GenesisBlock = genesis
	bc.Tip = genesis
}

// AppendBlock adds a new block to the chain
func (bc *BlockChain) AppendBlock(block *Block) error {
	if err := block.Validate(); err != nil {
		return err
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Verify previous hash
	if block.Header.PreviousHash != bc.Tip.Header.Hash {
		return fmt.Errorf("block previous hash does not match chain tip")
	}

	// Update block index and append to chain
	bc.Blocks = append(bc.Blocks, block)
	bc.blockIndex[block.Header.Hash] = block
	bc.Tip = block

	// Remove transactions from mempool
	txnMap := make(map[string]bool)
	for _, txn := range block.Transactions {
		txnMap[txn.ID] = true
	}
	newPending := make([]Transaction, 0)
	for _, txn := range bc.PendingTxns {
		if !txnMap[txn.ID] {
			newPending = append(newPending, txn)
		}
	}
	bc.PendingTxns = newPending

	return nil
}

// GetBlock retrieves a block by hash
func (bc *BlockChain) GetBlock(hash string) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	block, exists := bc.blockIndex[hash]
	if !exists {
		return nil, fmt.Errorf("block not found: %s", hash)
	}
	return block, nil
}

// GetBlockByHeight retrieves a block by height (index)
func (bc *BlockChain) GetBlockByHeight(index uint64) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if index >= uint64(len(bc.Blocks)) {
		return nil, fmt.Errorf("block height out of range: %d", index)
	}
	return bc.Blocks[index], nil
}

// Height returns the current blockchain height
func (bc *BlockChain) Height() uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if len(bc.Blocks) == 0 {
		return 0
	}
	return bc.Blocks[len(bc.Blocks)-1].Header.Index
}

// Length returns the number of blocks in chain
func (bc *BlockChain) Length() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.Blocks)
}

// ComputeMerkleRoot computes the Merkle root of transactions
func ComputeMerkleRoot(transactions []Transaction) string {
	if len(transactions) == 0 {
		return hashData("")
	}

	hashes := make([][]byte, len(transactions))
	for i, txn := range transactions {
		hash := sha256.Sum256([]byte(txn.Hash()))
		hashes[i] = hash[:]
	}

	for len(hashes) > 1 {
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}
		newHashes := make([][]byte, 0)
		for i := 0; i < len(hashes); i += 2 {
			combined := append(hashes[i], hashes[i+1]...)
			hash := sha256.Sum256(combined)
			newHashes = append(newHashes, hash[:])
		}
		hashes = newHashes
	}

	return hex.EncodeToString(hashes[0])
}

func hashData(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SetProofVerifier installs a custom verifier implementation.
func (bc *BlockChain) SetProofVerifier(verifier ProofVerifier) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if verifier == nil {
		bc.proofVerifier = &PermissiveProofVerifier{}
		return
	}
	bc.proofVerifier = verifier
}

// SetVerificationPolicy updates active FL proof gates.
func (bc *BlockChain) SetVerificationPolicy(policy VerificationPolicy) error {
	if policy.MinConfidenceBps > 10000 {
		return fmt.Errorf("min confidence bps out of range: %d", policy.MinConfidenceBps)
	}
	if !policy.AllowConsensusProof && !policy.AllowZKProof && !policy.AllowTEEProof {
		return fmt.Errorf("at least one proof type must be allowed")
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.VerificationPolicy = policy
	return nil
}

// GetVerificationPolicy returns the active verification policy.
func (bc *BlockChain) GetVerificationPolicy() VerificationPolicy {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.VerificationPolicy
}

// RefreshVerificationPolicyFromState syncs governance-applied policy from state database.
func (bc *BlockChain) RefreshVerificationPolicyFromState() {
	value, err := bc.StateDB.Get("verification_policy:active")
	if err != nil {
		return
	}
	policyMap, ok := value.(map[string]interface{})
	if !ok {
		return
	}

	current := bc.GetVerificationPolicy()
	updated := current
	updated.RequireProof = asBoolDefault(policyMap["require_proof"], current.RequireProof)
	updated.MinConfidenceBps = asUint32Default(policyMap["min_confidence_bps"], current.MinConfidenceBps)
	updated.RejectOnVerificationFailure = asBoolDefault(policyMap["reject_on_verification_failure"], current.RejectOnVerificationFailure)
	updated.AllowConsensusProof = asBoolDefault(policyMap["allow_consensus_proof"], current.AllowConsensusProof)
	updated.AllowZKProof = asBoolDefault(policyMap["allow_zk_proof"], current.AllowZKProof)
	updated.AllowTEEProof = asBoolDefault(policyMap["allow_tee_proof"], current.AllowTEEProof)

	_ = bc.SetVerificationPolicy(updated)
}

// EvaluateFLRoundVerification runs proof checks and policy gates, mutating roundData with verification fields.
func (bc *BlockChain) EvaluateFLRoundVerification(roundData map[string]interface{}) error {
	if roundData == nil {
		return fmt.Errorf("round data is required")
	}

	bc.mu.RLock()
	policy := bc.VerificationPolicy
	verifier := bc.proofVerifier
	bc.mu.RUnlock()

	roundID := asStringDefault(roundData["round_id"], "")
	modelHash := asStringDefault(roundData["model_hash"], "")
	proofType := strings.ToLower(asStringDefault(roundData["proof_type"], "consensus"))
	proofHash := asStringDefault(roundData["proof_hash"], "")

	proofBytes := []byte{}
	if raw, ok := roundData["consensus_proof"].([]byte); ok {
		proofBytes = raw
	}
	if proofHash == "" && len(proofBytes) > 0 {
		proofHash = hashData(string(proofBytes))
	}

	verified := true
	reason := "ok"
	confidenceBps := asUint32Default(roundData["verification_confidence_bps"], 8000)

	if !policy.allowsProofType(proofType) {
		verified = false
		reason = "proof type not allowed"
		confidenceBps = 0
	}

	if policy.RequireProof && proofHash == "" && len(proofBytes) == 0 {
		verified = false
		reason = "proof required but missing"
		confidenceBps = 0
	}

	if verified && verifier != nil {
		result, verifierConfidence, err := verifier.VerifyProof(ProofVerificationRequest{
			RoundID:   roundID,
			ModelHash: modelHash,
			ProofType: proofType,
			ProofHash: proofHash,
			ProofData: proofBytes,
			Payload:   roundData,
		})
		if err != nil {
			verified = false
			reason = fmt.Sprintf("proof verifier error: %v", err)
			confidenceBps = 0
		} else {
			verified = result
			confidenceBps = verifierConfidence
			if !verified {
				reason = "proof verification failed"
			}
		}
	}

	if verified && confidenceBps < policy.MinConfidenceBps {
		verified = false
		reason = fmt.Sprintf("confidence below policy minimum: %d < %d", confidenceBps, policy.MinConfidenceBps)
	}

	roundData["proof_type"] = proofType
	roundData["proof_hash"] = proofHash
	roundData["verification_passed"] = verified
	roundData["verification_confidence_bps"] = confidenceBps
	roundData["verification_reason"] = reason

	if policy.RejectOnVerificationFailure && !verified {
		return fmt.Errorf("verification rejected by policy: %s", reason)
	}

	return nil
}

func (p VerificationPolicy) allowsProofType(proofType string) bool {
	switch strings.ToLower(proofType) {
	case "consensus":
		return p.AllowConsensusProof
	case "zk", "zk-proof", "zk-proofs", "zkp":
		return p.AllowZKProof
	case "tee", "tpm", "attestation":
		return p.AllowTEEProof
	default:
		return false
	}
}

// BuildFLVerificationMetadata derives verification metadata from FL transaction payload.
func BuildFLVerificationMetadata(roundData map[string]interface{}, blockHeight uint64, ts int64) map[string]interface{} {
	roundID := fmt.Sprintf("%v", roundData["round_id"])
	proofType := asStringDefault(roundData["proof_type"], "consensus")
	proofHash := asStringDefault(roundData["proof_hash"], "")
	if proofHash == "" {
		proofHash = asStringDefault(roundData["model_hash"], "")
	}

	confidenceBps := asUint32Default(roundData["verification_confidence_bps"], 8000)
	verified := asBoolDefault(roundData["verification_passed"], true)

	return map[string]interface{}{
		"round_id":                    roundID,
		"proof_type":                  proofType,
		"proof_hash":                  proofHash,
		"verification_passed":         verified,
		"verification_confidence_bps": confidenceBps,
		"submitted_at":                asInt64Default(roundData["timestamp"], ts),
		"recorded_at":                 ts,
		"block_height":                blockHeight,
	}
}

// GetFLVerificationMetrics scans chain transactions to produce aggregate verification metrics.
func (bc *BlockChain) GetFLVerificationMetrics() FLVerificationMetrics {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	metrics := FLVerificationMetrics{}
	var confidenceSum uint64

	for _, block := range bc.Blocks {
		for _, txn := range block.Transactions {
			if txn.Type != TxTypeFlRound {
				continue
			}

			metrics.TotalRounds++
			roundID := fmt.Sprintf("%v", txn.Data["round_id"])
			if roundID != "" {
				metrics.LastRoundID = roundID
			}

			proofType := asStringDefault(txn.Data["proof_type"], "consensus")
			metrics.LastProofType = proofType

			if asBoolDefault(txn.Data["verification_passed"], true) {
				metrics.VerifiedRounds++
			} else {
				metrics.FailedRounds++
			}

			confidenceSum += uint64(asUint32Default(txn.Data["verification_confidence_bps"], 8000))
		}
	}

	if metrics.TotalRounds > 0 {
		metrics.AverageConfidenceBps = safeUint32(confidenceSum / metrics.TotalRounds)
		metrics.VerifiedRatio = float64(metrics.VerifiedRounds) / float64(metrics.TotalRounds)
	}

	return metrics
}

// CheckFLVerificationPolicy validates the embedded verification metadata in FL round
// transactions against the current active policy. Called by the block validator so
// that any block received from a peer is checked against the local node's policy.
func (bc *BlockChain) CheckFLVerificationPolicy(txns []Transaction) error {
	policy := bc.GetVerificationPolicy()
	if !policy.RejectOnVerificationFailure {
		return nil
	}
	for _, txn := range txns {
		if txn.Type != TxTypeFlRound {
			continue
		}
		roundID := asStringDefault(txn.Data["round_id"], txn.ID)
		passed := asBoolDefault(txn.Data["verification_passed"], true)
		confidenceBps := asUint32Default(txn.Data["verification_confidence_bps"], 8000)
		if !passed {
			return fmt.Errorf("FL round %q: verification_passed=false, policy rejects", roundID)
		}
		if confidenceBps < policy.MinConfidenceBps {
			return fmt.Errorf("FL round %q: confidence %d below policy minimum %d", roundID, confidenceBps, policy.MinConfidenceBps)
		}
	}
	return nil
}

func asStringDefault(v interface{}, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func asBoolDefault(v interface{}, fallback bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return fallback
}

func asUint32Default(v interface{}, fallback uint32) uint32 {
	switch n := v.(type) {
	case uint32:
		return n
	case uint64:
		if n <= math.MaxUint32 {
			return uint32(n)
		}
	case int:
		if n >= 0 && uint64(n) <= math.MaxUint32 {
			return uint32(n)
		}
	case float64:
		if n >= 0 && n <= math.MaxUint32 {
			return uint32(n)
		}
	}
	return fallback
}

func asInt64Default(v interface{}, fallback int64) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case uint64:
		if n <= math.MaxInt64 {
			return int64(n)
		}
	case float64:
		if n >= math.MinInt64 && n <= math.MaxInt64 {
			return int64(n)
		}
	}
	return fallback
}

func safeUint32(v uint64) uint32 {
	if v > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(v)
}
