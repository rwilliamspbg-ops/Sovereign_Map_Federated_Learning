package modelcheckpoints

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/storage/ipfs"
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

// Store manages model checkpoint artifacts with multiple backends.
type Store struct {
	RootDir    string
	Backend    BackendType
	IPFSClient *ipfs.Client
}

// Put stores checkpoint bytes and returns metadata with content digest.
func (s Store) Put(ctx context.Context, version string, payload []byte) (ArtifactMeta, error) {
	if version == "" {
		return ArtifactMeta{}, fmt.Errorf("version is required")
	}
	if len(payload) == 0 {
		return ArtifactMeta{}, fmt.Errorf("payload is empty")
	}

	digestBytes := sha256.Sum256(payload)
	digest := hex.EncodeToString(digestBytes[:])

	switch s.Backend {
	case BackendIPFS:
		return s.putIPFS(ctx, version, digest, payload)
	case BackendFilesystem:
		return s.putFilesystem(version, digest, payload)
	default:
		return ArtifactMeta{}, fmt.Errorf("unsupported backend: %s", s.Backend)
	}
}

// putFilesystem stores checkpoint to local filesystem.
func (s Store) putFilesystem(version, digest string, payload []byte) (ArtifactMeta, error) {
	if s.RootDir == "" {
		s.RootDir = "storage/model_checkpoints/data"
	}

	if err := os.MkdirAll(s.RootDir, 0o750); err != nil {
		return ArtifactMeta{}, fmt.Errorf("create checkpoint root: %w", err)
	}

	file := filepath.Join(s.RootDir, version+".ckpt")
	if err := os.WriteFile(file, payload, 0o600); err != nil {
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

// putIPFS stores checkpoint to IPFS content-addressed storage.
func (s Store) putIPFS(ctx context.Context, version, digest string, payload []byte) (ArtifactMeta, error) {
	if s.IPFSClient == nil {
		return ArtifactMeta{}, fmt.Errorf("IPFS client not configured")
	}

	cid, err := s.IPFSClient.Add(ctx, payload)
	if err != nil {
		return ArtifactMeta{}, fmt.Errorf("ipfs add: %w", err)
	}

	// Pin to ensure persistence
	if err := s.IPFSClient.Pin(ctx, cid); err != nil {
		return ArtifactMeta{}, fmt.Errorf("ipfs pin: %w", err)
	}

	return ArtifactMeta{
		Version:   version,
		Digest:    digest,
		URI:       "ipfs://" + cid,
		Backend:   BackendIPFS,
		SizeBytes: int64(len(payload)),
		CreatedAt: time.Now().UTC(),
	}, nil
}

// Get retrieves checkpoint by version and backend URI.
func (s Store) Get(ctx context.Context, meta ArtifactMeta) ([]byte, error) {
	switch meta.Backend {
	case BackendIPFS:
		return s.getIPFS(ctx, meta.URI)
	case BackendFilesystem:
		return s.getFilesystem(meta.URI)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", meta.Backend)
	}
}

// getFilesystem retrieves checkpoint from local filesystem.
func (s Store) getFilesystem(uri string) ([]byte, error) {
	path := uri[7:] // Strip "file://" prefix
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}
	return data, nil
}

// getIPFS retrieves checkpoint from IPFS.
func (s Store) getIPFS(ctx context.Context, uri string) ([]byte, error) {
	if s.IPFSClient == nil {
		return nil, fmt.Errorf("IPFS client not configured")
	}

	cid := uri[7:] // Strip "ipfs://" prefix
	data, err := s.IPFSClient.Get(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("ipfs get: %w", err)
	}
	return data, nil
}
