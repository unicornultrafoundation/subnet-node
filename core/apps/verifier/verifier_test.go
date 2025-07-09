package verifier

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	proto "github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	pvtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/app/verifier"
)

// ProviderNode represents a provider node for testing
type ProviderNode struct {
	host       host.Host
	verifierID peer.ID
	providerId int64
	appId      int64
	pow        *Pow
}

// NewProviderNode creates a new provider node for testing
func NewProviderNode(providerId, appId int64) (*ProviderNode, error) {
	h, err := libp2p.New()
	if err != nil {
		return nil, err
	}

	return &ProviderNode{
		host:       h,
		providerId: providerId,
		appId:      appId,
	}, nil
}

// Connect connects the provider to a verifier
func (p *ProviderNode) Connect(verifierHost host.Host) error {
	p.verifierID = verifierHost.ID()
	return p.host.Connect(context.Background(), peer.AddrInfo{
		ID:    verifierHost.ID(),
		Addrs: verifierHost.Addrs(),
	})
}

// SetupPoWHandler sets up the handler for PoW requests
func (p *ProviderNode) SetupPoWHandler(p2pInstance *p2p.P2P) {
	p.pow = NewPow(NodeProvider, p.host, p2pInstance)
	p.host.SetStreamHandler(atypes.ProtocolAppPoWRequest, p.pow.OnPoWRequest)
}

// SendUsageReport sends a usage report to the verifier
func (p *ProviderNode) SendUsageReport(cpu float64, memory, storage, uploadBytes, downloadBytes int64) error {
	// Create usage report
	usageReport := &pvtypes.UsageReport{
		AppId:         p.appId,
		ProviderId:    p.providerId,
		PeerId:        p.host.ID().String(),
		Cpu:           int64(cpu),
		Memory:        memory,
		Storage:       storage,
		UploadBytes:   uploadBytes,
		DownloadBytes: downloadBytes,
		Timestamp:     time.Now().Unix(),
	}

	// Open stream to verifier
	stream, err := p.host.NewStream(context.Background(), p.verifierID, atypes.ProtocolAppVerifierUsageReport)
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}
	defer stream.Close()

	// Marshal and send the usage report
	data, err := proto.Marshal(usageReport)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %v", err)
	}

	_, err = stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %v", err)
	}

	return nil
}

// TriggerPeriodicCheck triggers the periodic check process manually for testing
func (v *Verifier) TriggerPeriodicCheck() ([]*pvtypes.SignedUsage, error) {
	usagesByAppId, usageReportIds, uniquePeerIds, err := v.queryUsageReports()
	if err != nil {
		return nil, fmt.Errorf("failed to query usage reports: %v", err)
	}

	v.pow.RequestPoWFromPeers(uniquePeerIds, 10)
	time.Sleep(1 * time.Second) // Reduced sleep time for testing

	signedUsages, err := v.processUsageReports(usagesByAppId)
	if err != nil {
		return nil, fmt.Errorf("failed to process usage reports: %v", err)
	}

	v.pow.Clear()

	err = v.saveAndSendSignedUsages(signedUsages, usageReportIds)
	if err != nil {
		return nil, fmt.Errorf("failed to save and send signed usages: %v", err)
	}

	return signedUsages, nil
}

func newMockAccountServiceWithChainID(chainID int64) *account.AccountService {
	// Create a mock config
	cfg := config.NewC(nil)

	// Set the required config values
	cfg.Settings = map[string]any{
		"account": map[string]any{
			"private_key": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", // Test private key
			"chainid":     chainID,
			"rpc":         "http://localhost:8545", // Mock RPC endpoint
		},
		"apps": map[string]any{
			"subnet_app_store":   "0x2222222222222222222222222222222222222222",
			"subnet_provider":    "0x1111111111111111111111111111111111111111",
			"subnet_ip_registry": "0x3333333333333333333333333333333333333333",
		},
		"provider": map[string]any{
			"id": "0x1", // Provider ID in hex
		},
	}

	// Create the account service using the config
	mockAcc, err := account.NewAccountService(cfg)
	if err != nil {
		panic(err) // In test code, it's okay to panic on setup
	}

	return mockAcc
}

