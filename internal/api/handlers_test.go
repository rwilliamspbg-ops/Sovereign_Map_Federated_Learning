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
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/monitoring"
)

type mockStatusReader struct {
	status map[string]interface{}
}

func (m mockStatusReader) GetRuntimeStatus() map[string]interface{} {
	copy := make(map[string]interface{}, len(m.status))
	for k, v := range m.status {
		copy[k] = v
	}
	return copy
}

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
	tmpDir := t.TempDir()
	capPath := filepath.Join(tmpDir, "capabilities.json")
	bridgePath := filepath.Join(tmpDir, "bridge-policies.json")

	if err := os.WriteFile(capPath, []byte(`{"version":"1.0.0","runtime":"mohawk-proto-v1","thinker_clauses":{"enabled":true,"escalation_label":"thinker-review"}}`), 0o644); err != nil {
		t.Fatalf("write capabilities: %v", err)
	}
	if err := os.WriteFile(bridgePath, []byte(`{"version":"v1","routes":[]}`), 0o644); err != nil {
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
		t.Fatalf("protocol_version = %v, want 1.0.0", payload["protocol_version"])
	}

	apiSection, ok := payload["api"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing api section")
	}
	if got, want := mustStringSlice(t, apiSection["base_paths"], "api.base_paths"), []string{"/api", "/api/v1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("api.base_paths = %v, want %v", got, want)
	}

	if got, want := mustStringSlice(t, apiSection["open_endpoints"], "api.open_endpoints"), []string{"GET /health", "GET /readyz", "GET /api/v1/status", "GET /api/v1/readiness", "GET /api/v1/consensus/status", "GET /api/v1/capabilities", "GET /api/v1/verification_policy", "GET /api/v1/trust_snapshot"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("api.open_endpoints = %v, want %v", got, want)
	}

	if got, want := mustStringSlice(t, apiSection["auth_protected_endpoints"], "api.auth_protected_endpoints"), []string{"POST /api/v1/proof/verify", "POST /api/v1/proof/hybrid/verify", "GET /api/v1/ledger", "GET /api/v1/ledger/reconcile", "POST /api/v1/verification_policy"}; !reflect.DeepEqual(got, want) {
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

	capabilitiesSection, ok := payload["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatalf("capabilities missing")
	}
	thinkerClauses, ok := capabilitiesSection["thinker_clauses"].(map[string]interface{})
	if !ok {
		t.Fatalf("capabilities.thinker_clauses missing")
	}
	if thinkerClauses["enabled"] != true {
		t.Fatalf("capabilities.thinker_clauses.enabled = %v, want true", thinkerClauses["enabled"])
	}
	if thinkerClauses["escalation_label"] != "thinker-review" {
		t.Fatalf("capabilities.thinker_clauses.escalation_label = %v, want thinker-review", thinkerClauses["escalation_label"])
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
	if entries[0].EntryID == "" || entries[0].EntryHash == "" {
		t.Fatal("entry hashes and deterministic IDs must be populated")
	}
}

func TestProofLedgerIdempotentReplay(t *testing.T) {
	l := NewProofLedger(10)
	opts := LedgerRecordOptions{StreamID: "proof.snark", IdempotencyKey: "idem-1"}
	first, replay := l.RecordWithOptions("snark_verify", []byte("proof"), "verifier", true, 1, nil, opts)
	if replay {
		t.Fatal("first write must not be replay")
	}
	second, replay := l.RecordWithOptions("snark_verify", []byte("proof"), "verifier", true, 1, nil, opts)
	if !replay {
		t.Fatal("second write with same idempotency key must be replay")
	}
	if l.Len() != 1 {
		t.Fatalf("ledger len = %d, want 1", l.Len())
	}
	if first.EntryID != second.EntryID {
		t.Fatalf("entry IDs differ for idempotent replay: %s vs %s", first.EntryID, second.EntryID)
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

func TestProofLedgerReconcileHealthy(t *testing.T) {
	l := NewProofLedger(10)
	l.RecordWithOptions("snark_verify", []byte("proof-1"), "verifier", true, 1, nil, LedgerRecordOptions{StreamID: "proof.snark", IdempotencyKey: "k1"})
	l.RecordWithOptions("snark_verify", []byte("proof-2"), "verifier", true, 1, nil, LedgerRecordOptions{StreamID: "proof.snark", IdempotencyKey: "k2"})

	report := l.Reconcile()
	if !report.Healthy {
		t.Fatalf("expected healthy reconcile report, issues: %v", report.Issues)
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

func TestGetLedgerReconcileEndpoint(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"encoding":"raw","proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	post := httptest.NewRequest(http.MethodPost, "/api/v1/proof/verify", strings.NewReader(body))
	post.Header.Set("Content-Type", "application/json")
	post.Header.Set("Authorization", "Bearer test-token")
	post.Header.Set("X-API-Role", "verifier")
	post.Header.Set("Idempotency-Key", "test-proof-1")
	mux.ServeHTTP(httptest.NewRecorder(), post)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ledger/reconcile", nil)
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
	healthy, ok := resp["healthy"].(bool)
	if !ok || !healthy {
		t.Fatalf("expected healthy=true, got %v", resp)
	}
}

func TestVerifyProofIdempotentReplay(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"encoding":"raw","proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`

	post1 := httptest.NewRequest(http.MethodPost, "/api/v1/proof/verify", strings.NewReader(body))
	post1.Header.Set("Content-Type", "application/json")
	post1.Header.Set("Authorization", "Bearer test-token")
	post1.Header.Set("X-API-Role", "verifier")
	post1.Header.Set("Idempotency-Key", "idem-proof-1")
	w1 := httptest.NewRecorder()
	mux.ServeHTTP(w1, post1)

	post2 := httptest.NewRequest(http.MethodPost, "/api/v1/proof/verify", strings.NewReader(body))
	post2.Header.Set("Content-Type", "application/json")
	post2.Header.Set("Authorization", "Bearer test-token")
	post2.Header.Set("X-API-Role", "verifier")
	post2.Header.Set("Idempotency-Key", "idem-proof-1")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, post2)

	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Fatalf("verify responses status mismatch: first=%d second=%d", w1.Code, w2.Code)
	}

	var first map[string]interface{}
	if err := json.Unmarshal(w1.Body.Bytes(), &first); err != nil {
		t.Fatalf("first decode: %v", err)
	}
	var second map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &second); err != nil {
		t.Fatalf("second decode: %v", err)
	}
	if first["replay"] != false {
		t.Fatalf("first replay = %v, want false", first["replay"])
	}
	if second["replay"] != true {
		t.Fatalf("second replay = %v, want true", second["replay"])
	}
}

func TestCockroachBackendFallbackMetadata(t *testing.T) {
	t.Setenv("MOHAWK_LEDGER_BACKEND", "cockroach")
	t.Setenv("MOHAWK_LEDGER_SQL_DSN", "")

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/capabilities", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	observability, ok := resp["observability"].(map[string]interface{})
	if !ok {
		t.Fatalf("observability missing: %v", resp)
	}
	ledgerState, ok := observability["ledger_state"].(map[string]interface{})
	if !ok {
		t.Fatalf("ledger_state missing: %v", observability)
	}
	if ledgerState["storage_mode"] != "cockroach-compatible-inmemory" {
		t.Fatalf("storage_mode = %v, want cockroach-compatible-inmemory fallback", ledgerState["storage_mode"])
	}
	if ledgerState["has_error"] != true {
		t.Fatalf("has_error should be true for fallback: %v", ledgerState)
	}
}

func TestGetLedgerIncludesStorageMetadata(t *testing.T) {
	configureProofAuthForTests(t)

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

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
	if resp["storage_mode"] == "" {
		t.Fatalf("storage_mode must be present: %v", resp)
	}
}

func TestReadinessEndpointHealthyWithInMemoryBackend(t *testing.T) {
	t.Setenv("MOHAWK_LEDGER_BACKEND", "inmemory")

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/readiness", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if resp["ready"] != true {
		t.Fatalf("ready = %v, want true", resp["ready"])
	}
}

func TestReadinessEndpointDegradedOnSQLInitFailure(t *testing.T) {
	t.Setenv("MOHAWK_LEDGER_BACKEND", "cockroach")
	t.Setenv("MOHAWK_LEDGER_SQL_DSN", "")

	h := NewHandler(nil, nil, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if resp["ready"] != false {
		t.Fatalf("ready = %v, want false", resp["ready"])
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

func TestGetConsensusStatusEndpoint(t *testing.T) {
	h := NewHandler(nil, nil, nil, nil)
	h.SetConsensusReaders(
		mockStatusReader{status: map[string]interface{}{"state": "voting", "active_node_count": 8}},
		mockStatusReader{status: map[string]interface{}{"round_number": 12, "async_mode": true}},
	)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/consensus/status", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}

	consensus, ok := payload["consensus"].(map[string]interface{})
	if !ok {
		t.Fatalf("consensus missing: %v", payload)
	}
	if consensus["available"] != true {
		t.Fatalf("consensus.available = %v, want true", consensus["available"])
	}
	if consensus["state"] != "voting" {
		t.Fatalf("consensus.state = %v, want voting", consensus["state"])
	}

	aggregation, ok := payload["aggregation"].(map[string]interface{})
	if !ok {
		t.Fatalf("aggregation missing: %v", payload)
	}
	if aggregation["available"] != true {
		t.Fatalf("aggregation.available = %v, want true", aggregation["available"])
	}
	if aggregation["async_mode"] != true {
		t.Fatalf("aggregation.async_mode = %v, want true", aggregation["async_mode"])
	}
}

func TestGetMetricsIncludesChurnAndStaleness(t *testing.T) {
	collector := monitoring.NewCollector(100)
	collector.RecordNodeJoin("node-a")
	collector.RecordNodeJoin("node-b")
	collector.RecordNodeLeave("node-c")
	collector.RecordAsyncStaleness("node-d", 0.4, map[string]string{"source": "async"})
	collector.RecordAsyncStaleness("node-e", 0.6, map[string]string{"source": "async"})

	h := NewHandler(nil, nil, collector, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}

	if payload["churn_joins_total"] != float64(2) {
		t.Fatalf("churn_joins_total = %v, want 2", payload["churn_joins_total"])
	}
	if payload["churn_leaves_total"] != float64(1) {
		t.Fatalf("churn_leaves_total = %v, want 1", payload["churn_leaves_total"])
	}
	if payload["async_staleness_avg_seconds"] != 0.5 {
		t.Fatalf("async_staleness_avg_seconds = %v, want 0.5", payload["async_staleness_avg_seconds"])
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

func TestGetTrustSnapshotIncludesPolicyHistory(t *testing.T) {
	h := NewHandler(nil, nil, nil, nil)
	chain := blockchain.NewBlockChain()
	h.SetBlockchain(chain)

	if err := chain.StateDB.Set("governance_verification_audit:proposal-1:1700000000", map[string]interface{}{
		"proposal_id": "proposal-1",
		"timestamp":   int64(1700000000),
		"new_policy":  map[string]interface{}{"min_confidence_bps": uint32(8000)},
	}); err != nil {
		t.Fatalf("seed first history entry: %v", err)
	}
	if err := chain.StateDB.Set("api_verification_policy_audit:1700000100", map[string]interface{}{
		"proposal_id": "api-direct",
		"timestamp":   int64(1700000100),
		"source":      "api",
		"new_policy":  map[string]interface{}{"min_confidence_bps": uint32(9200)},
	}); err != nil {
		t.Fatalf("seed second history entry: %v", err)
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trust_snapshot", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if _, ok := payload["trust_status"].(map[string]interface{}); !ok {
		t.Fatalf("trust_status missing: %v", payload)
	}
	history, ok := payload["policy_history"].([]interface{})
	if !ok {
		t.Fatalf("policy_history missing: %v", payload)
	}
	if len(history) != 2 {
		t.Fatalf("policy_history len = %d, want 2", len(history))
	}
	first := history[0].(map[string]interface{})
	if first["proposal_id"] != "api-direct" {
		t.Fatalf("first history entry = %v, want api-direct first", first)
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

	auditEntries := chain.StateDB.GetAll()
	foundAudit := false
	for key, value := range auditEntries {
		if !strings.HasPrefix(key, "api_verification_policy_audit:") {
			continue
		}
		entry, ok := value.(map[string]interface{})
		if ok && entry["source"] == "api" {
			foundAudit = true
			break
		}
	}
	if !foundAudit {
		t.Fatal("expected api verification policy audit entry to be written")
	}
}
