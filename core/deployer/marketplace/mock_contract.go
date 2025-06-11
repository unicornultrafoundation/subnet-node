package marketplace

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"

	clusterTypes "github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

// MockMarketplaceContract is a mock implementation of the marketplace contract
type MockMarketplaceContract struct {
	logger *zap.Logger
}

// NewMockMarketplaceContract creates a new mock marketplace contract
func NewMockMarketplaceContract(logger *zap.Logger) *MockMarketplaceContract {
	return &MockMarketplaceContract{
		logger: logger,
	}
}

// SubmitBid submits a bid to the marketplace
func (c *MockMarketplaceContract) SubmitBid(ctx context.Context, auth *bind.TransactOpts, deploymentID string, amount *big.Int, duration *big.Int) (*types.Transaction, error) {
	c.logger.Info("Mock: Submitting bid",
		zap.String("deploymentID", deploymentID),
		zap.String("amount", amount.String()),
		zap.String("duration", duration.String()))

	// Create a mock transaction with a hash
	tx := types.NewTransaction(
		0, // nonce
		common.HexToAddress("0x0000000000000000000000000000000000000000"), // to
		big.NewInt(0), // value
		0,             // gas
		big.NewInt(0), // gasPrice
		[]byte{},      // data
	)
	return tx, nil
}

// SubscribeToDeploymentRequested subscribes to deployment requested events
func (c *MockMarketplaceContract) SubscribeToDeploymentRequested(ctx context.Context, handler func(*clusterTypes.DeploymentRequestedEvent) error) error {
	c.logger.Info("Mock: Subscribing to deployment requested events")
	return nil
}

// SubscribeToBidSubmitted subscribes to bid submitted events
func (c *MockMarketplaceContract) SubscribeToBidSubmitted(ctx context.Context, handler func(*clusterTypes.BidSubmittedEvent) error) error {
	c.logger.Info("Mock: Subscribing to bid submitted events")
	return nil
}

// SubscribeToProviderSelected subscribes to provider selected events
func (c *MockMarketplaceContract) SubscribeToProviderSelected(ctx context.Context, handler func(*clusterTypes.ProviderSelectedEvent) error) error {
	c.logger.Info("Mock: Subscribing to provider selected events")
	return nil
}

// SubscribeToDeploymentCompleted subscribes to deployment completed events
func (c *MockMarketplaceContract) SubscribeToDeploymentCompleted(ctx context.Context, handler func(*clusterTypes.DeploymentCompletedEvent) error) error {
	c.logger.Info("Mock: Subscribing to deployment completed events")
	return nil
}

// SubscribeToDeploymentTerminated subscribes to deployment terminated events
func (c *MockMarketplaceContract) SubscribeToDeploymentTerminated(ctx context.Context, handler func(*clusterTypes.DeploymentTerminatedEvent) error) error {
	c.logger.Info("Mock: Subscribing to deployment terminated events")
	return nil
}

