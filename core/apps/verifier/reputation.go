package verifier

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// Protocol IDs for reputation exchange
const (
	ProtocolPeerScoreQuery    = protocol.ID("/subnet/peer-score/query/1.0.0")
	ProtocolPeerScoreResponse = protocol.ID("/subnet/peer-score/response/1.0.0")

	// Datastore key prefix for peer scores
	PeerScorePrefix = "/peer_scores/"
)

// PeerScore represents a peer's reputation score
type PeerScore struct {
	PeerID    string    `json:"peer_id"`
	AppID     int64     `json:"app_id"`
	Score     int       `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

// StorePeerScore stores a peer's score in the datastore and publishes it to the DHT
func (v *Verifier) StorePeerScore(appID int64, peerID string, score int) error {
	peerScore := &PeerScore{
		PeerID:    peerID,
		AppID:     appID,
		Score:     score,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(peerScore)
	if err != nil {
		return fmt.Errorf("failed to marshal peer score: %v", err)
	}

	// Store locally
	key := datastore.NewKey(fmt.Sprintf("%s%d/%s", PeerScorePrefix, appID, peerID))
	err = v.ds.Put(context.Background(), key, data)
	if err != nil {
		return fmt.Errorf("failed to store peer score in datastore: %v", err)
	}

	// Publish to DHT for other verifiers to discover
	err = v.publishPeerScoreToDHT(appID, peerID, peerScore)
	if err != nil {
		log.Warnf("Failed to publish peer score to DHT: %v", err)
		// Continue even if DHT publish fails - local storage succeeded
	}

	log.Infof("Stored score %d for peer %s (app %d)", score, peerID, appID)
	return nil
}

// GetPeerScore retrieves a peer's score from the datastore
func (v *Verifier) GetPeerScore(appID int64, peerID string) (*PeerScore, error) {
	key := datastore.NewKey(fmt.Sprintf("%s%d/%s", PeerScorePrefix, appID, peerID))
	data, err := v.ds.Get(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("failed to get peer score from datastore: %v", err)
	}

	peerScore := &PeerScore{}
	err = json.Unmarshal(data, peerScore)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal peer score: %v", err)
	}

	return peerScore, nil
}

// RegisterPeerScoreHandlers registers handlers for peer score P2P protocols
func (v *Verifier) RegisterPeerScoreHandlers() {
	// Handler for peer score queries - open to any peer
	v.ps.SetStreamHandler(ProtocolPeerScoreQuery, func(s network.Stream) {
		defer s.Close()

		// Read query
		data, err := io.ReadAll(s)
		if err != nil {
			log.Errorf("Failed to read peer score query: %v", err)
			s.Reset()
			return
		}

		// Parse query
		var query struct {
			AppID  int64  `json:"app_id"`
			PeerID string `json:"peer_id"`
		}

		if err := json.Unmarshal(data, &query); err != nil {
			log.Errorf("Failed to unmarshal peer score query: %v", err)
			return
		}

		log.Infof("Received peer score query for peer %s (app %d) from %s",
			query.PeerID, query.AppID, s.Conn().RemotePeer())

		// Get peer score
		peerScore, err := v.GetPeerScore(query.AppID, query.PeerID)
		if err != nil {
			log.Warnf("Failed to get peer score: %v", err)
			peerScore = &PeerScore{
				PeerID:    query.PeerID,
				AppID:     query.AppID,
				Score:     0,
				Timestamp: time.Now(),
			}
		}

		// Send response
		responseData, err := json.Marshal(peerScore)
		if err != nil {
			log.Errorf("Failed to marshal peer score response: %v", err)
			return
		}

		// Create response stream
		responseStream, err := v.ps.NewStream(context.Background(), s.Conn().RemotePeer(), ProtocolPeerScoreResponse)
		if err != nil {
			log.Errorf("Failed to create response stream: %v", err)
			return
		}
		defer responseStream.Close()

		_, err = responseStream.Write(responseData)
		if err != nil {
			log.Errorf("Failed to send peer score response: %v", err)
			responseStream.Reset()
			return
		}

		log.Infof("Sent peer score %d for peer %s (app %d) to %s",
			peerScore.Score, peerScore.PeerID, peerScore.AppID, s.Conn().RemotePeer())
	})

	// Handler for peer score responses
	v.ps.SetStreamHandler(ProtocolPeerScoreResponse, func(s network.Stream) {
		defer s.Close()

		// Read response
		data, err := io.ReadAll(s)
		if err != nil {
			log.Errorf("Failed to read peer score response: %v", err)
			s.Reset()
			return
		}

		// Parse response
		peerScore := &PeerScore{}
		if err := json.Unmarshal(data, peerScore); err != nil {
			log.Errorf("Failed to unmarshal peer score response: %v", err)
			return
		}

		log.Infof("Received peer score %d for peer %s (app %d) from %s",
			peerScore.Score, peerScore.PeerID, peerScore.AppID, s.Conn().RemotePeer())
	})
}

// QueryPeerScoreFromVerifier queries a peer's score from another verifier
func (v *Verifier) QueryPeerScoreFromVerifier(verifierID peer.ID, appID int64, peerID string) (*PeerScore, error) {
	// Create query
	query := struct {
		AppID  int64  `json:"app_id"`
		PeerID string `json:"peer_id"`
	}{
		AppID:  appID,
		PeerID: peerID,
	}

	queryData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	// Create stream to verifier
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := v.ps.NewStream(ctx, verifierID, ProtocolPeerScoreQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream to verifier %s: %v", verifierID, err)
	}

	// Send query
	_, err = s.Write(queryData)
	if err != nil {
		s.Reset()
		return nil, fmt.Errorf("failed to send query: %v", err)
	}
	s.Close()

	// Wait for response
	responseStream, err := v.ps.NewStream(ctx, verifierID, ProtocolPeerScoreResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create response stream: %v", err)
	}
	defer responseStream.Close()

	// Read response
	responseData, err := io.ReadAll(responseStream)
	if err != nil {
		responseStream.Reset()
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	peerScore := &PeerScore{}
	if err := json.Unmarshal(responseData, peerScore); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	log.Infof("Received peer score %d for peer %s (app %d) from verifier %s",
		peerScore.Score, peerScore.PeerID, peerScore.AppID, verifierID)

	return peerScore, nil
}

// QueryPeerScoreFromDHT queries a peer's score from the DHT
func (v *Verifier) QueryPeerScoreFromDHT(appID int64, peerID string) ([]*PeerScore, error) {
	// Skip if DHT is not available
	if v.dht == nil {
		return nil, fmt.Errorf("DHT not available")
	}

	// Create DHT key prefix for this peer score
	dhtKeyPrefix := fmt.Sprintf("/subnet/peer-score/%d/%s", appID, peerID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find providers for this key
	peerScores := []*PeerScore{}

	// This is a simplified approach - in a real implementation, you would use a proper DHT query
	// For now, we'll just try to get values with known verifier IDs
	connectedPeers := v.ps.Network().Peers()
	for _, verifierID := range connectedPeers {
		dhtKey := fmt.Sprintf("%s/%s", dhtKeyPrefix, verifierID.String())

		// Get value from DHT using the ValueStore interface
		data, err := v.dht.GetValue(ctx, dhtKey)
		if err != nil {
			log.Debugf("Failed to get peer score from DHT for key %s: %v", dhtKey, err)
			continue
		}

		// Parse peer score
		peerScore := &PeerScore{}
		if err := json.Unmarshal(data, peerScore); err != nil {
			log.Warnf("Failed to unmarshal peer score from DHT: %v", err)
			continue
		}

		peerScores = append(peerScores, peerScore)
		log.Debugf("Found peer score in DHT: %+v", peerScore)
	}

	if len(peerScores) == 0 {
		return nil, fmt.Errorf("no peer scores found in DHT")
	}

	return peerScores, nil
}

// QueryPeerScoreFromAllVerifiers queries a peer's score from all known verifiers
// This now uses the DHT as the primary source and falls back to direct queries
func (v *Verifier) QueryPeerScoreFromAllVerifiers(appID int64, peerID string) (int, error) {
	// First try to get scores from DHT
	peerScores, err := v.QueryPeerScoreFromDHT(appID, peerID)

	// If DHT query successful, calculate average score
	if err == nil && len(peerScores) > 0 {
		totalScore := 0
		for _, score := range peerScores {
			totalScore += score.Score
		}
		return totalScore / len(peerScores), nil
	}

	// Fall back to local score if DHT query fails
	log.Debugf("DHT query failed, falling back to local score: %v", err)
	localScore, err := v.GetPeerScore(appID, peerID)
	if err != nil {
		return 0, fmt.Errorf("no scores available: %v", err)
	}

	return localScore.Score, nil
}

// publishPeerScoreToDHT publishes a peer score to the DHT for other verifiers to discover
func (v *Verifier) publishPeerScoreToDHT(appID int64, peerID string, peerScore *PeerScore) error {
	// Skip if DHT is not available
	if v.dht == nil {
		return fmt.Errorf("DHT not available")
	}

	// Create DHT key for this peer score
	dhtKey := fmt.Sprintf("/subnet/peer-score/%d/%s/%s", appID, peerID, v.ps.ID().String())

	// Marshal the data
	data, err := json.Marshal(peerScore)
	if err != nil {
		return fmt.Errorf("failed to marshal peer score for DHT: %v", err)
	}

	// Put value in DHT with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use the dual DHT's PutValue method
	err = v.dht.PutValue(ctx, dhtKey, data)
	if err != nil {
		return fmt.Errorf("failed to put peer score in DHT: %v", err)
	}

	log.Debugf("Published peer score for %s (app %d) to DHT", peerID, appID)
	return nil
}

// StartPeriodicScorePublishing starts a goroutine that periodically publishes all peer scores to the DHT
func (v *Verifier) StartPeriodicScorePublishing(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				v.publishAllScoresToDHT()
			}
		}
	}()

	log.Infof("Started periodic score publishing with interval %s", interval)
}

// publishAllScoresToDHT publishes all stored peer scores to the DHT
func (v *Verifier) publishAllScoresToDHT() {
	// Query all peer scores from local datastore
	query := query.Query{
		Prefix: PeerScorePrefix,
	}

	results, err := v.ds.Query(context.Background(), query)
	if err != nil {
		log.Errorf("Failed to query peer scores from datastore: %v", err)
		return
	}

	defer results.Close()

	count := 0
	for result := range results.Next() {
		if result.Error != nil {
			log.Warnf("Error in datastore query result: %v", result.Error)
			continue
		}

		// Parse key to get appID and peerID
		key := result.Key
		parts := strings.Split(strings.TrimPrefix(key, PeerScorePrefix), "/")
		if len(parts) != 2 {
			log.Warnf("Invalid peer score key format: %s", key)
			continue
		}

		appID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			log.Warnf("Invalid appID in key %s: %v", key, err)
			continue
		}

		peerID := parts[1]

		// Parse peer score
		peerScore := &PeerScore{}
		if err := json.Unmarshal(result.Value, peerScore); err != nil {
			log.Warnf("Failed to unmarshal peer score: %v", err)
			continue
		}

		// Publish to DHT
		err = v.publishPeerScoreToDHT(appID, peerID, peerScore)
		if err != nil {
			log.Warnf("Failed to publish peer score to DHT: %v", err)
			continue
		}

		count++
	}

	log.Infof("Published %d peer scores to DHT", count)
}

// getKnownVerifiers returns a list of known verifier peer IDs
// This is a placeholder - in a real implementation, you would get this from a registry or discovery mechanism
func (v *Verifier) getKnownVerifiers() ([]peer.ID, error) {
	// In this simplified approach, we consider all connected peers as potential verifiers
	return v.ps.Network().Peers(), nil
}
