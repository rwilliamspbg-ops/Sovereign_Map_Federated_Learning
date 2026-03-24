package tpm

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
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

type quoteCacheShard struct {
	mu      sync.RWMutex
	entries map[string]CachedQuote
}

type nonceCacheShard struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

var (
	quoteCacheShards [cacheShardCount]quoteCacheShard
	nonceCacheShards [cacheShardCount]nonceCacheShard
	nonceCleanerOnce sync.Once
)

const quoteBindingSalt = "sovereign-map-tpm-quote-salt-v1"
const quoteVersion = "tpm-quote-v2"
const maxQuoteAge = 10 * time.Minute
const maxClockSkew = 30 * time.Second
const nonceReplayTTL = 15 * time.Minute
const quoteCacheTTL = 30 * time.Second
const cacheShardCount = 64
const nonceCleanupInterval = 30 * time.Second

func init() {
	for i := 0; i < cacheShardCount; i++ {
		quoteCacheShards[i].entries = make(map[string]CachedQuote)
		nonceCacheShards[i].entries = make(map[string]time.Time)
	}
}

// GetVerifiedQuote implements a cache-aside pattern to bypass the 429ms TPM bottleneck.
func GetVerifiedQuote(nodeID string) ([]byte, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return nil, fmt.Errorf("node id cannot be empty")
	}

	shard := quoteShard(nodeID)
	readLockWaitStart := time.Now()
	shard.mu.RLock()
	observeQuoteCacheLockWait("read", time.Since(readLockWaitStart))
	entry, found := shard.entries[nodeID]
	shard.mu.RUnlock()

	if found && time.Now().Before(entry.ExpiresAt) {
		observeQuoteCacheHit()
		return entry.Quote, nil
	}
	observeQuoteCacheMiss()

	// Fallback to the hardware call (identified as a major performance bottleneck).
	quote, err := generateTPMQuoteForNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quote for node %s: %w", nodeID, err)
	}

	if err := verifyQuoteFormatAndDigest(nodeID, quote); err != nil {
		return nil, fmt.Errorf("generated quote failed self-verification: %w", err)
	}

	writeLockWaitStart := time.Now()
	shard.mu.Lock()
	observeQuoteCacheLockWait("write", time.Since(writeLockWaitStart))
	shard.entries[nodeID] = CachedQuote{
		Quote:     quote,
		ExpiresAt: time.Now().Add(quoteCacheTTL),
	}
	shard.mu.Unlock()

	return quote, nil
}

// PrewarmVerifiedQuotes pre-populates quote cache entries for active nodes at round start.
func PrewarmVerifiedQuotes(nodeIDs []string) int {
	if len(nodeIDs) == 0 {
		return 0
	}

	seen := make(map[string]struct{}, len(nodeIDs))
	warmed := 0
	for _, id := range nodeIDs {
		nodeID := strings.TrimSpace(id)
		if nodeID == "" {
			continue
		}
		if _, exists := seen[nodeID]; exists {
			continue
		}
		seen[nodeID] = struct{}{}
		if _, err := GetVerifiedQuote(nodeID); err == nil {
			warmed++
		}
	}

	observePrewarm(len(seen), warmed)

	return warmed
}

// isQuoteCached is used by tests in this package.
func isQuoteCached(nodeID string) bool {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return false
	}
	shard := quoteShard(nodeID)
	shard.mu.RLock()
	entry, ok := shard.entries[nodeID]
	shard.mu.RUnlock()
	return ok && time.Now().Before(entry.ExpiresAt)
}

// Verify implements the exported verification function used by the worker pool.
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

	nonceKey := nodeID + ":" + parsed.nonce
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

	message := buildQuoteDigestMessage(parsed.timestampNanos, parsed.nonce, parsed.entropy, parsed.nodeBinding)
	recomputed := sha256.Sum256(message)
	if parsed.digest != hex.EncodeToString(recomputed[:]) {
		return fmt.Errorf("quote digest mismatch")
	}

	return nil
}

// GenerateTPMQuote is a stub for the expensive hardware call.
func GenerateTPMQuote() ([]byte, error) {
	return generateTPMQuoteForNode("global")
}

func generateTPMQuoteForNode(nodeID string) ([]byte, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
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

	message := buildQuoteDigestMessage(timestampPart, noncePart, entropyPart, nodeBinding)
	digest := sha256.Sum256(message)
	digestHex := hex.EncodeToString(digest[:])

	quote := buildQuotePayload(timestampPart, noncePart, entropyPart, nodeBinding, digestHex)
	return quote, nil
}

func hashNodeBinding(nodeID string) string {
	sum := sha256.Sum256([]byte(nodeID))
	return hex.EncodeToString(sum[:8])
}

func checkAndRememberNonce(nonceKey string) error {
	ensureNonceCleanerStarted()

	shard := nonceShard(nonceKey)
	lockWaitStart := time.Now()
	shard.mu.Lock()
	observeNonceCacheLockWait(time.Since(lockWaitStart))
	defer shard.mu.Unlock()

	now := time.Now()
	if expiry, exists := shard.entries[nonceKey]; exists && now.Before(expiry) {
		observeNonceReplayRejected()
		return fmt.Errorf("replayed quote nonce detected")
	}

	shard.entries[nonceKey] = now.Add(nonceReplayTTL)
	return nil
}

func ensureNonceCleanerStarted() {
	nonceCleanerOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(nonceCleanupInterval)
			defer ticker.Stop()
			for range ticker.C {
				cleanupExpiredNonceEntries()
			}
		}()
	})
}

func cleanupExpiredNonceEntries() {
	now := time.Now()
	removed := 0
	for i := 0; i < cacheShardCount; i++ {
		shard := &nonceCacheShards[i]
		shard.mu.Lock()
		for key, expiry := range shard.entries {
			if now.After(expiry) {
				delete(shard.entries, key)
				removed++
			}
		}
		shard.mu.Unlock()
	}
	observeNonceCleanup(removed)
}

func quoteShard(key string) *quoteCacheShard {
	idx := cacheShardIndex(key)
	return &quoteCacheShards[idx]
}

func nonceShard(key string) *nonceCacheShard {
	idx := cacheShardIndex(key)
	return &nonceCacheShards[idx]
}

func cacheShardIndex(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32() % cacheShardCount)
}

func buildQuoteDigestMessage(timestampPart int64, noncePart, entropyPart, nodeBinding string) []byte {
	buf := make([]byte, 0, 32+len(noncePart)+len(entropyPart)+len(nodeBinding)+len(quoteBindingSalt)+5)
	buf = strconv.AppendInt(buf, timestampPart, 10)
	buf = append(buf, '|')
	buf = append(buf, noncePart...)
	buf = append(buf, '|')
	buf = append(buf, entropyPart...)
	buf = append(buf, '|')
	buf = append(buf, nodeBinding...)
	buf = append(buf, '|')
	buf = append(buf, quoteBindingSalt...)
	return buf
}

func buildQuotePayload(timestampPart int64, noncePart, entropyPart, nodeBinding, digestHex string) []byte {
	buf := make([]byte, 0, len(quoteVersion)+len(noncePart)+len(entropyPart)+len(nodeBinding)+len(digestHex)+32+6)
	buf = append(buf, quoteVersion...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, timestampPart, 10)
	buf = append(buf, ':')
	buf = append(buf, noncePart...)
	buf = append(buf, ':')
	buf = append(buf, entropyPart...)
	buf = append(buf, ':')
	buf = append(buf, nodeBinding...)
	buf = append(buf, ':')
	buf = append(buf, digestHex...)
	return buf
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
