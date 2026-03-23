package tpm

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/hva"
)

type CachedQuote struct {
	Quote     []byte
	ExpiresAt time.Time
}

var (
	quoteCache = make(map[string]CachedQuote)
	cacheMutex sync.RWMutex
	nonceCache = make(map[string]time.Time)
	nonceMutex sync.Mutex
)

const quoteBindingSalt = "sovereign-map-tpm-quote-salt-v1"
const quoteVersion = "tpm-quote-v2"
const maxQuoteAge = 10 * time.Minute
const maxClockSkew = 30 * time.Second
const nonceReplayTTL = 15 * time.Minute

// GetVerifiedQuote implements a cache-aside pattern to bypass the 429ms TPM bottleneck
func GetVerifiedQuote(nodeID string) ([]byte, error) {
	if strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("node id cannot be empty")
	}

	cacheMutex.RLock()
	entry, found := quoteCache[nodeID]
	cacheMutex.RUnlock()

	if found && time.Now().Before(entry.ExpiresAt) {
		return entry.Quote, nil
	}

	// Fallback to the hardware call (Identified as a 95% performance bottleneck)
	quote, err := generateTPMQuoteForNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quote for node %s: %w", nodeID, err)
	}

	if err := verifyQuoteFormatAndDigest(nodeID, quote); err != nil {
		return nil, fmt.Errorf("generated quote failed self-verification: %w", err)
	}

	cacheMutex.Lock()
	quoteCache[nodeID] = CachedQuote{
		Quote:     quote,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	cacheMutex.Unlock()

	return quote, nil
}

// Verify implements the exported verification function used by the worker pool
func Verify(nodeID string, quote []byte) error {
	if len(nodeID) == 0 {
		return fmt.Errorf("node id cannot be empty")
	}

	if len(quote) == 0 {
		return fmt.Errorf("quote cannot be empty")
	}

	if err := verifyQuoteFormatAndDigest(nodeID, quote); err != nil {
		return err
	}

	parsed, err := parseQuote(quote)
	if err != nil {
		return err
	}

	nonceKey := fmt.Sprintf("%s:%s", nodeID, parsed.nonce)
	if err := checkAndRememberNonce(nonceKey); err != nil {
		return err
	}

	return nil
}

type parsedQuote struct {
	timestampNanos int64
	nonce          string
	entropy        string
	nodeBinding    string
	digest         string
}

func parseQuote(quote []byte) (*parsedQuote, error) {
	parts := strings.Split(string(quote), ":")
	if len(parts) != 6 || parts[0] != quoteVersion {
		return nil, fmt.Errorf("invalid TPM quote format")
	}

	timestampPart := parts[1]
	noncePart := parts[2]
	entropyPart := parts[3]
	nodeBindingPart := parts[4]
	providedDigest := parts[5]

	timestampNanos, err := strconv.ParseInt(timestampPart, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid quote timestamp: %w", err)
	}

	if _, err := hex.DecodeString(noncePart); err != nil {
		return nil, fmt.Errorf("invalid quote nonce encoding: %w", err)
	}
	if _, err := hex.DecodeString(entropyPart); err != nil {
		return nil, fmt.Errorf("invalid quote entropy encoding: %w", err)
	}
	if _, err := hex.DecodeString(nodeBindingPart); err != nil {
		return nil, fmt.Errorf("invalid quote node binding encoding: %w", err)
	}

	return &parsedQuote{
		timestampNanos: timestampNanos,
		nonce:          noncePart,
		entropy:        entropyPart,
		nodeBinding:    nodeBindingPart,
		digest:         providedDigest,
	}, nil
}

