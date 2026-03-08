package modelcheckpoints

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackendType identifies artifact backend implementation.
type BackendType string

const (
	BackendFilesystem BackendType = "filesystem"
	BackendIPFS       BackendType = "ipfs"
	BackendObject     BackendType = "object-storage"
)

// ArtifactMeta tracks one stored model checkpoint artifact.
type ArtifactMeta struct {
	Version   string
	Digest    string
	URI       string
	Backend   BackendType
	SizeBytes int64
	CreatedAt time.Time
}

// Store manages model checkpoint artifacts.
type Store struct {
	RootDir string
}

// Put stores checkpoint bytes and returns metadata with content digest.
func (s Store) Put(version string, payload []byte) (ArtifactMeta, error) {
	if version == "" {
		return ArtifactMeta{}, fmt.Errorf("version is required")
	}
	if len(payload) == 0 {
		return ArtifactMeta{}, fmt.Errorf("payload is empty")
	}
	if s.RootDir == "" {
		s.RootDir = "storage/model_checkpoints/data"
	}

	if err := os.MkdirAll(s.RootDir, 0o755); err != nil {
		return ArtifactMeta{}, fmt.Errorf("create checkpoint root: %w", err)
	}

	digestBytes := sha256.Sum256(payload)
	digest := hex.EncodeToString(digestBytes[:])
	file := filepath.Join(s.RootDir, version+".ckpt")
	if err := os.WriteFile(file, payload, 0o644); err != nil {
		return ArtifactMeta{}, fmt.Errorf("write checkpoint file: %w", err)
	}

	return ArtifactMeta{
		Version:   version,
		Digest:    digest,
		URI:       "file://" + file,
		Backend:   BackendFilesystem,
		SizeBytes: int64(len(payload)),
		CreatedAt: time.Now().UTC(),
	}, nil
}
