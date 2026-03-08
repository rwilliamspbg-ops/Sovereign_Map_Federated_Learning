package network

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMeshTwoNodeBootstrapAndPublish(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	meshA, err := NewMesh(ctx, MeshConfig{
		NodeID:      "node-a",
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
		Topic:       "fl.rounds",
	})
	if err != nil {
		t.Fatalf("create mesh A: %v", err)
	}
	defer func() { _ = meshA.Stop() }()

	listen := meshA.ListenAddrs()
	if len(listen) == 0 {
		t.Fatalf("mesh A has no listen addresses")
	}
	bootstrapAddr := fmt.Sprintf("%s/p2p/%s", listen[0], meshA.PeerID())

	meshB, err := NewMesh(ctx, MeshConfig{
		NodeID:         "node-b",
		ListenAddrs:    []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{bootstrapAddr},
		Topic:          "fl.rounds",
	})
	if err != nil {
		t.Fatalf("create mesh B: %v", err)
	}
	defer func() { _ = meshB.Stop() }()

	deadline := time.Now().Add(8 * time.Second)
	connected := false
	for time.Now().Before(deadline) {
		if meshA.ActivePeers() > 0 && meshB.ActivePeers() > 0 {
			connected = true
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
	if !connected {
		t.Fatalf("meshes did not connect: peersA=%d peersB=%d", meshA.ActivePeers(), meshB.ActivePeers())
	}

	if err := meshB.Publish(ctx, []byte("smoke-test")); err != nil {
		t.Fatalf("publish gossip: %v", err)
	}
}
