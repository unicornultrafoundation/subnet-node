package ipfs

import (
	"bytes"
	"fmt"
	"io"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// Client implements the IPFSClient interface using go-ipfs-api
type Client struct {
	shell *shell.Shell
}

// NewClient creates a new IPFS client
func NewClient(ipfsURL string) types.IPFSClient {
	return &Client{
		shell: shell.NewShell(ipfsURL),
	}
}

// Get retrieves content from IPFS by hash
func (c *Client) Get(hash string) ([]byte, error) {
	// Get the content from IPFS
	reader, err := c.shell.Cat(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get content from IPFS: %w", err)
	}
	defer reader.Close()

	// Read all content
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read IPFS content: %w", err)
	}

	return data, nil
}

// Add adds content to IPFS and returns the hash
func (c *Client) Add(data []byte) (string, error) {
	// Add the content to IPFS
	hash, err := c.shell.Add(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to add content to IPFS: %w", err)
	}

	return hash, nil
}

// Verify that Client implements IPFSClient interface
var _ types.IPFSClient = (*Client)(nil)
