package contracts

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	bidenginetypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// BidMarketContractImpl implements BidMarketContract interface
type BidMarketContractImpl struct {
	client     *ethclient.Client
	contract   *BidMarket
	address    common.Address
	transactor *bind.TransactOpts
}

// NewBidMarketContract creates a new BidMarketContract instance
func NewBidMarketContract(client *ethclient.Client, address common.Address, transactor *bind.TransactOpts) (*BidMarketContractImpl, error) {
	contract, err := NewBidMarket(address, client)
	if err != nil {
		return nil, err
	}

	return &BidMarketContractImpl{
		client:     client,
		contract:   contract,
		address:    address,
		transactor: transactor,
	}, nil
}

// GetOrder retrieves an order by ID
func (b *BidMarketContractImpl) GetOrder(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error) {
	callOpts := &bind.CallOpts{Context: ctx}

	orderData, err := b.contract.Orders(callOpts, orderID)
	if err != nil {
		return nil, err
	}

	return &bidenginetypes.Order{
		ID:                        orderID,
		MachineType:               orderData.MachineType,
		Owner:                     orderData.Owner,
		Status:                    bidenginetypes.OrderStatus(orderData.Status),
		CreatedAt:                 orderData.CreatedAt,
		Duration:                  orderData.Duration,
		MinBidPrice:               orderData.MinBidPrice,
		MaxBidPrice:               orderData.MaxBidPrice,
		AcceptedBidPricePerSecond: orderData.AcceptedBidPricePerSecond,
		ParentOrderId:             orderData.ParentOrderId,
		PaymentToken:              orderData.PaymentToken,
		CpuCores:                  orderData.CpuCores,
		GpuCores:                  orderData.GpuCores,
		GpuMemory:                 orderData.GpuMemory,
		MemoryMB:                  orderData.MemoryMB,
		DiskGB:                    orderData.DiskGB,
		UploadMbps:                orderData.UploadMbps,
		DownloadMbps:              orderData.DownloadMbps,
		Region:                    orderData.Region,
		Specs:                     orderData.Specs,
		AcceptedProviderId:        orderData.AcceptedProviderId,
		AcceptedMachineId:         orderData.AcceptedMachineId,
		StartAt:                   orderData.StartAt,
		ExpiredAt:                 orderData.ExpiredAt,
		LastPaidAt:                orderData.LastPaidAt,
	}, nil
}

// GetOrderCount retrieves the total number of orders
func (b *BidMarketContractImpl) GetOrderCount(ctx context.Context) (*big.Int, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	return b.contract.OrderCount(callOpts)
}

// OrderCount retrieves the total number of orders (alias for GetOrderCount)
func (b *BidMarketContractImpl) OrderCount(ctx context.Context) (*big.Int, error) {
	return b.GetOrderCount(ctx)
}

// Orders retrieves order details by ID (alias for GetOrder)
func (b *BidMarketContractImpl) Orders(ctx context.Context, orderID *big.Int) (*bidenginetypes.Order, error) {
	return b.GetOrder(ctx, orderID)
}

// GetBids retrieves all bids for an order
func (b *BidMarketContractImpl) GetBids(ctx context.Context, orderID *big.Int) ([]bidenginetypes.Bid, error) {
	callOpts := &bind.CallOpts{Context: ctx}

	bidsData, err := b.contract.GetBids(callOpts, orderID)
	if err != nil {
		return nil, err
	}

	bids := make([]bidenginetypes.Bid, len(bidsData))
	for i, bidData := range bidsData {
		bids[i] = bidenginetypes.Bid{
			Provider:       bidData.Provider,
			PricePerSecond: bidData.PricePerSecond,
			Status:         bidenginetypes.BidStatus(bidData.Status),
			CreatedAt:      bidData.CreatedAt,
			ProviderId:     bidData.ProviderId,
			MachineId:      bidData.MachineId,
		}
	}

	return bids, nil
}

// IsBiddingOpen checks if bidding is still open for an order
func (b *BidMarketContractImpl) IsBiddingOpen(ctx context.Context, orderID *big.Int) (bool, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	return b.contract.IsBiddingOpen(callOpts, orderID)
}

// GetRemainingBidTime gets the remaining time for bidding
func (b *BidMarketContractImpl) GetRemainingBidTime(ctx context.Context, orderID *big.Int) (*big.Int, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	return b.contract.GetRemainingBidTime(callOpts, orderID)
}

