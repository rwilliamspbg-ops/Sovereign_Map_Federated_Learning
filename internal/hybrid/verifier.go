package hybrid

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	internalpkg "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal"
)

type ProofScheme string

const (
	SchemeSNARK ProofScheme = "snark"
	SchemeSTARK ProofScheme = "stark"
)

type HybridMode string

const (
	ModeAny         HybridMode = "any"
	ModeBoth        HybridMode = "both"
	ModePreferSNARK HybridMode = "prefer_snark"
)

type VerifyRequest struct {
	Mode         HybridMode `json:"mode"`
	SNARKProof   []byte     `json:"snark_proof"`
	STARKProof   []byte     `json:"stark_proof"`
	STARKBackend string     `json:"stark_backend"`
}

type VerifyResult struct {
	SNARKValid   bool   `json:"snark_valid"`
	STARKValid   bool   `json:"stark_valid"`
	Accepted     bool   `json:"accepted"`
	Policy       string `json:"policy"`
	STARKBackend string `json:"stark_backend"`
}

type SNARKVerifier interface {
	Verify(proof []byte) (bool, error)
}

type STARKVerifier interface {
	BackendName() string
	Verify(proof []byte) (bool, error)
}

var (
	registryMu         sync.RWMutex
	starkBackends                    = map[string]STARKVerifier{}
	defaultSNARKBridge SNARKVerifier = snarkVerifier{}
)

func init() {
	RegisterSTARKBackend(friVerifier{})
	RegisterSTARKBackend(winterfellVerifier{})
	if cmd := strings.TrimSpace(os.Getenv("MOHAWK_STARK_VERIFY_CMD")); cmd != "" {
		RegisterSTARKBackend(externalCommandVerifier{command: cmd})
	}
}

