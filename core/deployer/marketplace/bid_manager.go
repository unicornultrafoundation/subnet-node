package marketplace

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/manifest"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/payment"
	clusterTypes "github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

type BidStatus string

const (
	BidStatusPending  BidStatus = "pending"
	BidStatusAccepted BidStatus = "accepted"
	BidStatusRejected BidStatus = "rejected"
)

type BidState struct {
	DeploymentID string
	Provider     common.Address
	Amount       uint64
	Duration     time.Duration
	CreatedAt    time.Time
	Status       BidStatus
}

// BidManager represents the bid manager
type BidManager struct {
	deploymentMgr clusterTypes.DeploymentManagerInterface
	providerAddr  common.Address
	contract      clusterTypes.ContractInterface
	client        *ethclient.Client
	logger        *zap.Logger
	mu            sync.RWMutex
	bids          map[string][]*BidState
	priceCalc     *payment.PriceCalculator
	ipfsClient    clusterTypes.IPFSClient
}

// NewBidManager creates a new bid manager
func NewBidManager(deploymentMgr clusterTypes.DeploymentManagerInterface, providerAddr common.Address, contract clusterTypes.ContractInterface, client *ethclient.Client, logger *zap.Logger, priceConfig *payment.PricingConfig, ipfsClient clusterTypes.IPFSClient) *BidManager {
	return &BidManager{
		deploymentMgr: deploymentMgr,
		providerAddr:  providerAddr,
		contract:      contract,
		client:        client,
		logger:        logger,
		bids:          make(map[string][]*BidState),
		priceCalc:     payment.NewPriceCalculator(priceConfig),
		ipfsClient:    ipfsClient,
	}
}

// SubmitBid submits a bid for a deployment
func (m *BidManager) SubmitBid(ctx context.Context, deploymentID string, sdlHash string) (*BidState, error) {
	// Get SDL from IPFS
	sdl, err := m.getSDL(ctx, sdlHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get SDL: %w", err)
	}

	// Verify resource availability
	available, err := m.verifyResources(ctx, sdl)
	if err != nil {
		return nil, fmt.Errorf("failed to verify resources: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("insufficient resources available")
	}

	// Calculate bid amount from SDL
	amount, err := m.calculateBid(ctx, sdl)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate bid: %w", err)
	}

	// Get duration from deployment profile
	duration, err := m.getDurationFromSDL(sdl)
	if err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	// Submit bid to contract
	tx, err := m.contract.SubmitBid(ctx, nil, deploymentID, amount, big.NewInt(int64(duration)))
	if err != nil {
		return nil, fmt.Errorf("failed to submit bid: %w", err)
	}

	// Wait for transaction confirmation
	if err := m.waitForTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to confirm transaction: %w", err)
	}

	bid := &BidState{
		DeploymentID: deploymentID,
		Provider:     m.providerAddr,
		Amount:       amount.Uint64(),
		Duration:     time.Duration(duration) * time.Second,
		CreatedAt:    time.Now(),
	}

	// Add bid to state
	m.AddBid(deploymentID, bid)

	return bid, nil
}

// getSDL gets the SDL from IPFS
func (m *BidManager) getSDL(ctx context.Context, sdlHash string) (*manifest.SDL, error) {
	// Get SDL data from IPFS
	data, err := m.ipfsClient.Get(sdlHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get SDL from IPFS: %w", err)
	}

	// Parse SDL
	parser := manifest.NewParser()
	sdl, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SDL: %w", err)
	}

	return sdl, nil
}

