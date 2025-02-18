package verifier

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
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

const reportInterval = 30 * time.Minute // Define the minimum interval between reports
const maxFailures = 5                   // Maximum allowed failures before marking as fraudulent

// Default average usage values for a basic app in bytes
var defaultAverageUsage = &atypes.ResourceUsage{
	UsedCpu:           big.NewInt(10000000),   // Default CPU usage
	UsedGpu:           big.NewInt(2147483648), // Default GPU usage
	UsedMemory:        big.NewInt(2147483648), // Default Memory usage in bytes
	UsedStorage:       big.NewInt(2147483648), // Default Storage usage in bytes
	UsedUploadBytes:   big.NewInt(2147483648), // Default Upload usage in bytes
	UsedDownloadBytes: big.NewInt(2147483648), // Default Download usage in bytes
}

// Verifier is a struct that provides methods to verify resource usage
type Verifier struct {
	ds              datastore.Datastore
	p2p             *p2p.P2P
	acc             *account.AccountService
	previousUsages  map[string]*atypes.ResourceUsage
	fraudulentNodes map[string]bool
	ipAddresses     map[string]struct{}
	failureCounts   map[string]int
	mu              sync.Mutex
}

// NewVerifier creates a new instance of Verifier
func NewVerifier(ds datastore.Datastore, P2P *p2p.P2P, acc *account.AccountService) *Verifier {
	v := &Verifier{
		ds:              ds,
		p2p:             P2P,
		acc:             acc,
		previousUsages:  make(map[string]*atypes.ResourceUsage),
		fraudulentNodes: make(map[string]bool),
		ipAddresses:     make(map[string]struct{}),
		failureCounts:   make(map[string]int),
	}
	return v
}