// SubmitBid submits a new bid
func (b *BidMarketContractImpl) SubmitBid(ctx context.Context, orderID *big.Int, pricePerSecond *big.Int, providerID *big.Int, machineID *big.Int) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    b.transactor.From,
		Signer:  b.transactor.Signer,
	}

	return b.contract.SubmitBid(transactOpts, orderID, pricePerSecond, providerID, machineID)
}

// GetBidIndexFromTransaction gets the bid index from a transaction receipt by parsing the BidSubmitted event
func (b *BidMarketContractImpl) GetBidIndexFromTransaction(ctx context.Context, tx *types.Transaction, orderID *big.Int) (*big.Int, error) {
	receipt, err := bind.WaitMined(ctx, b.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Create filterer to parse events
	filterer, err := NewBidMarketFilterer(b.address, b.client)
	if err != nil {
		return nil, fmt.Errorf("failed to create filterer: %w", err)
	}

	// Parse logs to find BidSubmitted event
	for _, vLog := range receipt.Logs {
		event, err := filterer.ParseBidSubmitted(*vLog)
		if err == nil && event.OrderId.Cmp(orderID) == 0 {
			return event.BidIndex, nil
		}
	}

	return nil, fmt.Errorf("BidSubmitted event not found in transaction receipt")
}

// CancelBid cancels a bid
func (b *BidMarketContractImpl) CancelBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    b.transactor.From,
		Signer:  b.transactor.Signer,
	}

	return b.contract.CancelBid(transactOpts, orderID, bidIndex)
}

// AcceptBid accepts a bid
func (b *BidMarketContractImpl) AcceptBid(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    b.transactor.From,
		Signer:  b.transactor.Signer,
	}

	return b.contract.AcceptBid(transactOpts, orderID, bidIndex)
}

// CancelOrder cancels an order
func (b *BidMarketContractImpl) CancelOrder(ctx context.Context, orderID *big.Int) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    b.transactor.From,
		Signer:  b.transactor.Signer,
	}

	return b.contract.CancelOrder(transactOpts, orderID)
}

// CloseOrder closes an order
func (b *BidMarketContractImpl) CloseOrder(ctx context.Context, orderID *big.Int, reason string) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    b.transactor.From,
		Signer:  b.transactor.Signer,
	}

	return b.contract.CloseOrder(transactOpts, orderID, reason)
}

// ExtendOrder extends an order
func (b *BidMarketContractImpl) ExtendOrder(ctx context.Context, orderID *big.Int, amount *big.Int) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    b.transactor.From,
		Signer:  b.transactor.Signer,
	}

	return b.contract.Extend(transactOpts, orderID, amount)
}

// GetUsedResource gets the used resources for a machine
func (b *BidMarketContractImpl) GetUsedResource(ctx context.Context, providerID *big.Int, machineID *big.Int) (*bidenginetypes.ResourceUsage, error) {
	callOpts := &bind.CallOpts{Context: ctx}

	usage, err := b.contract.GetUsedResource(callOpts, providerID, machineID)
	if err != nil {
		return nil, err
	}

	return &bidenginetypes.ResourceUsage{
		CPUUsed:    usage.CpuCores,
		GPUUsed:    usage.GpuCores,
		MemoryUsed: usage.MemoryMB,
		DiskUsed:   usage.DiskGB,
	}, nil
}

// ReleaseOrderResource releases resources for an order
func (b *BidMarketContractImpl) ReleaseOrderResource(ctx context.Context, orderID *big.Int) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    b.transactor.From,
		Signer:  b.transactor.Signer,
	}

	return b.contract.ReleaseOrderResource(transactOpts, orderID)
}

