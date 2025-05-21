package api

import (
	"context"

	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
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

// GetProviderReputationScore retrieves the reputation score for a provider
func (api *VerifierAPI) GetProviderReputationScore(ctx context.Context, providerId string) (int, error) {
	score, err := api.verifier.GetReputationScore(ctx, providerId)
	if err != nil {
		return 0, err
	}
	return score, nil
}

// GetReputationScoreRanges retrieves reputation scores grouped by ranges
func (api *VerifierAPI) GetReputationScoreRanges(ctx context.Context) ([]atypes.ReputationScoreRange, error) {
	// Call the implementation in the verifier
	verifierRanges, err := api.verifier.GetReputationScoreRanges(ctx)
	if err != nil {
		return nil, err
	}

	return verifierRanges, nil

}

func (api *VerifierAPI) GetVerifierIds(ctx context.Context) ([]string, error) {
	return api.verifier.GetVerifierIds(ctx)
}
