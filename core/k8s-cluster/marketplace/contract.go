package marketplace

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
	clusterTypes "github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// Ensure MarketplaceContract implements types.ContractInterface
var _ types.ContractInterface = (*MarketplaceContract)(nil)

// ContractEvent represents a contract event
type ContractEvent interface {
	GetDeploymentID() string
	GetRequester() common.Address
	GetProvider() common.Address
	GetAmount() *big.Int
	GetDuration() uint64
}

// DeploymentRequestedEvent represents a deployment requested event
type DeploymentRequestedEvent struct {
	DeploymentID string
	Requester    common.Address
	SDLHash      string
}

func (e *DeploymentRequestedEvent) GetDeploymentID() string      { return e.DeploymentID }
func (e *DeploymentRequestedEvent) GetRequester() common.Address { return e.Requester }
func (e *DeploymentRequestedEvent) GetProvider() common.Address  { return common.Address{} }
func (e *DeploymentRequestedEvent) GetAmount() *big.Int          { return big.NewInt(0) }
func (e *DeploymentRequestedEvent) GetDuration() uint64          { return 0 }

// BidSubmittedEvent represents a bid submitted event
type BidSubmittedEvent struct {
	DeploymentID string
	Provider     common.Address
	Amount       *big.Int
	Duration     uint64
}

func (e *BidSubmittedEvent) GetDeploymentID() string      { return e.DeploymentID }
func (e *BidSubmittedEvent) GetRequester() common.Address { return common.Address{} }
func (e *BidSubmittedEvent) GetProvider() common.Address  { return e.Provider }
func (e *BidSubmittedEvent) GetAmount() *big.Int          { return e.Amount }
func (e *BidSubmittedEvent) GetDuration() uint64          { return e.Duration }

// ProviderSelectedEvent represents a provider selected event
type ProviderSelectedEvent struct {
	DeploymentID string
	Provider     common.Address
	Amount       *big.Int
}

func (e *ProviderSelectedEvent) GetDeploymentID() string      { return e.DeploymentID }
func (e *ProviderSelectedEvent) GetRequester() common.Address { return common.Address{} }
func (e *ProviderSelectedEvent) GetProvider() common.Address  { return e.Provider }
func (e *ProviderSelectedEvent) GetAmount() *big.Int          { return e.Amount }
func (e *ProviderSelectedEvent) GetDuration() uint64          { return 0 }

// DeploymentCompletedEvent represents a deployment completed event
type DeploymentCompletedEvent struct {
	DeploymentID string
	Provider     common.Address
}

func (e *DeploymentCompletedEvent) GetDeploymentID() string      { return e.DeploymentID }
func (e *DeploymentCompletedEvent) GetRequester() common.Address { return common.Address{} }
func (e *DeploymentCompletedEvent) GetProvider() common.Address  { return e.Provider }
func (e *DeploymentCompletedEvent) GetAmount() *big.Int          { return big.NewInt(0) }
func (e *DeploymentCompletedEvent) GetDuration() uint64          { return 0 }

// DeploymentTerminatedEvent represents a deployment terminated event
type DeploymentTerminatedEvent struct {
	DeploymentID string
	Provider     common.Address
	Requester    common.Address
}

func (e *DeploymentTerminatedEvent) GetDeploymentID() string      { return e.DeploymentID }
func (e *DeploymentTerminatedEvent) GetRequester() common.Address { return e.Requester }
func (e *DeploymentTerminatedEvent) GetProvider() common.Address  { return e.Provider }
func (e *DeploymentTerminatedEvent) GetAmount() *big.Int          { return big.NewInt(0) }
func (e *DeploymentTerminatedEvent) GetDuration() uint64          { return 0 }

// MarketplaceContract represents the marketplace smart contract
type MarketplaceContract struct {
	client       *ethclient.Client
	contract     *bind.BoundContract
	contractAddr common.Address
	abi          abi.ABI
	bidTracker   types.BidTrackerInterface
	paymentMgr   types.PaymentManagerInterface
	logger       *zap.Logger
}

