package apps

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
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

func (s *Service) RegisterSignProtocol() error {
	// Define the signing logic
	signHandler := func(stream network.Stream) ([]byte, error) {
		signRequest, err := receiveSignRequest(stream)

		if err != nil {
			return []byte{}, fmt.Errorf("failed to receive sign request: %w", err)
		}

		var signature []byte
		switch data := signRequest.Data.(type) {
		case *pbapp.SignatureRequest_Usage:
			usage := ProtoToResourceUsage(data.Usage)

			// Verify the resource usage before signing
			if err := s.verifier.VerifyResourceUsage(usage, stream); err != nil {
				return []byte{}, fmt.Errorf("failed to verify resource usage: %w", err)
			}

			signature, err = s.signResourceUsage(usage)

			if err != nil {
				return []byte{}, fmt.Errorf("failed to sign resource usage: %w", err)
			}
		default:
			return []byte{}, fmt.Errorf("unsupported signature request type: %w", err)
		}

		return signature, nil
	}

	// Create the listener for the signing protocol
	listener := p2p.NewSignProtocolListener(PROTOCOL_ID, signHandler)

	// Register the listener in the ListenersP2P
	err := s.P2P.ListenersP2P.Register(listener)
	if err != nil {
		return err
	}

	log.Debugf("Registered signing protocol: %s", PROTOCOL_ID)
	return nil
}
