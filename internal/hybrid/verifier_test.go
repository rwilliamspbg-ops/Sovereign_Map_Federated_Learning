// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package hybrid

import (
	"bytes"
	"strings"
	"testing"
)

func TestFRIProofRoundTrip(t *testing.T) {
	content := []byte("federated-gradient-round-42-aaaaaaa") // 35 bytes >= 32
	proof := GenFRIProof(content)
	v := friVerifier{}
	ok, err := v.Verify(proof)
	if err != nil {
		t.Fatalf("FRI verify error: %v", err)
	}
	if !ok {
		t.Fatal("FRI proof should be valid")
	}
}

func TestFRIProofRejectsTampered(t *testing.T) {
	proof := GenFRIProof([]byte("original"))
	proof[33] ^= 0xff
	v := friVerifier{}
	ok, _ := v.Verify(proof)
	if ok {
		t.Fatal("tampered FRI proof should be rejected")
	}
}

func TestFRIProofRejectsEmpty(t *testing.T) {
	v := friVerifier{}
	ok, err := v.Verify(nil)
	if ok || err == nil {
		t.Fatal("empty FRI proof should return error")
	}
}

func TestFRIProofRejectsTooShort(t *testing.T) {
	v := friVerifier{}
	ok, err := v.Verify([]byte("short"))
	if ok || err == nil {
		t.Fatal("short FRI proof should return error")
	}
}

func TestWinterfellProofRoundTrip(t *testing.T) {
	transcript := []byte("model-state-hash-abcdef-000000000000000000000000000000000000000000") // 66 bytes >= 64
	proof := GenWinterfellProof(transcript)
	v := winterfellVerifier{}
	ok, err := v.Verify(proof)
	if err != nil {
		t.Fatalf("winterfell verify error: %v", err)
	}
	if !ok {
		t.Fatal("winterfell proof should be valid")
	}
}

func TestWinterfellProofRejectsTampered(t *testing.T) {
	proof := GenWinterfellProof([]byte("transcript"))
	proof[33] ^= 0xff
	v := winterfellVerifier{}
	ok, _ := v.Verify(proof)
	if ok {
		t.Fatal("tampered winterfell proof should be rejected")
	}
}

func TestWinterfellProofRejectsEmpty(t *testing.T) {
	v := winterfellVerifier{}
	ok, err := v.Verify(nil)
	if ok || err == nil {
		t.Fatal("empty winterfell proof should return error")
	}
}

func TestAvailableSTARKBackends(t *testing.T) {
	backends := AvailableSTARKBackends()
	if len(backends) == 0 {
		t.Fatal("expected at least one registered STARK backend")
	}
	// simulated_fri and winterfell_mock must always be present
	has := func(name string) bool {
		for _, b := range backends {
			if b == name {
				return true
			}
		}
		return false
	}
	if !has("simulated_fri") {
		t.Error("simulated_fri backend missing")
	}
	if !has("winterfell_mock") {
		t.Error("winterfell_mock backend missing")
	}
}

func TestRegisterSTARKBackend(t *testing.T) {
	// nil / blank name must be silently ignored (no panic, no new entry)
	before := len(AvailableSTARKBackends())
	RegisterSTARKBackend(nil)
	after := len(AvailableSTARKBackends())
	if after != before {
		t.Errorf("RegisterSTARKBackend(nil) changed backend count: %d -> %d", before, after)
	}
}

func TestVerifyHybridModeAnyAcceptsOneGood(t *testing.T) {
	snark := make([]byte, 128)
	// FRI proof that intentionally fails (too short) => stark bad, snark irrelevant
	// Use a valid FRI proof so stark is good => accepted in ModeAny even if snark fails.
	content := []byte("test-transcript-padded-to-satisfy-min-size-requirement") // 54 bytes >= 32
	friProof := GenFRIProof(content)
	res, err := VerifyHybrid(VerifyRequest{
		Mode:         ModeAny,
		SNARKProof:   snark,
		STARKProof:   friProof,
		STARKBackend: "simulated_fri",
	})
	if err != nil {
		t.Fatalf("VerifyHybrid ModeAny: %v", err)
	}
	if !res.Accepted {
		t.Error("ModeAny should accept when STARK is valid")
	}
	if !res.STARKValid {
		t.Error("STARKValid should be true")
	}
}

func TestVerifyHybridModeBothRequiresBoth(t *testing.T) {
	snark := make([]byte, 128)
	badSTARK := []byte("too-short")
	res, _ := VerifyHybrid(VerifyRequest{
		Mode:         ModeBoth,
		SNARKProof:   snark,
		STARKProof:   badSTARK,
		STARKBackend: "simulated_fri",
	})
	if res.Accepted {
		t.Error("ModeBoth must reject when STARK is invalid")
	}
}

func TestVerifyHybridUnknownBackendReturnsError(t *testing.T) {
	_, err := VerifyHybrid(VerifyRequest{
		Mode:         ModeAny,
		SNARKProof:   make([]byte, 128),
		STARKProof:   []byte("proof"),
		STARKBackend: "nonexistent_backend_xyz",
	})
	if err == nil {
		t.Error("expected error for unknown STARK backend")
	}
	if !strings.Contains(err.Error(), "unknown stark backend") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBackendNames(t *testing.T) {
	fri := friVerifier{}
	if fri.BackendName() != "simulated_fri" {
		t.Errorf("friVerifier BackendName = %q, want simulated_fri", fri.BackendName())
	}
	wf := winterfellVerifier{}
	if wf.BackendName() != "winterfell_mock" {
		t.Errorf("winterfellVerifier BackendName = %q, want winterfell_mock", wf.BackendName())
	}
}

func TestGenFRIProofSize(t *testing.T) {
	content := []byte("data")
	proof := GenFRIProof(content)
	if len(proof) != 32+len(content) {
		t.Errorf("FRI proof length = %d, want %d", len(proof), 32+len(content))
	}
	// root must not be all zeros
	if bytes.Equal(proof[:32], make([]byte, 32)) {
		t.Error("FRI root is all zeros")
	}
}

func TestGenWinterfellProofSize(t *testing.T) {
	transcript := bytes.Repeat([]byte("x"), 64)
	proof := GenWinterfellProof(transcript)
	if len(proof) != 32+len(transcript) {
		t.Errorf("Winterfell proof length = %d, want %d", len(proof), 32+len(transcript))
	}
}