// NewMarketplaceContract creates a new marketplace contract instance
func NewMarketplaceContract(
	client *ethclient.Client,
	contractAddr common.Address,
	bidTracker types.BidTrackerInterface,
	paymentMgr types.PaymentManagerInterface,
	logger *zap.Logger,
) (*MarketplaceContract, error) {
	if client == nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "client cannot be nil",
		}
	}
	if bidTracker == nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "bidTracker cannot be nil",
		}
	}
	if paymentMgr == nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "paymentMgr cannot be nil",
		}
	}
	if logger == nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "logger cannot be nil",
		}
	}

	// Load contract ABI
	abi, err := loadContractABI()
	if err != nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "failed to load contract ABI",
			Err:     err,
		}
	}

	// Create bound contract
	contract := bind.NewBoundContract(contractAddr, abi, client, client, client)

	return &MarketplaceContract{
		client:       client,
		contract:     contract,
		contractAddr: contractAddr,
		abi:          abi,
		bidTracker:   bidTracker,
		paymentMgr:   paymentMgr,
		logger:       logger,
	}, nil
}

// loadContractABI loads the contract ABI from file
func loadContractABI() (abi.ABI, error) {
	// Load ABI from the compiled contract
	abiPath := filepath.Join("core", "k8s-cluster", "marketplace", "abi", "marketplace.json")
	abiData, err := ioutil.ReadFile(abiPath)
	if err != nil {
		return abi.ABI{}, errors.Wrap(err, "failed to read ABI file")
	}

	var abiJSON struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := json.Unmarshal(abiData, &abiJSON); err != nil {
		return abi.ABI{}, errors.Wrap(err, "failed to parse ABI JSON")
	}

	return abi.JSON(strings.NewReader(string(abiJSON.ABI)))
}

