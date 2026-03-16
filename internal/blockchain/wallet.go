// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Wallet represents a blockchain wallet with a cryptographic key pair.
// It can create and sign transactions, and derive a stable on-chain address
// from its Ed25519 public key.
type Wallet struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	address    string
}

// WalletLedger wraps a StateDatabase to provide balance management for addresses.
type WalletLedger struct {
	state *StateDatabase
}

// NewWallet generates a new wallet with a fresh Ed25519 key pair.
func NewWallet() (*Wallet, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key pair: %w", err)
	}
	return &Wallet{
		privateKey: privKey,
		publicKey:  pubKey,
		address:    deriveAddress(pubKey),
	}, nil
}

// WalletFromSeed restores a wallet from a 32-byte private key seed.
func WalletFromSeed(seed []byte) (*Wallet, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid seed: need %d bytes, got %d", ed25519.SeedSize, len(seed))
	}
	privKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privKey.Public().(ed25519.PublicKey)
	return &Wallet{
		privateKey: privKey,
		publicKey:  pubKey,
		address:    deriveAddress(pubKey),
	}, nil
}

// Address returns the wallet's blockchain address (0x-prefixed hex, 20 bytes).
func (w *Wallet) Address() string {
	return w.address
}

// PublicKeyBytes returns the raw 32-byte Ed25519 public key.
func (w *Wallet) PublicKeyBytes() []byte {
	return []byte(w.publicKey)
}

// Seed returns the 32-byte private key seed suitable for persistent storage.
func (w *Wallet) Seed() []byte {
	return w.privateKey.Seed()
}

// Sign signs arbitrary data with the wallet's private key.
func (w *Wallet) Sign(data []byte) []byte {
	return ed25519.Sign(w.privateKey, data)
}

// VerifySignature verifies that sig was produced by the private key corresponding
// to pubKeyBytes over data.
func VerifySignature(data, sig, pubKeyBytes []byte) bool {
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return false
	}
	return ed25519.Verify(ed25519.PublicKey(pubKeyBytes), data, sig)
}

// SignTransaction attaches a valid signature to txn.
// The signing message is a deterministic serialisation of the transaction fields.
func (w *Wallet) SignTransaction(txn *Transaction) {
	txn.Signature = w.Sign(txSigningMessage(txn))
}

// CreateTransfer creates a signed TxTypeTransfer transaction.
func (w *Wallet) CreateTransfer(to string, amount, nonce uint64) (*Transaction, error) {
	if to == "" {
		return nil, fmt.Errorf("recipient address is empty")
	}
	if amount == 0 {
		return nil, fmt.Errorf("transfer amount must be > 0")
	}
	txn := &Transaction{
		ID:        fmt.Sprintf("tx-transfer-%s-%d-%d", shortAddr(w.address), nonce, time.Now().UnixNano()),
		Type:      TxTypeTransfer,
		From:      w.address,
		To:        to,
		Amount:    amount,
		Nonce:     nonce,
		Gas:  21000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Data:      map[string]interface{}{},
	}
	w.SignTransaction(txn)
	return txn, nil
}

// CreateStakeTx creates a signed TxTypeStake transaction.
func (w *Wallet) CreateStakeTx(amount, nonce uint64) (*Transaction, error) {
	if amount == 0 {
		return nil, fmt.Errorf("stake amount must be > 0")
	}
	txn := &Transaction{
		ID:        fmt.Sprintf("tx-stake-%s-%d-%d", shortAddr(w.address), nonce, time.Now().UnixNano()),
		Type:      TxTypeStake,
		From:      w.address,
		Amount:    amount,
		Nonce:     nonce,
		Gas:  50000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Data:      map[string]interface{}{"stake_amount": amount},
	}
	w.SignTransaction(txn)
	return txn, nil
}

// CreateUnstakeTx creates a signed TxTypeUnstake transaction.
func (w *Wallet) CreateUnstakeTx(amount, nonce uint64) (*Transaction, error) {
	if amount == 0 {
		return nil, fmt.Errorf("unstake amount must be > 0")
	}
	txn := &Transaction{
		ID:        fmt.Sprintf("tx-unstake-%s-%d-%d", shortAddr(w.address), nonce, time.Now().UnixNano()),
		Type:      TxTypeUnstake,
		From:      w.address,
		Amount:    amount,
		Nonce:     nonce,
		Gas:  50000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Data:      map[string]interface{}{"unstake_amount": amount},
	}
	w.SignTransaction(txn)
	return txn, nil
}

