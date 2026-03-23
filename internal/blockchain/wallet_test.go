// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Wallet construction
// ---------------------------------------------------------------------------

func TestNewWallet(t *testing.T) {
	w, err := NewWallet()
	if err != nil {
		t.Fatalf("NewWallet: %v", err)
	}
	addr := w.Address()
	if !strings.HasPrefix(addr, "0x") {
		t.Errorf("address %q must start with 0x", addr)
	}
	if len(addr) != 42 { // "0x" + 40 hex chars
		t.Errorf("expected address length 42, got %d", len(addr))
	}
}

func TestWalletFromSeed(t *testing.T) {
	w1, err := NewWallet()
	if err != nil {
		t.Fatalf("NewWallet: %v", err)
	}
	seed := w1.Seed()
	w2, err := WalletFromSeed(seed)
	if err != nil {
		t.Fatalf("WalletFromSeed: %v", err)
	}
	if w1.Address() != w2.Address() {
		t.Errorf("restored wallet address mismatch: %s vs %s", w1.Address(), w2.Address())
	}
}

func TestWalletFromSeedBadSize(t *testing.T) {
	_, err := WalletFromSeed([]byte("too-short"))
	if err == nil {
		t.Error("expected error for wrong seed size")
	}
}

func TestTwoWalletsDistinct(t *testing.T) {
	w1, _ := NewWallet()
	w2, _ := NewWallet()
	if w1.Address() == w2.Address() {
		t.Error("two independently generated wallets must not share an address")
	}
}

// ---------------------------------------------------------------------------
// Signing
// ---------------------------------------------------------------------------

func TestSignAndVerify(t *testing.T) {
	w, _ := NewWallet()
	msg := []byte("hello sovereign blockchain")
	sig := w.Sign(msg)
	if len(sig) == 0 {
		t.Fatal("signature is empty")
	}
	if !VerifySignature(msg, sig, w.PublicKeyBytes()) {
		t.Error("valid signature did not verify")
	}
}

func TestVerifyWrongKey(t *testing.T) {
	w1, _ := NewWallet()
	w2, _ := NewWallet()
	sig := w1.Sign([]byte("data"))
	if VerifySignature([]byte("data"), sig, w2.PublicKeyBytes()) {
		t.Error("signature verified with wrong public key")
	}
}

func TestVerifyBadPublicKey(t *testing.T) {
	w, _ := NewWallet()
	sig := w.Sign([]byte("data"))
	if VerifySignature([]byte("data"), sig, []byte("short")) {
		t.Error("should not verify with invalid public key bytes")
	}
}

// ---------------------------------------------------------------------------
// Transaction creation
// ---------------------------------------------------------------------------

func TestCreateTransfer(t *testing.T) {
	sender, _ := NewWallet()
	recipient, _ := NewWallet()

	txn, err := sender.CreateTransfer(recipient.Address(), 500, 1)
	if err != nil {
		t.Fatalf("CreateTransfer: %v", err)
	}
	if txn.Type != TxTypeTransfer {
		t.Errorf("expected TxTypeTransfer, got %s", txn.Type)
	}
	if txn.From != sender.Address() {
		t.Errorf("unexpected From: %s", txn.From)
	}
	if txn.To != recipient.Address() {
		t.Errorf("unexpected To: %s", txn.To)
	}
	if txn.Amount != 500 {
		t.Errorf("expected amount 500, got %d", txn.Amount)
	}
	if err := txn.Validate(); err != nil {
		t.Errorf("created transfer is invalid: %v", err)
	}
	if !VerifyTxSignature(txn, sender.PublicKeyBytes()) {
		t.Error("transfer signature did not verify")
	}
}

func TestCreateTransferZeroAmount(t *testing.T) {
	sender, _ := NewWallet()
	recipient, _ := NewWallet()
	_, err := sender.CreateTransfer(recipient.Address(), 0, 1)
	if err == nil {
		t.Error("expected error for zero-amount transfer")
	}
}

