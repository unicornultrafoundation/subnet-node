package bidengine

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
)

// BidEngine is responsible for managing bids on the BidMarket contract
type BidEngine struct {
	cfg           *config.C
	client        *ethclient.Client
	auth          *bind.TransactOpts
	log           *logrus.Logger
	provider      *contracts.Provider
	bidMarket     *contracts.BidMarket
	providerAddr  common.Address
	bidMarketAddr common.Address
	activeBids    map[string]*Bid
	bidsMutex     sync.RWMutex
	bidConfig     BidConfig
	providerId    *big.Int // Our provider ID

	shutdown chan struct{}
	wg       sync.WaitGroup
}

// NewBidEngine creates a new instance of the BidEngine
func NewBidEngine(cfg *config.C, log *logrus.Logger, acc account.Service) (*BidEngine, error) {
	client := acc.GetClient()

	// Get contract addresses from config
	providerAddrStr := cfg.GetString("contracts.provider", "")
	if providerAddrStr == "" {
		return nil, fmt.Errorf("contracts.provider address is not set in config")
	}
	providerAddr := common.HexToAddress(providerAddrStr)

	bidMarketAddrStr := cfg.GetString("contracts.bid_market", "")
	if bidMarketAddrStr == "" {
		return nil, fmt.Errorf("contracts.bid_market address is not set in config")
	}
	bidMarketAddr := common.HexToAddress(bidMarketAddrStr)

	// Initialize contracts
	provider, err := contracts.NewProvider(providerAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider contract: %v", err)
	}

	bidMarket, err := contracts.NewBidMarket(bidMarketAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create bid market contract: %v", err)
	}

	auth, err := acc.NewKeyedTransactor()
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	return &BidEngine{
		cfg:           cfg,
		client:        client,
		log:           log,
		provider:      provider,
		bidMarket:     bidMarket,
		providerAddr:  providerAddr,
		bidMarketAddr: bidMarketAddr,
		activeBids:    make(map[string]*Bid),
		shutdown:      make(chan struct{}),
		auth:          auth,
	}, nil
}

// WatchNewOrders subscribes to new order events on the blockchain and processes them
func (b *BidEngine) watchNewOrders(ctx context.Context) error {
	// Create a filter for new orders
	orderCreatedCh := make(chan *contracts.BidMarketOrderCreated)
	orderSub, err := b.bidMarket.WatchOrderCreated(&bind.WatchOpts{
		Context: ctx,
	}, orderCreatedCh, nil) // You might want to add filters if available

	if err != nil {
		return fmt.Errorf("failed to watch for new orders: %v", err)
	}

	b.log.Info("Started watching for new orders on blockchain")
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		defer orderSub.Unsubscribe()

		for {
			select {
			case err := <-orderSub.Err():
				b.log.WithError(err).Error("Order subscription error")

				// Try to resubscribe after a delay
				time.Sleep(time.Second * 10)
				newSub, err := b.bidMarket.WatchOrderCreated(&bind.WatchOpts{
					Context: ctx,
				}, orderCreatedCh, nil)

				if err != nil {
					b.log.WithError(err).Error("Failed to resubscribe to order events")
					continue
				}
				orderSub = newSub

			case event := <-orderCreatedCh:
				// Process the new order
				b.processNewOrder(ctx, event)

			case <-b.shutdown:
				b.log.Info("Stopping order watcher")
				return

			case <-ctx.Done():
				b.log.Info("Context canceled, stopping order watcher")
				return
			}
		}
	}()

	return nil
}

