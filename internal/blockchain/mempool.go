// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"fmt"
	"sort"
	"sync"
)

// Mempool manages pending transactions waiting for block inclusion
type Mempool struct {
	mu             sync.RWMutex
	transactions   map[string]*Transaction
	txnsByAddress  map[string][]*Transaction
	nonces         map[string]uint64
	maxSize        int
	maxTxnPerBlock int
}

// NewMempool creates a new mempool
func NewMempool() *Mempool {
	return &Mempool{
		transactions:   make(map[string]*Transaction),
		txnsByAddress:  make(map[string][]*Transaction),
		nonces:         make(map[string]uint64),
		maxSize:        100000,
		maxTxnPerBlock: 1000,
	}
}

// AddTransaction adds a transaction to the mempool
func (mp *Mempool) AddTransaction(txn *Transaction) error {
	if err := txn.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Check for duplicate
	if _, exists := mp.transactions[txn.ID]; exists {
		return fmt.Errorf("transaction already in mempool: %s", txn.ID)
	}

	// Check mempool size
	if len(mp.transactions) >= mp.maxSize {
		return fmt.Errorf("mempool full")
	}

	// Check nonce to prevent replay attacks
	expectedNonce := mp.nonces[txn.From]
	if txn.Nonce != expectedNonce {
		return fmt.Errorf("invalid nonce: got %d, expected %d", txn.Nonce, expectedNonce)
	}

	// Add transaction
	mp.transactions[txn.ID] = txn
	mp.txnsByAddress[txn.From] = append(mp.txnsByAddress[txn.From], txn)
	mp.nonces[txn.From] = txn.Nonce + 1

	return nil
}

// RemoveTransaction removes a transaction from the mempool
func (mp *Mempool) RemoveTransaction(txnID string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	txn, exists := mp.transactions[txnID]
	if !exists {
		return fmt.Errorf("transaction not in mempool: %s", txnID)
	}

	delete(mp.transactions, txnID)

	// Remove from address list
	addressTxns := mp.txnsByAddress[txn.From]
	for i, t := range addressTxns {
		if t.ID == txnID {
			mp.txnsByAddress[txn.From] = append(addressTxns[:i], addressTxns[i+1:]...)
			break
		}
	}

	return nil
}

// GetTransactionsForBlock selects transactions for inclusion in a block
// Prioritizes by gas price and timestamp
func (mp *Mempool) GetTransactionsForBlock(maxTxns int) []Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if maxTxns == 0 {
		maxTxns = mp.maxTxnPerBlock
	}

	// Convert to slice and sort by gas price (descending) then timestamp (ascending)
	txns := make([]*Transaction, 0, len(mp.transactions))
	for _, txn := range mp.transactions {
		txns = append(txns, txn)
	}

	sort.Slice(txns, func(i, j int) bool {
		if txns[i].GasPrice != txns[j].GasPrice {
			return txns[i].GasPrice > txns[j].GasPrice
		}
		return txns[i].Timestamp < txns[j].Timestamp
	})

	// Limit to max transactions
	if len(txns) > maxTxns {
		txns = txns[:maxTxns]
	}

	// Convert back to Transaction slice
	result := make([]Transaction, len(txns))
	for i, txn := range txns {
		result[i] = *txn
	}

	return result
}

// Size returns the number of transactions in the mempool
func (mp *Mempool) Size() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return len(mp.transactions)
}

// GetTransaction returns a transaction by ID
func (mp *Mempool) GetTransaction(txnID string) (*Transaction, error) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	txn, exists := mp.transactions[txnID]
	if !exists {
		return nil, fmt.Errorf("transaction not found: %s", txnID)
	}
	return txn, nil
}

// GetAddressNonce returns the next nonce for an address
func (mp *Mempool) GetAddressNonce(address string) uint64 {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.nonces[address]
}

// Clear removes all transactions from the mempool
func (mp *Mempool) Clear() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.transactions = make(map[string]*Transaction)
	mp.txnsByAddress = make(map[string][]*Transaction)
	mp.nonces = make(map[string]uint64)
}

// GetPending returns all pending transactions
func (mp *Mempool) GetPending() []Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	txns := make([]Transaction, 0, len(mp.transactions))
	for _, txn := range mp.transactions {
		txns = append(txns, *txn)
	}
	return txns
}
