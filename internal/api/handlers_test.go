package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
)

func mustStringSlice(t *testing.T, value interface{}, field string) []string {
	t.Helper()
	raw, ok := value.([]interface{})
	if !ok {
		t.Fatalf("%s is not a list: %T", field, value)
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("%s contains non-string value: %T", field, v)
		}
		out = append(out, s)
	}
	return out
}

func configureProofAuthForTests(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "api_token")
	if err := os.WriteFile(tokenPath, []byte("test-token"), 0o600); err != nil {
		t.Fatalf("write token file: %v", err)
	}
	t.Setenv("MOHAWK_API_AUTH_MODE", "file-only")
	t.Setenv("MOHAWK_API_TOKEN_FILE", tokenPath)
	t.Setenv("MOHAWK_API_ENFORCE_ROLES", "true")
	t.Setenv("MOHAWK_API_PROOF_ALLOWED_ROLES", "verifier,admin")
}

func TestDecodePayloadBase64(t *testing.T) {
	raw := []byte("proof-bytes")
	payload := base64.StdEncoding.EncodeToString(raw)
	decoded, err := decodePayload(payload, "base64")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if string(decoded) != string(raw) {
		t.Fatalf("decoded mismatch: got %q want %q", string(decoded), string(raw))
	}
}

func TestDecodePayloadHex(t *testing.T) {
	decoded, err := decodePayload("0x616263", "hex")
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if string(decoded) != "abc" {
		t.Fatalf("decoded mismatch: got %q want %q", string(decoded), "abc")
	}
}

