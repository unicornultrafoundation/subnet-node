package apps

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
)

func (s *Service) ReportUsage(ctx context.Context, usage *atypes.ResourceUsage, signature []byte) (common.Hash, error) {
	// Create a new transactor
	key, err := s.accountService.NewKeyedTransactor()
	if err != nil {
		return common.Hash{}, err
	}

	// Call the ReportUsage function from the ABI
	tx, err := s.accountService.AppStore().ReportUsage(
		key, usage.AppId, usage.ProviderId, usage.PeerId,
		usage.UsedCpu, usage.UsedGpu, usage.UsedMemory, usage.UsedStorage,
		usage.UsedUploadBytes, usage.UsedDownloadBytes, usage.Duration,
		usage.Timestamp,
		signature)
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
