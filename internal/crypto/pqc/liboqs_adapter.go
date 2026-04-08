// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package pqc

// AdapterStatus reports whether liboqs-backed PQC primitives are active.
type AdapterStatus struct {
	Adapter          string `json:"adapter"`
	Available        bool   `json:"available"`
	KEMAlgorithm     string `json:"kem_algorithm"`
	HybridKEMEnabled bool   `json:"hybrid_kem_enabled"`
	BuildTagEnabled  bool   `json:"build_tag_enabled"`
	Reason           string `json:"reason,omitempty"`
}

// Status returns runtime availability for the Go liboqs adapter.
func Status() AdapterStatus {
	return AdapterStatus{
		Adapter:          "go-liboqs-hybrid-kem",
		Available:        false,
		KEMAlgorithm:     "ML-KEM-768",
		HybridKEMEnabled: false,
		BuildTagEnabled:  false,
		Reason:           "compiled without liboqs build tags (use -tags liboqs with cgo and liboqs installed)",
	}
}