// SubscribeToDeploymentRequested subscribes to deployment requested events
func (c *MarketplaceContract) SubscribeToDeploymentRequested(ctx context.Context, handler func(*types.DeploymentRequestedEvent) error) error {
	event := c.abi.Events["DeploymentRequested"]
	if event.ID == (common.Hash{}) {
		return &types.ContractError{
			Code:    types.ErrCodeEventNotFound,
			Message: "DeploymentRequested event not found in contract ABI",
		}
	}

	topics := [][]common.Hash{{event.ID}}
	logs := make(chan ethtypes.Log)
	sub, err := c.client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	}, logs)
	if err != nil {
		return &types.ContractError{
			Code:    types.ErrCodeTransactionFailed,
			Message: "failed to subscribe to logs",
			Err:     err,
		}
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case err := <-sub.Err():
				if err != nil {
					c.logger.Error("subscription error", zap.Error(err))
					return
				}
			case log := <-logs:
				event := new(DeploymentRequestedEvent)
				if err := c.contract.UnpackLog(event, "DeploymentRequested", log); err != nil {
					c.logger.Error("failed to unpack log", zap.Error(err))
					continue
				}
				typesEvent := &types.DeploymentRequestedEvent{
					BaseEvent: types.BaseEvent{
						Timestamp: time.Now(),
					},
					DeploymentID: event.DeploymentID,
					Requester:    event.Requester,
					SDLHash:      event.SDLHash,
				}
				if err := handler(typesEvent); err != nil {
					c.logger.Error("failed to handle event", zap.Error(err))
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// SubscribeToBidSubmitted subscribes to bid submitted events
func (c *MarketplaceContract) SubscribeToBidSubmitted(ctx context.Context, handler func(*clusterTypes.BidSubmittedEvent) error) error {
	event := c.abi.Events["BidSubmitted"]
	topics := [][]common.Hash{{event.ID}}
	logs := make(chan ethtypes.Log)
	sub, err := c.client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	}, logs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to logs")
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case err := <-sub.Err():
				if err != nil {
					// Handle subscription error
					return
				}
			case log := <-logs:
				event := new(BidSubmittedEvent)
				if err := c.contract.UnpackLog(event, "BidSubmitted", log); err != nil {
					continue
				}
				typesEvent := &clusterTypes.BidSubmittedEvent{
					BaseEvent: clusterTypes.BaseEvent{
						Timestamp: time.Now(),
					},
					DeploymentID: event.DeploymentID,
					Provider:     event.Provider,
					Amount:       event.Amount,
					Duration:     time.Duration(event.Duration) * time.Second,
				}
				if err := handler(typesEvent); err != nil {
					// Handle event processing error
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// SubscribeToProviderSelected subscribes to provider selected events
func (c *MarketplaceContract) SubscribeToProviderSelected(ctx context.Context, handler func(*clusterTypes.ProviderSelectedEvent) error) error {
	event := c.abi.Events["ProviderSelected"]
	topics := [][]common.Hash{{event.ID}}
	logs := make(chan ethtypes.Log)
	sub, err := c.client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	}, logs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to logs")
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case err := <-sub.Err():
				if err != nil {
					// Handle subscription error
					return
				}
			case log := <-logs:
				event := new(ProviderSelectedEvent)
				if err := c.contract.UnpackLog(event, "ProviderSelected", log); err != nil {
					continue
				}
				typesEvent := &clusterTypes.ProviderSelectedEvent{
					BaseEvent: clusterTypes.BaseEvent{
						Timestamp: time.Now(),
					},
					DeploymentID: event.DeploymentID,
					Provider:     event.Provider,
					Amount:       event.Amount,
				}
				if err := handler(typesEvent); err != nil {
					// Handle event processing error
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// SubscribeToDeploymentCompleted subscribes to deployment completed events
func (c *MarketplaceContract) SubscribeToDeploymentCompleted(ctx context.Context, handler func(*clusterTypes.DeploymentCompletedEvent) error) error {
	event := c.abi.Events["DeploymentCompleted"]
	topics := [][]common.Hash{{event.ID}}
	logs := make(chan ethtypes.Log)
	sub, err := c.client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	}, logs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to logs")
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case err := <-sub.Err():
				if err != nil {
					// Handle subscription error
					return
				}
			case log := <-logs:
				event := new(DeploymentCompletedEvent)
				if err := c.contract.UnpackLog(event, "DeploymentCompleted", log); err != nil {
					continue
				}
				typesEvent := &clusterTypes.DeploymentCompletedEvent{
					BaseEvent: clusterTypes.BaseEvent{
						Timestamp: time.Now(),
					},
					DeploymentID: event.DeploymentID,
					Provider:     event.Provider,
				}
				if err := handler(typesEvent); err != nil {
					// Handle event processing error
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// SubscribeToDeploymentTerminated subscribes to deployment terminated events
func (c *MarketplaceContract) SubscribeToDeploymentTerminated(ctx context.Context, handler func(*clusterTypes.DeploymentTerminatedEvent) error) error {
	event := c.abi.Events["DeploymentTerminated"]
	topics := [][]common.Hash{{event.ID}}
	logs := make(chan ethtypes.Log)
	sub, err := c.client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	}, logs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to logs")
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			select {
			case err := <-sub.Err():
				if err != nil {
					// Handle subscription error
					return
				}
			case log := <-logs:
				event := new(DeploymentTerminatedEvent)
				if err := c.contract.UnpackLog(event, "DeploymentTerminated", log); err != nil {
					continue
				}
				typesEvent := &clusterTypes.DeploymentTerminatedEvent{
					BaseEvent: clusterTypes.BaseEvent{
						Timestamp: time.Now(),
					},
					DeploymentID: event.DeploymentID,
					Provider:     event.Provider,
					Requester:    event.Requester,
				}
				if err := handler(typesEvent); err != nil {
					// Handle event processing error
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// RequestDeployment requests a new deployment
func (c *MarketplaceContract) RequestDeployment(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
	sdlHash string,
	minBid *big.Int,
	maxBid *big.Int,
	duration *big.Int,
) (*ethtypes.Transaction, error) {
	if opts == nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "opts cannot be nil",
		}
	}
	if deploymentID == "" {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "deploymentID cannot be empty",
		}
	}
	if sdlHash == "" {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "sdlHash cannot be empty",
		}
	}
	if minBid == nil || minBid.Sign() <= 0 {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "minBid must be positive",
		}
	}
	if maxBid == nil || maxBid.Sign() <= 0 {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "maxBid must be positive",
		}
	}
	if minBid.Cmp(maxBid) > 0 {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "minBid must be less than or equal to maxBid",
		}
	}
	if duration == nil || duration.Sign() <= 0 {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "duration must be positive",
		}
	}

	tx, err := c.contract.Transact(opts, "requestDeployment", deploymentID, sdlHash, minBid, maxBid, duration)
	if err != nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeTransactionFailed,
			Message: "failed to request deployment",
			Err:     err,
		}
	}

	c.logger.Info("deployment requested",
		zap.String("deploymentID", deploymentID),
		zap.String("sdlHash", sdlHash),
		zap.String("minBid", minBid.String()),
		zap.String("maxBid", maxBid.String()),
		zap.String("duration", duration.String()),
		zap.String("txHash", tx.Hash().String()),
	)

	return tx, nil
}

