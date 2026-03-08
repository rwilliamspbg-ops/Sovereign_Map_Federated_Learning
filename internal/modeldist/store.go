package modeldist

import "fmt"

// CheckpointRef identifies a distributed model checkpoint.
type CheckpointRef struct {
	Version    string
	Digest     string
	StorageURI string
}

// Validate checks minimal invariants before a checkpoint is shared with peers.
func (c CheckpointRef) Validate() error {
	if c.Version == "" || c.Digest == "" || c.StorageURI == "" {
		return fmt.Errorf("checkpoint reference is incomplete")
	}
	return nil
}
