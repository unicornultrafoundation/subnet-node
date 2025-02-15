package apps

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
)

// Verifier is a struct that provides methods to verify resource usage
type Verifier struct{}

// NewVerifier creates a new instance of Verifier
func NewVerifier() *Verifier {
	return &Verifier{}
}

// VerifyResourceUsage verifies the resource usage before signing
func (v *Verifier) VerifyResourceUsage(usage *ResourceUsage, stream network.Stream) error {
	if usage.PeerId != stream.Conn().RemotePeer().String() {
		return fmt.Errorf("peer ID mismatch: expected %s, got %s", stream.Conn().RemotePeer().String(), usage.PeerId)
	}
	return nil
}