// SubmitBid submits a bid for a deployment
func (c *MarketplaceContract) SubmitBid(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
	amount *big.Int,
	duration *big.Int,
) (*ethtypes.Transaction, error) {
	if opts == nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "opts cannot be nil",
		}
	}
	if deploymentID == "" {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "deploymentID cannot be empty",
		}
	}
	if amount == nil || amount.Sign() <= 0 {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "amount must be positive",
		}
	}
	if duration == nil || duration.Sign() <= 0 {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "duration must be positive",
		}
	}

	tx, err := c.contract.Transact(opts, "submitBid", deploymentID, amount, duration)
	if err != nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeTransactionFailed,
			Message: "failed to submit bid",
			Err:     err,
		}
	}

	// Add bid to tracker
	err = c.bidTracker.AddBid(ctx, deploymentID, opts.From, amount, time.Duration(duration.Int64())*time.Second)
	if err != nil {
		c.logger.Error("failed to add bid to tracker",
			zap.Error(err),
			zap.String("deploymentID", deploymentID),
			zap.String("provider", opts.From.String()),
			zap.String("amount", amount.String()),
			zap.String("duration", duration.String()),
		)
		// Don't return error here as the transaction was successful
	}

	c.logger.Info("bid submitted",
		zap.String("deploymentID", deploymentID),
		zap.String("provider", opts.From.String()),
		zap.String("amount", amount.String()),
		zap.String("duration", duration.String()),
		zap.String("txHash", tx.Hash().String()),
	)

	return tx, nil
}

// SelectProvider selects a provider for a deployment
func (c *MarketplaceContract) SelectProvider(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
	provider common.Address,
	bidIndex *big.Int,
) (*ethtypes.Transaction, error) {
	tx, err := c.contract.Transact(opts, "selectBid", deploymentID, provider, bidIndex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send transaction")
	}

	return tx, nil
}

// CompleteDeployment marks a deployment as completed
func (c *MarketplaceContract) CompleteDeployment(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
) (*ethtypes.Transaction, error) {
	// Verify deployment status
	deployment, err := c.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deployment")
	}
	if deployment.Status != clusterTypes.DeploymentStatusRunning {
		return nil, errors.New("deployment is not running")
	}

	// Release payment to provider
	tx, err := c.contract.Transact(opts, "releasePayment", deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send transaction")
	}

	return tx, nil
}

