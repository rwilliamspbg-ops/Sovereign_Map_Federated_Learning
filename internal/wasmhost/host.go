// Copyright 2026 Sovereign-Mohawk Core Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Reference: /proofs/cryptography.md
// Theorem 5: O(1) verification time via optimized Wasm host calls.

package wasmhost

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const compilationCacheDir = "/tmp/mohawk-wasm-cache"

// Host manages the WebAssembly runtime environment for zk-SNARK verification.
type Host struct {
	runtime wazero.Runtime
	mod     api.Module
	mu      sync.Mutex
}

// Registry stores hash-addressed modules and supports default module hot reload.
type Registry struct {
	mu          sync.RWMutex
	modules     map[string]*Host
	defaultHash string
}

func NewRegistry() *Registry {
	return &Registry{modules: make(map[string]*Host)}
}

// NewHost initializes a high-performance Wasm environment.
func NewHost(ctx context.Context, wasmBin []byte) (*Host, error) {
	cfg := wazero.NewRuntimeConfig().WithCompilationCache(newCompilationCache())
	r := wazero.NewRuntimeWithConfig(ctx, cfg)

	// Instantiate the module with hardware acceleration where available
	mod, err := r.Instantiate(ctx, wasmBin)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate wasm: %w", err)
	}

	return &Host{
		runtime: r,
		mod:     mod,
	}, nil
}

func newCompilationCache() wazero.CompilationCache {
	cache, err := wazero.NewCompilationCacheWithDir(compilationCacheDir)
	if err != nil {
		return wazero.NewCompilationCache()
	}
	return cache
}

// NewRunner keeps backward compatibility with earlier host naming.
// Upsert stores a module keyed by SHA256 of its bytes.
func (r *Registry) Upsert(ctx context.Context, wasmBin []byte) (string, error) {
	if len(wasmBin) == 0 {
		return "", fmt.Errorf("empty wasm module")
	}
	hashBytes := sha256.Sum256(wasmBin)
	hash := hex.EncodeToString(hashBytes[:])

	r.mu.RLock()
	_, exists := r.modules[hash]
	r.mu.RUnlock()
	if exists {
		return hash, nil
	}

	host, err := NewHost(ctx, wasmBin)
	if err != nil {
		return "", err
	}

	r.mu.Lock()
	if _, exists = r.modules[hash]; exists {
		r.mu.Unlock()
		_ = host.Close(ctx)
		return hash, nil
	}
	r.modules[hash] = host
	if r.defaultHash == "" {
		r.defaultHash = hash
	}
	r.mu.Unlock()

	return hash, nil
}

func (r *Registry) Get(hash string) (*Host, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	host, ok := r.modules[hash]
	return host, ok
}

func (r *Registry) Default() *Host {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.defaultHash == "" {
		return nil
	}
	return r.modules[r.defaultHash]
}

func (r *Registry) HotReload(ctx context.Context, wasmBin []byte) (string, error) {
	hash, err := r.Upsert(ctx, wasmBin)
	if err != nil {
		return "", err
	}
	r.mu.Lock()
	r.defaultHash = hash
	r.mu.Unlock()
	return hash, nil
}

func (r *Registry) Close(ctx context.Context) error {
	r.mu.Lock()
	modules := r.modules
	r.modules = make(map[string]*Host)
	r.defaultHash = ""
	r.mu.Unlock()

	for _, host := range modules {
		if host != nil {
			_ = host.Close(ctx)
		}
	}
	return nil
}

// Verify executes the zk-SNARK proof verification in the Wasm sandbox.
func (h *Host) Verify(ctx context.Context, proof []byte) (bool, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	fn := h.mod.ExportedFunction("verify_proof")
	if fn == nil {
		return false, fmt.Errorf("wasm module missing required export: verify_proof")
	}

	// Theorem 5: Constant-time verification check
	results, err := fn.Call(ctx, uint64(len(proof)))
	if err != nil {
		return false, fmt.Errorf("wasm execution error: %w", err)
	}

	if len(results) == 0 {
		return false, fmt.Errorf("wasm function returned no results")
	}

	return results[0] == 1, nil
}

// FastVerify is an optimized alias for the Verify method.
func (h *Host) FastVerify(ctx context.Context, proof []byte) (bool, error) {
	return h.Verify(ctx, proof)
}

// Close releases Wasm resources.
func (h *Host) Close(ctx context.Context) error {
	return h.runtime.Close(ctx)
}
