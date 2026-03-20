package chainruntime

import (
	"path/filepath"
	"testing"
)

func TestLoadBridgePolicySetAndValidate(t *testing.T) {
	path := filepath.Join("..", "..", "bridge-policies.json")
	set, err := LoadBridgePolicySet(path)
	if err != nil {
		t.Fatalf("expected policy load success: %v", err)
	}

	if got := set.Version(); got != "v1" {
		t.Fatalf("expected version v1, got %q", got)
	}

	err = set.ValidateTransfer(BridgeTransferRequest{
		SourceChain:    "ethereum",
		TargetChain:    "polygon",
		Asset:          "USDC",
		Amount:         10,
		FinalityBlocks: 12,
	})
	if err != nil {
		t.Fatalf("expected valid transfer, got %v", err)
	}

	err = set.ValidateTransfer(BridgeTransferRequest{
		SourceChain:    "ethereum",
		TargetChain:    "polygon",
		Asset:          "DAI",
		Amount:         10,
		FinalityBlocks: 12,
	})
	if err == nil {
		t.Fatalf("expected asset validation error")
	}
}