// CreateFlRoundTx creates a signed TxTypeFlRound transaction for a training round.
func (w *Wallet) CreateFlRoundTx(roundData map[string]interface{}, nonce uint64) (*Transaction, error) {
	txn := &Transaction{
		ID:        fmt.Sprintf("tx-flround-%s-%d-%d", shortAddr(w.address), nonce, time.Now().UnixNano()),
		Type:      TxTypeFlRound,
		From:      w.address,
		Nonce:     nonce,
		Gas:  200000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Data:      roundData,
	}
	w.SignTransaction(txn)
	return txn, nil
}

// VerifyTxSignature checks that a transaction's embedded signature is valid
// against the given public key bytes.
func VerifyTxSignature(txn *Transaction, pubKeyBytes []byte) bool {
	return VerifySignature(txSigningMessage(txn), txn.Signature, pubKeyBytes)
}

// ---------------------------------------------------------------------------
// WalletLedger – balance management via StateDatabase
// ---------------------------------------------------------------------------

// NewWalletLedger creates a ledger backed by the given state database.
func NewWalletLedger(state *StateDatabase) *WalletLedger {
	return &WalletLedger{state: state}
}

// GetBalance returns the token balance for address. Returns 0 if not found.
func (l *WalletLedger) GetBalance(address string) uint64 {
	val, err := l.state.Get(balanceKey(address))
	if err != nil {
		return 0
	}
	switch v := val.(type) {
	case uint64:
		return v
	case float64:
		return uint64(v)
	case int:
		return uint64(v)
	}
	return 0
}

// SetBalance overwrites the balance for address.
func (l *WalletLedger) SetBalance(address string, amount uint64) error {
	return l.state.Set(balanceKey(address), amount)
}

// Credit increases address's balance by amount.
func (l *WalletLedger) Credit(address string, amount uint64) error {
	bal := l.GetBalance(address)
	return l.SetBalance(address, bal+amount)
}

// Debit decreases address's balance by amount. Returns an error if funds
// are insufficient.
func (l *WalletLedger) Debit(address string, amount uint64) error {
	bal := l.GetBalance(address)
	if bal < amount {
		return fmt.Errorf("insufficient balance for %s: have %d, need %d", address, bal, amount)
	}
	return l.SetBalance(address, bal-amount)
}

// Transfer moves amount from sender to recipient atomically.
func (l *WalletLedger) Transfer(from, to string, amount uint64) error {
	if from == to {
		return fmt.Errorf("cannot transfer to self")
	}
	if err := l.Debit(from, amount); err != nil {
		return err
	}
	return l.Credit(to, amount)
}

// ApplyTransaction updates ledger state for a finalised transaction.
// Called when a block is committed to update all wallet balances.
func (l *WalletLedger) ApplyTransaction(txn *Transaction) error {
	switch txn.Type {
	case TxTypeTransfer:
		return l.Transfer(txn.From, txn.To, txn.Amount)
	case TxTypeReward:
		return l.Credit(txn.To, txn.Amount)
	default:
		// Stake/Unstake, SmartContract, FlRound etc. have separate handlers.
		return nil
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// deriveAddress computes a blockchain address from an Ed25519 public key.
// Address = "0x" + hex(sha256(pubKey)[:20])
func deriveAddress(pubKey ed25519.PublicKey) string {
	hash := sha256.Sum256(pubKey)
	return "0x" + hex.EncodeToString(hash[:20])
}

// txSigningMessage builds the canonical signing payload for a transaction.
func txSigningMessage(txn *Transaction) []byte {
	msg := fmt.Sprintf("%s:%s:%s:%s:%d:%d:%d",
		txn.ID, txn.Type, txn.From, txn.To, txn.Nonce, txn.Amount, txn.Timestamp)
	return []byte(msg)
}

// balanceKey returns the StateDatabase key for an address's balance.
func balanceKey(address string) string {
	return "balance:" + address
}

// shortAddr returns an 8-char prefix of an address for use in IDs.
func shortAddr(address string) string {
	if len(address) > 10 {
		return address[2:10]
	}
	return address
}
