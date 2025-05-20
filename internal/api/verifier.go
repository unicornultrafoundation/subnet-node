package api

import (
	"context"
	"fmt"
	"math/big"
	"strings"

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

// ReputationScoreResponse represents a response containing a provider's reputation score
type ReputationScoreResponse struct {
	AppID      int64  `json:"app_id"`
	ProviderID string `json:"provider_id"`
	VerifierID string `json:"verifier_id"`
	Score      int    `json:"score"`
}

// GetProviderReputationScore retrieves the reputation score for a provider
func (api *VerifierAPI) GetProviderReputationScore(ctx context.Context, appId string, providerId string, verifierId string) (*ReputationScoreResponse, error) {
	appId = strings.TrimPrefix(appId, "0x")
	appIdInt, ok := new(big.Int).SetString(appId, 16)

	if !ok {
		return nil, fmt.Errorf("invalid app id: %s", appId)
	}

	appIdInt64 := appIdInt.Int64()

	score, err := api.verifier.GetReputationScore(ctx, appIdInt64, providerId, verifierId)
	if err != nil {
		return nil, err
	}

	return &ReputationScoreResponse{
		AppID:      appIdInt64,
		ProviderID: providerId,
		VerifierID: verifierId,
		Score:      score,
	}, nil
}

func (api *VerifierAPI) GetVerifierIds(ctx context.Context) ([]string, error) {
	return api.verifier.GetVerifierIds(ctx)
}

// // ListProvidersByReputation lists providers sorted by their reputation scores
// func (api *VerifierAPI) ListProvidersByReputation(ctx context.Context, req *ListProvidersRequest) (*ListProvidersResponse, error) {
// 	providers, err := api.verifier.ListProvidersByReputation(req.AppID, req.MinScore, req.MaxScore, req.Limit, req.SortOrder)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &ListProvidersResponse{
// 		AppID:     req.AppID,
// 		Providers: providers,
// 	}, nil
// }
