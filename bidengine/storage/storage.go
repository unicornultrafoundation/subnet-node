package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// Storage provides persistent storage for BidEngine data
type Storage struct {
	ds     ds.Datastore
	logger *logrus.Logger
}

// NewStorage creates a new storage instance
func NewStorage(datastore ds.Datastore, logger *logrus.Logger) *Storage {
	return &Storage{
		ds:     datastore,
		logger: logger,
	}
}

// Storage keys
const (
	orderPrefix      = "/bidengine/orders/"
	bidPrefix        = "/bidengine/bids/"
	machinePrefix    = "/bidengine/machines/"
	marketPrefix     = "/bidengine/market/"
	allocationPrefix = "/bidengine/allocations/"
)

// OrderStorage represents stored order data
type OrderStorage struct {
	Order      *Order      `json:"order"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	LastSynced time.Time   `json:"last_synced"`
	BidCount   int         `json:"bid_count"`
	Status     OrderStatus `json:"status"`
}

// BidStorage represents stored bid data
type BidStorage struct {
	Bid       *Bid      `json:"bid"`
	OrderID   string    `json:"order_id"`
	BidIndex  int       `json:"bid_index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    BidStatus `json:"status"`
}

// MachineStorage represents stored machine data
type MachineStorage struct {
	Machine    *Machine  `json:"machine"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastUsed   time.Time `json:"last_used"`
	UsageCount int       `json:"usage_count"`
}

// MarketDataStorage represents stored market data
type MarketDataStorage struct {
	MarketData *MarketData `json:"market_data"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// Error definitions
var (
	ErrOrderNotFound   = fmt.Errorf("order not found")
	ErrBidNotFound     = fmt.Errorf("bid not found")
	ErrMachineNotFound = fmt.Errorf("machine not found")
)

// Status constants
const (
	OrderStatusOpen      = types.OrderStatusOpen
	OrderStatusClosed    = types.OrderStatusClosed
	OrderStatusExpired   = types.OrderStatusExpired
	OrderStatusCancelled = types.OrderStatusCancelled
	OrderStatusAccepted  = types.OrderStatusAccepted

	BidStatusActive    = types.BidStatusActive
	BidStatusAccepted  = types.BidStatusAccepted
	BidStatusCancelled = types.BidStatusCancelled
	BidStatusExpired   = types.BidStatusExpired
)

// SaveOrder saves an order to storage
func (s *Storage) SaveOrder(ctx context.Context, order *Order) error {
	orderID := order.ID.String()
	key := ds.NewKey(orderPrefix + orderID)

	storage := &OrderStorage{
		Order:      order,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastSynced: time.Now(),
		BidCount:   0,
		Status:     order.Status,
	}

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	err = s.ds.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	s.logger.Debug("Saved order to storage", "orderID", orderID)
	return nil
}

// GetOrder retrieves an order from storage
func (s *Storage) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	key := ds.NewKey(orderPrefix + orderID)

	data, err := s.ds.Get(ctx, key)
	if err != nil {
		if err == ds.ErrNotFound {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	var storage OrderStorage
	err = json.Unmarshal(data, &storage)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return storage.Order, nil
}

// ListOrders retrieves all orders from storage
func (s *Storage) ListOrders(ctx context.Context) ([]*Order, error) {
	q := query.Query{
		Prefix: orderPrefix,
	}

	results, err := s.ds.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer results.Close()

	var orders []*Order
	for result := range results.Next() {
		if result.Error != nil {
			s.logger.Warn("Error reading order from storage", "error", result.Error)
			continue
		}

		var storage OrderStorage
		err := json.Unmarshal(result.Value, &storage)
		if err != nil {
			s.logger.Warn("Failed to unmarshal order", "error", err)
			continue
		}

		orders = append(orders, storage.Order)
	}

	return orders, nil
}

// UpdateOrder updates an existing order in storage
func (s *Storage) UpdateOrder(ctx context.Context, order *Order) error {
	orderID := order.ID.String()
	key := ds.NewKey(orderPrefix + orderID)

	// Get existing data to preserve creation time
	existingData, err := s.ds.Get(ctx, key)
	var storage OrderStorage

	switch err {
	case nil:
		// Update existing order
		err = json.Unmarshal(existingData, &storage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal existing order: %w", err)
		}
		storage.UpdatedAt = time.Now()
		storage.LastSynced = time.Now()
	case ds.ErrNotFound:
		// Create new order
		storage = OrderStorage{
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			LastSynced: time.Now(),
		}
	default:
		return fmt.Errorf("failed to get existing order: %w", err)
	}

	storage.Order = order
	storage.Status = order.Status

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	err = s.ds.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	s.logger.Debug("Updated order in storage", "orderID", orderID)
	return nil
}

// DeleteOrder removes an order from storage
func (s *Storage) DeleteOrder(ctx context.Context, orderID string) error {
	key := ds.NewKey(orderPrefix + orderID)

	err := s.ds.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	s.logger.Debug("Deleted order from storage", "orderID", orderID)
	return nil
}

// SaveBid saves a bid to storage
func (s *Storage) SaveBid(ctx context.Context, bid *Bid, orderID string, bidIndex int) error {
	key := ds.NewKey(bidPrefix + orderID + "/" + strconv.Itoa(bidIndex))

	storage := &BidStorage{
		Bid:       bid,
		OrderID:   orderID,
		BidIndex:  bidIndex,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    bid.Status,
	}

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal bid: %w", err)
	}

	err = s.ds.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to save bid: %w", err)
	}

	s.logger.Debug("Saved bid to storage", "bidIndex", bidIndex, "orderID", orderID)
	return nil
}