// verifyResources verifies resource availability
func (m *BidManager) verifyResources(ctx context.Context, sdl *manifest.SDL) (bool, error) {
	// Get resource requirements from SDL
	requirements := sdl.Profiles.Compute["default"].Resources

	// Get available resources from deployment manager
	available := m.deploymentMgr.GetResourceUsage()

	// Parse resource requirements
	cpuReq, err := parseResourceValue(requirements.CPU.Request)
	if err != nil {
		return false, fmt.Errorf("failed to parse CPU requirement: %w", err)
	}

	memReq, err := parseResourceValue(requirements.Memory.Request)
	if err != nil {
		return false, fmt.Errorf("failed to parse memory requirement: %w", err)
	}

	storageReq := 0.0
	for _, vol := range requirements.Storage {
		size, err := parseResourceValue(vol.Size)
		if err != nil {
			return false, fmt.Errorf("failed to parse storage requirement: %w", err)
		}
		storageReq += size
	}

	// Debug log for all resources
	fmt.Printf("[DEBUG] CPU: req=%f, available=%f\n", cpuReq, available["cpu"])
	fmt.Printf("[DEBUG] Memory: req=%f, available=%f\n", memReq, available["memory"])
	fmt.Printf("[DEBUG] Storage: req=%f, available=%f\n", storageReq, available["storage"])
	if m.logger != nil {
		m.logger.Debug("Resource check",
			zap.Float64("cpuReq", cpuReq),
			zap.Float64("availableCPU", available["cpu"]),
			zap.Float64("memReq", memReq),
			zap.Float64("availableMemory", available["memory"]),
			zap.Float64("storageReq", storageReq),
			zap.Float64("availableStorage", available["storage"]))
	}

	// Check if we have enough resources
	if cpuReq > available["cpu"] {
		return false, nil
	}
	if memReq > available["memory"] {
		return false, nil
	}
	// Only check storage if storageReq > 0
	if storageReq > 0 {
		if storageReq > available["storage"] {
			return false, nil
		}
	}

	return true, nil
}

// calculateBid calculates the bid amount from SDL
func (m *BidManager) calculateBid(ctx context.Context, sdl *manifest.SDL) (*big.Int, error) {
	// Get duration
	duration, err := m.getDurationFromSDL(sdl)
	if err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	// Create deployment request event
	event := &clusterTypes.DeploymentRequestedEvent{
		BaseEvent: clusterTypes.BaseEvent{
			Timestamp: time.Now(),
		},
		DeploymentID: sdl.Deployment["default"].Profile,
		SDLHash:      sdl.VersionStr,
		Duration:     time.Duration(duration) * time.Second,
	}

	// Calculate price
	price, err := m.priceCalc.CalculatePrice(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price: %w", err)
	}

	// Convert to wei (1e18)
	amount := big.NewInt(price)
	return amount, nil
}

// getDurationFromSDL extracts duration from SDL
func (m *BidManager) getDurationFromSDL(sdl *manifest.SDL) (uint64, error) {
	// Default to 24 hours if not specified
	duration := uint64(24 * time.Hour.Seconds())
	return duration, nil
}

// waitForTransaction waits for a transaction to be confirmed
func (m *BidManager) waitForTransaction(ctx context.Context, tx *types.Transaction) error {
	// If client is nil, we're in mock mode
	if m.client == nil {
		return nil
	}

	// Wait for transaction receipt
	receipt, err := m.client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
	}

	return nil
}

// parseResourceValue parses a resource value string into a float64
func parseResourceValue(value string) (float64, error) {
	// Remove any whitespace
	value = strings.TrimSpace(value)

	// Split into number and unit
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty resource value")
	}

	// Parse number
	num, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid resource value: %w", err)
	}

	// Convert to base unit if needed
	if len(parts) > 1 {
		unit := strings.ToLower(parts[1])
		switch unit {
		case "ki", "kib":
			num *= 1024
		case "mi", "mib":
			num *= 1024 * 1024
		case "gi", "gib":
			num *= 1024 * 1024 * 1024
		case "ti", "tib":
			num *= 1024 * 1024 * 1024 * 1024
		case "k", "kb":
			num *= 1000
		case "m", "mb":
			num *= 1000 * 1000
		case "g", "gb":
			num *= 1000 * 1000 * 1000
		case "t", "tb":
			num *= 1000 * 1000 * 1000 * 1000
		}
	}

	return num, nil
}

// GetBids gets all bids for a deployment
func (m *BidManager) GetBids(deploymentID string) []*BidState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.bids[deploymentID]
}

// AddBid adds a bid to the state
func (m *BidManager) AddBid(deploymentID string, bid *BidState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bids[deploymentID] = append(m.bids[deploymentID], bid)
}

// UpdateBidStatus updates the status of a bid
func (m *BidManager) UpdateBidStatus(deploymentID string, provider common.Address, status BidStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, bid := range m.bids[deploymentID] {
		if bid.Provider == provider {
			bid.Status = status
			break
		}
	}
}

// SetIPFSClient sets the IPFS client
func (m *BidManager) SetIPFSClient(client clusterTypes.IPFSClient) {
	m.ipfsClient = client
}
