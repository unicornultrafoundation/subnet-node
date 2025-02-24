package apps

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	pvtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/app/verifier"
)

func (s *Service) ReportUsage(ctx context.Context, sign *pvtypes.SignatureResponse) (common.Hash, error) {
	// Create a new transactor
	key, err := s.accountService.NewKeyedTransactor()
	if err != nil {
		return common.Hash{}, err
	}

	usage := sign.SignedUsage

	// Call the ReportUsage function from the ABI
	tx, err := s.accountService.AppStore().ReportUsage(
		key, big.NewInt(usage.AppId), big.NewInt(usage.ProviderId), usage.PeerId,
		big.NewInt(usage.Cpu), big.NewInt(usage.Gpu), big.NewInt(usage.Memory), big.NewInt(usage.Storage),
		big.NewInt(usage.UploadBytes), big.NewInt(usage.DownloadBytes), big.NewInt(usage.Duration),
		big.NewInt(usage.Timestamp),
		usage.Signature)
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