// processNewOrder handles a new order event from the blockchain
func (b *BidEngine) processNewOrder(ctx context.Context, event *contracts.BidMarketOrderCreated) {
	b.log.WithFields(logrus.Fields{
		"orderId":   event.OrderId,
		"requester": event.Owner.String(),
	}).Info("New order detected on blockchain")

	// Fetch complete order details from the blockchain
	orderDetails, err := b.bidMarket.Orders(&bind.CallOpts{
		Context: ctx,
	}, event.OrderId)

	if err != nil {
		b.log.WithError(err).WithField("orderId", event.OrderId).Error("Failed to fetch order details")
		return
	}

	// Convert contract order to our internal OrderInfo struct
	order := OrderInfo{
		OrderID:     event.OrderId,
		RequesterID: event.Owner,
		Requirements: &BidRequirements{
			MinCPUCores:      orderDetails.CpuCores,
			MinMemoryMB:      orderDetails.MemoryMB,
			MinDiskGB:        orderDetails.DiskGB,
			MinGPUCores:      orderDetails.GpuCores,
			MinUploadSpeed:   orderDetails.UploadMbps,
			MinDownloadSpeed: orderDetails.DownloadMbps,
		},
		MaxPrice:     orderDetails.MaxBidPrice,
		MinPrice:     orderDetails.MinBidPrice,
		CreatedAt:    time.Unix(orderDetails.CreatedAt.Int64(), 0),
		ExpirationAt: time.Unix(orderDetails.ExpiredAt.Int64(), 0),
		Status:       orderDetails.Status,
	}

	// Check if we already have a bid on this order
	bidKey := order.OrderID.String()
	b.bidsMutex.RLock()
	_, alreadyBid := b.activeBids[bidKey]
	b.bidsMutex.RUnlock()

	if alreadyBid {
		b.log.WithField("orderId", order.OrderID).Debug("Already bid on this order")
		return
	}

	// Check if the order is eligible for bidding
	eligible, err := b.checkOrderEligibility(ctx, order)
	if err != nil {
		b.log.WithError(err).WithField("orderId", order.OrderID).Error("Error checking order eligibility")
		return
	}

	if !eligible {
		b.log.WithField("orderId", order.OrderID).Debug("Order not eligible for bidding")
		return
	}

	// Calculate bid amount
	bidAmount := b.calculateBidAmount(order)

	// Find a suitable machine
	machineId := b.findSuitableMachine(ctx, order.Requirements)
	if machineId == nil {
		b.log.WithField("orderId", order.OrderID).Warn("No suitable machine found for order")
		return
	}

	// Log our intention
	b.log.WithFields(logrus.Fields{
		"orderId":   order.OrderID,
		"maxPrice":  order.MaxPrice,
		"bidAmount": bidAmount,
	}).Info("Placing bid on new order")

	// Place the bid
	_, err = b.PlaceBid(ctx, order, b.providerId, machineId, bidAmount, order.Requirements)
	if err != nil {
		b.log.WithError(err).WithField("orderId", order.OrderID).Error("Failed to place bid")
		return
	}
}

// Start initializes and starts the bid engine processes
func (b *BidEngine) Start(ctx context.Context) error {
	b.log.Info("Starting BidEngine...")

	// Load bid configuration
	b.loadBidConfig()

	// Start monitoring for active bids
	b.wg.Add(1)
	go b.monitorActiveBids(ctx)

	// Start watching for new orders directly from the blockchain
	if err := b.watchNewOrders(ctx); err != nil {
		return fmt.Errorf("failed to start watching orders: %v", err)
	}

	return nil
}

// Stop gracefully stops the bid engine
func (b *BidEngine) Stop(ctx context.Context) error {
	b.log.Info("Stopping BidEngine...")
	close(b.shutdown)

	// Wait for all goroutines to finish with a timeout
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		b.log.Info("BidEngine stopped gracefully")
	case <-ctx.Done():
		b.log.Warn("BidEngine stop timed out, some goroutines may still be running")
	}

	return nil
}

