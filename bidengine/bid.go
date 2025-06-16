package bidengine

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/sirupsen/logrus"
)

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
		OrderId:      order.OrderID,
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

	// Save bid to datastore
	if b.store != nil {
		if err := b.store.SaveBid(ctx, bid); err != nil {
			b.log.WithError(err).WithField("bidId", bid.ID).Error("Failed to save bid to datastore")
			// Continue even if datastore save fails
		} else {
			b.log.WithField("bidId", bid.ID).Debug("Bid saved to datastore")
		}
	}

	b.log.WithFields(logrus.Fields{
		"orderId":    order.OrderID,
		"providerId": providerId,
		"machineId":  machineId,
		"amount":     amount.String(),
		"txHash":     bid.TxHash,
	}).Info("Bid placed successfully on order")

	return bid, nil
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
				// Delete from datastore before removing from memory
				if b.store != nil {
					if err := b.store.DeleteBid(ctx, key); err != nil {
						b.log.WithError(err).WithField("bidId", bid.ID).Warn("Failed to delete bid from datastore")
					}
				}

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

			// Save updated status to datastore
			if b.store != nil {
				if err := b.store.SaveBid(ctx, bid); err != nil {
					b.log.WithError(err).WithField("bidId", bid.ID).Error("Failed to update bid status in datastore")
				}
			}

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
