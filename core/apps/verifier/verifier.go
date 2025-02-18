package verifier

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
)

var log = logrus.WithField("service", "app-verifier")

const reportInterval = 1 * time.Minute // Define the minimum interval between reports

// Verifier is a struct that provides methods to verify resource usage
type Verifier struct {
	ds             datastore.Datastore
	p2p            *p2p.P2P
	acc            *account.AccountService
	previousUsages map[string]*atypes.ResourceUsage
	mu             sync.Mutex
}

// NewVerifier creates a new instance of Verifier
func NewVerifier(ds datastore.Datastore, P2P *p2p.P2P, acc *account.AccountService) *Verifier {
	v := &Verifier{
		ds:             ds,
		p2p:            P2P,
		acc:            acc,
		previousUsages: make(map[string]*atypes.ResourceUsage),
	}
	return v
}

// VerifyResourceUsage verifies the resource usage before signing
func (v *Verifier) VerifyResourceUsage(ctx context.Context, usage *atypes.ResourceUsage, stream network.Stream) error {
	if err := v.verifyPeerID(usage, stream); err != nil {
		return err
	}

	if err := v.verifyDuration(usage); err != nil {
		return err
	}

	if err := v.verifyReportInterval(usage); err != nil {
		return err
	}

	return nil
}

func (v *Verifier) verifyPeerID(usage *atypes.ResourceUsage, stream network.Stream) error {
	if usage.PeerId != stream.Conn().RemotePeer().String() {
		return fmt.Errorf("peer ID mismatch: expected %s, got %s", stream.Conn().RemotePeer().String(), usage.PeerId)
	}
	return nil
}

func (v *Verifier) verifyDuration(usage *atypes.ResourceUsage) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	previousUsage, exists := v.previousUsages[usage.PeerId]
	if exists {
		currentTime := time.Unix(usage.Timestamp.Int64(), 0)
		prevTime := time.Unix(previousUsage.Timestamp.Int64(), 0)
		if int64(currentTime.Sub(prevTime).Seconds()) < usage.Duration.Int64() {
			return fmt.Errorf("invalid duration: current duration %s is not greater than previous duration %s", usage.Duration.String(), previousUsage.Duration.String())
		}
	}
	return nil
}

func (v *Verifier) verifyReportInterval(usage *atypes.ResourceUsage) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	previousUsage, exists := v.previousUsages[usage.PeerId]
	if exists {
		currentTime := time.Unix(usage.Timestamp.Int64(), 0)
		prevTime := time.Unix(previousUsage.Timestamp.Int64(), 0)
		if currentTime.Sub(prevTime) < reportInterval {
			return fmt.Errorf("reports must be spaced by at least %s", reportInterval.String())
		}
	}
	return nil
}

func (v *Verifier) updatePreviousUsage(usage *atypes.ResourceUsage) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.previousUsages[usage.PeerId] = usage
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
			if err := s.VerifyResourceUsage(context.Background(), usage, stream); err != nil {
				return []byte{}, fmt.Errorf("failed to verify resource usage: %w", err)
			}

			signature, err = s.signResourceUsage(usage)

			if err != nil {
				return []byte{}, fmt.Errorf("failed to sign resource usage: %w", err)
			}
			s.updatePreviousUsage(usage)
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
	typedData, err := atypes.ConvertUsageToTypedData(usage, s.acc.GetChainID(), s.acc.AppStoreAddr())
	if err != nil {
		return []byte{}, fmt.Errorf("failed to get usage typed data: %v", err)
	}
	return s.acc.SignTypedData(typedData)
}
