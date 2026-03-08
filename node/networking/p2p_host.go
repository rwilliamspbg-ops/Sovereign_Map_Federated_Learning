package networking

import (
	"context"
	"fmt"

	meshruntime "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/node/network"
)

// HostConfig controls mesh host initialization.
type HostConfig struct {
	NodeID         string
	ListenAddrs    []string
	BootstrapPeers []string
	DiscoveryTag   string
	Topic          string
	EnableRelay    bool
}

// P2PHost adapts the mesh runtime for higher-level networking workflows.
type P2PHost struct {
	NodeID string
	Mesh   *meshruntime.Mesh
	NAT    *NATService
}

// NewP2PHost creates mesh runtime and optional NAT traversal service.
func NewP2PHost(ctx context.Context, cfg HostConfig) (*P2PHost, error) {
	mesh, err := meshruntime.NewMesh(ctx, meshruntime.MeshConfig{
		NodeID:         cfg.NodeID,
		ListenAddrs:    cfg.ListenAddrs,
		BootstrapPeers: cfg.BootstrapPeers,
		DiscoveryTag:   cfg.DiscoveryTag,
		Topic:          cfg.Topic,
	})
	if err != nil {
		return nil, fmt.Errorf("create mesh: %w", err)
	}

	ph := &P2PHost{NodeID: cfg.NodeID, Mesh: mesh}
	if cfg.EnableRelay {
		natSvc, natErr := NewNATService(ctx, mesh.Host(), true)
		if natErr != nil {
			_ = mesh.Stop()
			return nil, fmt.Errorf("create NAT service: %w", natErr)
		}
		ph.NAT = natSvc
	}

	return ph, nil
}

// Stop closes mesh resources.
func (p *P2PHost) Stop() error {
	if p == nil || p.Mesh == nil {
		return nil
	}
	return p.Mesh.Stop()
}
