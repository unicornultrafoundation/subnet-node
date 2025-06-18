package bidengine

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
)

// PlaceBid places a bid on a specific order in the BidMarket contract
func (b *Service) PlaceBid(ctx context.Context, order *OrderInfo, providerId, machineId, pricePerSecond *uint256.Int, requirements *BidRequirements) (*Bid, error) {
	// Check if machine is active
	isActive, err := b.provider.IsMachineActive(&bind.CallOpts{Context: ctx}, providerId.ToBig(), machineId.ToBig())
	if err != nil {
		return nil, fmt.Errorf("failed to check if machine is active: %v", err)
	}

	if !isActive {
		return nil, fmt.Errorf("machine %v is not active", machineId)
	}

	// Validate machine specs against requirements
	valid, err := b.provider.ValidateMachineRequirements(
		&bind.CallOpts{Context: ctx},
		order.Requirements.MachineType.ToBig(), // machineType
		providerId.ToBig(),
		machineId.ToBig(),
		requirements.MinCPUCores.ToBig(),
		requirements.MinMemoryMB.ToBig(),
		requirements.MinDiskGB.ToBig(),
		requirements.MinGPUCores.ToBig(),
		requirements.MinUploadSpeed.ToBig(),
		requirements.MinDownloadSpeed.ToBig(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to validate machine requirements: %v", err)
	}

	if !valid {
		return nil, fmt.Errorf("machine %v does not meet requirements", machineId)
	}

	// Place bid on the blockchain
	tx, err := b.bidMarket.SubmitBid(b.auth, order.OrderID, providerId, machineId, pricePerSecond)
	if err != nil {
		return nil, fmt.Errorf("failed to place bid on blockchain: %v", err)
	}

	bid := &Bid{
		ID:           order.OrderID, // Using order ID as bid ID for tracking
		OrderId:      order.OrderID,
		ProviderId:   providerId,
		MachineId:    machineId,
		PricePerSec:  pricePerSecond,
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

	// Save bid to datastore
	if err := b.store.SaveBid(ctx, bid); err != nil {
		b.log.WithError(err).WithField("bidId", bid.ID).Error("Failed to save bid to datastore")
		// Continue even if datastore save fails
	} else {
		b.log.WithField("bidId", bid.ID).Debug("Bid saved to datastore")
	}

	b.log.WithFields(logrus.Fields{
		"orderId":        order.OrderID,
		"providerId":     providerId,
		"machineId":      machineId,
		"pricePerSecond": pricePerSecond.String(),
		"txHash":         bid.TxHash,
	}).Info("Bid placed successfully on order")

	return bid, nil
}

// monitorActiveBids periodically checks the status of active bids
func (b *Service) monitorActiveBids(ctx context.Context) {
	// Start watching for bid acceptances
	go b.watchBidAcceptances(ctx)

	// Check for non-accepted bids every 5 minutes and clean them from memory
	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()

	b.log.Info("Started monitoring active bids")

	for {
		select {
		case <-cleanupTicker.C:
			b.cleanupNonAcceptedBids()
		case <-b.shutdown:
			b.log.Info("Stopping bid monitoring")
			return
		case <-ctx.Done():
			b.log.Info("Context canceled, stopping bid monitoring")
			return
		}
	}
}

// cleanupNonAcceptedBids removes bids that are not in accepted status from memory
func (b *Service) cleanupNonAcceptedBids() {
	b.bidsMutex.Lock()
	defer b.bidsMutex.Unlock()

	removed := 0
	for key, bid := range b.activeBids {
		// Keep only accepted bids in memory
		if bid.Status != BidStatusAccepted {
			delete(b.activeBids, key)
			removed++

			b.log.WithFields(logrus.Fields{
				"bidId":   bid.ID,
				"orderId": bid.OrderId,
				"status":  bid.Status,
			}).Debug("Removed non-accepted bid from cache")
		}
	}

	if removed > 0 {
		b.log.WithField("count", removed).Info("Cleaned up non-accepted bids from memory cache")
	}
}

// watchBidAcceptances watches for bid acceptance events on the blockchain
func (b *Service) watchBidAcceptances(ctx context.Context) {
	bidAcceptedCh := make(chan *contracts.BidMarketBidAccepted)
	acceptSub, err := b.bidMarket.WatchBidAccepted(&bind.WatchOpts{
		Context: ctx,
	}, bidAcceptedCh, b.providerId) // You might want to add filters for our provider ID

	if err != nil {
		b.log.WithError(err).Error("Failed to watch for bid acceptance events")
		return
	}

	b.log.Info("Started watching for bid acceptance events")
	defer acceptSub.Unsubscribe()

	for {
		select {
		case err := <-acceptSub.Err():
			b.log.WithError(err).Error("Bid acceptance subscription error")

			// Try to resubscribe after a delay
			time.Sleep(time.Second * 10)
			newSub, err := b.bidMarket.WatchBidAccepted(&bind.WatchOpts{
				Context: ctx,
			}, bidAcceptedCh, b.providerId)

			if err != nil {
				b.log.WithError(err).Error("Failed to resubscribe to bid acceptance events")
				continue
			}
			acceptSub = newSub

		case event := <-bidAcceptedCh:
			// Process bid acceptance
			b.processBidAcceptance(ctx, event)

		case <-b.shutdown:
			b.log.Info("Stopping bid acceptance watcher")
			return

		case <-ctx.Done():
			b.log.Info("Context canceled, stopping bid acceptance watcher")
			return
		}
	}
}

// processBidAcceptance handles a bid acceptance event from the blockchain
func (b *Service) processBidAcceptance(ctx context.Context, event *contracts.BidMarketBidAccepted) {
	b.log.WithFields(logrus.Fields{
		"orderId":    event.OrderId,
		"providerId": event.ProviderId,
		"machineId":  event.MachineId,
		"amount":     event.PricePerSecond.String(),
		"timestamp":  time.Now().Format(time.RFC3339),
	}).Info("Bid accepted by requester")

	// Find the bid in our active bids or from the store
	bidKey := event.OrderId.String()

	// First try to get from active bids (in memory)
	b.bidsMutex.RLock()
	bid, exists := b.activeBids[bidKey]
	b.bidsMutex.RUnlock()

	// If not found in memory, try to load from store
	if !exists {
		var err error
		bid, err = b.store.GetBidByID(ctx, bidKey)
		if err != nil {
			b.log.WithError(err).WithField("orderId", event.OrderId).Error("Failed to retrieve bid from store")
			return
		}

		if bid == nil {
			b.log.WithField("orderId", event.OrderId).Warn("Received acceptance for unknown bid (not found in memory or store)")
			return
		}

		// Add to active bids since we found it in the store
		b.bidsMutex.Lock()
		b.activeBids[bidKey] = bid
		b.bidsMutex.Unlock()

		b.log.WithField("orderId", event.OrderId).Info("Restored bid from datastore for processing acceptance")
	}

	order, err := b.getOrder(ctx, uint256.MustFromBig(event.OrderId))
	if err != nil {
		b.log.WithError(err).Error("Failed to retrieve order information")
		return
	}

	// Validate machine ID matches
	if bid.MachineId.Cmp(order.AcceptedMachineId) != 0 {
		b.log.WithFields(logrus.Fields{
			"expectedMachine": bid.MachineId,
			"actualMachine":   order.AcceptedMachineId,
			"orderId":         event.OrderId,
		}).Warn("Machine ID mismatch in bid acceptance")
		return
	}

	// Acquire write lock to update bid status
	b.bidsMutex.Lock()
	// Update bid status
	bid.Status = BidStatusAccepted
	bid.UpdatedAt = time.Now()
	b.bidsMutex.Unlock()

	// Save updated status to datastore
	if err := b.store.SaveBid(ctx, bid); err != nil {
		b.log.WithError(err).WithField("bidId", bid.ID).Error("Failed to update accepted bid in datastore")
	}

	// Trigger any post-acceptance actions (e.g., prepare resources, notify system)
	go b.handleAcceptedBid(ctx, bid)
}

// handleAcceptedBid performs necessary actions after a bid is accepted
func (b *Service) handleAcceptedBid(_ context.Context, bid *Bid) {
	// Implement resource provisioning, notification to resource manager, etc.
	b.log.WithFields(logrus.Fields{
		"orderId":    bid.OrderId,
		"providerId": bid.ProviderId,
		"machineId":  bid.MachineId,
	}).Info("Processing accepted bid")

	// TODO: Implement resource provisioning logic
	// This could include:
	// 1. Notify resource manager to prepare the machine
	// 2. Set up the environment according to requirements
	// 3. Start monitoring for service-related events
}
