package ipfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

// ClientConfig stores endpoint details for content-addressed storage.
type ClientConfig struct {
	APIEndpoint string
	Timeout     time.Duration
}

// Client wraps IPFS API for model checkpoint storage.
type Client struct {
	shell   *shell.Shell
	timeout time.Duration
}

// NewClient creates an IPFS client connected to the specified API endpoint.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = "localhost:5001"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	sh := shell.NewShell(cfg.APIEndpoint)
	sh.SetTimeout(cfg.Timeout)

	return &Client{
		shell:   sh,
		timeout: cfg.Timeout,
	}, nil
}

// Add uploads checkpoint bytes and returns IPFS CID.
func (c *Client) Add(ctx context.Context, data []byte) (string, error) {
	reader := bytes.NewReader(data)
	cid, err := c.shell.Add(reader)
	if err != nil {
		return "", fmt.Errorf("ipfs add: %w", err)
	}
	return cid, nil
}

// Get retrieves checkpoint bytes by CID.
func (c *Client) Get(ctx context.Context, cid string) ([]byte, error) {
	reader, err := c.shell.Cat(cid)
	if err != nil {
		return nil, fmt.Errorf("ipfs cat %s: %w", cid, err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read ipfs data: %w", err)
	}
	return data, nil
}

// Pin ensures checkpoint CID is permanently retained.
func (c *Client) Pin(ctx context.Context, cid string) error {
	if err := c.shell.Pin(cid); err != nil {
		return fmt.Errorf("ipfs pin %s: %w", cid, err)
	}
	return nil
}

// Unpin removes checkpoint from permanent storage.
func (c *Client) Unpin(ctx context.Context, cid string) error {
	if err := c.shell.Unpin(cid); err != nil {
		return fmt.Errorf("ipfs unpin %s: %w", cid, err)
	}
	return nil
}

// Stat returns size and block info for checkpoint CID.
func (c *Client) Stat(ctx context.Context, cid string) (*shell.ObjectStats, error) {
	stat, err := c.shell.ObjectStat(cid)
	if err != nil {
		return nil, fmt.Errorf("ipfs stat %s: %w", cid, err)
	}
	return stat, nil
}

// IsOnline checks IPFS daemon connectivity.
func (c *Client) IsOnline(ctx context.Context) bool {
	// Use ID command to check connectivity
	_, err := c.shell.ID()
	return err == nil
}
