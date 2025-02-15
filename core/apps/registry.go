package apps

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (s *Service) ReportUsage(ctx context.Context, appId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration *big.Int, signature []byte) (common.Hash, error) {
	providerId := big.NewInt(s.accountService.ProviderID())
	// Create a new transactor
	key, err := s.accountService.NewKeyedTransactor()
	if err != nil {
		return common.Hash{}, err
	}

	peerId := s.peerId.String()

	// Call the ClaimReward function from the ABI
	tx, err := s.accountService.AppStore().ReportUsage(key, providerId, appId, peerId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration, signature)
	if err != nil {
		return common.Hash{}, err
	}

	// Wait for the transaction to be mined
	receipt, err := bind.WaitMined(ctx, s.ethClient, tx)
	if err != nil {
		return common.Hash{}, err
	}

	if receipt.Status != 1 {
		return common.Hash{}, fmt.Errorf("transaction failed: %s", tx.Hash().Hex())
	}

	return tx.Hash(), nil
}
