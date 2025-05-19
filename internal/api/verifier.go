package api

import (
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

// PeerScoreListResponse represents a list of peer reputation scores
type PeerScoreListResponse struct {
	Scores []*PeerScoreResponse `json:"scores"`
	Total  int                  `json:"total"`
}

// PeerScoreListParams represents parameters for listing peer scores
type PeerScoreListParams struct {
	AppId      hexutil.Big `json:"app_id"`
	MinScore   *int        `json:"min_score,omitempty"`    // Optional minimum score filter
	MaxScore   *int        `json:"max_score,omitempty"`    // Optional maximum score filter
	Limit      int         `json:"limit,omitempty"`        // Maximum number of results to return
	Offset     int         `json:"offset,omitempty"`       // Offset for pagination
	SortByDesc bool        `json:"sort_by_desc,omitempty"` // Sort by score descending
}

// GetProviderScore retrieves a provider's reputation score
// func (api *VerifierAPI) GetProviderScore(ctx context.Context, appID *hexutil.Big, providerID string) (*PeerScoreResponse, error) {
// 	if api.verifier == nil {
// 		return nil, errors.New("verifier service is not enabled")
// 	}

// 	// Convert hexutil.Big to int64
// 	appIDInt64 := appID.ToInt().Int64()

// 	// Get the score from the verifier
// 	score, err := api.verifier.GetProviderScore(appIDInt64, providerID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get provider score: %v", err)
// 	}

// 	// Create the response
// 	response := &PeerScoreResponse{
// 		PeerID:    providerID,
// 		AppID:     appID,
// 		Score:     score,
// 		Timestamp: time.Now().Unix(),
// 	}

// 	return response, nil
// }

// // ListProviderScores lists provider reputation scores with optional filtering
// func (api *VerifierAPI) ListProviderScores(ctx context.Context, params PeerScoreListParams) (*PeerScoreListResponse, error) {
// 	if api.verifier == nil {
// 		return nil, errors.New("verifier service is not enabled")
// 	}

// 	// Convert hexutil.Big to int64
// 	appIDInt64 := params.AppId.ToInt().Int64()

// 	// Query all scores from the verifier
// 	allScores, err := api.verifier.QueryProviderScores()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query provider scores: %v", err)
// 	}

// 	// Filter scores by appID
// 	appScores, exists := allScores[appIDInt64]
// 	if !exists {
// 		// Return empty response if no scores for this app
// 		return &PeerScoreListResponse{
// 			Scores: []*PeerScoreResponse{},
// 			Total:  0,
// 		}, nil
// 	}

// 	// Convert to response format and apply filters
// 	var scores []*PeerScoreResponse
// 	for providerID, score := range appScores {
// 		// Apply score filters if provided
// 		if params.MinScore != nil && score < *params.MinScore {
// 			continue
// 		}
// 		if params.MaxScore != nil && score > *params.MaxScore {
// 			continue
// 		}

// 		scores = append(scores, &PeerScoreResponse{
// 			PeerID:    providerID,
// 			AppID:     &params.AppId,
// 			Score:     score,
// 			Timestamp: time.Now().Unix(),
// 		})
// 	}

// 	// Sort the scores
// 	if params.SortByDesc {
// 		sort.Slice(scores, func(i, j int) bool {
// 			return scores[i].Score > scores[j].Score
// 		})
// 	} else {
// 		sort.Slice(scores, func(i, j int) bool {
// 			return scores[i].Score < scores[j].Score
// 		})
// 	}

// 	// Apply pagination
// 	total := len(scores)
// 	if params.Offset > 0 && params.Offset < len(scores) {
// 		scores = scores[params.Offset:]
// 	}
// 	if params.Limit > 0 && params.Limit < len(scores) {
// 		scores = scores[:params.Limit]
// 	}

// 	return &PeerScoreListResponse{
// 		Scores: scores,
// 		Total:  total,
// 	}, nil
// }