// TerminateDeployment terminates a deployment
func (c *MarketplaceContract) TerminateDeployment(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
) (*ethtypes.Transaction, error) {
	// Verify deployment status
	deployment, err := c.GetDeployment(ctx, deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deployment")
	}
	if deployment.Status != clusterTypes.DeploymentStatusRunning {
		return nil, errors.New("deployment is not running")
	}

	// Refund payment to requester
	tx, err := c.contract.Transact(opts, "refundPayment", deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send transaction")
	}

	return tx, nil
}

// LockPayment locks funds in escrow for a deployment
func (c *MarketplaceContract) LockPayment(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
	amount *big.Int,
	provider common.Address,
) (*ethtypes.Transaction, error) {
	// Verify payment with payment manager
	payment, err := c.paymentMgr.GetPayment(deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get payment")
	}
	if payment.Amount.Cmp(amount) != 0 {
		return nil, errors.New("payment amount mismatch")
	}

	tx, err := c.contract.Transact(opts, "selectBid", deploymentID, provider, big.NewInt(0)) // Use first bid
	if err != nil {
		return nil, errors.Wrap(err, "failed to send transaction")
	}

	return tx, nil
}

// ReleasePayment releases funds from escrow to the provider
func (c *MarketplaceContract) ReleasePayment(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
) (*ethtypes.Transaction, error) {
	// Verify escrow with payment manager
	escrow, err := c.paymentMgr.GetEscrow(deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get escrow")
	}
	if escrow == nil {
		return nil, errors.New("no escrow found")
	}

	tx, err := c.contract.Transact(opts, "releasePayment", deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send transaction")
	}

	return tx, nil
}

// RefundPayment refunds funds from escrow to the client
func (c *MarketplaceContract) RefundPayment(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
) (*ethtypes.Transaction, error) {
	// Verify escrow with payment manager
	escrow, err := c.paymentMgr.GetEscrow(deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get escrow")
	}
	if escrow == nil {
		return nil, errors.New("no escrow found")
	}

	tx, err := c.contract.Transact(opts, "refundPayment", deploymentID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send transaction")
	}

	return tx, nil
}

// InitiateDispute initiates a dispute for a payment
func (c *MarketplaceContract) InitiateDispute(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
	reason string,
) (*ethtypes.Transaction, error) {
	tx, err := c.contract.Transact(opts, "initiateDispute", deploymentID, reason)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send transaction")
	}

	return tx, nil
}

// GetDeployment retrieves deployment information
func (c *MarketplaceContract) GetDeployment(
	ctx context.Context,
	deploymentID string,
) (*clusterTypes.Deployment, error) {
	var result struct {
		Requester         common.Address
		SDLHash           string
		MinBid            *big.Int
		MaxBid            *big.Int
		Duration          *big.Int
		CreatedAt         *big.Int
		Status            uint8
		SelectedProvider  common.Address
		SelectedBidAmount *big.Int
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var output []interface{}
	if err := c.contract.Call(callOpts, &output, "getDeployment", deploymentID); err != nil {
		return nil, errors.Wrap(err, "failed to call contract")
	}

	if err := c.abi.UnpackIntoInterface(&result, "getDeployment", output[0].([]byte)); err != nil {
		return nil, errors.Wrap(err, "failed to unpack output")
	}

	return &clusterTypes.Deployment{
		ID:        deploymentID,
		Requester: result.Requester,
		Provider:  result.SelectedProvider,
		Status:    clusterTypes.DeploymentStatus(result.Status),
		CreatedAt: time.Unix(result.CreatedAt.Int64(), 0),
		UpdatedAt: time.Now(),
	}, nil
}

// GetBids retrieves all bids for a deployment
func (c *MarketplaceContract) GetBids(
	ctx context.Context,
	deploymentID string,
) ([]*clusterTypes.BidSubmittedEvent, error) {
	var result struct {
		Providers []common.Address
		Amounts   []*big.Int
		Durations []*big.Int
		CreatedAt []*big.Int
		Statuses  []uint8
	}

	callOpts := &bind.CallOpts{Context: ctx}
	var output []interface{}
	if err := c.contract.Call(callOpts, &output, "getBids", deploymentID); err != nil {
		return nil, errors.Wrap(err, "failed to call contract")
	}

	if err := c.abi.UnpackIntoInterface(&result, "getBids", output[0].([]byte)); err != nil {
		return nil, errors.Wrap(err, "failed to unpack output")
	}

	bids := make([]*clusterTypes.BidSubmittedEvent, len(result.Providers))
	for i := range result.Providers {
		bids[i] = &clusterTypes.BidSubmittedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Unix(result.CreatedAt[i].Int64(), 0),
			},
			DeploymentID: deploymentID,
			Provider:     result.Providers[i],
			Amount:       result.Amounts[i],
			Duration:     time.Duration(result.Durations[i].Uint64()) * time.Second,
		}
	}

	return bids, nil
}

