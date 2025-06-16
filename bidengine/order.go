package bidengine

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
)

// watchNewOrders subscribes to new order events on the blockchain and processes them
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
			Region:           orderDetails.Region,
			MachineType:      orderDetails.MachineType,
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