// WatchOrderCreated watches for order created events
func (b *BidMarketContractImpl) WatchOrderCreated(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	watchOpts := &bind.WatchOpts{Context: ctx}

	eventSink := make(chan *BidMarketOrderCreated)
	sub, err := b.contract.WatchOrderCreated(watchOpts, eventSink, nil)
	if err != nil {
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case event := <-eventSink:
				sink <- &bidenginetypes.OrderEvent{
					Type:      bidenginetypes.OrderEventNew,
					OrderID:   event.OrderId,
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"owner":    event.Owner,
						"duration": event.Duration,
					},
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// WatchOrderClosed watches for order closed events
func (b *BidMarketContractImpl) WatchOrderClosed(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	watchOpts := &bind.WatchOpts{Context: ctx}

	eventSink := make(chan *BidMarketOrderClosed)
	sub, err := b.contract.WatchOrderClosed(watchOpts, eventSink, nil)
	if err != nil {
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case event := <-eventSink:
				sink <- &bidenginetypes.OrderEvent{
					Type:      bidenginetypes.OrderEventClosed,
					OrderID:   event.OrderId,
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"refundAmount": event.RefundAmount,
						"reason":       event.Reason,
					},
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// WatchOrderExpired watches for order expired events
func (b *BidMarketContractImpl) WatchOrderExpired(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	watchOpts := &bind.WatchOpts{Context: ctx}

	eventSink := make(chan *BidMarketBidTimeExpired)
	sub, err := b.contract.WatchBidTimeExpired(watchOpts, eventSink, nil)
	if err != nil {
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case event := <-eventSink:
				sink <- &bidenginetypes.OrderEvent{
					Type:      bidenginetypes.OrderEventExpired,
					OrderID:   event.OrderId,
					Timestamp: time.Now(),
					Data:      map[string]interface{}{},
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// WatchBidSubmitted watches for bid submitted events
func (b *BidMarketContractImpl) WatchBidSubmitted(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	watchOpts := &bind.WatchOpts{Context: ctx}

	eventSink := make(chan *BidMarketBidSubmitted)
	sub, err := b.contract.WatchBidSubmitted(watchOpts, eventSink, nil, nil)
	if err != nil {
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case event := <-eventSink:
				sink <- &bidenginetypes.OrderEvent{
					Type:      bidenginetypes.OrderEventUpdated,
					OrderID:   event.OrderId,
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"providerId": event.ProviderId,
						"machineId":  event.MachineId,
						"bidIndex":   event.BidIndex,
					},
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// WatchBidAccepted watches for bid accepted events
func (b *BidMarketContractImpl) WatchBidAccepted(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	watchOpts := &bind.WatchOpts{Context: ctx}

	eventSink := make(chan *BidMarketBidAccepted)
	sub, err := b.contract.WatchBidAccepted(watchOpts, eventSink, nil, nil)
	if err != nil {
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case event := <-eventSink:
				sink <- &bidenginetypes.OrderEvent{
					Type:      bidenginetypes.OrderEventUpdated,
					OrderID:   event.OrderId,
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"providerId":     event.ProviderId,
						"machineId":      event.MachineId,
						"bidIndex":       event.BidIndex,
						"pricePerSecond": event.PricePerSecond,
					},
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// WatchBidCancelled watches for bid cancelled events
func (b *BidMarketContractImpl) WatchBidCancelled(ctx context.Context, sink chan<- *bidenginetypes.OrderEvent) error {
	watchOpts := &bind.WatchOpts{Context: ctx}

	eventSink := make(chan *BidMarketBidCancelled)
	sub, err := b.contract.WatchBidCancelled(watchOpts, eventSink, nil, nil)
	if err != nil {
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case event := <-eventSink:
				sink <- &bidenginetypes.OrderEvent{
					Type:      bidenginetypes.OrderEventUpdated,
					OrderID:   event.OrderId,
					Timestamp: time.Now(),
					Data: map[string]interface{}{
						"providerId": event.ProviderId,
						"machineId":  event.MachineId,
						"bidIndex":   event.BidIndex,
					},
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// OrderBids retrieves a single bid by orderID and bidIndex
func (b *BidMarketContractImpl) OrderBids(ctx context.Context, orderID *big.Int, bidIndex *big.Int) (*bidenginetypes.Bid, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	bidData, err := b.contract.OrderBids(callOpts, orderID, bidIndex)
	if err != nil {
		return nil, err
	}
	return &bidenginetypes.Bid{
		Provider:       bidData.Provider,
		PricePerSecond: bidData.PricePerSecond,
		Status:         bidenginetypes.BidStatus(bidData.Status),
		CreatedAt:      bidData.CreatedAt,
		ProviderId:     bidData.ProviderId,
		MachineId:      bidData.MachineId,
		Id:             bidIndex,
	}, nil
}

// ProviderContractImpl implements ProviderContract interface
type ProviderContractImpl struct {
	client     *ethclient.Client
	contract   *Provider
	address    common.Address
	transactor *bind.TransactOpts
}

// NewProviderContract creates a new ProviderContract instance
func NewProviderContract(client *ethclient.Client, address common.Address, transactor *bind.TransactOpts) (*ProviderContractImpl, error) {
	contract, err := NewProvider(address, client)
	if err != nil {
		return nil, err
	}

	return &ProviderContractImpl{
		client:     client,
		contract:   contract,
		address:    address,
		transactor: transactor,
	}, nil
}

// GetProvider retrieves provider information
func (p *ProviderContractImpl) GetProvider(ctx context.Context, providerID *big.Int) (*bidenginetypes.Provider, error) {
	callOpts := &bind.CallOpts{Context: ctx}

	providerData, err := p.contract.GetProvider(callOpts, providerID)
	if err != nil {
		return nil, err
	}

	return &bidenginetypes.Provider{
		Operator:           providerData.Operator,
		Registered:         providerData.Registered,
		Reputation:         providerData.Reputation,
		MachineCount:       providerData.MachineCount,
		CreatedAt:          providerData.CreatedAt,
		UpdatedAt:          providerData.UpdatedAt,
		TotalStaked:        providerData.TotalStaked,
		PendingWithdrawals: providerData.PendingWithdrawals,
		SlashedAmount:      providerData.SlashedAmount,
		TokenId:            providerData.TokenId,
		Metadata:           providerData.Metadata,
		IsSlashed:          providerData.IsSlashed,
		IsActive:           providerData.IsActive,
		Verified:           providerData.Verified,
	}, nil
}

// GetProviderOwner gets the owner of a provider
func (p *ProviderContractImpl) GetProviderOwner(ctx context.Context, providerID *big.Int) (common.Address, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	return p.contract.GetProviderOwner(callOpts, providerID)
}

// IsProviderOperatorOrOwner checks if an account is operator or owner of a provider
func (p *ProviderContractImpl) IsProviderOperatorOrOwner(ctx context.Context, providerID *big.Int, account common.Address) (bool, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	return p.contract.IsProviderOperatorOrOwner(callOpts, providerID, account)
}

// IsVerified checks if a provider is verified
func (p *ProviderContractImpl) IsVerified(ctx context.Context, providerID *big.Int) (bool, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	return p.contract.IsVerified(callOpts, providerID)
}

// GetMachines retrieves all machines for a provider
func (p *ProviderContractImpl) GetMachines(ctx context.Context, providerID *big.Int) ([]*bidenginetypes.Machine, error) {
	callOpts := &bind.CallOpts{Context: ctx}

	machinesData, err := p.contract.GetMachines(callOpts, providerID)
	if err != nil {
		return nil, err
	}

	machines := make([]*bidenginetypes.Machine, len(machinesData))
	for i, machineData := range machinesData {
		machines[i] = &bidenginetypes.Machine{
			ID:                   big.NewInt(int64(i)),
			Active:               machineData.Active,
			MachineType:          machineData.MachineType,
			Region:               machineData.Region,
			CpuCores:             machineData.CpuCores,
			GpuCores:             machineData.GpuCores,
			GpuMemory:            machineData.GpuMemory,
			MemoryMB:             machineData.MemoryMB,
			DiskGB:               machineData.DiskGB,
			UploadSpeed:          machineData.UploadSpeed,
			DownloadSpeed:        machineData.DownloadSpeed,
			CreatedAt:            machineData.CreatedAt,
			UpdatedAt:            machineData.UpdatedAt,
			StakeAmount:          machineData.StakeAmount,
			RemovedAt:            machineData.RemovedAt,
			UnlockTime:           machineData.UnlockTime,
			WithdrawalProcessed:  machineData.WithdrawalProcessed,
			Metadata:             machineData.Metadata,
			CpuPricePerSecond:    machineData.CpuPricePerSecond,
			GpuPricePerSecond:    machineData.GpuPricePerSecond,
			MemoryPricePerSecond: machineData.MemoryPricePerSecond,
			DiskPricePerSecond:   machineData.DiskPricePerSecond,
		}
	}

	return machines, nil
}

// GetActiveMachinesPaginated gets active machines with pagination
func (p *ProviderContractImpl) GetActiveMachinesPaginated(ctx context.Context, providerID *big.Int, start *big.Int, limit *big.Int) ([]bidenginetypes.Machine, error) {
	callOpts := &bind.CallOpts{Context: ctx}

	machinesData, err := p.contract.GetActiveMachinesPaginated(callOpts, providerID, start, limit)
	if err != nil {
		return nil, err
	}

	machines := make([]bidenginetypes.Machine, len(machinesData))
	for i, machineData := range machinesData {
		machines[i] = bidenginetypes.Machine{
			Active:               machineData.Active,
			MachineType:          machineData.MachineType,
			Region:               machineData.Region,
			CpuCores:             machineData.CpuCores,
			GpuCores:             machineData.GpuCores,
			GpuMemory:            machineData.GpuMemory,
			MemoryMB:             machineData.MemoryMB,
			DiskGB:               machineData.DiskGB,
			UploadSpeed:          machineData.UploadSpeed,
			DownloadSpeed:        machineData.DownloadSpeed,
			CreatedAt:            machineData.CreatedAt,
			UpdatedAt:            machineData.UpdatedAt,
			StakeAmount:          machineData.StakeAmount,
			RemovedAt:            machineData.RemovedAt,
			UnlockTime:           machineData.UnlockTime,
			WithdrawalProcessed:  machineData.WithdrawalProcessed,
			Metadata:             machineData.Metadata,
			CpuPricePerSecond:    machineData.CpuPricePerSecond,
			GpuPricePerSecond:    machineData.GpuPricePerSecond,
			MemoryPricePerSecond: machineData.MemoryPricePerSecond,
			DiskPricePerSecond:   machineData.DiskPricePerSecond,
		}
	}

	return machines, nil
}

// IsMachineActive checks if a machine is active
func (p *ProviderContractImpl) IsMachineActive(ctx context.Context, providerID *big.Int, machineID *big.Int) (bool, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	return p.contract.IsMachineActive(callOpts, providerID, machineID)
}

// GetMachineResourcePrice gets the resource prices for a machine
func (p *ProviderContractImpl) GetMachineResourcePrice(ctx context.Context, providerID *big.Int, machineID *big.Int) (*bidenginetypes.ResourceUsage, error) {
	callOpts := &bind.CallOpts{Context: ctx}

	prices, err := p.contract.GetMachineResourcePrice(callOpts, providerID, machineID)
	if err != nil {
		return nil, err
	}

	return &bidenginetypes.ResourceUsage{
		CPUUsed:    prices.CpuPricePerSecond,
		GPUUsed:    prices.GpuPricePerSecond,
		MemoryUsed: prices.MemoryPricePerSecond,
		DiskUsed:   prices.DiskPricePerSecond,
	}, nil
}

// AddMachine adds a new machine
func (p *ProviderContractImpl) AddMachine(ctx context.Context, providerID *big.Int, machine bidenginetypes.Machine) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    p.transactor.From,
		Signer:  p.transactor.Signer,
	}

	return p.contract.AddMachine(transactOpts, providerID, machine.MachineType, machine.Region,
		machine.CpuCores, machine.GpuCores, machine.GpuMemory, machine.MemoryMB, machine.DiskGB,
		machine.UploadSpeed, machine.DownloadSpeed, machine.Metadata,
		machine.CpuPricePerSecond, machine.GpuPricePerSecond, machine.MemoryPricePerSecond, machine.DiskPricePerSecond)
}

// UpdateMachine updates an existing machine
func (p *ProviderContractImpl) UpdateMachine(ctx context.Context, providerID *big.Int, machineID *big.Int, machine bidenginetypes.Machine) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    p.transactor.From,
		Signer:  p.transactor.Signer,
	}

	return p.contract.UpdateMachine(transactOpts, providerID, machineID,
		machine.CpuCores, machine.GpuCores, machine.GpuMemory, machine.MemoryMB, machine.DiskGB,
		machine.UploadSpeed, machine.DownloadSpeed, machine.Metadata,
		machine.CpuPricePerSecond, machine.GpuPricePerSecond, machine.MemoryPricePerSecond, machine.DiskPricePerSecond)
}

// RemoveMachine removes a machine
func (p *ProviderContractImpl) RemoveMachine(ctx context.Context, providerID *big.Int, machineID *big.Int) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    p.transactor.From,
		Signer:  p.transactor.Signer,
	}

	return p.contract.RemoveMachine(transactOpts, providerID, machineID)
}

// SetMachineResourcePrice sets the resource prices for a machine
func (p *ProviderContractImpl) SetMachineResourcePrice(ctx context.Context, providerID *big.Int, machineID *big.Int, prices *bidenginetypes.ResourceUsage) (*types.Transaction, error) {
	transactOpts := &bind.TransactOpts{
		Context: ctx,
		From:    p.transactor.From,
		Signer:  p.transactor.Signer,
	}

	return p.contract.SetMachineResourcePrice(transactOpts, providerID, machineID,
		prices.CPUUsed, prices.GPUUsed, prices.MemoryUsed, prices.DiskUsed)
}

// ValidateMachineRequirements validates if a machine meets requirements
func (p *ProviderContractImpl) ValidateMachineRequirements(ctx context.Context, machineType *big.Int, providerID *big.Int, machineID *big.Int, requirements *bidenginetypes.ResourceUsage) (bool, error) {
	callOpts := &bind.CallOpts{Context: ctx}

	return p.contract.ValidateMachineRequirements(callOpts, machineType, providerID, machineID,
		requirements.CPUUsed, requirements.MemoryUsed, requirements.DiskUsed,
		requirements.GPUUsed, requirements.NetworkUsed, requirements.NetworkUsed)
}