// GetBids retrieves all bids for an order from storage
func (s *Storage) GetBids(ctx context.Context, orderID string) ([]*Bid, error) {
	q := query.Query{
		Prefix: bidPrefix + orderID + "/",
	}

	results, err := s.ds.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to query bids: %w", err)
	}
	defer results.Close()

	var bids []*Bid
	for result := range results.Next() {
		if result.Error != nil {
			s.logger.Warn("Error reading bid from storage", "error", result.Error)
			continue
		}

		var storage BidStorage
		err := json.Unmarshal(result.Value, &storage)
		if err != nil {
			s.logger.Warn("Failed to unmarshal bid", "error", err)
			continue
		}

		bids = append(bids, storage.Bid)
	}

	return bids, nil
}

// UpdateBid updates an existing bid in storage
func (s *Storage) UpdateBid(ctx context.Context, bid *Bid, orderID string, bidIndex int) error {
	key := ds.NewKey(bidPrefix + orderID + "/" + strconv.Itoa(bidIndex))

	// Get existing data to preserve creation time
	existingData, err := s.ds.Get(ctx, key)
	var storage BidStorage

	switch err {
	case nil:
		// Update existing bid
		err = json.Unmarshal(existingData, &storage)
		if err != nil {
			return fmt.Errorf("failed to unmarshal existing bid: %w", err)
		}
		storage.UpdatedAt = time.Now()
	case ds.ErrNotFound:
		// Create new bid
		storage = BidStorage{
			OrderID:   orderID,
			BidIndex:  bidIndex,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	default:
		return fmt.Errorf("failed to get existing bid: %w", err)
	}

	storage.Bid = bid
	storage.Status = bid.Status

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal bid: %w", err)
	}

	err = s.ds.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to update bid: %w", err)
	}

	s.logger.Debug("Updated bid in storage", "bidIndex", bidIndex, "orderID", orderID)
	return nil
}

// SaveMachine saves a machine to storage
func (s *Storage) SaveMachine(ctx context.Context, machine *Machine) error {
	machineID := machine.ID.String()
	key := ds.NewKey(machinePrefix + machineID)

	storage := &MachineStorage{
		Machine:    machine,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		LastUsed:   time.Now(),
		UsageCount: 0,
	}

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal machine: %w", err)
	}

	err = s.ds.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to save machine: %w", err)
	}

	s.logger.Debug("Saved machine to storage", "machineID", machineID)
	return nil
}

