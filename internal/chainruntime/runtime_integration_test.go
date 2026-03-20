package chainruntime

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
	meshruntime "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/node/network"
)

func TestRuntimeMeshPropagationTxAndBlock(t *testing.T) {
	topic := fmt.Sprintf("chainruntime-it-%d", time.Now().UnixNano())
	discoveryTag := fmt.Sprintf("chainruntime-discovery-%d", time.Now().UnixNano())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	meshA, err := meshruntime.NewMesh(ctx, meshruntime.MeshConfig{
		NodeID:       "node-a",
		ListenAddrs:  []string{"/ip4/127.0.0.1/tcp/0"},
		DiscoveryTag: discoveryTag,
		Topic:        topic,
	})
	if err != nil {
		t.Fatalf("create meshA: %v", err)
	}
	defer func() { _ = meshA.Stop() }()

	bootstrap := fmt.Sprintf("%s/p2p/%s", meshA.ListenAddrs()[0], meshA.PeerID())
	meshB, err := meshruntime.NewMesh(ctx, meshruntime.MeshConfig{
		NodeID:         "node-b",
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{bootstrap},
		DiscoveryTag:   discoveryTag,
		Topic:          topic,
	})
	if err != nil {
		t.Fatalf("create meshB: %v", err)
	}
	defer func() { _ = meshB.Stop() }()

	if err := waitUntil(ctx, func() bool { return meshA.ActivePeers() > 0 && meshB.ActivePeers() > 0 }); err != nil {
		t.Fatalf("meshes did not connect: %v", err)
	}

	policies, err := LoadBridgePolicySet(filepath.Join("..", "..", "bridge-policies.json"))
	if err != nil {
		t.Fatalf("load bridge policies: %v", err)
	}

	chainA := blockchain.NewBlockChain()
	chainB := blockchain.NewBlockChain()
	runtimeA, err := New(meshA, chainA, policies)
	if err != nil {
		t.Fatalf("runtimeA create: %v", err)
	}
	runtimeB, err := New(meshB, chainB, policies)
	if err != nil {
		t.Fatalf("runtimeB create: %v", err)
	}

	if err := runtimeA.Start(ctx); err != nil {
		t.Fatalf("start runtimeA: %v", err)
	}
	if err := runtimeB.Start(ctx); err != nil {
		t.Fatalf("start runtimeB: %v", err)
	}

	// Allow gossipsub control plane to settle before sending first tx.
	time.Sleep(1500 * time.Millisecond)

	ledgerA := blockchain.NewWalletLedger(chainA.StateDB)
	_ = ledgerA.SetBalance("0xsender", 2_000)
	ledgerB := blockchain.NewWalletLedger(chainB.StateDB)
	_ = ledgerB.SetBalance("0xsender", 2_000)

	txn := &blockchain.Transaction{
		ID:        "tx-integration-1",
		Type:      blockchain.TxTypeTransfer,
		From:      "0xsender",
		To:        "0xrecipient",
		Nonce:     0,
		Amount:    100,
		Gas:       21000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("integration-signature"),
	}
	if err := runtimeA.SubmitTransaction(ctx, txn); err != nil {
		t.Fatalf("submit tx: %v", err)
	}

	if err := waitUntil(ctx, func() bool { return chainB.Mempool.Size() == 1 }); err != nil {
		t.Fatalf("tx did not propagate to runtimeB mempool: %v", err)
	}

	roundData := map[string]interface{}{
		"round_id":   "it-round-1",
		"model_hash": "integration-model",
	}
	block, err := runtimeA.ProposeAndBroadcastBlock(ctx, meshA.PeerID(), roundData)
	if err != nil {
		t.Fatalf("propose and broadcast block: %v", err)
	}

	if err := waitUntil(ctx, func() bool { return chainB.Height() == 1 }); err != nil {
		t.Fatalf("block did not propagate to runtimeB: %v", err)
	}

	if _, err := chainB.GetBlock(block.Header.Hash); err != nil {
		t.Fatalf("runtimeB missing propagated block: %v", err)
	}
}

func waitUntil(ctx context.Context, condition func() bool) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