// VerifyResourceUsage verifies the resource usage before signing
func (v *Verifier) VerifyResourceUsage(ctx context.Context, usage *atypes.ResourceUsage, stream network.Stream) error {
	if v.isFraudulentNode(usage.PeerId) {
		return fmt.Errorf("fraudulent node detected: %s", usage.PeerId)
	}

	if err := v.verifyPeerID(usage, stream); err != nil {
		v.incrementFailureCount(usage.PeerId)
		return err
	}

	if err := v.verifyDuration(usage); err != nil {
		v.incrementFailureCount(usage.PeerId)
		return err
	}

	if err := v.verifyReportInterval(usage); err != nil {
		v.incrementFailureCount(usage.PeerId)
		return err
	}

	if err := v.verifySuddenHighUsage(usage); err != nil {
		v.incrementFailureCount(usage.PeerId)
		return err
	}

	if err := v.verifyRelayer(stream); err != nil {
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

	if usage.Duration.Int64() > int64(2*reportInterval.Seconds()) {
		return fmt.Errorf("invalid duration: duration %s exceeds twice the report interval %s", usage.Duration.String(), reportInterval.String())
	}

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

func (v *Verifier) verifySuddenHighUsage(usage *atypes.ResourceUsage) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	averageUsage, err := v.calculateAverageUsage(usage.AppId)
	if err != nil {
		return fmt.Errorf("failed to calculate average usage: %v", err)
	}

	// Handle cases where previous usage is zero
	if averageUsage.UsedCpu.Cmp(big.NewInt(0)) == 0 {
		averageUsage.UsedCpu = defaultAverageUsage.UsedCpu
	}
	if averageUsage.UsedGpu.Cmp(big.NewInt(0)) == 0 {
		averageUsage.UsedGpu = defaultAverageUsage.UsedGpu
	}
	if averageUsage.UsedMemory.Cmp(big.NewInt(0)) == 0 {
		averageUsage.UsedMemory = defaultAverageUsage.UsedMemory
	}
	if averageUsage.UsedStorage.Cmp(big.NewInt(0)) == 0 {
		averageUsage.UsedStorage = defaultAverageUsage.UsedStorage
	}
	if averageUsage.UsedUploadBytes.Cmp(big.NewInt(0)) == 0 {
		averageUsage.UsedUploadBytes = defaultAverageUsage.UsedUploadBytes
	}
	if averageUsage.UsedDownloadBytes.Cmp(big.NewInt(0)) == 0 {
		averageUsage.UsedDownloadBytes = defaultAverageUsage.UsedDownloadBytes
	}

	if usage.UsedCpu.Cmp(big.NewInt(0).Mul(averageUsage.UsedCpu, big.NewInt(2))) > 0 ||
		usage.UsedGpu.Cmp(big.NewInt(0).Mul(averageUsage.UsedGpu, big.NewInt(2))) > 0 ||
		usage.UsedMemory.Cmp(big.NewInt(0).Mul(averageUsage.UsedMemory, big.NewInt(2))) > 0 ||
		usage.UsedStorage.Cmp(big.NewInt(0).Mul(averageUsage.UsedStorage, big.NewInt(2))) > 0 ||
		usage.UsedUploadBytes.Cmp(big.NewInt(0).Mul(averageUsage.UsedUploadBytes, big.NewInt(2))) > 0 ||
		usage.UsedDownloadBytes.Cmp(big.NewInt(0).Mul(averageUsage.UsedDownloadBytes, big.NewInt(2))) > 0 {
		return fmt.Errorf("sudden high resource usage detected for peer ID: %s", usage.PeerId)
	}

	return nil
}

func (v *Verifier) verifyRelayer(stream network.Stream) error {
	// Check if the peer is coming through a relayer
	if stream.Conn().Stat().Direction == network.DirInbound && stream.Conn().RemoteMultiaddr().String() != "" {
		return fmt.Errorf("peer is coming through a relayer: %s", stream.Conn().RemotePeer().String())
	}
	return nil
}

func (v *Verifier) calculateAverageUsage(appId *big.Int) (*atypes.ResourceUsage, error) {
	var totalCpu, totalGpu, totalMemory, totalStorage, totalUpload, totalDownload int64
	var count int64

	// Get a list of all peer IDs for the given app ID
	peerIDs := make([]string, 0)
	for peerID, usage := range v.previousUsages {
		if usage.AppId.Cmp(appId) == 0 {
			peerIDs = append(peerIDs, peerID)
		}
	}

	// Shuffle the list of peer IDs
	rand.Shuffle(len(peerIDs), func(i, j int) {
		peerIDs[i], peerIDs[j] = peerIDs[j], peerIDs[i]
	})

	// Select a random subset of peer IDs
	numPeers := len(peerIDs)
	if numPeers > 30 {
		numPeers = 30
	}
	selectedPeers := peerIDs[:numPeers]

	// Calculate the average usage for the selected peers
	for _, peerID := range selectedPeers {
		usage := v.previousUsages[peerID]
		totalCpu += usage.UsedCpu.Int64()
		totalGpu += usage.UsedGpu.Int64()
		totalMemory += usage.UsedMemory.Int64()
		totalStorage += usage.UsedStorage.Int64()
		totalUpload += usage.UsedUploadBytes.Int64()
		totalDownload += usage.UsedDownloadBytes.Int64()
		count++
	}

	if count == 0 {
		// Return default average usage if no peers are available
		return defaultAverageUsage, nil
	}

	averageUsage := &atypes.ResourceUsage{
		UsedCpu:           big.NewInt(totalCpu / count),
		UsedGpu:           big.NewInt(totalGpu / count),
		UsedMemory:        big.NewInt(totalMemory / count),
		UsedStorage:       big.NewInt(totalStorage / count),
		UsedUploadBytes:   big.NewInt(totalUpload / count),
		UsedDownloadBytes: big.NewInt(totalDownload / count),
	}

	return averageUsage, nil
}

func (v *Verifier) updatePreviousUsage(usage *atypes.ResourceUsage) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.previousUsages[usage.PeerId] = usage
}

func (v *Verifier) incrementFailureCount(peerId string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.failureCounts[peerId]++
	if v.failureCounts[peerId] >= maxFailures {
		v.markAsFraudulent(peerId)
	}
}

func (v *Verifier) resetFailureCount(peerId string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.failureCounts[peerId] = 0
}

func (v *Verifier) markAsFraudulent(peerId string) {
	v.fraudulentNodes[peerId] = true
}

func (v *Verifier) isFraudulentNode(peerId string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.fraudulentNodes[peerId]
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
			s.resetFailureCount(usage.PeerId)

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
