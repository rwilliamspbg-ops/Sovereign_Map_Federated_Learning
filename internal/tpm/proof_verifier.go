// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package tpm

import (
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
)

// TPMProofVerifier implements blockchain.ProofVerifier by delegating TEE/TPM proof
// types to an AttestationManager.  For "tee" and "tpm" proof types it generates a
// fresh attestation for the requesting node and immediately verifies it through
// VerifyAttestation.  All other proof types fall through to permissive pass-through
// semantics (trusted-by-consensus) so the verifier can be installed without
// disrupting non-TPM FL rounds.
type TPMProofVerifier struct {
	manager *AttestationManager
}

// NewTPMProofVerifier creates a ProofVerifier backed by the given AttestationManager.
// The manager must be non-nil; enabling the manager is controlled by its own
// enabled flag so callers can still install the verifier in environments without
// real TPM hardware.
func NewTPMProofVerifier(manager *AttestationManager) *TPMProofVerifier {
	if manager == nil {
		panic("tpm: NewTPMProofVerifier requires a non-nil AttestationManager")
	}
	return &TPMProofVerifier{manager: manager}
}

// VerifyProof implements blockchain.ProofVerifier.
//
// For "tee" / "tpm" proof types:
//   - The ProofHash field is used as the nonce for attestation generation (falls back
//     to RoundID if empty).
//   - Node ID is derived from the Payload["node_id"] or Payload["from"] key.
//   - Successful attestation yields confidence 9500 bps.
//   - Any attestation failure yields (false, 2500, nil) — a soft error that allows
//     the governance RejectOnVerificationFailure flag to decide whether to reject.
//
// For all other proof types (consensus, zk, unknown):
//   - Returns the same defaults as PermissiveProofVerifier so existing FL rounds are
//     unaffected.
func (v *TPMProofVerifier) VerifyProof(req blockchain.ProofVerificationRequest) (bool, uint32, error) {
	switch req.ProofType {
	case "tee", "tpm", "attestation":
		nodeID := payloadString(req.Payload, "node_id", "")
		if nodeID == "" {
			nodeID = payloadString(req.Payload, "from", "unknown")
		}

		nonce := []byte(req.ProofHash)
		if len(nonce) == 0 {
			nonce = []byte(req.RoundID)
		}

		report, err := v.manager.GenerateAttestation(nodeID, nonce)
		if err != nil {
			// Attestation generation failure is treated as a low-confidence soft fail
			// so the VerificationPolicy gate (not us) decides whether to reject the round.
			return false, 2500, nil
		}

		valid, err := v.manager.VerifyAttestation(report)
		if err != nil || !valid {
			return false, 2500, nil
		}

		return true, 9500, nil

	default:
		// consensus / zk / unknown: permissive pass-through matching PermissiveProofVerifier
		if req.ProofHash == "" && len(req.ProofData) == 0 {
			return true, 6500, nil
		}
		return true, 9000, nil
	}
}

// payloadString extracts a string value from a round payload, returning fallback
// if the key is missing or the value is not a string.
func payloadString(payload map[string]interface{}, key, fallback string) string {
	if payload == nil {
		return fallback
	}
	v, ok := payload[key]
	if !ok {
		return fallback
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return fallback
	}
	return s
}
