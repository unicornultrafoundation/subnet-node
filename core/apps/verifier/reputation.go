package verifier

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/p2p"
)

var reputationLog = logrus.WithField("service", "reputation-service")

const (
	ReputationCacheExpiry = 30 * time.Minute
	MaxResponses          = 10
	ResponseTimeout       = 10 * time.Second
)

type ReputationQuery struct {
	AppID      int64  `json:"app_id"`
	ProviderID string `json:"provider_id"`
	Timestamp  int64  `json:"timestamp"`
	RequestID  string `json:"request_id"`
	VerifierID string `json:"verifier_id"`
}

type ReputationResponse struct {
	AppID      int64  `json:"app_id"`
	ProviderID string `json:"provider_id"`
	Score      int    `json:"score"`
	Timestamp  int64  `json:"timestamp"`
	VerifierID string `json:"verifier_id"`
	RequestID  string `json:"request_id"`
}

type ReputationService struct {
	ds             datastore.Datastore
	p2p            *p2p.P2P
	host           host.Host
	scoreCache     *lru.Cache
	pendingQueries sync.Map
}

type ReputationQueryResponse struct {
	Responses []ReputationResponse
	Error     error
	Done      chan struct{}
}

// NewReputationService creates a new instance of ReputationService
func NewReputationService(ds datastore.Datastore, host host.Host, p2p *p2p.P2P) *ReputationService {
	cache, _ := lru.New(1024)
	rs := &ReputationService{
		ds:         ds,
		p2p:        p2p,
		host:       host,
		scoreCache: cache,
	}

	return rs
}

// Register registers protocol handlers for reputation service
func (rs *ReputationService) Register() error {
	rs.host.SetStreamHandler(atypes.ProtocolAppVerifierProviderScoreRequest, rs.handleReputationQuery)
	rs.host.SetStreamHandler(atypes.ProtocolAppVerifierProviderScoreResponse, rs.handleReputationResponse)

	reputationLog.Info("Reputation service registered protocol handlers")
	return nil
}

// QueryReputationScore queries reputation scores for a provider from other verifiers
func (rs *ReputationService) QueryReputationScore(appID int64, providerID string, verifierID string) (int, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%d-%s", appID, providerID)
	if cachedScore, found := rs.scoreCache.Get(cacheKey); found {
		reputationLog.Debugf("Found cached reputation score for app %d provider %s: %d",
			appID, providerID, cachedScore.(int))
		return cachedScore.(int), nil
	}

	// Get local score first
	localScore, err := rs.getLocalScore(appID, providerID)
	if err != nil {
		reputationLog.Warnf("Failed to get local reputation score: %v", err)
		// Continue anyway, we'll query peers
	}

	// Query peers for their scores
	peerScores, err := rs.queryReputationScores(appID, providerID, verifierID)
	if err != nil {
		return localScore, fmt.Errorf("failed to query peers: %v", err)
	}

	// Aggregate scores
	aggregateScore := rs.aggregateScores(localScore, peerScores)

	// Cache the result
	rs.scoreCache.Add(cacheKey, aggregateScore)

	return aggregateScore, nil
}

// getLocalScore retrieves the local reputation score for a provider
func (rs *ReputationService) getLocalScore(appID int64, providerID string) (int, error) {
	key := datastore.NewKey(fmt.Sprintf("/provider_scores/%d/%s", appID, providerID))

	scoreBytes, err := rs.ds.Get(context.Background(), key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return 0, nil // No score yet, return 0
		}
		return 0, fmt.Errorf("failed to get score from datastore: %v", err)
	}

	var score int
	err = json.Unmarshal(scoreBytes, &score)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal score: %v", err)
	}

	return score, nil
}

