package api

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/unicornultrafoundation/subnet-node/core/apps/verifier"
)

// VerifierAPI provides RPC methods for interacting with the verifier service
type VerifierAPI struct {
	verifier *verifier.Verifier
}

// NewVerifierAPI creates a new instance of VerifierAPI
func NewVerifierAPI(verifier *verifier.Verifier) *VerifierAPI {
	return &VerifierAPI{verifier: verifier}
}

// PeerScoreResponse represents a peer's reputation score
type PeerScoreResponse struct {
	PeerID    string       `json:"peer_id"`
	AppID     *hexutil.Big `json:"app_id"`
	Score     int          `json:"score"`
	Timestamp int64        `json:"timestamp"`
}

// GetPeerScore retrieves a peer's reputation score for a specific app
func (api *VerifierAPI) GetPeerScore(ctx context.Context, appId hexutil.Big, peerId string) (*PeerScoreResponse, error) {
	if api.verifier == nil {
		return nil, fmt.Errorf("verifier service is not enabled")
	}

	// Get peer score from local datastore
	peerScore, err := api.verifier.GetPeerScore(appId.ToInt().Int64(), peerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get peer score: %v", err)
	}

	return &PeerScoreResponse{
		PeerID:    peerScore.PeerID,
		AppID:     (*hexutil.Big)(big.NewInt(peerScore.AppID)),
		Score:     peerScore.Score,
		Timestamp: peerScore.Timestamp.Unix(),
	}, nil
}

// GetAggregatedPeerScore retrieves a peer's aggregated reputation score from all known verifiers
func (api *VerifierAPI) GetAggregatedPeerScore(ctx context.Context, appId hexutil.Big, peerId string) (*PeerScoreResponse, error) {
	if api.verifier == nil {
		return nil, fmt.Errorf("verifier service is not enabled")
	}

	// Query score from all verifiers
	score, err := api.verifier.QueryPeerScoreFromAllVerifiers(appId.ToInt().Int64(), peerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated peer score: %v", err)
	}

	return &PeerScoreResponse{
		PeerID:    peerId,
		AppID:     &appId,
		Score:     score,
		Timestamp: 0, // Aggregated score doesn't have a specific timestamp
	}, nil
}
