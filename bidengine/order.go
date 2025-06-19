package bidengine

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// watchNewOrders subscribes to new order events on the blockchain and processes them
func (b *Service) watchNewOrders(ctx context.Context) error {
	// Create a filter for new orders
	orderCreatedCh := make(chan *contracts.BidMarketOrderCreated)
	orderSub, err := b.bidMarket.WatchOrderCreated(&bind.WatchOpts{
		Context: ctx,
	}, orderCreatedCh) // You might want to add filters if available

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
				}, orderCreatedCh)

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

// getOrder retrieves order details from the blockchain and converts to internal format
func (b *Service) getOrder(ctx context.Context, orderId *uint256.Int) (*types.OrderInfo, error) {
	// Fetch complete order details from the blockchain
	return b.bidMarket.GetOrder(&bind.CallOpts{
		Context: ctx,
	}, orderId)
}

// processNewOrder handles a new order event from the blockchain
func (b *Service) processNewOrder(ctx context.Context, event *contracts.BidMarketOrderCreated) {
	b.log.WithFields(logrus.Fields{
		"orderId":   event.OrderId,
		"requester": event.Owner.String(),
	}).Info("New order detected on blockchain")

	// Get order details using the new function
	order, err := b.getOrder(ctx, uint256.MustFromBig(event.OrderId))
	if err != nil {
		b.log.WithError(err).WithField("orderId", event.OrderId).Error("Failed to get order details")
		return
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

	// Find a suitable machine
	machine := b.findSuitableMachine(ctx, order.Requirements)
	if machine == nil {
		b.log.WithField("orderId", order.OrderID).Warn("No suitable machine found for order")
		return
	}

	bidAmount := b.calculateBidAmount(order, machine)

	// Log our intention
	b.log.WithFields(logrus.Fields{
		"orderId":   order.OrderID,
		"maxPrice":  order.MaxPrice,
		"bidAmount": bidAmount,
	}).Info("Placing bid on new order")

	// Place the bid
	_, err = b.PlaceBid(ctx, order, b.providerId, machine.ID, bidAmount, order.Requirements)
	if err != nil {
		b.log.WithError(err).WithField("orderId", order.OrderID).Error("Failed to place bid")
		return
	}
}

// checkOrderEligibility checks if we can and should bid on this order
func (b *Service) checkOrderEligibility(_ context.Context, order *types.OrderInfo) (bool, error) {
	// 1. Check if requirements meet our minimum criteria
	if order.Requirements.MinCPUCores.Cmp(b.bidConfig.MinRequirements.MinCPUCores) < 0 ||
		order.Requirements.MinMemoryMB.Cmp(b.bidConfig.MinRequirements.MinMemoryMB) < 0 ||
		order.Requirements.MinDiskGB.Cmp(b.bidConfig.MinRequirements.MinDiskGB) < 0 {
		return false, nil
	}

	// 3. Check if we have available resources to fulfill this order
	// In a real implementation, you'd check against your available machines

	return true, nil
}

// calculateBidAmount calculates the amount to bid based on the order and our config
func (b *Service) calculateBidAmount(order *types.OrderInfo, machine *types.Machine) *uint256.Int {
	// Calculate a bid amount between min and max percentages of the max price
	minAmount := new(uint256.Int).Mul(order.MaxPrice, uint256.NewInt(uint64(b.bidConfig.MinBidPercent)))
	minAmount = minAmount.Div(minAmount, uint256.NewInt(100))

	maxAmount := new(uint256.Int).Mul(order.MaxPrice, uint256.NewInt(uint64(b.bidConfig.MaxBidPercent)))
	maxAmount = maxAmount.Div(maxAmount, uint256.NewInt(100))

	// Choose a bid amount between min and max, potentially based on competition
	// For simplicity, we'll just use the min amount for now
	bidAmount := new(uint256.Int).Set(minAmount)

	// Ensure our bid is at least our cost estimate plus minimum profit
	costEstimate := b.calculateMachineCost(order.Requirements, machine, 1)
	minProfitableAmount := new(uint256.Int).Add(costEstimate, uint256.NewInt(uint64(b.cfg.GetInt("bidengine.min_profit", 1000000000))))

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

// GetOrder gets order details by order ID
func (b *Service) GetOrder(ctx context.Context, orderId *uint256.Int) (*types.OrderInfo, error) {
	order, err := b.getOrder(ctx, orderId)
	if err != nil {
		b.log.WithError(err).WithField("orderId", orderId).Error("Failed to get order details")
		return nil, err
	}

	b.log.WithFields(logrus.Fields{
		"orderId":   orderId,
		"requester": order.RequesterID.String(),
		"cpuCores":  order.Requirements.MinCPUCores,
		"memoryMB":  order.Requirements.MinMemoryMB,
		"diskGB":    order.Requirements.MinDiskGB,
		"gpuCores":  order.Requirements.MinGPUCores,
		"maxPrice":  order.MaxPrice,
		"minPrice":  order.MinPrice,
		"expiresAt": order.ExpirationAt,
		"status":    order.Status,
	}).Debug("Retrieved order details")

	return order, nil
}