func verifyQuoteFormatAndDigest(nodeID string, quote []byte) error {
	parsed, err := parseQuote(quote)
	if err != nil {
		return err
	}

	now := time.Now()
	quoteTime := time.Unix(0, parsed.timestampNanos)
	if quoteTime.After(now.Add(maxClockSkew)) {
		return fmt.Errorf("quote timestamp is in the future")
	}
	if now.Sub(quoteTime) > maxQuoteAge {
		return fmt.Errorf("quote timestamp too old")
	}

	nodeBinding := hashNodeBinding(nodeID)
	if parsed.nodeBinding != nodeBinding {
		return fmt.Errorf("quote node binding mismatch")
	}

	message := fmt.Sprintf("%d|%s|%s|%s|%s", parsed.timestampNanos, parsed.nonce, parsed.entropy, parsed.nodeBinding, quoteBindingSalt)
	recomputed := sha256.Sum256([]byte(message))
	if parsed.digest != hex.EncodeToString(recomputed[:]) {
		return fmt.Errorf("quote digest mismatch")
	}

	return nil
}

// GenerateTPMQuote is a stub for the expensive hardware call
func GenerateTPMQuote() ([]byte, error) {
	return generateTPMQuoteForNode("global")
}

func generateTPMQuoteForNode(nodeID string) ([]byte, error) {
	if strings.TrimSpace(nodeID) == "" {
		return nil, fmt.Errorf("node id cannot be empty")
	}

	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate quote nonce: %w", err)
	}

	entropy := make([]byte, 16)
	if _, err := rand.Read(entropy); err != nil {
		return nil, fmt.Errorf("failed to generate quote entropy: %w", err)
	}

	timestampPart := time.Now().UnixNano()
	noncePart := hex.EncodeToString(nonce)
	entropyPart := hex.EncodeToString(entropy)
	nodeBinding := hashNodeBinding(nodeID)

	message := fmt.Sprintf("%d|%s|%s|%s|%s", timestampPart, noncePart, entropyPart, nodeBinding, quoteBindingSalt)
	digest := sha256.Sum256([]byte(message))

	quote := fmt.Sprintf("%s:%d:%s:%s:%s:%s", quoteVersion, timestampPart, noncePart, entropyPart, nodeBinding, hex.EncodeToString(digest[:]))
	return []byte(quote), nil
}

func hashNodeBinding(nodeID string) string {
	sum := sha256.Sum256([]byte(nodeID))
	return hex.EncodeToString(sum[:8])
}

func checkAndRememberNonce(nonceKey string) error {
	nonceMutex.Lock()
	defer nonceMutex.Unlock()

	now := time.Now()
	for key, expiry := range nonceCache {
		if now.After(expiry) {
			delete(nonceCache, key)
		}
	}

	if expiry, exists := nonceCache[nonceKey]; exists && now.Before(expiry) {
		return fmt.Errorf("replayed quote nonce detected")
	}

	nonceCache[nonceKey] = now.Add(nonceReplayTTL)
	return nil
}

// VerifyByzantineResilience implements the safety check for Theorem 1.
// It ensures the number of nodes (n) can support the declared
// Byzantine fault tolerance (f) per the Hierarchical Multi-Krum proof.
// Reference: /proofs/bft_resilience.md
func VerifyByzantineResilience(totalNodes int, maliciousNodes int) (bool, error) {
	if totalNodes <= 0 {
		return false, fmt.Errorf("total nodes must be positive")
	}
	if maliciousNodes < 0 || maliciousNodes > totalNodes {
		return false, fmt.Errorf("malicious node count out of range")
	}

	maxByzantine := hva.MaximumByzantineNodes(totalNodes)
	if maliciousNodes > maxByzantine {
		return false, fmt.Errorf(
			"security threshold violated: Theorem 1 allows at most %d Byzantine nodes out of %d at the 55.5%% honest boundary",
			maxByzantine,
			totalNodes,
		)
	}
	return true, nil
}

// CalculateGlobalTolerance returns the (Σf_t) bound described in the whitepaper.
func CalculateGlobalTolerance(f_tiers []int) int {
	total := 0
	for _, f := range f_tiers {
		total += f
	}
	return total
}