// GetDeployment gets a deployment by ID
func (c *MockMarketplaceContract) GetDeployment(ctx context.Context, deploymentID string) (*clusterTypes.Deployment, error) {
	c.logger.Info("Mock: Getting deployment", zap.String("deploymentID", deploymentID))
	return &clusterTypes.Deployment{
		ID:        deploymentID,
		Status:    clusterTypes.DeploymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// GetBid gets a bid by deployment ID and provider
func (c *MockMarketplaceContract) GetBid(ctx context.Context, deploymentID string, provider common.Address) (*clusterTypes.BidSubmittedEvent, error) {
	c.logger.Info("Mock: Getting bid",
		zap.String("deploymentID", deploymentID),
		zap.String("provider", provider.Hex()))
	return &clusterTypes.BidSubmittedEvent{
		BaseEvent: clusterTypes.BaseEvent{
			Timestamp: time.Now(),
		},
		DeploymentID: deploymentID,
		Provider:     provider,
		Amount:       big.NewInt(0),
		Duration:     time.Hour * 24,
		Status:       clusterTypes.BidStatusActive,
	}, nil
}

// GetBids gets all bids for a deployment
func (c *MockMarketplaceContract) GetBids(ctx context.Context, deploymentID string) ([]*clusterTypes.BidSubmittedEvent, error) {
	c.logger.Info("Mock: Getting bids", zap.String("deploymentID", deploymentID))
	return []*clusterTypes.BidSubmittedEvent{}, nil
}

// GetProvider gets a provider by address
func (c *MockMarketplaceContract) GetProvider(ctx context.Context, provider common.Address) (*clusterTypes.Provider, error) {
	c.logger.Info("Mock: Getting provider", zap.String("provider", provider.Hex()))
	return &clusterTypes.Provider{
		Address: provider,
		Status:  clusterTypes.ProviderStatusActive,
	}, nil
}

// GetProviders gets all providers
func (c *MockMarketplaceContract) GetProviders(ctx context.Context) ([]*clusterTypes.Provider, error) {
	c.logger.Info("Mock: Getting providers")
	return []*clusterTypes.Provider{}, nil
}

// GetRequester gets a requester by address
func (c *MockMarketplaceContract) GetRequester(ctx context.Context, requester common.Address) (*clusterTypes.Requester, error) {
	c.logger.Info("Mock: Getting requester", zap.String("requester", requester.Hex()))
	return &clusterTypes.Requester{
		Address: requester,
		Status:  clusterTypes.RequesterStatusActive,
	}, nil
}

// GetRequesters gets all requesters
func (c *MockMarketplaceContract) GetRequesters(ctx context.Context) ([]*clusterTypes.Requester, error) {
	c.logger.Info("Mock: Getting requesters")
	return []*clusterTypes.Requester{}, nil
}

// ParsePaymentReceived is a stub for the mock contract
func (c *MockMarketplaceContract) ParsePaymentReceived(log types.Log) (*clusterTypes.PaymentReceivedEvent, error) {
	c.logger.Info("Mock: ParsePaymentReceived")
	return &clusterTypes.PaymentReceivedEvent{}, nil
}

// CompleteDeployment is a stub for the mock contract
func (c *MockMarketplaceContract) CompleteDeployment(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*types.Transaction, error) {
	c.logger.Info("Mock: CompleteDeployment", zap.String("deploymentID", deploymentID))
	return &types.Transaction{}, nil
}

// TerminateDeployment is a stub for the mock contract
func (c *MockMarketplaceContract) TerminateDeployment(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*types.Transaction, error) {
	c.logger.Info("Mock: TerminateDeployment", zap.String("deploymentID", deploymentID))
	return &types.Transaction{}, nil
}

// LockPayment is a stub for the mock contract
func (c *MockMarketplaceContract) LockPayment(ctx context.Context, opts *bind.TransactOpts, deploymentID string, amount *big.Int, provider common.Address) (*types.Transaction, error) {
	c.logger.Info("Mock: LockPayment", zap.String("deploymentID", deploymentID))
	return &types.Transaction{}, nil
}

// ReleasePayment is a stub for the mock contract
func (c *MockMarketplaceContract) ReleasePayment(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*types.Transaction, error) {
	c.logger.Info("Mock: ReleasePayment", zap.String("deploymentID", deploymentID))
	return &types.Transaction{}, nil
}

// RefundPayment is a stub for the mock contract
func (c *MockMarketplaceContract) RefundPayment(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*types.Transaction, error) {
	c.logger.Info("Mock: RefundPayment", zap.String("deploymentID", deploymentID))
	return &types.Transaction{}, nil
}

// InitiateDispute is a stub for the mock contract
func (c *MockMarketplaceContract) InitiateDispute(ctx context.Context, opts *bind.TransactOpts, deploymentID string, reason string) (*types.Transaction, error) {
	c.logger.Info("Mock: InitiateDispute", zap.String("deploymentID", deploymentID), zap.String("reason", reason))
	return &types.Transaction{}, nil
}

// ReleaseEscrow is a stub for the mock contract
func (c *MockMarketplaceContract) ReleaseEscrow(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*types.Transaction, error) {
	c.logger.Info("Mock: ReleaseEscrow", zap.String("deploymentID", deploymentID))
	return &types.Transaction{}, nil
}

// RefundEscrow is a stub for the mock contract
func (c *MockMarketplaceContract) RefundEscrow(ctx context.Context, opts *bind.TransactOpts, deploymentID string) (*types.Transaction, error) {
	c.logger.Info("Mock: RefundEscrow", zap.String("deploymentID", deploymentID))
	return &types.Transaction{}, nil
}

// RequestDeployment is a stub for the mock contract
func (c *MockMarketplaceContract) RequestDeployment(ctx context.Context, opts *bind.TransactOpts, deploymentID string, sdlHash string, minBid *big.Int, maxBid *big.Int, duration *big.Int) (*types.Transaction, error) {
	c.logger.Info("Mock: RequestDeployment", zap.String("deploymentID", deploymentID))
	return &types.Transaction{}, nil
}

// SelectProvider is a stub for the mock contract
func (c *MockMarketplaceContract) SelectProvider(ctx context.Context, opts *bind.TransactOpts, deploymentID string, provider common.Address, bidIndex *big.Int) (*types.Transaction, error) {
	c.logger.Info("Mock: SelectProvider", zap.String("deploymentID", deploymentID), zap.String("provider", provider.Hex()))
	return &types.Transaction{}, nil
}

// GetDeploymentRequestedEvents is a stub for the mock contract
func (c *MockMarketplaceContract) GetDeploymentRequestedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.DeploymentRequestedEvent, error) {
	c.logger.Info("Mock: GetDeploymentRequestedEvents")
	return []*clusterTypes.DeploymentRequestedEvent{}, nil
}

// GetBidSubmittedEvents is a stub for the mock contract
func (c *MockMarketplaceContract) GetBidSubmittedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.BidSubmittedEvent, error) {
	c.logger.Info("Mock: GetBidSubmittedEvents")
	return []*clusterTypes.BidSubmittedEvent{}, nil
}

// GetProviderSelectedEvents is a stub for the mock contract
func (c *MockMarketplaceContract) GetProviderSelectedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.ProviderSelectedEvent, error) {
	c.logger.Info("Mock: GetProviderSelectedEvents")
	return []*clusterTypes.ProviderSelectedEvent{}, nil
}

// GetDeploymentCompletedEvents is a stub for the mock contract
func (c *MockMarketplaceContract) GetDeploymentCompletedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.DeploymentCompletedEvent, error) {
	c.logger.Info("Mock: GetDeploymentCompletedEvents")
	return []*clusterTypes.DeploymentCompletedEvent{}, nil
}

// GetDeploymentTerminatedEvents is a stub for the mock contract
func (c *MockMarketplaceContract) GetDeploymentTerminatedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.DeploymentTerminatedEvent, error) {
	c.logger.Info("Mock: GetDeploymentTerminatedEvents")
	return []*clusterTypes.DeploymentTerminatedEvent{}, nil
}