// StoreProviderScore stores a provider's reputation score
func (rs *ReputationService) StoreProviderScore(appID int64, providerID string, score int) error {
	key := datastore.NewKey(fmt.Sprintf("/provider_scores/%d/%s", appID, providerID))

	scoreBytes, err := json.Marshal(score)
	if err != nil {
		return fmt.Errorf("failed to marshal score: %v", err)
	}

	err = rs.ds.Put(context.Background(), key, scoreBytes)
	if err != nil {
		return fmt.Errorf("failed to store score in datastore: %v", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("%d-%s", appID, providerID)
	rs.scoreCache.Add(cacheKey, score)

	return nil
}

// queryReputationScores queries other verifier nodes for reputation scores
func (rs *ReputationService) queryReputationScores(appID int64, providerID string, verifierID string) ([]ReputationResponse, error) {
	// Generate a unique request ID
	requestID := fmt.Sprintf("%d-%s-%d", appID, providerID, time.Now().UnixNano())

	// Create a query
	query := ReputationQuery{
		AppID:      appID,
		ProviderID: providerID,
		Timestamp:  time.Now().Unix(),
		RequestID:  requestID,
		VerifierID: verifierID,
	}

	// Create a result to track responses
	result := &ReputationQueryResponse{
		Responses: make([]ReputationResponse, 0),
		Done:      make(chan struct{}),
	}

	// Store the pending query
	rs.pendingQueries.Store(requestID, result)

	// Set a timeout for the query
	ctx, cancel := context.WithTimeout(context.Background(), ResponseTimeout)
	defer cancel()

	// Determine target peers
	var targetPeers []peer.ID
	if verifierID != "" {
		peerID, err := peer.Decode(verifierID)
		if err != nil {
			return nil, fmt.Errorf("invalid verifier ID: %v", err)
		}
		// Check if the peer is connected
		if rs.host.Network().Connectedness(peerID) != network.Connected {
			return nil, fmt.Errorf("not connected to verifier %s", verifierID)
		}
		targetPeers = []peer.ID{peerID}
	} else {
		targetPeers = rs.host.Network().Peers()
	}

	if len(targetPeers) == 0 {
		return nil, fmt.Errorf("no connected peers to query")
	}

	// Convert query to bytes
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	// Send query to target peers
	var wg sync.WaitGroup
	for _, peerID := range targetPeers {
		wg.Add(1)
		go func(peer peer.ID) {
			defer wg.Done()
			stream, err := rs.host.NewStream(ctx, peer, atypes.ProtocolAppVerifierProviderScoreRequest)
			if err != nil {
				reputationLog.Warnf("Failed to open stream to peer %s: %v", peer, err)
				return
			}
			defer stream.Close()

			// Write query to stream
			_, err = stream.Write(queryBytes)
			if err != nil {
				reputationLog.Warnf("Failed to write query to peer %s: %v", peer, err)
				stream.Reset()
				return
			}

			reputationLog.Debugf("Sent reputation query to peer %s", peer)
		}(peerID)
	}

	// Wait for all query sends to complete
	wg.Wait()

	// Wait for responses or timeout
	select {
	case <-ctx.Done():
		reputationLog.Debugf("Query timed out after waiting %s", ResponseTimeout)
	case <-result.Done:
		reputationLog.Debugf("Query completed with %d responses", len(result.Responses))
	}

	// Cleanup
	rs.pendingQueries.Delete(requestID)

	return result.Responses, result.Error
}

// handleReputationQuery handles incoming reputation queries
func (rs *ReputationService) handleReputationQuery(stream network.Stream) {
	defer stream.Close()

	// Read the query
	buf, err := readFullStream(stream)
	if err != nil {
		reputationLog.Errorf("Failed to read reputation query: %v", err)
		stream.Reset()
		return
	}

	var query ReputationQuery
	err = json.Unmarshal(buf, &query)
	if err != nil {
		reputationLog.Errorf("Failed to unmarshal reputation query: %v", err)
		return
	}

	// Check timestamp to prevent replay attacks
	if abs(time.Now().Unix()-query.Timestamp) > 60 {
		reputationLog.Warnf("Rejecting stale query from %s, timestamp difference: %d seconds",
			stream.Conn().RemotePeer(), abs(time.Now().Unix()-query.Timestamp))
		return
	}

	// Check if the query is directed to this verifier
	if query.VerifierID != "" && query.VerifierID != rs.host.ID().String() {
		reputationLog.Debugf("Ignoring query not directed to this verifier: %s", query.VerifierID)
		return
	}

	reputationLog.Debugf("Received reputation query from %s for app %d provider %s",
		stream.Conn().RemotePeer(), query.AppID, query.ProviderID)

	// Get the local score
	score, err := rs.getLocalScore(query.AppID, query.ProviderID)
	if err != nil {
		reputationLog.Errorf("Failed to get local score: %v", err)
		return
	}

	// Create response
	response := ReputationResponse{
		AppID:      query.AppID,
		ProviderID: query.ProviderID,
		Score:      score,
		Timestamp:  time.Now().Unix(),
		VerifierID: rs.host.ID().String(),
		RequestID:  query.RequestID,
	}

	// Send response
	rs.sendReputationResponse(stream.Conn().RemotePeer(), response)
}

// handleReputationResponse handles incoming reputation responses
func (rs *ReputationService) handleReputationResponse(stream network.Stream) {
	defer stream.Close()

	// Read the response
	buf, err := readFullStream(stream)
	if err != nil {
		reputationLog.Errorf("Failed to read reputation response: %v", err)
		stream.Reset()
		return
	}

	var response ReputationResponse
	err = json.Unmarshal(buf, &response)
	if err != nil {
		reputationLog.Errorf("Failed to unmarshal reputation response: %v", err)
		return
	}

	// Check timestamp to prevent replay attacks
	if abs(time.Now().Unix()-response.Timestamp) > 60 {
		reputationLog.Warnf("Rejecting stale response from %s, timestamp difference: %d seconds",
			stream.Conn().RemotePeer(), abs(time.Now().Unix()-response.Timestamp))
		return
	}

	// Verify the response is from the sender
	if response.VerifierID != stream.Conn().RemotePeer().String() {
		reputationLog.Warnf("Reputation response verifier ID doesn't match sender: %s vs %s",
			response.VerifierID, stream.Conn().RemotePeer())
		return
	}

	reputationLog.Debugf("Received reputation response from %s for request %s with score %d",
		stream.Conn().RemotePeer(), response.RequestID, response.Score)

	// Look up the pending query
	if result, found := rs.pendingQueries.Load(response.RequestID); found {
		queryResult := result.(*ReputationQueryResponse)

		// Add the response
		queryResult.Responses = append(queryResult.Responses, response)

		// If we have enough responses, mark the query as done
		if len(queryResult.Responses) >= MaxResponses {
			close(queryResult.Done)
		}
	} else {
		reputationLog.Warnf("Received response for unknown request ID: %s", response.RequestID)
	}
}

// sendReputationResponse sends a reputation response to a peer
func (rs *ReputationService) sendReputationResponse(peerID peer.ID, response ReputationResponse) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := rs.host.NewStream(ctx, peerID, atypes.ProtocolAppVerifierProviderScoreResponse)
	if err != nil {
		reputationLog.Errorf("Failed to open stream to peer %s: %v", peerID, err)
		return
	}
	defer stream.Close()

	// Convert response to bytes
	responseBytes, err := json.Marshal(response)
	if err != nil {
		reputationLog.Errorf("Failed to marshal response: %v", err)
		return
	}

	// Write response to stream
	_, err = stream.Write(responseBytes)
	if err != nil {
		reputationLog.Errorf("Failed to write response to peer %s: %v", peerID, err)
		stream.Reset()
		return
	}

	reputationLog.Debugf("Sent reputation response to peer %s", peerID)
}

// aggregateScores aggregates local and peer scores
func (rs *ReputationService) aggregateScores(localScore int, peerResponses []ReputationResponse) int {
	if len(peerResponses) == 0 {
		return localScore // Return local score if no peer responses
	}

	// Sum up all scores
	totalScore := localScore
	validResponses := 1 // Count local score as one valid response

	for _, response := range peerResponses {
		totalScore += response.Score
		validResponses++
	}

	// Calculate average
	if validResponses > 0 {
		return totalScore / validResponses
	}

	return 0
}

// readFullStream reads the entire contents of a stream
func readFullStream(stream network.Stream) ([]byte, error) {
	const maxSize = 1 << 20 // 1MB max message size

	var buf []byte
	for len(buf) < maxSize {
		// Read in chunks
		chunk := make([]byte, 1024)
		n, err := stream.Read(chunk)
		if err != nil {
			if err.Error() == "EOF" {
				// End of stream, append final chunk and return
				buf = append(buf, chunk[:n]...)
				return buf, nil
			}
			return nil, err
		}

		// Append the chunk to our buffer
		buf = append(buf, chunk[:n]...)
	}

	return buf, fmt.Errorf("message exceeds maximum size of 1MB")
}
