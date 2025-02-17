package verifier

import (
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/p2p"

	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
)

var log = logrus.WithField("service", "app-verifier")

// Verifier is a struct that provides methods to verify resource usage
type Verifier struct {
	ds  datastore.Datastore
	p2p *p2p.P2P
	acc *account.AccountService
}

// NewVerifier creates a new instance of Verifier
func NewVerifier(ds datastore.Datastore, P2P *p2p.P2P, acc *account.AccountService) *Verifier {
	return &Verifier{
		ds:  ds,
		p2p: P2P,
		acc: acc,
	}
}

// VerifyResourceUsage verifies the resource usage before signing
func (v *Verifier) verifyResourceUsage(usage *atypes.ResourceUsage, stream network.Stream) error {
	if usage.PeerId != stream.Conn().RemotePeer().String() {
		return fmt.Errorf("peer ID mismatch: expected %s, got %s", stream.Conn().RemotePeer().String(), usage.PeerId)
	}
	return nil
}

func (s *Verifier) Register() error {
	// Define the signing logic
	signHandler := func(stream network.Stream) ([]byte, error) {
		signRequest, err := atypes.ReceiveSignRequest(stream)

		if err != nil {
			return []byte{}, fmt.Errorf("failed to receive sign request: %w", err)
		}

		var signature []byte
		switch data := signRequest.Data.(type) {
		case *pbapp.SignatureRequest_Usage:
			usage := atypes.ProtoToResourceUsage(data.Usage)

			// Verify the resource usage before signing
			if err := s.verifyResourceUsage(usage, stream); err != nil {
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
	listener := p2p.NewSignProtocolListener(atypes.PROTOCOL_ID, signHandler)

	// Register the listener in the ListenersP2P
	err := s.p2p.ListenersP2P.Register(listener)
	if err != nil {
		return err
	}

	log.Debugf("Registered signing protocol: %s", atypes.PROTOCOL_ID)
	return nil
}

func (s *Verifier) signResourceUsage(usage *atypes.ResourceUsage) ([]byte, error) {
	filledUsage := atypes.FillDefaultResourceUsage(usage)
	typedData, err := atypes.ConvertUsageToTypedData(filledUsage, s.acc.GetChainID(), s.acc.AppStoreAddr())
	if err != nil {
		return []byte{}, fmt.Errorf("failed to get usage typed data: %v", err)
	}

	typedDataHash, _, err := atypes.TypedDataAndHash(*typedData)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to hash typed data: %v", err)
	}

	return s.acc.Sign(typedDataHash)
}