func TestGetCapabilities(t *testing.T) {
	tmpDir := t.TempDir()
	capPath := filepath.Join(tmpDir, "capabilities.json")
	bridgePath := filepath.Join(tmpDir, "bridge-policies.json")

	if err := os.WriteFile(capPath, []byte(`{"version":"1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write capabilities: %v", err)
	}
	if err := os.WriteFile(bridgePath, []byte(`{"version":"v1"}`), 0o644); err != nil {
		t.Fatalf("write bridge policies: %v", err)
	}

	t.Setenv("MOHAWK_CAPABILITIES_PATH", capPath)
	t.Setenv("MOHAWK_BRIDGE_POLICIES_PATH", bridgePath)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if payload["protocol_version"] != "1.0.0" {
		t.Fatalf("unexpected protocol version: %v", payload["protocol_version"])
	}

	apiSection, ok := payload["api"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing api section: %v", payload)
	}

	authSection, ok := apiSection["auth"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing auth metadata in capabilities: %v", apiSection)
	}
	if authSection["default_mode"] != "file-only" {
		t.Fatalf("unexpected auth default mode: %v", authSection["default_mode"])
	}

	observability, ok := payload["observability"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing observability metadata in capabilities: %v", payload)
	}
	if _, ok := observability["proof_metrics"].([]interface{}); !ok {
		t.Fatalf("proof_metrics missing in observability metadata: %v", observability)
	}
	if _, ok := observability["ledger_metrics"].([]interface{}); !ok {
		t.Fatalf("ledger_metrics missing in observability metadata: %v", observability)
	}
}

func TestCapabilitiesContractV1(t *testing.T) {
	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}

	if payload["protocol_version"] != "1.0.0" {
		t.Fatalf("protocol_version = %v, want 1.0.0", payload["protocol_version"])
	}

	apiSection, ok := payload["api"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing api section")
	}
	if got, want := mustStringSlice(t, apiSection["base_paths"], "api.base_paths"), []string{"/api", "/api/v1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("api.base_paths = %v, want %v", got, want)
	}

	if got, want := mustStringSlice(t, apiSection["open_endpoints"], "api.open_endpoints"), []string{"GET /health", "GET /api/v1/status", "GET /api/v1/capabilities", "GET /api/v1/verification_policy"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("api.open_endpoints = %v, want %v", got, want)
	}

	if got, want := mustStringSlice(t, apiSection["auth_protected_endpoints"], "api.auth_protected_endpoints"), []string{"POST /api/v1/proof/verify", "POST /api/v1/proof/hybrid/verify", "GET /api/v1/ledger", "POST /api/v1/verification_policy"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("api.auth_protected_endpoints = %v, want %v", got, want)
	}

	proofPayload, ok := apiSection["proof_payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("api.proof_payload missing")
	}
	if got, want := mustStringSlice(t, proofPayload["fields"], "api.proof_payload.fields"), []string{"proof", "encoding", "public_input"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("api.proof_payload.fields = %v, want %v", got, want)
	}

	if got, want := mustStringSlice(t, proofPayload["supported_encodings"], "api.proof_payload.supported_encodings"), []string{"base64", "hex", "raw"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("api.proof_payload.supported_encodings = %v, want %v", got, want)
	}

	hybridPayload, ok := apiSection["hybrid_payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("api.hybrid_payload missing")
	}
	if got, want := mustStringSlice(t, hybridPayload["supported_modes"], "api.hybrid_payload.supported_modes"), []string{"any", "both", "prefer_snark"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("api.hybrid_payload.supported_modes = %v, want %v", got, want)
	}

	authSection, ok := apiSection["auth"].(map[string]interface{})
	if !ok {
		t.Fatalf("api.auth missing")
	}
	if authSection["default_mode"] != "file-only" {
		t.Fatalf("api.auth.default_mode = %v, want file-only", authSection["default_mode"])
	}
	if authSection["default_token_file"] != "/run/secrets/mohawk_api_token" {
		t.Fatalf("api.auth.default_token_file = %v, want /run/secrets/mohawk_api_token", authSection["default_token_file"])
	}

	observability, ok := payload["observability"].(map[string]interface{})
	if !ok {
		t.Fatalf("observability missing")
	}
	if got := mustStringSlice(t, observability["proof_metrics"], "observability.proof_metrics"); len(got) == 0 {
		t.Fatal("observability.proof_metrics must not be empty")
	}
	if got := mustStringSlice(t, observability["ledger_metrics"], "observability.ledger_metrics"); len(got) == 0 {
		t.Fatal("observability.ledger_metrics must not be empty")
	}
}

func TestVerifyHybridProofEndpoint(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"mode":"any","encoding":"raw","snark_proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","stark_proof":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/proof/hybrid/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "verifier")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want 200", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if _, ok := payload["accepted"]; !ok {
		t.Fatalf("response missing accepted field: %v", payload)
	}
}

func TestVerifyProofEndpointRejectsUnauthorized(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"encoding":"raw","proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/proof/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want 401", w.Code)
	}
}

func TestVerifyProofEndpointRejectsInvalidJSON(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/proof/verify", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "verifier")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want 400", w.Code)
	}
}

func TestVerifyProofEndpointRejectsInvalidPayload(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"encoding":"base64","proof":"%%%not-base64%%%"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/proof/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "verifier")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want 400", w.Code)
	}
}

func TestVerifyProofEndpointRejectsRoleMismatch(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"encoding":"raw","proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/proof/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "observer")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status code = %d, want 403", w.Code)
	}
}

func TestVerifyHybridProofRejectsInvalidJSON(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/proof/hybrid/verify", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "verifier")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want 400", w.Code)
	}
}

func TestVerifyHybridProofRejectsInvalidSNARKPayload(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"mode":"any","encoding":"base64","snark_proof":"###","stark_proof":"YmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYmJiYg=="}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/proof/hybrid/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "verifier")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want 400", w.Code)
	}
}

func TestVerifyHybridProofRejectsInvalidToken(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"mode":"any","encoding":"raw","snark_proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","stark_proof":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/proof/hybrid/verify", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong-token")
	req.Header.Set("X-API-Role", "verifier")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want 401", w.Code)
	}
}

// ─── Ledger tests ─────────────────────────────────────────────────────────────

func TestProofLedgerRecord(t *testing.T) {
	l := NewProofLedger(10)
	l.Record("snark_verify", []byte("proof"), "verifier", true, 5, nil)
	if l.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", l.Len())
	}
	entries := l.Entries()
	if entries[0].EventType != "snark_verify" {
		t.Errorf("EventType = %q, want snark_verify", entries[0].EventType)
	}
	if !entries[0].Accepted {
		t.Error("expected Accepted = true")
	}
	if entries[0].ID != 1 {
		t.Errorf("ID = %d, want 1", entries[0].ID)
	}
	if entries[0].ProofHash == "" {
		t.Error("ProofHash must not be empty")
	}
}

func TestProofLedgerRingBuffer(t *testing.T) {
	cap := 3
	l := NewProofLedger(cap)
	for i := 0; i < 5; i++ {
		l.Record("snark_verify", []byte("proof"), "admin", true, 1, nil)
	}
	if l.Len() != cap {
		t.Fatalf("expected %d entries (ring buffer), got %d", cap, l.Len())
	}
	// Oldest entry should have been evicted; last 3 IDs should be 3,4,5.
	entries := l.Entries()
	if entries[0].ID != 3 {
		t.Errorf("oldest ID = %d, want 3 after ring overflow", entries[0].ID)
	}
	if entries[cap-1].ID != 5 {
		t.Errorf("newest ID = %d, want 5", entries[cap-1].ID)
	}
}

func TestProofLedgerRetentionHighVolume(t *testing.T) {
	capacity := 128
	total := 10000
	l := NewProofLedger(capacity)

	for i := 0; i < total; i++ {
		l.Record("snark_verify", []byte("proof"), "verifier", i%2 == 0, int64(i%7), nil)
	}

	if l.Len() != capacity {
		t.Fatalf("ledger len = %d, want %d", l.Len(), capacity)
	}

	entries := l.Entries()
	wantFirst := uint64(total - capacity + 1)
	wantLast := uint64(total)
	if entries[0].ID != wantFirst {
		t.Fatalf("oldest id = %d, want %d", entries[0].ID, wantFirst)
	}
	if entries[len(entries)-1].ID != wantLast {
		t.Fatalf("newest id = %d, want %d", entries[len(entries)-1].ID, wantLast)
	}
}

func TestProofLedgerConcurrentRecordRetention(t *testing.T) {
	capacity := 64
	workers := 8
	perWorker := 250
	total := workers * perWorker
	l := NewProofLedger(capacity)

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				l.Record("hybrid_verify", []byte("proof"), "admin", (worker+i)%2 == 0, 1, nil)
			}
		}(w)
	}
	wg.Wait()

	if l.Len() != capacity {
		t.Fatalf("ledger len = %d, want %d", l.Len(), capacity)
	}

	entries := l.Entries()
	if entries[0].ID != uint64(total-capacity+1) {
		t.Fatalf("oldest id = %d, want %d", entries[0].ID, total-capacity+1)
	}
	if entries[len(entries)-1].ID != uint64(total) {
		t.Fatalf("newest id = %d, want %d", entries[len(entries)-1].ID, total)
	}
}

func TestGetLedgerEndpoint(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Emit a hybrid verify event so the ledger has at least one entry.
	hybridBody := `{"mode":"any","encoding":"raw","snark_proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","stark_proof":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`
	post := httptest.NewRequest(http.MethodPost, "/api/v1/proof/hybrid/verify", strings.NewReader(hybridBody))
	post.Header.Set("Content-Type", "application/json")
	post.Header.Set("Authorization", "Bearer test-token")
	post.Header.Set("X-API-Role", "verifier")
	mux.ServeHTTP(httptest.NewRecorder(), post)

	// GET /api/v1/ledger with valid auth.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ledger", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "verifier")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	count, ok := resp["count"].(float64)
	if !ok || count < 1 {
		t.Fatalf("expected count >= 1 in ledger response; got %v", resp)
	}
}

func TestGetLedgerEndpointRejectsUnauthorized(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ledger", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestGetLedgerEndpointRejectsInvalidToken(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ledger", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	req.Header.Set("X-API-Role", "verifier")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestGetLedgerEndpointRejectsRoleMismatch(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ledger", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "observer")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestGetTrustStatusIncludesBlockchainVerificationState(t *testing.T) {
	h := NewHandler(nil, nil, nil, nil)
	chain := blockchain.NewBlockChain()
	if err := chain.SetVerificationPolicy(blockchain.VerificationPolicy{
		RequireProof:                true,
		MinConfidenceBps:            9100,
		RejectOnVerificationFailure: true,
		AllowConsensusProof:         true,
		AllowZKProof:                false,
		AllowTEEProof:               true,
	}); err != nil {
		t.Fatalf("set verification policy: %v", err)
	}
	h.SetBlockchain(chain)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/trust_status", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}

	if payload["trust_mode"] != "p2p-reputation+governed-proof-verification" {
		t.Fatalf("trust_mode = %v", payload["trust_mode"])
	}

	policy, ok := payload["verification_policy"].(map[string]interface{})
	if !ok {
		t.Fatalf("verification_policy missing: %v", payload)
	}
	if got := policy["require_proof"]; got != true {
		t.Fatalf("require_proof = %v, want true", got)
	}
	if got := policy["min_confidence_bps"]; got != float64(9100) {
		t.Fatalf("min_confidence_bps = %v, want 9100", got)
	}
	if got := policy["allow_zk_proof"]; got != false {
		t.Fatalf("allow_zk_proof = %v, want false", got)
	}

	verification, ok := payload["fl_verification"].(map[string]interface{})
	if !ok {
		t.Fatalf("fl_verification missing: %v", payload)
	}
	if got := verification["total_rounds"]; got != float64(0) {
		t.Fatalf("total_rounds = %v, want 0", got)
	}
	if got := verification["average_confidence_bps"]; got != float64(0) {
		t.Fatalf("average_confidence_bps = %v, want 0", got)
	}
}

func TestGetVerificationPolicyEndpoint(t *testing.T) {
	h := NewHandler(nil, nil, nil, nil)
	chain := blockchain.NewBlockChain()
	h.SetBlockchain(chain)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/verification_policy", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if _, ok := payload["verification_policy"].(map[string]interface{}); !ok {
		t.Fatalf("verification_policy missing: %v", payload)
	}
}

func TestUpdateVerificationPolicyEndpointRequiresAdmin(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	chain := blockchain.NewBlockChain()
	h.SetBlockchain(chain)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"require_proof":true,"min_confidence_bps":9200,"reject_on_verification_failure":true,"allow_consensus_proof":true,"allow_zk_proof":false,"allow_tee_proof":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/verification_policy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "verifier")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestUpdateVerificationPolicyEndpoint(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	chain := blockchain.NewBlockChain()
	h.SetBlockchain(chain)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"require_proof":true,"min_confidence_bps":9200,"reject_on_verification_failure":true,"allow_consensus_proof":true,"allow_zk_proof":false,"allow_tee_proof":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/verification_policy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-API-Role", "admin")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 body=%s", w.Code, w.Body.String())
	}

	policy := chain.GetVerificationPolicy()
	if !policy.RequireProof || policy.MinConfidenceBps != 9200 || policy.AllowZKProof {
		t.Fatalf("unexpected policy after update: %+v", policy)
	}

	value, err := chain.StateDB.Get("verification_policy:active")
	if err != nil {
		t.Fatalf("persisted policy missing: %v", err)
	}
	persisted, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("persisted policy wrong type: %T", value)
	}
	if persisted["min_confidence_bps"] != uint32(9200) {
		t.Fatalf("persisted min_confidence_bps = %v", persisted["min_confidence_bps"])
	}
}
