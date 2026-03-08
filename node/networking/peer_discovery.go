package networking

import (
	meshruntime "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/node/network"
)

// DiscoveryResult captures newly discovered peers.
type DiscoveryResult struct {
	PeerIDs []string
	Count   int
}

// DiscoveryService provides a lightweight discovery view for a mesh node.
type DiscoveryService struct {
	mesh *meshruntime.Mesh
}

// NewDiscoveryService creates a discovery service bound to mesh runtime.
func NewDiscoveryService(mesh *meshruntime.Mesh) *DiscoveryService {
	return &DiscoveryService{mesh: mesh}
}

// Snapshot returns current discovery information.
func (d *DiscoveryService) Snapshot() DiscoveryResult {
	if d == nil || d.mesh == nil {
		return DiscoveryResult{}
	}
	// Active peer IDs are not directly exposed by mesh yet; return count as reliable signal.
	count := d.mesh.ActivePeers()
	return DiscoveryResult{Count: count}
}