// PlaceBid places a bid on a specific order in the BidMarket contract
func (b *BidEngine) PlaceBid(ctx context.Context, order OrderInfo, providerId, machineId, amount *big.Int, requirements *BidRequirements) (*Bid, error) {
	// Check if machine is active
	isActive, err := b.provider.IsMachineActive(&bind.CallOpts{Context: ctx}, providerId, machineId)
	if err != nil {
		return nil, fmt.Errorf("failed to check if machine is active: %v", err)
	}

	if !isActive {
		return nil, fmt.Errorf("machine %v is not active", machineId)
	}

	// Validate machine specs against requirements
	valid, err := b.provider.ValidateMachineRequirements(
		&bind.CallOpts{Context: ctx},
		big.NewInt(0), // machineType
		providerId,
		machineId,
		requirements.MinCPUCores,
		requirements.MinMemoryMB,
		requirements.MinDiskGB,
		requirements.MinGPUCores,
		requirements.MinUploadSpeed,
		requirements.MinDownloadSpeed,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to validate machine requirements: %v", err)
	}

	if !valid {
		return nil, fmt.Errorf("machine %v does not meet requirements", machineId)
	}

	// Place bid on the blockchain
	tx, err := b.bidMarket.SubmitBid(b.auth, order.OrderID, providerId, machineId, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to place bid on blockchain: %v", err)
	}

	bid := &Bid{
		ID:           order.OrderID, // Using order ID as bid ID for tracking
		ProviderId:   providerId,
		MachineId:    machineId,
		Amount:       amount,
		Status:       BidStatusPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpirationAt: time.Now().Add(time.Minute * 5), // Set expiration to 5 minutes
		Requirements: requirements,
		TxHash:       tx.Hash().String(),
	}

	// Wait for transaction confirmation (optional)
	_, err = bind.WaitMined(ctx, b.client, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm bid transaction: %v", err)
	}

	// Add to active bids - use order ID as key for easier tracking
	bidKey := order.OrderID.String()
	b.bidsMutex.Lock()
	b.activeBids[bidKey] = bid
	b.bidsMutex.Unlock()

	b.log.WithFields(logrus.Fields{
		"orderId":    order.OrderID,
		"providerId": providerId,
		"machineId":  machineId,
		"amount":     amount.String(),
		"txHash":     bid.TxHash,
	}).Info("Bid placed successfully on order")

	return bid, nil
}

// GetActiveBids returns all active bids
func (b *BidEngine) GetActiveBids() []*Bid {
	b.bidsMutex.RLock()
	defer b.bidsMutex.RUnlock()

	bids := make([]*Bid, 0, len(b.activeBids))
	for _, bid := range b.activeBids {
		bids = append(bids, bid)
	}

	return bids
}

// loadBidConfig loads the bid configuration from the config file
func (b *BidEngine) loadBidConfig() {
	b.bidConfig = BidConfig{
		MinBidPercent: b.cfg.GetInt("bidengine.min_bid_percent", 70),
		MaxBidPercent: b.cfg.GetInt("bidengine.max_bid_percent", 95),
		PriceFactor:   float64(b.cfg.GetInt("bidengine.price_factor", 1)),
		MinRequirements: &BidRequirements{
			MinCPUCores:      big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.cpu_cores", 1))),
			MinMemoryMB:      big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.memory_mb", 1024))),
			MinDiskGB:        big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.disk_gb", 10))),
			MinGPUCores:      big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.gpu_cores", 0))),
			MinUploadSpeed:   big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.upload_speed", 0))),
			MinDownloadSpeed: big.NewInt(int64(b.cfg.GetInt("bidengine.min_requirements.download_speed", 0))),
		},
		MaxProfit: new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.max_profit", 1000000000000000000))), // Default 1 ETH

		// Resource cost parameters with defaults in wei
		CpuCorePrice:       new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.cpu_core", 1000000000000000))),     // 0.001 ETH default
		MemoryGBPrice:      new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.memory_gb", 2000000000000000))),    // 0.002 ETH default
		DiskGBPrice:        new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.disk_gb", 100000000000000))),       // 0.0001 ETH default
		GpuCorePrice:       new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.gpu_core", 5000000000000000))),     // 0.005 ETH default
		UploadSpeedPrice:   new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.upload_speed", 50000000000000))),   // 0.00005 ETH default
		DownloadSpeedPrice: new(big.Int).SetInt64(int64(b.cfg.GetInt("bidengine.resource_cost.download_speed", 50000000000000))), // 0.00005 ETH default

		// Initialize machine type cost multipliers with defaults
		// Type 0 (shared resources): 1.0x multiplier (base cost)
		// Type 1 (VMs): 1.5x multiplier
		// Type 2 (bare metal): 2.5x multiplier
		MachineTypeMultipliers: make(map[int64]*big.Int),
	}

	// Load machine type multipliers from config
	// Default: shared=1.0, VM=1.5, bare metal=2.5
	b.bidConfig.MachineTypeMultipliers[1] = big.NewInt(int64(b.cfg.GetInt("bidengine.machine_type_multipliers.shared", 10)))
	b.bidConfig.MachineTypeMultipliers[2] = big.NewInt(int64(b.cfg.GetInt("bidengine.machine_type_multipliers.vm", 15)))
	b.bidConfig.MachineTypeMultipliers[3] = big.NewInt(int64(b.cfg.GetInt("bidengine.machine_type_multipliers.bare_metal", 25)))
	// You can add more machine types as needed
}