func RegisterSTARKBackend(verifier STARKVerifier) {
	if verifier == nil || verifier.BackendName() == "" {
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	starkBackends[verifier.BackendName()] = verifier
}

func AvailableSTARKBackends() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, 0, len(starkBackends))
	for name := range starkBackends {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func resolveSTARKBackend(name string) (STARKVerifier, string, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if name == "" {
		name = strings.TrimSpace(os.Getenv("MOHAWK_DEFAULT_STARK_BACKEND"))
		if name == "" {
			name = "simulated_fri"
		}
	}
	verifier, ok := starkBackends[name]
	if !ok {
		return nil, "", fmt.Errorf("unknown stark backend %q", name)
	}
	return verifier, name, nil
}

func VerifyHybrid(req VerifyRequest) (VerifyResult, error) {
	if req.Mode == "" {
		req.Mode = ModePreferSNARK
	}
	starkVerifier, starkBackend, err := resolveSTARKBackend(req.STARKBackend)
	if err != nil {
		return VerifyResult{}, err
	}
	snarkOK, snarkErr := defaultSNARKBridge.Verify(req.SNARKProof)
	starkOK, starkErr := starkVerifier.Verify(req.STARKProof)

	result := VerifyResult{
		SNARKValid:   snarkOK,
		STARKValid:   starkOK,
		Policy:       string(req.Mode),
		STARKBackend: starkBackend,
	}

	switch req.Mode {
	case ModeBoth:
		result.Accepted = snarkOK && starkOK
	case ModeAny:
		result.Accepted = snarkOK || starkOK
	case ModePreferSNARK:
		result.Accepted = snarkOK || starkOK
	default:
		return VerifyResult{}, fmt.Errorf("unsupported hybrid mode: %s", req.Mode)
	}

	if !result.Accepted {
		return result, errors.Join(
			fmt.Errorf("hybrid policy %q rejected proof set", req.Mode),
			snarkErr,
			starkErr,
		)
	}
	return result, nil
}

type snarkVerifier struct{}

func (snarkVerifier) Verify(proof []byte) (bool, error) {
	if len(proof) == 0 {
		return false, fmt.Errorf("snark proof missing")
	}
	if len(proof) < 128 {
		proof = append(proof, make([]byte, 128-len(proof))...)
	}
	ok, err := internalpkg.VerifyProof(proof, nil)
	if err != nil {
		return false, fmt.Errorf("snark verify failed: %w", err)
	}
	return ok, nil
}

type friVerifier struct{}

func (friVerifier) BackendName() string { return "simulated_fri" }

func (friVerifier) Verify(proof []byte) (bool, error) {
	const minProofBytes = 64
	if len(proof) == 0 {
		return false, fmt.Errorf("stark proof missing")
	}
	if len(proof) < minProofBytes {
		return false, fmt.Errorf("stark proof too short: got %d bytes, need %d (root[32]+content[32+])", len(proof), minProofBytes)
	}
	root := proof[0:32]
	content := proof[32:]
	expected := sha256.Sum256(content)
	if string(root) != string(expected[:]) {
		return false, fmt.Errorf("FRI commitment mismatch: root does not match SHA256(transcript)")
	}
	return true, nil
}

func GenFRIProof(content []byte) []byte {
	root := sha256.Sum256(content)
	result := make([]byte, 32+len(content))
	copy(result[:32], root[:])
	copy(result[32:], content)
	return result
}

type winterfellVerifier struct{}

func (winterfellVerifier) BackendName() string { return "winterfell_mock" }

func (winterfellVerifier) Verify(proof []byte) (bool, error) {
	const minProofBytes = 96
	const domainSep = "winterfell-v1:"
	if len(proof) == 0 {
		return false, fmt.Errorf("winterfell stark proof missing")
	}
	if len(proof) < minProofBytes {
		return false, fmt.Errorf("winterfell proof too short: got %d bytes, need %d (root[32]+transcript[64+])", len(proof), minProofBytes)
	}
	root := proof[0:32]
	transcript := proof[32:]
	h := sha256.New()
	h.Write([]byte(domainSep))
	h.Write(transcript)
	expected := h.Sum(nil)
	if string(root) != string(expected) {
		return false, fmt.Errorf("winterfell commitment mismatch: root does not match SHA256(%q || transcript)", domainSep)
	}
	return true, nil
}

func GenWinterfellProof(transcript []byte) []byte {
	const domainSep = "winterfell-v1:"
	h := sha256.New()
	h.Write([]byte(domainSep))
	h.Write(transcript)
	root := h.Sum(nil)
	result := make([]byte, 32+len(transcript))
	copy(result[:32], root)
	copy(result[32:], transcript)
	return result
}

type externalCommandVerifier struct {
	command string
}

func (externalCommandVerifier) BackendName() string { return "external_cmd" }

func (v externalCommandVerifier) Verify(proof []byte) (bool, error) {
	if len(proof) == 0 {
		return false, fmt.Errorf("external stark proof missing")
	}
	if strings.TrimSpace(v.command) == "" {
		return false, fmt.Errorf("external stark verify command is not configured")
	}
	timeout := 5 * time.Second
	if raw := strings.TrimSpace(os.Getenv("MOHAWK_STARK_VERIFY_TIMEOUT")); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tokens := strings.Fields(v.command)
	if len(tokens) == 0 {
		return false, fmt.Errorf("external stark verify command is not configured")
	}
	for _, token := range tokens {
		if strings.ContainsAny(token, "\n\r;&|`$><") {
			return false, fmt.Errorf("external stark verify command contains unsupported shell control characters")
		}
	}

	cmd := exec.CommandContext(ctx, tokens[0], tokens[1:]...) // #nosec G204,G702 -- tokens are validated and shell control characters are rejected
	cmd.Stdin = strings.NewReader(string(proof))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("external stark backend failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	response := strings.TrimSpace(strings.ToLower(string(out)))
	if response == "" || response == "ok" || response == "valid" || response == "true" {
		return true, nil
	}
	if strings.Contains(response, "invalid") || strings.Contains(response, "false") {
		return false, fmt.Errorf("external stark backend reported invalid proof")
	}
	return true, nil
}