func TestCreateTransferEmptyRecipient(t *testing.T) {
	sender, _ := NewWallet()
	_, err := sender.CreateTransfer("", 100, 1)
	if err == nil {
		t.Error("expected error for empty recipient")
	}
}

func TestCreateStakeTx(t *testing.T) {
	w, _ := NewWallet()
	txn, err := w.CreateStakeTx(10000, 1)
	if err != nil {
		t.Fatalf("CreateStakeTx: %v", err)
	}
	if txn.Type != TxTypeStake {
		t.Errorf("expected TxTypeStake, got %s", txn.Type)
	}
	if err := txn.Validate(); err != nil {
		t.Errorf("stake tx invalid: %v", err)
	}
}

func TestCreateStakeTxZero(t *testing.T) {
	w, _ := NewWallet()
	_, err := w.CreateStakeTx(0, 1)
	if err == nil {
		t.Error("expected error for zero-amount stake")
	}
}

func TestCreateUnstakeTx(t *testing.T) {
	w, _ := NewWallet()
	txn, err := w.CreateUnstakeTx(5000, 2)
	if err != nil {
		t.Fatalf("CreateUnstakeTx: %v", err)
	}
	if txn.Type != TxTypeUnstake {
		t.Errorf("expected TxTypeUnstake, got %s", txn.Type)
	}
	if err := txn.Validate(); err != nil {
		t.Errorf("unstake tx invalid: %v", err)
	}
}

func TestCreateFlRoundTx(t *testing.T) {
	w, _ := NewWallet()
	data := map[string]interface{}{
		"round":    uint64(1),
		"accuracy": 0.92,
	}
	txn, err := w.CreateFlRoundTx(data, 3)
	if err != nil {
		t.Fatalf("CreateFlRoundTx: %v", err)
	}
	if txn.Type != TxTypeFlRound {
		t.Errorf("expected TxTypeFlRound, got %s", txn.Type)
	}
	if err := txn.Validate(); err != nil {
		t.Errorf("fl round tx invalid: %v", err)
	}
	if !VerifyTxSignature(txn, w.PublicKeyBytes()) {
		t.Error("fl round tx signature did not verify")
	}
}

// ---------------------------------------------------------------------------
// WalletLedger
// ---------------------------------------------------------------------------

func newTestLedger(t *testing.T) *WalletLedger {
	t.Helper()
	state := NewStateDatabase()
	return NewWalletLedger(state)
}

func TestGetBalanceUnknown(t *testing.T) {
	l := newTestLedger(t)
	if l.GetBalance("0xdeadbeef") != 0 {
		t.Error("unknown address should have zero balance")
	}
}

func TestSetAndGetBalance(t *testing.T) {
	l := newTestLedger(t)
	w, _ := NewWallet()
	if err := l.SetBalance(w.Address(), 1000); err != nil {
		t.Fatalf("SetBalance: %v", err)
	}
	if bal := l.GetBalance(w.Address()); bal != 1000 {
		t.Errorf("expected 1000, got %d", bal)
	}
}

func TestCredit(t *testing.T) {
	l := newTestLedger(t)
	w, _ := NewWallet()
	if err := l.Credit(w.Address(), 500); err != nil {
		t.Fatalf("Credit: %v", err)
	}
	if err := l.Credit(w.Address(), 300); err != nil {
		t.Fatalf("Credit: %v", err)
	}
	if bal := l.GetBalance(w.Address()); bal != 800 {
		t.Errorf("expected 800, got %d", bal)
	}
}

func TestDebit(t *testing.T) {
	l := newTestLedger(t)
	w, _ := NewWallet()
	if err := l.SetBalance(w.Address(), 1000); err != nil {
		t.Fatalf("SetBalance: %v", err)
	}
	if err := l.Debit(w.Address(), 400); err != nil {
		t.Fatalf("Debit: %v", err)
	}
	if bal := l.GetBalance(w.Address()); bal != 600 {
		t.Errorf("expected 600, got %d", bal)
	}
}

