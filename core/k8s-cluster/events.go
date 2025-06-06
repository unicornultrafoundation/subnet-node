package k8scluster

import (
	"context"
	"fmt"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// Subscribe subscribes to events
func (s *Service) Subscribe() error {
	// Subscribe to deployment requested events
	if err := s.eventBus.Subscribe(string(types.MarketplaceEventTypeDeploymentRequested), func(ctx context.Context, event types.MarketplaceEvent) error {
		if e, ok := event.(*types.DeploymentRequestedEvent); ok {
			return s.handleDeploymentRequestedEvent(ctx, e)
		}
		return fmt.Errorf("invalid event type for deployment requested event")
	}); err != nil {
		return fmt.Errorf("failed to subscribe to deployment requested events: %w", err)
	}

	// Subscribe to bid submitted events
	if err := s.eventBus.Subscribe(string(types.MarketplaceEventTypeBidSubmitted), func(ctx context.Context, event types.MarketplaceEvent) error {
		if e, ok := event.(*types.BidSubmittedEvent); ok {
			return s.handleBidSubmittedEvent(ctx, e)
		}
		return fmt.Errorf("invalid event type for bid submitted event")
	}); err != nil {
		return fmt.Errorf("failed to subscribe to bid submitted events: %w", err)
	}

	// Subscribe to provider selected events
	if err := s.eventBus.Subscribe(string(types.MarketplaceEventTypeProviderSelected), func(ctx context.Context, event types.MarketplaceEvent) error {
		if e, ok := event.(*types.ProviderSelectedEvent); ok {
			return s.handleProviderSelectedEvent(ctx, e)
		}
		return fmt.Errorf("invalid event type for provider selected event")
	}); err != nil {
		return fmt.Errorf("failed to subscribe to provider selected events: %w", err)
	}

	// Subscribe to deployment approved events
	if err := s.eventBus.Subscribe(string(types.MarketplaceEventTypeDeploymentApproved), func(ctx context.Context, event types.MarketplaceEvent) error {
		if e, ok := event.(*types.DeploymentApprovedEvent); ok {
			return s.handleDeploymentApprovedEvent(ctx, e)
		}
		return fmt.Errorf("invalid event type for deployment approved event")
	}); err != nil {
		return fmt.Errorf("failed to subscribe to deployment approved events: %w", err)
	}

	// Subscribe to deployment completed events
	if err := s.eventBus.Subscribe(string(types.MarketplaceEventTypeDeploymentCompleted), func(ctx context.Context, event types.MarketplaceEvent) error {
		if e, ok := event.(*types.DeploymentCompletedEvent); ok {
			return s.handleDeploymentCompletedEvent(ctx, e)
		}
		return fmt.Errorf("invalid event type for deployment completed event")
	}); err != nil {
		return fmt.Errorf("failed to subscribe to deployment completed events: %w", err)
	}

	// Subscribe to deployment terminated events
	if err := s.eventBus.Subscribe(string(types.MarketplaceEventTypeDeploymentTerminated), func(ctx context.Context, event types.MarketplaceEvent) error {
		if e, ok := event.(*types.DeploymentTerminatedEvent); ok {
			return s.handleDeploymentTerminatedEvent(ctx, e)
		}
		return fmt.Errorf("invalid event type for deployment terminated event")
	}); err != nil {
		return fmt.Errorf("failed to subscribe to deployment terminated events: %w", err)
	}

	return nil
}
