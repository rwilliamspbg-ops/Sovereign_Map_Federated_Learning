package network

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	ma "github.com/multiformats/go-multiaddr"
)

// MeshConfig controls host, discovery, and gossip behavior for a node.
type MeshConfig struct {
	NodeID         string
	ListenAddrs    []string
	BootstrapPeers []string
	DiscoveryTag   string
	Topic          string
}

// Mesh wires libp2p host + discovery + gossipsub for runtime messaging.
type Mesh struct {
	mu      sync.RWMutex
	host    host.Host
	topic   *pubsub.Topic
	mdnsSvc mdns.Service
	joined  int
}

// NewMesh initializes the mesh with startup gossip topic join.
func NewMesh(ctx context.Context, cfg MeshConfig) (*Mesh, error) {
	if len(cfg.ListenAddrs) == 0 {
		cfg.ListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}
	}
	if strings.TrimSpace(cfg.Topic) == "" {
		cfg.Topic = "fl.rounds"
	}
	if strings.TrimSpace(cfg.DiscoveryTag) == "" {
		cfg.DiscoveryTag = "sovereign-map-testnet"
	}

	listenAddrs := make([]ma.Multiaddr, 0, len(cfg.ListenAddrs))
	for _, s := range cfg.ListenAddrs {
		addr, err := ma.NewMultiaddr(strings.TrimSpace(s))
		if err != nil {
			return nil, fmt.Errorf("invalid listen multiaddr %q: %w", s, err)
		}
		listenAddrs = append(listenAddrs, addr)
	}

	// Enable default transports (TCP + QUIC automatically included)
	h, err := libp2p.New(
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.DefaultTransports,
	)
	if err != nil {
		return nil, fmt.Errorf("create libp2p host: %w", err)
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		_ = h.Close()
		return nil, fmt.Errorf("create gossipsub: %w", err)
	}

	topic, err := ps.Join(cfg.Topic)
	if err != nil {
		_ = h.Close()
		return nil, fmt.Errorf("join topic %q: %w", cfg.Topic, err)
	}

	m := &Mesh{host: h, topic: topic}
	m.mdnsSvc = mdns.NewMdnsService(h, cfg.DiscoveryTag, &discoveryNotifee{mesh: m})
	if err := m.mdnsSvc.Start(); err != nil {
		_ = topic.Close()
		_ = h.Close()
		return nil, fmt.Errorf("start mdns discovery: %w", err)
	}

	m.joined = m.ConnectBootstrap(ctx, cfg.BootstrapPeers)
	return m, nil
}

// ConnectBootstrap dials static peers and returns successful connections.
func (m *Mesh) ConnectBootstrap(ctx context.Context, peers []string) int {
	joined := 0
	for _, raw := range peers {
		info, err := parsePeerAddr(raw)
		if err != nil {
			continue
		}
		m.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstorePermanentAddrTTL)
		if err := m.host.Connect(ctx, *info); err == nil {
			joined++
		}
	}

	m.mu.Lock()
	m.joined += joined
	m.mu.Unlock()
	return joined
}

// Publish sends data over gossip to the joined topic.
func (m *Mesh) Publish(ctx context.Context, payload []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.topic == nil {
		return fmt.Errorf("mesh topic is not initialized")
	}
	return m.topic.Publish(ctx, payload)
}

// ActivePeers reports currently connected peers count.
func (m *Mesh) ActivePeers() int {
	return len(m.host.Network().Peers())
}

// JoinedBootstrap reports successful static bootstrap joins during startup.
func (m *Mesh) JoinedBootstrap() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.joined
}

// ListenAddrs returns host listen addresses.
func (m *Mesh) ListenAddrs() []string {
	addrs := m.host.Addrs()
	out := make([]string, 0, len(addrs))
	for _, a := range addrs {
		out = append(out, a.String())
	}
	return out
}

// Stop closes pubsub topic and libp2p host.
func (m *Mesh) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.topic != nil {
		_ = m.topic.Close()
	}
	if m.host != nil {
		return m.host.Close()
	}
	return nil
}

type discoveryNotifee struct {
	mesh *Mesh
}

func (n *discoveryNotifee) HandlePeerFound(info peer.AddrInfo) {
	_ = n.mesh.host.Connect(context.Background(), info)
}

func parsePeerAddr(raw string) (*peer.AddrInfo, error) {
	addr, err := ma.NewMultiaddr(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	info, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return nil, err
	}
	return info, nil
}

const peerstorePermanentAddrTTL = 24 * time.Hour
