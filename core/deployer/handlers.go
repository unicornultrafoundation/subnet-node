package k8scluster

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
	"go.uber.org/zap"
)

// handleDeploymentRequestedEvent handles deployment requested events
func (s *Service) handleDeploymentRequestedEvent(ctx context.Context, event *types.DeploymentRequestedEvent) error {
	s.logger.Info("Handling DeploymentRequestedEvent",
		zap.String("deploymentID", event.DeploymentID),
		zap.String("requester", event.Requester.Hex()),
		zap.String("sdlHash", event.SDLHash))

	// Submit bid to marketplace
	bid, err := s.bidManager.SubmitBid(ctx, event.DeploymentID, event.SDLHash)
	if err != nil {
		s.logger.Error("Failed to submit bid", zap.Error(err))
		return err
	}

	// Create and publish bid submitted event
	bidSubmittedEvent := &types.BidSubmittedEvent{
		BaseEvent: types.BaseEvent{
			Timestamp: time.Now(),
		},
		DeploymentID: event.DeploymentID,
		SDLHash:      event.SDLHash,
		Provider:     s.config.ProviderAddress, // Use provider address from config
		Requester:    event.Requester,
		Amount:       big.NewInt(int64(bid.Amount * 1e18)), // Convert float to big.Int with 18 decimals
		Duration:     time.Hour * 24,                       // Default duration of 24 hours
	}
	if err := s.eventBus.Publish(ctx, string(types.MarketplaceEventTypeBidSubmitted), bidSubmittedEvent); err != nil {
		s.logger.Error("Failed to publish bid submitted event", zap.Error(err))
	}

	return nil
}

// handleBidSubmittedEvent handles a bid submitted event
func (s *Service) handleBidSubmittedEvent(ctx context.Context, event *types.BidSubmittedEvent) error {
	s.logger.Info("Handling BidSubmittedEvent",
		zap.String("deploymentID", event.DeploymentID),
		zap.String("provider", event.Provider.Hex()),
		zap.String("amount", event.Amount.String()))

	return nil
}

// handleProviderSelectedEvent handles a provider selected event
func (s *Service) handleProviderSelectedEvent(ctx context.Context, event *types.ProviderSelectedEvent) error {
	s.logger.Info("Handling ProviderSelectedEvent",
		zap.String("deploymentID", event.DeploymentID),
		zap.String("provider", event.Provider.Hex()),
		zap.String("amount", event.Amount.String()))

	// Get the bid
	bid, err := s.bidTracker.GetBid(ctx, event.DeploymentID, event.Provider)
	if err != nil {
		return fmt.Errorf("failed to get bid: %w", err)
	}
	if bid == nil {
		s.logger.Error("Bid not found for provider", zap.String("deploymentID", event.DeploymentID), zap.String("provider", event.Provider.Hex()))
		return fmt.Errorf("bid not found for provider %s", event.Provider.Hex())
	}
	if s.ipfsClient == nil {
		s.logger.Error("ipfsClient is nil in handleProviderSelectedEvent")
		return fmt.Errorf("ipfsClient is nil")
	}
	if s.sdlParser == nil {
		s.logger.Error("sdlParser is nil in handleProviderSelectedEvent")
		return fmt.Errorf("sdlParser is nil")
	}

	// Get the manifest from IPFS
	manifestData, err := s.ipfsClient.Get(bid.SDLHash)
	if err != nil {
		return fmt.Errorf("failed to get manifest from IPFS: %w", err)
	}

	// Parse the manifest
	sdl, err := s.sdlParser.Parse(bytes.NewReader(manifestData))
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Create deployment
	if err := s.deploymentMgr.CreateDeployment(ctx, event.DeploymentID, bid.Requester, sdl); err != nil {
		s.logger.Error("Failed to create deployment", zap.Error(err))
		return err
	}

	return nil
}

// handleDeploymentCompletedEvent handles a deployment completed event
func (s *Service) handleDeploymentCompletedEvent(ctx context.Context, event *types.DeploymentCompletedEvent) error {
	s.logger.Info("Handling DeploymentCompletedEvent",
		zap.String("deploymentID", event.DeploymentID),
		zap.String("provider", event.Provider.Hex()),
		zap.String("status", string(event.Status)))

	// Get the deployment
	deployment, err := s.deploymentMgr.GetDeployment(event.DeploymentID)
	if err != nil {
		s.logger.Error("Failed to get deployment", zap.Error(err))
		return err
	}

	// Update deployment status
	deployment.Status = types.DeploymentStatusCompleted
	deployment.UpdatedAt = time.Now()

	// Update deployment
	if err := s.deploymentMgr.UpdateDeployment(ctx, deployment); err != nil {
		s.logger.Error("Failed to update deployment", zap.Error(err))
		return err
	}

	return nil
}

// handleDeploymentTerminatedEvent handles a deployment terminated event
func (s *Service) handleDeploymentTerminatedEvent(ctx context.Context, event *types.DeploymentTerminatedEvent) error {
	s.logger.Info("Handling DeploymentTerminatedEvent",
		zap.String("deploymentID", event.DeploymentID),
		zap.String("provider", event.Provider.Hex()),
		zap.String("requester", event.Requester.Hex()))

	// Get the deployment
	deployment, err := s.deploymentMgr.GetDeployment(event.DeploymentID)
	if err != nil {
		s.logger.Error("Failed to get deployment", zap.Error(err))
		return err
	}

	// Update deployment status
	deployment.Status = types.DeploymentStatusTerminated
	deployment.UpdatedAt = time.Now()

	// Update deployment
	if err := s.deploymentMgr.UpdateDeployment(ctx, deployment); err != nil {
		s.logger.Error("Failed to update deployment", zap.Error(err))
		return err
	}

	return nil
}
