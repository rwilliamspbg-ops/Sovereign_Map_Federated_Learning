//go:build liboqs && cgo

// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package pqc

/*
#cgo pkg-config: liboqs
#include <oqs/oqs.h>
*/
import "C"

// Status returns runtime availability when compiled with liboqs support.
func Status() AdapterStatus {
	return AdapterStatus{
		Adapter:          "go-liboqs-hybrid-kem",
		Available:        true,
		KEMAlgorithm:     "ML-KEM-768",
		HybridKEMEnabled: true,
		BuildTagEnabled:  true,
	}
}
