package wasmhost

import (
	"context"
	"errors"
)

type Runner struct {
	wasmBinary []byte
}

func NewRunner(ctx context.Context, wasmBin []byte) (*Runner, error) {
	if len(wasmBin) < 8 {
		return nil, errors.New("invalid WASM binary")
	}
	return &Runner{wasmBinary: wasmBin}, nil
}

func (r *Runner) Verify(ctx context.Context, proof []byte) (bool, error) {
	// Mock verification for now
	return len(proof) == 200, nil
}

func (r *Runner) Close(ctx context.Context) error {
	return nil
}