func TestDebitInsufficientFunds(t *testing.T) {
	l := newTestLedger(t)
	w, _ := NewWallet()
	if err := l.SetBalance(w.Address(), 100); err != nil {
		t.Fatalf("SetBalance: %v", err)
	}
	if err := l.Debit(w.Address(), 200); err == nil {
		t.Error("expected insufficient-funds error")
	}
}

func TestTransfer(t *testing.T) {
	l := newTestLedger(t)
	alice, _ := NewWallet()
	bob, _ := NewWallet()
	if err := l.SetBalance(alice.Address(), 1000); err != nil {
		t.Fatalf("SetBalance: %v", err)
	}
	if err := l.Transfer(alice.Address(), bob.Address(), 300); err != nil {
		t.Fatalf("Transfer: %v", err)
	}
	if bal := l.GetBalance(alice.Address()); bal != 700 {
		t.Errorf("alice: expected 700, got %d", bal)
	}
	if bal := l.GetBalance(bob.Address()); bal != 300 {
		t.Errorf("bob: expected 300, got %d", bal)
	}
}

func TestTransferInsufficientFunds(t *testing.T) {
	l := newTestLedger(t)
	alice, _ := NewWallet()
	bob, _ := NewWallet()
	l.SetBalance(alice.Address(), 50)
	if err := l.Transfer(alice.Address(), bob.Address(), 100); err == nil {
		t.Error("expected error for insufficient balance")
	}
	// Balance must be unchanged on failure
	if bal := l.GetBalance(alice.Address()); bal != 50 {
		t.Errorf("alice balance should be unchanged at 50, got %d", bal)
	}
}

func TestTransferToSelf(t *testing.T) {
	l := newTestLedger(t)
	w, _ := NewWallet()
	l.SetBalance(w.Address(), 500)
	if err := l.Transfer(w.Address(), w.Address(), 100); err == nil {
		t.Error("expected error for self-transfer")
	}
}

// ---------------------------------------------------------------------------
// ApplyTransaction
// ---------------------------------------------------------------------------

func TestApplyTransactionTransfer(t *testing.T) {
	l := newTestLedger(t)
	alice, _ := NewWallet()
	bob, _ := NewWallet()
	l.SetBalance(alice.Address(), 1000)

	txn, _ := alice.CreateTransfer(bob.Address(), 250, 1)
	if err := l.ApplyTransaction(txn); err != nil {
		t.Fatalf("ApplyTransaction(transfer): %v", err)
	}
	if l.GetBalance(alice.Address()) != 750 {
		t.Errorf("alice: expected 750")
	}
	if l.GetBalance(bob.Address()) != 250 {
		t.Errorf("bob: expected 250")
	}
}

func TestApplyTransactionReward(t *testing.T) {
	l := newTestLedger(t)
	w, _ := NewWallet()
	l.SetBalance(w.Address(), 0)

	txn := &Transaction{
		ID:        "reward-tx-1",
		Type:      TxTypeReward,
		To:        w.Address(),
		Amount:    10000,
		Timestamp: 1234567890,
		Signature: []byte("dummy"),
	}
	if err := l.ApplyTransaction(txn); err != nil {
		t.Fatalf("ApplyTransaction(reward): %v", err)
	}
	if l.GetBalance(w.Address()) != 10000 {
		t.Errorf("expected 10000 after reward")
	}
}

func TestApplyTransactionNoop(t *testing.T) {
	l := newTestLedger(t)
	w, _ := NewWallet()
	l.SetBalance(w.Address(), 500)

	// FL round txns don't change balances — should be a no-op
	txn, _ := w.CreateFlRoundTx(map[string]interface{}{}, 1)
	if err := l.ApplyTransaction(txn); err != nil {
		t.Fatalf("ApplyTransaction(flround): %v", err)
	}
	if l.GetBalance(w.Address()) != 500 {
		t.Errorf("fl round should not change balance")
	}
}
