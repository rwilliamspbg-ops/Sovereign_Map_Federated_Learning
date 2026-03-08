package networking

import (
	"context"
	"fmt"

	meshruntime "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/node/network"
)

// GossipTopic identifies a pubsub channel used for FL round coordination.
type GossipTopic struct {
	Name string
	mesh *meshruntime.Mesh
}

// NewGossipTopic creates a publish adapter for a mesh topic.
func NewGossipTopic(name string, mesh *meshruntime.Mesh) *GossipTopic {
	return &GossipTopic{Name: name, mesh: mesh}
}

// Publish emits a gossip message to peers on this topic.
func (g *GossipTopic) Publish(ctx context.Context, payload []byte) error {
	if g == nil || g.mesh == nil {
		return fmt.Errorf("gossip topic mesh is not initialized")
	}
	return g.mesh.Publish(ctx, payload)
}