// GetMachine retrieves a machine from storage
func (s *Storage) GetMachine(ctx context.Context, machineID string) (*Machine, error) {
	key := ds.NewKey(machinePrefix + machineID)

	data, err := s.ds.Get(ctx, key)
	if err != nil {
		if err == ds.ErrNotFound {
			return nil, ErrMachineNotFound
		}
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	var storage MachineStorage
	err = json.Unmarshal(data, &storage)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal machine: %w", err)
	}

	return storage.Machine, nil
}

// ListMachines retrieves all machines from storage
func (s *Storage) ListMachines(ctx context.Context) ([]*Machine, error) {
	q := query.Query{
		Prefix: machinePrefix,
	}

	results, err := s.ds.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to query machines: %w", err)
	}
	defer results.Close()

	var machines []*Machine
	for result := range results.Next() {
		if result.Error != nil {
			s.logger.Warn("Error reading machine from storage", "error", result.Error)
			continue
		}

		var storage MachineStorage
		err := json.Unmarshal(result.Value, &storage)
		if err != nil {
			s.logger.Warn("Failed to unmarshal machine", "error", err)
			continue
		}

		machines = append(machines, storage.Machine)
	}

	return machines, nil
}

// SaveMarketData saves market data to storage
func (s *Storage) SaveMarketData(ctx context.Context, marketData *MarketData) error {
	key := ds.NewKey(marketPrefix + "current")

	storage := &MarketDataStorage{
		MarketData: marketData,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal market data: %w", err)
	}

	err = s.ds.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to save market data: %w", err)
	}

	s.logger.Debug("Saved market data to storage")
	return nil
}

// GetMarketData retrieves market data from storage
func (s *Storage) GetMarketData(ctx context.Context) (*MarketData, error) {
	key := ds.NewKey(marketPrefix + "current")

	data, err := s.ds.Get(ctx, key)
	if err != nil {
		if err == ds.ErrNotFound {
			return nil, nil // Return nil if no market data exists
		}
		return nil, fmt.Errorf("failed to get market data: %w", err)
	}

	var storage MarketDataStorage
	err = json.Unmarshal(data, &storage)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal market data: %w", err)
	}

	return storage.MarketData, nil
}

// SaveResourceAllocation saves a resource allocation to storage
func (s *Storage) SaveResourceAllocation(ctx context.Context, allocation *ResourceAllocation) error {
	orderID := allocation.OrderID.String()
	key := ds.NewKey(allocationPrefix + orderID)

	storage := &ResourceAllocation{
		OrderID:   allocation.OrderID,
		Usage:     allocation.Usage,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal resource allocation: %w", err)
	}

	err = s.ds.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to save resource allocation: %w", err)
	}

	s.logger.Debug("Saved resource allocation to storage", "orderID", orderID)
	return nil
}

// ListResourceAllocations retrieves all resource allocations from storage
func (s *Storage) ListResourceAllocations(ctx context.Context) ([]*ResourceAllocation, error) {
	q := query.Query{
		Prefix: allocationPrefix,
	}

	results, err := s.ds.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to query resource allocations: %w", err)
	}
	defer results.Close()

	var allocations []*ResourceAllocation
	for result := range results.Next() {
		if result.Error != nil {
			s.logger.Warn("Error reading resource allocation from storage", "error", result.Error)
			continue
		}

		var allocation ResourceAllocation
		err := json.Unmarshal(result.Value, &allocation)
		if err != nil {
			s.logger.Warn("Failed to unmarshal resource allocation", "error", err)
			continue
		}

		allocations = append(allocations, &allocation)
	}

	return allocations, nil
}

// Type aliases to make storage compatible with types package
type (
	Order              = types.Order
	Bid                = types.Bid
	Machine            = types.Machine
	MarketData         = types.MarketData
	ResourceUsage      = types.ResourceUsage
	ResourceAllocation = types.ResourceAllocation
	OrderStatus        = types.OrderStatus
	BidStatus          = types.BidStatus
	Logger             = types.Logger
)