// GetDeploymentRequestedEvents retrieves deployment requested events
func (c *MarketplaceContract) GetDeploymentRequestedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.DeploymentRequestedEvent, error) {
	event := c.abi.Events["DeploymentRequested"]
	topics := [][]common.Hash{{event.ID}}
	logs, err := c.client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to filter logs")
	}

	events := make([]*clusterTypes.DeploymentRequestedEvent, 0, len(logs))
	for _, log := range logs {
		event := new(DeploymentRequestedEvent)
		if err := c.contract.UnpackLog(event, "DeploymentRequested", log); err != nil {
			continue
		}
		typesEvent := &clusterTypes.DeploymentRequestedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: event.DeploymentID,
			Requester:    event.Requester,
			SDLHash:      event.SDLHash,
		}
		events = append(events, typesEvent)
	}

	return events, nil
}

// GetBidSubmittedEvents retrieves bid submitted events
func (c *MarketplaceContract) GetBidSubmittedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.BidSubmittedEvent, error) {
	event := c.abi.Events["BidSubmitted"]
	topics := [][]common.Hash{{event.ID}}
	logs, err := c.client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to filter logs")
	}

	events := make([]*clusterTypes.BidSubmittedEvent, 0, len(logs))
	for _, log := range logs {
		event := new(BidSubmittedEvent)
		if err := c.contract.UnpackLog(event, "BidSubmitted", log); err != nil {
			continue
		}
		typesEvent := &clusterTypes.BidSubmittedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: event.DeploymentID,
			Provider:     event.Provider,
			Amount:       event.Amount,
			Duration:     time.Duration(event.Duration) * time.Second,
		}
		events = append(events, typesEvent)
	}

	return events, nil
}

// GetProviderSelectedEvents retrieves provider selected events
func (c *MarketplaceContract) GetProviderSelectedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.ProviderSelectedEvent, error) {
	event := c.abi.Events["ProviderSelected"]
	topics := [][]common.Hash{{event.ID}}
	logs, err := c.client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to filter logs")
	}

	events := make([]*clusterTypes.ProviderSelectedEvent, 0, len(logs))
	for _, log := range logs {
		event := new(ProviderSelectedEvent)
		if err := c.contract.UnpackLog(event, "ProviderSelected", log); err != nil {
			continue
		}
		typesEvent := &clusterTypes.ProviderSelectedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: event.DeploymentID,
			Provider:     event.Provider,
			Amount:       event.Amount,
		}
		events = append(events, typesEvent)
	}

	return events, nil
}

