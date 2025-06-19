package market

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/holiman/uint256"
	"github.com/unicornultrafoundation/subnet-node/bidengine/contracts"
	btypes "github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// Ensure MarketService implements MarketService
var _ btypes.MarketService = (*MarketService)(nil)

type MarketService struct {
	bidMarket *contracts.BidMarket
}

func NewMarketService(bidMarket *contracts.BidMarket) *MarketService {
	return &MarketService{
		bidMarket: bidMarket,
	}
}

// GetOrder retrieves order details from the blockchain
func (s *MarketService) GetOrder(opts *bind.CallOpts, orderId *uint256.Int) (*btypes.OrderInfo, error) {
	var result *btypes.OrderInfo
	bcOrder, err := s.bidMarket.Orders(opts, orderId.ToBig())
	if err != nil {
		return nil, err
	}
	result = &btypes.OrderInfo{
		OrderID:     orderId,
		RequesterID: bcOrder.Owner,
		Requirements: &btypes.BidRequirements{
			MinCPUCores:      uint256.MustFromBig(bcOrder.CpuCores),
			MinMemoryMB:      uint256.MustFromBig(bcOrder.MemoryMB),
			MinDiskGB:        uint256.MustFromBig(bcOrder.DiskGB),
			MinGPUCores:      uint256.MustFromBig(bcOrder.GpuCores),
			MinUploadSpeed:   uint256.MustFromBig(bcOrder.UploadMbps),
			MinDownloadSpeed: uint256.MustFromBig(bcOrder.DownloadMbps),
			Region:           uint256.MustFromBig(bcOrder.Region),
			MachineType:      uint256.MustFromBig(bcOrder.MachineType),
		},
		MaxPrice:     uint256.MustFromBig(bcOrder.MaxBidPrice),
		MinPrice:     uint256.MustFromBig(bcOrder.MinBidPrice),
		ExpirationAt: time.Unix(bcOrder.ExpiredAt.Int64(), 0),
		CreatedAt:    time.Unix(bcOrder.CreatedAt.Int64(), 0),
		Status:       bcOrder.Status,
	}
	return result, nil
}

func (s *MarketService) GetOrderInfo(opts *bind.CallOpts, orderId *uint256.Int) (*btypes.OrderInfo, error) {
	return s.GetOrder(opts, orderId)
}

func (s *MarketService) GetBids(opts *bind.CallOpts, orderId *uint256.Int) ([]*btypes.Bid, error) {
	bids, err := s.bidMarket.GetBids(opts, orderId.ToBig())
	if err != nil {
		return nil, err
	}

	var result []*btypes.Bid
	for i, bid := range bids {
		result = append(result, &btypes.Bid{
			ID:          uint256.NewInt(uint64(i)),
			OrderId:     orderId,
			ProviderId:  uint256.MustFromBig(bid.ProviderId),
			MachineId:   uint256.MustFromBig(bid.MachineId),
			PricePerSec: uint256.MustFromBig(bid.PricePerSecond),
			Status:      btypes.BidStatus(bid.Status),
			CreatedAt:   time.Unix(bid.CreatedAt.Int64(), 0),
		})
	}
	return result, nil
}

func (s *MarketService) WatchOrderCreated(opts *bind.WatchOpts, ch chan<- *contracts.BidMarketOrderCreated) (event.Subscription, error) {
	return s.bidMarket.WatchOrderCreated(opts, ch, nil)
}

func (s *MarketService) WatchBidAccepted(opts *bind.WatchOpts, ch chan<- *contracts.BidMarketBidAccepted, providerId *uint256.Int) (event.Subscription, error) {
	return s.bidMarket.WatchBidAccepted(opts, ch, nil, []*big.Int{providerId.ToBig()})
}

func (s *MarketService) SubmitBid(auth *bind.TransactOpts, orderId, providerId, machineId, pricePerSecond *uint256.Int) (*types.Transaction, error) {
	tx, err := s.bidMarket.SubmitBid(auth, orderId.ToBig(), providerId.ToBig(), machineId.ToBig(), pricePerSecond.ToBig())
	if err != nil {
		return nil, err
	}
	return tx, nil
}
