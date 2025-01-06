package apps

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/unicornultrafoundation/subnet-node/config"
)

func (s *Service) RegisterApp(ctx context.Context, name, symbol, peerId string, metadata AppMetadata, budget, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB *big.Int) (common.Hash, error) {
	// Create a new transactor
	key, err := s.accountService.NewKeyedTransactor()
	if err != nil {
		return common.Hash{}, err
	}

	key.Value = budget

	metadataBytes, err := json.Marshal(metadata)

	if err != nil {
		return common.Hash{}, err
	}

	metadataBase64 := base64.StdEncoding.EncodeToString(metadataBytes)
	// Call the CreateApp function from the ABI
	tx, err := s.subnetAppRegistry.CreateApp(key, name, symbol, peerId, budget, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadataBase64)
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

func (s *Service) CreateNode(ctx context.Context, name string, metadata string, nftId *big.Int) (common.Hash, error) {
	// Create a new transactor
	key, err := s.accountService.NewKeyedTransactor()
	if err != nil {
		return common.Hash{}, err
	}

	subnetRegistryAddr := s.cfg.GetString("apps.subnet_registry_contract", config.DefaultSubnetRegistryCOntract)

	tx, err := s.accountService.NftLicense().Approve(key, common.HexToAddress(subnetRegistryAddr), nftId)
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

	// Call the CreateNode function from the ABI
	tx, err = s.subnetRegistry.RegisterSubnet(key, nftId, s.peerId.String(), name, metadata)
	if err != nil {
		return common.Hash{}, err
	}

	// Wait for the transaction to be mined
	receipt, err = bind.WaitMined(ctx, s.ethClient, tx)
	if err != nil {
		return common.Hash{}, err
	}

	if receipt.Status != 1 {
		return common.Hash{}, fmt.Errorf("transaction failed: %s", tx.Hash().Hex())
	}

	return tx.Hash(), nil
}

func (s *Service) RegisterNode(ctx context.Context, appId *big.Int) (common.Hash, error) {
	_, err := s.subnetAppRegistry.GetAppNode(nil, appId, &s.subnetID)
	if err == nil {
		return common.Hash{}, nil
	}
	// Get subnetId from peerId
	subnetId, err := s.subnetRegistry.PeerToSubnet(nil, s.peerId.String())
	if err != nil {
		return common.Hash{}, err
	}

	// Create a new transactor
	key, err := s.accountService.NewKeyedTransactor()
	if err != nil {
		return common.Hash{}, err
	}

	// Call the RegisterNode function from the ABI
	tx, err := s.subnetAppRegistry.RegisterNode(key, subnetId, appId)
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

func (s *Service) ClaimReward(ctx context.Context, appId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration *big.Int, signature []byte) (common.Hash, error) {
	// Get subnetId from peerId
	subnetId, err := s.subnetRegistry.PeerToSubnet(nil, s.peerId.String())
	if err != nil {
		return common.Hash{}, err
	}

	// Create a new transactor
	key, err := s.accountService.NewKeyedTransactor()
	if err != nil {
		return common.Hash{}, err
	}

	// Call the ClaimReward function from the ABI
	tx, err := s.subnetAppRegistry.ClaimReward(key, subnetId, appId, usedCpu, usedGpu, usedMemory, usedStorage, usedUploadBytes, usedDownloadBytes, duration, signature)
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

func (s *Service) UpdateApp(ctx context.Context, appId *big.Int, name, peerId, metadata string, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB *big.Int) (common.Hash, error) {
	// Get the current app details
	app, err := s.subnetAppRegistry.GetApp(nil, appId)
	if err != nil {
		return common.Hash{}, err
	}

	// Use current values if new values are not provided
	if name == "" {
		name = app.Name
	}
	if peerId == "" {
		peerId = app.PeerId
	}
	if metadata == "" {
		metadata = app.Metadata
	}
	if maxNodes == nil {
		maxNodes = app.MaxNodes
	}
	if minCpu == nil {
		minCpu = app.MinCpu
	}
	if minGpu == nil {
		minGpu = app.MinGpu
	}
	if minMemory == nil {
		minMemory = app.MinMemory
	}
	if minUploadBandwidth == nil {
		minUploadBandwidth = app.MinUploadBandwidth
	}
	if minDownloadBandwidth == nil {
		minDownloadBandwidth = app.MinDownloadBandwidth
	}
	if pricePerCpu == nil {
		pricePerCpu = app.PricePerCpu
	}
	if pricePerGpu == nil {
		pricePerGpu = app.PricePerGpu
	}
	if pricePerMemoryGB == nil {
		pricePerMemoryGB = app.PricePerMemoryGB
	}
	if pricePerStorageGB == nil {
		pricePerStorageGB = app.PricePerStorageGB
	}
	if pricePerBandwidthGB == nil {
		pricePerBandwidthGB = app.PricePerBandwidthGB
	}

	// Create a new transactor
	key, err := s.accountService.NewKeyedTransactor()
	if err != nil {
		return common.Hash{}, err
	}

	// Call the UpdateApp function from the ABI
	tx, err := s.subnetAppRegistry.UpdateApp(key, appId, name, peerId, maxNodes, minCpu, minGpu, minMemory, minUploadBandwidth, minDownloadBandwidth, pricePerCpu, pricePerGpu, pricePerMemoryGB, pricePerStorageGB, pricePerBandwidthGB, metadata)
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
