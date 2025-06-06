package types

// IPFSClient defines the interface for IPFS operations
type IPFSClient interface {
	// Get retrieves content from IPFS by hash
	Get(hash string) ([]byte, error)

	// Add adds content to IPFS and returns the hash
	Add(data []byte) (string, error)
}
