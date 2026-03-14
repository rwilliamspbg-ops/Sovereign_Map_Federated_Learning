package networking

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	autonat "github.com/libp2p/go-libp2p/p2p/host/autonat"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
)

// NATStatus tracks NAT type and reachability.
type NATStatus struct {
	Reachability network.Reachability
	PublicAddrs  []string
	RelayCount   int
	LastCheck    time.Time
}

// NATService handles NAT traversal with AutoNAT and circuit relay.
type NATService struct {
	mu      sync.RWMutex
	host    host.Host
	autonat autonat.AutoNAT
	status  NATStatus
}

// NewNATService creates NAT traversal service with AutoNAT + relay.
func NewNATService(ctx context.Context, h host.Host, enableRelay bool) (*NATService, error) {
	// Enable AutoNAT for reachability detection
	an, err := autonat.New(h, autonat.EnableService(h.Network()))
	if err != nil {
		return nil, fmt.Errorf("create autonat: %w", err)
	}

	ns := &NATService{
		host:    h,
		autonat: an,
		status: NATStatus{
			Reachability: network.ReachabilityUnknown,
			LastCheck:    time.Now(),
		},
	}

	// Enable circuit relay if requested
	if enableRelay {
		_, err := relay.New(h)
		if err != nil {
			return nil, fmt.Errorf("enable relay: %w", err)
		}

		// Enable AutoRelay for automatic relay discovery
		_, err = autorelay.NewAutoRelay(h)
		if err != nil {
			return nil, fmt.Errorf("enable autorelay: %w", err)
		}
	}

	go ns.monitorReachability(ctx)

	return ns, nil
}

// monitorReachability periodically checks NAT status.
func (ns *NATService) monitorReachability(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ns.updateStatus()
		}
	}
}

// updateStatus refreshes NAT reachability status.
func (ns *NATService) updateStatus() {
	// Use host metrics instead of direct network reachability
	reachability := network.ReachabilityUnknown
	if len(ns.host.Network().Conns()) > 0 {
		reachability = network.ReachabilityPublic
	}

	publicAddrs := []string{}
	for _, addr := range ns.host.Addrs() {
		publicAddrs = append(publicAddrs, addr.String())
	}

	relayCount := 0
	for _, conn := range ns.host.Network().Conns() {
		protos := conn.RemoteMultiaddr().Protocols()
		for _, p := range protos {
			if p.Name == "p2p-circuit" {
				relayCount++
				break
			}
		}
	}

	ns.mu.Lock()
	ns.status = NATStatus{
		Reachability: reachability,
		PublicAddrs:  publicAddrs,
		RelayCount:   relayCount,
		LastCheck:    time.Now(),
	}
	ns.mu.Unlock()
}

// Status returns current NAT traversal status.
func (ns *NATService) Status() NATStatus {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.status
}

// IsPublic returns true if node is publicly reachable.
func (ns *NATService) IsPublic() bool {
	return ns.status.Reachability == network.ReachabilityPublic
}