// For convenience, also provide the default version
func newMockAccountService() *account.AccountService {
	return newMockAccountServiceWithChainID(1) // Default to chain ID 1
}

func TestVerifierBasic(t *testing.T) {
	// Setup verifier
	verifierHost, err := libp2p.New()
	require.NoError(t, err)
	defer verifierHost.Close()

	ds := ds_sync.MutexWrap(datastore.NewMapDatastore())
	p2pInstance := &p2p.P2P{}
	mockAcc := newMockAccountService()
	verifier := NewVerifier(ds, verifierHost, p2pInstance, mockAcc)
	require.NotNil(t, verifier)
	err = verifier.Register()
	require.NoError(t, err)

	// Setup provider
	provider, err := NewProviderNode(1, 1) // Provider ID 1, App ID 1
	require.NoError(t, err)
	defer provider.host.Close()

	// Setup PoW handler and connect to verifier
	provider.SetupPoWHandler(p2pInstance)
	// Add verifier to provider's verifier peers list BEFORE connecting
	provider.pow.AddVerifierPeer(verifierHost.ID().String())

	err = provider.Connect(verifierHost)
	require.NoError(t, err)

	// Add provider to verifier's qualified peers
	verifier.pow.qualifiedPeers[provider.host.ID().String()] = true
	verifier.pow.AddVerifierPeer(provider.host.ID().String())

	// Send multiple usage reports to build reputation
	for i := 0; i < 3; i++ {
		err = provider.SendUsageReport(
			50.0, // CPU: 50%
			1024, // Memory: 1024 MB
			2048, // Storage: 2GB
			1000, // Upload: 1000 bytes
			2000, // Download: 2000 bytes
		)
		require.NoError(t, err)
		time.Sleep(30 * time.Second) // Wait between reports to meet time threshold
	}

	// Trigger periodic check manually
	signedUsages, err := verifier.TriggerPeriodicCheck()
	require.NoError(t, err)
	require.NotEmpty(t, signedUsages, "Should have at least one signed usage")

	// Verify the signed usage content
	signedUsage := signedUsages[0]
	require.Equal(t, provider.appId, signedUsage.AppId)
	require.Equal(t, provider.providerId, signedUsage.ProviderId)
	require.Equal(t, provider.host.ID().String(), signedUsage.PeerId)
	require.NotEmpty(t, signedUsage.Signature, "Should have a signature")
	require.NotEmpty(t, signedUsage.Hash, "Should have a hash")
	require.Equal(t, int32(12), signedUsage.Score, "Provider should have a reputation score of 12")

	// Query and verify the signed usage was stored
	storedSignedUsages, err := verifier.getSignedUsages(provider.appId, provider.host.ID().String(), 1)

	require.NoError(t, err)
	require.NotEmpty(t, storedSignedUsages, "Should find at least one signed usage")
	require.Equal(t, signedUsage.Hash, storedSignedUsages[0].Hash, "Stored signed usage should match")
	require.Equal(t, int32(12), storedSignedUsages[0].Score, "Stored signed usage should have a score of 12")
}

func TestTriggerPeriodicCheck(t *testing.T) {
	// Create mock account
	mockAcc := newMockAccountService()

	// Create verifier with mock account
	verifierHost, err := libp2p.New()
	require.NoError(t, err)
	defer verifierHost.Close()

	ds := ds_sync.MutexWrap(datastore.NewMapDatastore())
	p2pInstance := &p2p.P2P{}

	verifier := NewVerifier(ds, verifierHost, p2pInstance, mockAcc)
	require.NotNil(t, verifier)
	err = verifier.Register()
	require.NoError(t, err)

	// Setup test provider
	provider, err := NewProviderNode(1, 1)
	require.NoError(t, err)
	defer provider.host.Close()

	// When you need to verify chain ID related functionality
	chainID := mockAcc.GetChainID()
	require.Equal(t, big.NewInt(1), chainID)
}
