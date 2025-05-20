package verifier

import (
	"context"
	"testing"

	"github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/p2p"
)

func init() {
	// Set log level to error to reduce noise during tests
	logrus.SetLevel(logrus.ErrorLevel)
}

func setupReputationService(t *testing.T) (*ReputationService, host.Host, *p2p.P2P) {
	// Create a new in-memory datastore
	ds := ds_sync.MutexWrap(datastore.NewMapDatastore())

	// Create a new libp2p host
	h, err := libp2p.New()
	require.NoError(t, err)

	// Create a new P2P instance
	p2pInstance := &p2p.P2P{}

	// Create a new ReputationService
	rs := NewReputationService(ds, h, p2pInstance)
	err = rs.Register()
	require.NoError(t, err)

	return rs, h, p2pInstance
}

func TestStoreAndRetrieveLocalScore(t *testing.T) {
	rs, h, _ := setupReputationService(t)
	defer h.Close()

	appID := int64(1)
	providerID := "provider1"
	score := 85

	// Store the score
	err := rs.StoreProviderScore(appID, providerID, score)
	assert.NoError(t, err)

	// Retrieve the score
	retrievedScore, err := rs.getLocalScore(appID, providerID)
	assert.NoError(t, err)
	assert.Equal(t, score, retrievedScore, "Retrieved score should match stored score")
}

func TestQueryReputationScoreWithNoPeers(t *testing.T) {
	rs, h, _ := setupReputationService(t)
	defer h.Close()

	appID := int64(1)
	providerID := "provider1"
	localScore := 90

	// Store a local score
	err := rs.StoreProviderScore(appID, providerID, localScore)
	assert.NoError(t, err)

	// Query reputation score (no peers connected)
	aggregateScore, err := rs.QueryReputationScore(appID, providerID, "")
	assert.NoError(t, err)
	assert.Equal(t, localScore, aggregateScore, "Should return local score when no peers are available")
}

func TestQueryReputationScoreWithPeers(t *testing.T) {
	// Setup two verifier nodes
	rs1, h1, _ := setupReputationService(t)
	rs2, h2, _ := setupReputationService(t)
	defer h1.Close()
	defer h2.Close()

	// Connect the two hosts
	err := h1.Connect(context.Background(), peer.AddrInfo{ID: h2.ID(), Addrs: h2.Addrs()})
	require.NoError(t, err)

	appID := int64(1)
	providerID := "provider1"
	score1 := 80
	score2 := 90

	// Store scores on both nodes
	err = rs1.StoreProviderScore(appID, providerID, score1)
	assert.NoError(t, err)
	err = rs2.StoreProviderScore(appID, providerID, score2)
	assert.NoError(t, err)

	// Query reputation score from rs1
	aggregateScore, err := rs1.QueryReputationScore(appID, providerID, "")
	assert.NoError(t, err)

	expectedScore := 80
	assert.Equal(t, expectedScore, aggregateScore, "Aggregate score should be the average of local and peer scores")
}

func TestQueryReputationScoreSpecificVerifier(t *testing.T) {
	// Setup two verifier nodes
	rs1, h1, _ := setupReputationService(t)
	rs2, h2, _ := setupReputationService(t)
	defer h1.Close()
	defer h2.Close()

	// Connect the two hosts
	err := h1.Connect(context.Background(), peer.AddrInfo{ID: h2.ID(), Addrs: h2.Addrs()})
	require.NoError(t, err)

	appID := int64(1)
	providerID := "provider1"
	score2 := 90

	// Store score on rs2
	err = rs2.StoreProviderScore(appID, providerID, score2)
	assert.NoError(t, err)

	// Query rs2's score from rs1
	queryScore, err := rs1.QueryReputationScore(appID, providerID, h2.ID().String())
	assert.NoError(t, err)

	// Expected score: average of rs1's score (0, since no local score) and rs2's score (90) = 90/2 = 45
	expectedScore := score2
	assert.Equal(t, expectedScore, queryScore, "Aggregate score should be average of local (0) and specific verifier's score")
}