// checkOrderEligibility checks if we can and should bid on this order
func (b *BidEngine) checkOrderEligibility(_ context.Context, order OrderInfo) (bool, error) {
	// 1. Check if requirements meet our minimum criteria
	if order.Requirements.MinCPUCores.Cmp(b.bidConfig.MinRequirements.MinCPUCores) < 0 ||
		order.Requirements.MinMemoryMB.Cmp(b.bidConfig.MinRequirements.MinMemoryMB) < 0 ||
		order.Requirements.MinDiskGB.Cmp(b.bidConfig.MinRequirements.MinDiskGB) < 0 {
		return false, nil
	}

	// 2. Check if the max price is worth our time
	// Calculate our minimum acceptable price based on resource requirements
	costEstimate := b.calculateResourceCost(order.Requirements)

	// If maximum order price doesn't give us enough profit, don't bid
	if new(big.Int).Sub(order.MaxPrice, costEstimate).Cmp(b.bidConfig.MaxProfit) < 0 {
		return false, nil
	}

	// 3. Check if we have available resources to fulfill this order
	// In a real implementation, you'd check against your available machines

	return true, nil
}

// calculateBidAmount calculates the amount to bid based on the order and our config
func (b *BidEngine) calculateBidAmount(order OrderInfo) *big.Int {
	// Calculate a bid amount between min and max percentages of the max price
	minAmount := new(big.Int).Mul(order.MaxPrice, big.NewInt(int64(b.bidConfig.MinBidPercent)))
	minAmount = minAmount.Div(minAmount, big.NewInt(100))

	maxAmount := new(big.Int).Mul(order.MaxPrice, big.NewInt(int64(b.bidConfig.MaxBidPercent)))
	maxAmount = maxAmount.Div(maxAmount, big.NewInt(100))

	// Choose a bid amount between min and max, potentially based on competition
	// For simplicity, we'll just use the min amount for now
	bidAmount := new(big.Int).Set(minAmount)

	// Ensure our bid is at least our cost estimate plus minimum profit
	costEstimate := b.calculateResourceCost(order.Requirements)
	minProfitableAmount := new(big.Int).Add(costEstimate, big.NewInt(int64(b.cfg.GetInt("bidengine.min_profit", 1000000000))))

	if bidAmount.Cmp(minProfitableAmount) < 0 {
		bidAmount = minProfitableAmount
	}

	// Ensure we don't exceed our maximum bid amount
	if bidAmount.Cmp(maxAmount) > 0 {
		b.log.WithFields(logrus.Fields{
			"calculatedBid": bidAmount,
			"maxBid":        maxAmount,
			"orderId":       order.OrderID,
		}).Info("Calculated bid exceeds maximum, capping at max amount")
		bidAmount = maxAmount
	}

	return bidAmount
}

