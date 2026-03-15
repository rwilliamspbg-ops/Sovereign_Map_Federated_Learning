package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func getFreeListenAddress(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate free port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func waitForHealthy(t *testing.T, baseURL string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}

	t.Fatalf("node-agent did not become healthy within %s", timeout)
}

func postJSON(t *testing.T, client *http.Client, url string, body string, token string, role string) (*http.Response, map[string]interface{}) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if role != "" {
		req.Header.Set("X-API-Role", role)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}

	payload := map[string]interface{}{}
	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	_ = json.Unmarshal(respBody, &payload)
	return resp, payload
}

func TestNodeAgentEndpointsEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	tokenDir := t.TempDir()
	tokenFile := filepath.Join(tokenDir, "mohawk_api_token")
	tokenValue := "integration-token"
	if err := os.WriteFile(tokenFile, []byte(tokenValue), 0o600); err != nil {
		t.Fatalf("write token file: %v", err)
	}

	listenAddr := getFreeListenAddress(t)
	baseURL := "http://" + listenAddr

	buildCtx, buildCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer buildCancel()

	binPath := filepath.Join(t.TempDir(), "node-agent-bin")
	buildCmd := exec.CommandContext(buildCtx, "go", "build", "-o", binPath, "./cmd/node-agent")
	buildCmd.Dir = repoRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build node-agent binary failed: %v\n%s", err, string(output))
	}

	runCtx, runCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer runCancel()

	cmd := exec.CommandContext(runCtx, binPath)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(),
		"MOHAWK_API_LISTEN="+listenAddr,
		"MOHAWK_API_AUTH_MODE=file-only",
		"MOHAWK_API_TOKEN_FILE="+tokenFile,
		"MOHAWK_API_ENFORCE_ROLES=true",
		"MOHAWK_API_HYBRID_ALLOWED_ROLES=verifier,admin",
		"MOHAWK_API_PROOF_ALLOWED_ROLES=verifier,admin",
		"MOHAWK_CAPABILITIES_PATH="+filepath.Join(repoRoot, "capabilities.json"),
		"MOHAWK_BRIDGE_POLICIES_PATH="+filepath.Join(repoRoot, "bridge-policies.json"),
	)

	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined

	if err := cmd.Start(); err != nil {
		t.Fatalf("start node-agent: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		waitDone := make(chan struct{})
		go func() {
			_ = cmd.Wait()
			close(waitDone)
		}()
		select {
		case <-waitDone:
		case <-time.After(5 * time.Second):
		}
	}()

	waitForHealthy(t, baseURL, 25*time.Second)

	client := &http.Client{Timeout: 5 * time.Second}

	capResp, err := client.Get(baseURL + "/api/v1/capabilities")
	if err != nil {
		t.Fatalf("capabilities request failed: %v", err)
	}
	if capResp.StatusCode != http.StatusOK {
		_ = capResp.Body.Close()
		t.Fatalf("capabilities status=%d output=%s", capResp.StatusCode, combined.String())
	}
	_ = capResp.Body.Close()

	verifyBody := `{"encoding":"raw","proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	unauthResp, _ := postJSON(t, client, baseURL+"/api/v1/proof/verify", verifyBody, "", "")
	if unauthResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized proof status=%d output=%s", unauthResp.StatusCode, combined.String())
	}

	authResp, verifyPayload := postJSON(t, client, baseURL+"/api/v1/proof/verify", verifyBody, tokenValue, "verifier")
	if authResp.StatusCode != http.StatusOK {
		t.Fatalf("authorized proof status=%d payload=%v output=%s", authResp.StatusCode, verifyPayload, combined.String())
	}
	if _, ok := verifyPayload["valid"]; !ok {
		t.Fatalf("proof response missing valid field: %v", verifyPayload)
	}

	hybridBody := `{"mode":"any","encoding":"raw","snark_proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","stark_proof":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`
	hybridResp, hybridPayload := postJSON(t, client, baseURL+"/api/v1/proof/hybrid/verify", hybridBody, tokenValue, "verifier")
	if hybridResp.StatusCode != http.StatusOK {
		t.Fatalf("hybrid status=%d payload=%v output=%s", hybridResp.StatusCode, hybridPayload, combined.String())
	}
	if _, ok := hybridPayload["accepted"]; !ok {
		t.Fatalf("hybrid response missing accepted field: %v", hybridPayload)
	}

	// ── Ledger checks ─────────────────────────────────────────────────────────

	// Unauthenticated ledger request must be rejected.
	ledgerUnauthReq, _ := http.NewRequest(http.MethodGet, baseURL+"/api/v1/ledger", nil)
	ledgerUnauthResp, err := client.Do(ledgerUnauthReq)
	if err != nil {
		t.Fatalf("ledger unauth request: %v", err)
	}
	_ = ledgerUnauthResp.Body.Close()
	if ledgerUnauthResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("ledger without token: status=%d, want 401", ledgerUnauthResp.StatusCode)
	}

	// Authenticated ledger request must return the events we just created.
	ledgerReq, _ := http.NewRequest(http.MethodGet, baseURL+"/api/v1/ledger", nil)
	ledgerReq.Header.Set("Authorization", "Bearer "+tokenValue)
	ledgerReq.Header.Set("X-API-Role", "verifier")
	ledgerResp, err := client.Do(ledgerReq)
	if err != nil {
		t.Fatalf("ledger request: %v", err)
	}
	var ledgerPayload map[string]interface{}
	ledgerBody, _ := io.ReadAll(ledgerResp.Body)
	_ = ledgerResp.Body.Close()
	_ = json.Unmarshal(ledgerBody, &ledgerPayload)

	if ledgerResp.StatusCode != http.StatusOK {
		t.Fatalf("ledger status=%d payload=%v output=%s", ledgerResp.StatusCode, ledgerPayload, combined.String())
	}
	count, _ := ledgerPayload["count"].(float64)
	if count < 1 {
		t.Fatalf("ledger count=%v, expected >= 1 after proof verifications", count)
	}

	if runCtx.Err() != nil {
		t.Fatalf("integration context expired: %v\nprocess output:\n%s", runCtx.Err(), combined.String())
	}
}