// GetDeploymentCompletedEvents retrieves deployment completed events
func (c *MarketplaceContract) GetDeploymentCompletedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.DeploymentCompletedEvent, error) {
	event := c.abi.Events["DeploymentCompleted"]
	topics := [][]common.Hash{{event.ID}}
	logs, err := c.client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to filter logs")
	}

	events := make([]*clusterTypes.DeploymentCompletedEvent, 0, len(logs))
	for _, log := range logs {
		event := new(DeploymentCompletedEvent)
		if err := c.contract.UnpackLog(event, "DeploymentCompleted", log); err != nil {
			continue
		}
		typesEvent := &clusterTypes.DeploymentCompletedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: event.DeploymentID,
			Provider:     event.Provider,
		}
		events = append(events, typesEvent)
	}

	return events, nil
}

// GetDeploymentTerminatedEvents retrieves deployment terminated events
func (c *MarketplaceContract) GetDeploymentTerminatedEvents(ctx context.Context, fromBlock, toBlock uint64) ([]*clusterTypes.DeploymentTerminatedEvent, error) {
	event := c.abi.Events["DeploymentTerminated"]
	topics := [][]common.Hash{{event.ID}}
	logs, err := c.client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{c.contractAddr},
		Topics:    topics,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to filter logs")
	}

	events := make([]*clusterTypes.DeploymentTerminatedEvent, 0, len(logs))
	for _, log := range logs {
		event := new(DeploymentTerminatedEvent)
		if err := c.contract.UnpackLog(event, "DeploymentTerminated", log); err != nil {
			continue
		}
		typesEvent := &clusterTypes.DeploymentTerminatedEvent{
			BaseEvent: clusterTypes.BaseEvent{
				Timestamp: time.Now(),
			},
			DeploymentID: event.DeploymentID,
			Provider:     event.Provider,
			Requester:    event.Requester,
		}
		events = append(events, typesEvent)
	}

	return events, nil
}

// ParsePaymentReceived parses a payment received event from a transaction log
func (c *MarketplaceContract) ParsePaymentReceived(log ethtypes.Log) (*types.PaymentReceivedEvent, error) {
	event := new(types.PaymentReceivedEvent)
	if err := c.contract.UnpackLog(event, "PaymentReceived", log); err != nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeEventNotFound,
			Message: "failed to parse PaymentReceived event",
			Err:     err,
		}
	}
	return event, nil
}

// ReleaseEscrow releases escrowed funds to the provider
func (c *MarketplaceContract) ReleaseEscrow(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
) (*ethtypes.Transaction, error) {
	if opts == nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "opts cannot be nil",
		}
	}

	// Call contract method
	tx, err := c.contract.Transact(opts, "releaseEscrow", deploymentID)
	if err != nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeTransactionFailed,
			Message: "failed to release escrow",
			Err:     err,
		}
	}

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(ctx, c.client, tx)
	if err != nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeTransactionFailed,
			Message: "failed to wait for transaction",
			Err:     err,
		}
	}

	if receipt.Status != ethtypes.ReceiptStatusSuccessful {
		return nil, &types.ContractError{
			Code:    types.ErrCodeTransactionFailed,
			Message: "transaction failed",
		}
	}

	return tx, nil
}

// RefundEscrow refunds escrowed funds to the requester
func (c *MarketplaceContract) RefundEscrow(
	ctx context.Context,
	opts *bind.TransactOpts,
	deploymentID string,
) (*ethtypes.Transaction, error) {
	if opts == nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeInvalidInput,
			Message: "opts cannot be nil",
		}
	}

	// Call contract method
	tx, err := c.contract.Transact(opts, "refundEscrow", deploymentID)
	if err != nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeTransactionFailed,
			Message: "failed to refund escrow",
			Err:     err,
		}
	}

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(ctx, c.client, tx)
	if err != nil {
		return nil, &types.ContractError{
			Code:    types.ErrCodeTransactionFailed,
			Message: "failed to wait for transaction",
			Err:     err,
		}
	}

	if receipt.Status != ethtypes.ReceiptStatusSuccessful {
		return nil, &types.ContractError{
			Code:    types.ErrCodeTransactionFailed,
			Message: "transaction failed",
		}
	}

	return tx, nil
}