// calculateResourceCost estimates the cost to fulfill the requirements based on configured prices
// and machine type
func (b *BidEngine) calculateResourceCost(requirements *BidRequirements) *big.Int {
	// Calculate base costs as before
	cpuCost := new(big.Int).Mul(requirements.MinCPUCores, b.bidConfig.CpuCorePrice)

	// Calculate memory cost (convert MB to GB for pricing)
	memGB := new(big.Int).Div(requirements.MinMemoryMB, big.NewInt(1024))
	memCost := new(big.Int).Mul(memGB, b.bidConfig.MemoryGBPrice)

	// Calculate disk cost
	diskCost := new(big.Int).Mul(requirements.MinDiskGB, b.bidConfig.DiskGBPrice)

	// Calculate GPU cost if applicable
	gpuCost := new(big.Int).Mul(requirements.MinGPUCores, b.bidConfig.GpuCorePrice)

	// Calculate network costs if applicable
	uploadCost := new(big.Int).Mul(requirements.MinUploadSpeed, b.bidConfig.UploadSpeedPrice)
	downloadCost := new(big.Int).Mul(requirements.MinDownloadSpeed, b.bidConfig.DownloadSpeedPrice)

	// Sum up all costs for base calculation
	totalCost := new(big.Int).Add(cpuCost, memCost)
	totalCost = new(big.Int).Add(totalCost, diskCost)
	totalCost = new(big.Int).Add(totalCost, gpuCost)
	totalCost = new(big.Int).Add(totalCost, uploadCost)
	totalCost = new(big.Int).Add(totalCost, downloadCost)

	// Apply machine type multiplier if specified
	// Default to 1.0 (no change) if machine type not found in config
	machineTypeId := requirements.MachineType.Int64()
	multiplier, exists := b.bidConfig.MachineTypeMultipliers[machineTypeId]
	if !exists {
		multiplier = big.NewInt(1) // Default multiplier if machine type is not configured
	}

	// Apply the multiplier to get the final cost
	if multiplier.Cmp(big.NewInt(1)) != 0 {
		// Convert to float for multiplication with multiplier, then back to big.Int
		costFloat := new(big.Float).SetInt(totalCost)
		costFloat.Mul(costFloat, new(big.Float).SetInt(multiplier))

		// Convert back to big.Int (truncating any fractional part)
		var finalCost big.Int
		costFloat.Int(&finalCost)
		totalCost = &finalCost
	}

	// Log the cost breakdown for debugging
	b.log.WithFields(logrus.Fields{
		"cpuCost":      cpuCost,
		"memCost":      memCost,
		"diskCost":     diskCost,
		"gpuCost":      gpuCost,
		"uploadCost":   uploadCost,
		"downloadCost": downloadCost,
		"machineType":  machineTypeId,
		"multiplier":   multiplier,
		"totalCost":    totalCost,
	}).Debug("Resource cost breakdown")

	return totalCost
}

