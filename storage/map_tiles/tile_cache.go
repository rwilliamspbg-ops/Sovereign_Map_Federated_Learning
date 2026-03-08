package maptiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

// TileCacheConfig controls tile storage and memory cache.
type TileCacheConfig struct {
	RootDir          string
	MemoryCacheSize  int
	MaxTileSizeBytes int64
}

// TileCacheStats tracks cache occupancy for map tiles.
type TileCacheStats struct {
	Entries    int
	MemoryHits int64
	DiskHits   int64
	Misses     int64
	TotalBytes int64
}

// TileCache manages tile storage with LRU memory cache + disk backend.
type TileCache struct {
	mu          sync.RWMutex
	config      TileCacheConfig
	memoryCache *lru.Cache[string, []byte]
	stats       TileCacheStats
}

// NewTileCache creates tile cache with specified configuration.
func NewTileCache(cfg TileCacheConfig) (*TileCache, error) {
	if cfg.RootDir == "" {
		cfg.RootDir = "storage/map_tiles/data"
	}
	if cfg.MemoryCacheSize == 0 {
		cfg.MemoryCacheSize = 1000
	}
	if cfg.MaxTileSizeBytes == 0 {
		cfg.MaxTileSizeBytes = 500 * 1024 // 500KB
	}

	if err := os.MkdirAll(cfg.RootDir, 0o755); err != nil {
		return nil, fmt.Errorf("create tile cache root: %w", err)
	}

	memCache, err := lru.New[string, []byte](cfg.MemoryCacheSize)
	if err != nil {
		return nil, fmt.Errorf("create LRU cache: %w", err)
	}

	return &TileCache{
		config:      cfg,
		memoryCache: memCache,
		stats:       TileCacheStats{},
	}, nil
}

// Get retrieves tile from cache (memory then disk).
func (c *TileCache) Get(ctx context.Context, tile TileCoordinate, format TileFormat) ([]byte, error) {
	key := TilePath(tile, format)

	// Check memory cache first
	c.mu.RLock()
	if data, ok := c.memoryCache.Get(key); ok {
		c.stats.MemoryHits++
		c.mu.RUnlock()
		return data, nil
	}
	c.mu.RUnlock()

	// Check disk
	path := filepath.Join(c.config.RootDir, key)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.mu.Lock()
			c.stats.Misses++
			c.mu.Unlock()
			return nil, fmt.Errorf("tile not found: %w", err)
		}
		return nil, fmt.Errorf("read tile: %w", err)
	}

	// Populate memory cache
	c.mu.Lock()
	c.memoryCache.Add(key, data)
	c.stats.DiskHits++
	c.mu.Unlock()

	return data, nil
}

// Put stores tile in cache (memory + disk).
func (c *TileCache) Put(ctx context.Context, tile TileCoordinate, format TileFormat, data []byte) error {
	if int64(len(data)) > c.config.MaxTileSizeBytes {
		return fmt.Errorf("tile size %d exceeds limit %d", len(data), c.config.MaxTileSizeBytes)
	}

	key := TilePath(tile, format)
	path := filepath.Join(c.config.RootDir, key)

	// Create directory structure
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create tile directory: %w", err)
	}

	// Write to disk
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write tile: %w", err)
	}

	// Store in memory cache
	c.mu.Lock()
	c.memoryCache.Add(key, data)
	c.stats.Entries++
	c.stats.TotalBytes += int64(len(data))
	c.mu.Unlock()

	return nil
}

// Delete removes tile from cache (memory + disk).
func (c *TileCache) Delete(ctx context.Context, tile TileCoordinate, format TileFormat) error {
	key := TilePath(tile, format)
	path := filepath.Join(c.config.RootDir, key)

	c.mu.Lock()
	c.memoryCache.Remove(key)
	c.mu.Unlock()

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete tile: %w", err)
	}

	return nil
}

// Stats returns current cache statistics.
func (c *TileCache) Stats() TileCacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// Clear empties memory cache (disk tiles remain).
func (c *TileCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.memoryCache.Purge()
}

// Prune removes tiles older than specified age from disk.
func (c *TileCache) Prune(ctx context.Context, maxAge time.Duration) (int, error) {
	pruned := 0
	cutoff := time.Now().Add(-maxAge)

	err := filepath.Walk(c.config.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				return err
			}
			pruned++
		}
		return nil
	})

	return pruned, err
}
