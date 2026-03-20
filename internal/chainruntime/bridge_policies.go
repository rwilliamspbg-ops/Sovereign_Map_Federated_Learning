package chainruntime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type bridgePolicyFile struct {
	Version string              `json:"version"`
	Routes  []bridgePolicyRoute `json:"routes"`
}

type bridgePolicyRoute struct {
	SourceChain string            `json:"source_chain"`
	TargetChain string            `json:"target_chain"`
	Policy      bridgeRoutePolicy `json:"policy"`
}

type bridgeRoutePolicy struct {
	ID                string   `json:"id"`
	AllowedAssets     []string `json:"allowed_assets"`
	MinAmount         float64  `json:"min_amount"`
	MaxAmount         float64  `json:"max_amount"`
	MinFinalityBlocks uint64   `json:"min_finality_blocks"`
}

// BridgeTransferRequest contains metadata for a bridge transfer pre-check.
type BridgeTransferRequest struct {
	SourceChain    string
	TargetChain    string
	Asset          string
	Amount         float64
	FinalityBlocks uint64
}

// BridgePolicySet indexes bridge routes for fast runtime checks.
type BridgePolicySet struct {
	version string
	routes  map[string]bridgeRoutePolicy
}

func LoadBridgePolicySet(path string) (*BridgePolicySet, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read bridge policy file: %w", err)
	}

	var file bridgePolicyFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse bridge policy file: %w", err)
	}

	set := &BridgePolicySet{
		version: file.Version,
		routes:  make(map[string]bridgeRoutePolicy, len(file.Routes)),
	}

	for _, route := range file.Routes {
		src := normalizeToken(route.SourceChain)
		dst := normalizeToken(route.TargetChain)
		if src == "" || dst == "" {
			continue
		}
		set.routes[routeKey(src, dst)] = route.Policy
	}

	if len(set.routes) == 0 {
		return nil, fmt.Errorf("bridge policy file contains no valid routes")
	}

	return set, nil
}

func (b *BridgePolicySet) Version() string {
	if b == nil {
		return ""
	}
	return b.version
}

func (b *BridgePolicySet) ValidateTransfer(req BridgeTransferRequest) error {
	if b == nil {
		return fmt.Errorf("bridge policies are not configured")
	}

	src := normalizeToken(req.SourceChain)
	dst := normalizeToken(req.TargetChain)
	asset := normalizeToken(req.Asset)
	if src == "" || dst == "" || asset == "" {
		return fmt.Errorf("source_chain, target_chain, and asset are required")
	}

	policy, ok := b.routes[routeKey(src, dst)]
	if !ok {
		return fmt.Errorf("route %s->%s is not allowed", src, dst)
	}

	if !assetAllowed(policy.AllowedAssets, asset) {
		return fmt.Errorf("asset %s is not allowed for route %s->%s", asset, src, dst)
	}

	if req.Amount < policy.MinAmount {
		return fmt.Errorf("amount %.4f below minimum %.4f", req.Amount, policy.MinAmount)
	}

	if req.Amount > policy.MaxAmount {
		return fmt.Errorf("amount %.4f above maximum %.4f", req.Amount, policy.MaxAmount)
	}

	if req.FinalityBlocks < policy.MinFinalityBlocks {
		return fmt.Errorf("finality blocks %d below minimum %d", req.FinalityBlocks, policy.MinFinalityBlocks)
	}

	return nil
}

func routeKey(sourceChain, targetChain string) string {
	return sourceChain + ":" + targetChain
}

func assetAllowed(allowed []string, asset string) bool {
	for _, entry := range allowed {
		if normalizeToken(entry) == asset {
			return true
		}
	}
	return false
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