// findSuitableMachine finds a suitable machine for the given requirements
func (b *BidEngine) findSuitableMachine(ctx context.Context, requirements *BidRequirements) *big.Int {
	// First get the provider ID from config or a previous initialization
	if b.providerId == nil {
		b.log.Error("Provider ID not set, cannot find suitable machine")
		return nil
	}

	// Get provider details to see how many machines there are
	provider, err := b.provider.GetProvider(&bind.CallOpts{Context: ctx}, b.providerId)
	if err != nil {
		b.log.WithError(err).Error("Failed to get provider details")
		return nil
	}

	// Check if the provider is active
	if !provider.IsActive {
		b.log.WithField("providerId", b.providerId).Warn("Provider is not active, cannot allocate machines")
		return nil
	}

	machineCount := provider.MachineCount.Int64()
	if machineCount == 0 {
		b.log.Warn("No machines found for provider")
		return nil
	}

	// Get all machines for our provider
	machines, err := b.provider.GetMachinesPaginated(
		&bind.CallOpts{Context: ctx},
		b.providerId,
		big.NewInt(0),            // Start index
		big.NewInt(machineCount), // End index (exclusive)
	)

	if err != nil {
		b.log.WithError(err).Error("Failed to get machines from provider contract")
		return nil
	}

	b.log.WithField("machineCount", len(machines)).Debug("Retrieved machines from provider contract")

	// Track already committed machines
	committedMachines := make(map[string]bool)
	b.bidsMutex.RLock()
	for _, bid := range b.activeBids {
		if bid.Status != BidStatusRejected && bid.Status != BidStatusExpired {
			committedMachines[bid.MachineId.String()] = true
		}
	}
	b.bidsMutex.RUnlock()

	// Define machine candidates with their scores
	type machineCandidate struct {
		id    *big.Int
		score float64
	}
	var candidates []machineCandidate

	// Find all machines that meet requirements
	for i, machine := range machines {
		// Skip inactive machines
		if !machine.Active {
			continue
		}

		machineId := big.NewInt(int64(i)) // Assuming machine IDs are sequential

		// Skip if machine is already committed
		if committedMachines[machineId.String()] {
			continue
		}

		// Check if machine meets requirements
		if machine.CpuCores.Cmp(requirements.MinCPUCores) < 0 ||
			machine.MemoryMB.Cmp(requirements.MinMemoryMB) < 0 ||
			machine.DiskGB.Cmp(requirements.MinDiskGB) < 0 ||
			machine.GpuCores.Cmp(requirements.MinGPUCores) < 0 ||
			machine.UploadSpeed.Cmp(requirements.MinUploadSpeed) < 0 ||
			machine.DownloadSpeed.Cmp(requirements.MinDownloadSpeed) < 0 ||
			machine.Region.Cmp(requirements.Region) != 0 ||
			machine.MachineType.Cmp(requirements.MachineType) != 0 {
			continue
		}

		// Calculate a score for this machine - lower is better
		// We aim for machines that just meet the requirements without wasting resources
		cpuRatio := float64(machine.CpuCores.Int64()) / float64(requirements.MinCPUCores.Int64())
		memRatio := float64(machine.MemoryMB.Int64()) / float64(requirements.MinMemoryMB.Int64())
		diskRatio := float64(machine.DiskGB.Int64()) / float64(requirements.MinDiskGB.Int64())

		// Calculate how much each resource exceeds requirements (0 is perfect)
		cpuExcess := cpuRatio - 1.0
		memExcess := memRatio - 1.0
		diskExcess := diskRatio - 1.0

		// Calculate a combined score - the smaller the better
		// We square the values to penalize larger excesses more
		score := cpuExcess*cpuExcess + memExcess*memExcess + diskExcess*diskExcess

		candidates = append(candidates, machineCandidate{
			id:    machineId,
			score: score,
		})
	}

	if len(candidates) == 0 {
		b.log.Warn("No suitable machines found for requirements")
		return nil
	}

	// Sort candidates by score (lower is better)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score < candidates[j].score
	})

	b.log.WithFields(logrus.Fields{
		"machineId":       candidates[0].id,
		"score":           candidates[0].score,
		"totalCandidates": len(candidates),
	}).Info("Selected suitable machine for bid")

	return candidates[0].id
}

// monitorActiveBids periodically checks the status of active bids
func (b *BidEngine) monitorActiveBids(ctx context.Context) {
	b.wg.Done()

	// Check bids more frequently (every 30 seconds) since they expire quickly
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.updateBidStatuses(ctx)
		case <-b.shutdown:
			return
		case <-ctx.Done():
			return
		}
	}
}

// updateBidStatuses updates the status of all active bids by checking their status on the blockchain
func (b *BidEngine) updateBidStatuses(ctx context.Context) {
	b.bidsMutex.Lock()
	defer b.bidsMutex.Unlock()

	for key, bid := range b.activeBids {
		// Skip checking status for already finalized bids
		if bid.Status == BidStatusExpired || bid.Status == BidStatusAccepted ||
			bid.Status == BidStatusRejected {
			// Remove expired or finalized bids after a certain time
			if time.Since(bid.UpdatedAt) > time.Hour {
				delete(b.activeBids, key)
				b.log.WithFields(logrus.Fields{
					"bidId":      bid.ID,
					"orderId":    bid.OrderId,
					"providerId": bid.ProviderId,
					"machineId":  bid.MachineId,
					"status":     bid.Status,
				}).Info("Removed finalized bid from active monitoring")
			}
			continue
		}

		// Check if bid has expired based on local time
		if time.Now().After(bid.ExpirationAt) {
			bid.Status = BidStatusExpired
			bid.UpdatedAt = time.Now()
			b.log.WithFields(logrus.Fields{
				"bidId":      bid.ID,
				"orderId":    bid.OrderId,
				"providerId": bid.ProviderId,
				"machineId":  bid.MachineId,
			}).Info("Bid expired after 5 minutes")
			continue
		}

		// Try to get bid status from blockchain
		// This is optional since we're using local expiration tracking
		// But it's good to check for acceptance/rejection
		// ...additional blockchain status check code if available...
	}
}
