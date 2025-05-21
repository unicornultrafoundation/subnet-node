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

	providerID := "provider1"
	score := 85

	// Store the score
	err := rs.StoreProviderScore(providerID, score)
	assert.NoError(t, err)

	// Retrieve the score
	retrievedScore, err := rs.getLocalScore(providerID)
	assert.NoError(t, err)
	assert.Equal(t, score, retrievedScore, "Retrieved score should match stored score")
}

func TestQueryReputationScoreWithNoPeers(t *testing.T) {
	rs, h, _ := setupReputationService(t)
	defer h.Close()

	providerID := "provider1"
	localScore := 90

	// Store a local score
	err := rs.StoreProviderScore(providerID, localScore)
	assert.NoError(t, err)

	// Query reputation score (no peers connected)
	aggregateScore, err := rs.QueryReputationScore(providerID)
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

	providerID := "provider1"
	score1 := 80
	score2 := 90

	// Store scores on both nodes
	err = rs1.StoreProviderScore(providerID, score1)
	assert.NoError(t, err)
	err = rs2.StoreProviderScore(providerID, score2)
	assert.NoError(t, err)

	// Query reputation score from rs1
	aggregateScore, err := rs1.QueryReputationScore(providerID)
	assert.NoError(t, err)

	expectedScore := (score1 + score2) / 2
	assert.Equal(t, expectedScore, aggregateScore, "Aggregate score should be the average of local and peer scores")
}

func TestReputationScoreRanges(t *testing.T) {
	// Create a reputation service
	rs, h, p2pInstance := setupReputationService(t)
	defer h.Close()

	// Create a verifier with the reputation service and the same datastore
	v := &Verifier{
		ds:  rs.ds,
		ps:  h,
		p2p: p2pInstance,
		rs:  rs,
	}

	// Define test providers and their scores
	providers := map[string]int{
		"provider1": 15, // Range: 0-20
		"provider2": 35, // Range: 21-40
		"provider3": 55, // Range: 41-60
		"provider4": 75, // Range: 61-80
		"provider5": 95, // Range: 81-100
		"provider6": 10, // Range: 0-20
		"provider7": 85, // Range: 81-100
	}

	// Store the scores using StoreProviderScore
	for providerID, score := range providers {
		err := rs.StoreProviderScore(providerID, score)
		require.NoError(t, err)
	}

	// Call GetReputationScoreRanges
	ctx := context.Background()
	ranges, err := v.GetReputationScoreRanges(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(ranges), "Should have 5 ranges")

	// Verify range definitions
	assert.Equal(t, 0, ranges[0].MinScore)
	assert.Equal(t, 20, ranges[0].MaxScore)
	assert.Equal(t, 21, ranges[1].MinScore)
	assert.Equal(t, 40, ranges[1].MaxScore)
	assert.Equal(t, 41, ranges[2].MinScore)
	assert.Equal(t, 60, ranges[2].MaxScore)
	assert.Equal(t, 61, ranges[3].MinScore)
	assert.Equal(t, 80, ranges[3].MaxScore)
	assert.Equal(t, 81, ranges[4].MinScore)
	assert.Equal(t, 100, ranges[4].MaxScore)

	// Verify counts
	assert.Equal(t, 2, ranges[0].Count, "Range 0-20 should have 2 providers")
	assert.Equal(t, 1, ranges[1].Count, "Range 21-40 should have 1 provider")
	assert.Equal(t, 1, ranges[2].Count, "Range 41-60 should have 1 provider")
	assert.Equal(t, 1, ranges[3].Count, "Range 61-80 should have 1 provider")
	assert.Equal(t, 2, ranges[4].Count, "Range 81-100 should have 2 providers")

	// Verify providers in each range
	assert.Contains(t, ranges[0].Providers, "provider1", "Range 0-20 should contain provider1")
	assert.Contains(t, ranges[0].Providers, "provider6", "Range 0-20 should contain provider6")
	assert.Contains(t, ranges[1].Providers, "provider2", "Range 21-40 should contain provider2")
	assert.Contains(t, ranges[2].Providers, "provider3", "Range 41-60 should contain provider3")
	assert.Contains(t, ranges[3].Providers, "provider4", "Range 61-80 should contain provider4")
	assert.Contains(t, ranges[4].Providers, "provider5", "Range 81-100 should contain provider5")
	assert.Contains(t, ranges[4].Providers, "provider7", "Range 81-100 should contain provider7")

	// Verify sorting within ranges (highest first)
	if len(ranges[0].Providers) >= 2 {
		// Provider1 (15) should come before Provider6 (10)
		assert.Equal(t, "provider1", ranges[0].Providers[0], "provider1 should be first in range 0-20")
		assert.Equal(t, "provider6", ranges[0].Providers[1], "provider6 should be second in range 0-20")
	}

	if len(ranges[4].Providers) >= 2 {
		// Provider5 (95) should come before Provider7 (85)
		assert.Equal(t, "provider5", ranges[4].Providers[0], "Provider5 should be first in range 81-100")
		assert.Equal(t, "provider7", ranges[4].Providers[1], "Provider7 should be second in range 81-100")
	}
}
