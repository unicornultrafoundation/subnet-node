package pricing

import (
	"context"
	"math/big"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

// Engine implements PricingEngine interface
type Engine struct {
	config *types.BidEngineConfig
	logger *logrus.Entry
}

// NewEngine creates a new Engine instance
func NewEngine(config *types.BidEngineConfig, logger *logrus.Entry) *Engine {
	return &Engine{
		config: config,
		logger: logger,
	}
}

// CalculateBidPrice calculates the optimal bid price for an order
func (p *Engine) CalculateBidPrice(ctx context.Context, order *types.Order, machine *types.Machine, marketData *types.MarketData) (*big.Int, error) {
	// Calculate base resource price
	basePrice, err := p.CalculateResourcePrice(ctx, machine, &types.ResourceUsage{
		CPUUsed:     order.CpuCores,
		GPUUsed:     order.GpuCores,
		MemoryUsed:  order.MemoryMB,
		DiskUsed:    order.DiskGB,
		NetworkUsed: order.UploadMbps,
	})
	if err != nil {
		return nil, err
	}

	// Adjust price based on market conditions
	marketAdjustedPrice := p.adjustPriceForMarket(basePrice, marketData)

	// Apply bidding strategy
	finalPrice, err := p.AdjustPriceForStrategy(ctx, marketAdjustedPrice, &p.config.BidStrategy)
	if err != nil {
		return nil, err
	}

	// Ensure price is within order constraints
	finalPrice = p.constrainPriceToOrderLimits(finalPrice, order)

	p.logger.Info("Calculated bid price",
		"orderID", order.ID,
		"basePrice", basePrice,
		"marketAdjustedPrice", marketAdjustedPrice,
		"finalPrice", finalPrice,
		"minBidPrice", order.MinBidPrice,
		"maxBidPrice", order.MaxBidPrice)

	return finalPrice, nil
}

// AnalyzeMarket analyzes market conditions for pricing
func (p *Engine) AnalyzeMarket(ctx context.Context, orders []*types.Order) (*types.MarketData, error) {
	if len(orders) == 0 {
		return &types.MarketData{
			AveragePricePerSecond: big.NewInt(0),
			MinPricePerSecond:     big.NewInt(0),
			MaxPricePerSecond:     big.NewInt(0),
			TotalOrders:           big.NewInt(0),
			ActiveOrders:          big.NewInt(0),
			LastUpdated:           time.Now(),
		}, nil
	}

	var totalPrice big.Int
	var minPrice *big.Int
	var maxPrice *big.Int
	activeOrders := 0

	for _, order := range orders {
		if order.Status == types.OrderStatusOpen {
			activeOrders++

			// Use min bid price as reference
			if minPrice == nil || order.MinBidPrice.Cmp(minPrice) < 0 {
				minPrice = order.MinBidPrice
			}
			if maxPrice == nil || order.MaxBidPrice.Cmp(maxPrice) > 0 {
				maxPrice = order.MaxBidPrice
			}

			totalPrice.Add(&totalPrice, order.MinBidPrice)
		}
	}

	var averagePrice big.Int
	if activeOrders > 0 {
		averagePrice.Div(&totalPrice, big.NewInt(int64(activeOrders)))
	}

	return &types.MarketData{
		AveragePricePerSecond: &averagePrice,
		MinPricePerSecond:     minPrice,
		MaxPricePerSecond:     maxPrice,
		TotalOrders:           big.NewInt(int64(len(orders))),
		ActiveOrders:          big.NewInt(int64(activeOrders)),
		LastUpdated:           time.Now(),
	}, nil
}

// CalculateResourcePrice calculates the price for specific resources
func (p *Engine) CalculateResourcePrice(ctx context.Context, machine *types.Machine, usage *types.ResourceUsage) (*big.Int, error) {
	var totalPrice big.Int

	// Calculate CPU cost
	if usage.CPUUsed != nil && machine.CpuPricePerSecond != nil {
		var cpuCost big.Int
		cpuCost.Mul(usage.CPUUsed, machine.CpuPricePerSecond)
		totalPrice.Add(&totalPrice, &cpuCost)
	}

	// Calculate GPU cost
	if usage.GPUUsed != nil && machine.GpuPricePerSecond != nil {
		var gpuCost big.Int
		gpuCost.Mul(usage.GPUUsed, machine.GpuPricePerSecond)
		totalPrice.Add(&totalPrice, &gpuCost)
	}

	// Calculate Memory cost
	if usage.MemoryUsed != nil && machine.MemoryPricePerSecond != nil {
		var memoryCost big.Int
		memoryCost.Mul(usage.MemoryUsed, machine.MemoryPricePerSecond)
		totalPrice.Add(&totalPrice, &memoryCost)
	}

	// Calculate Disk cost
	if usage.DiskUsed != nil && machine.DiskPricePerSecond != nil {
		var diskCost big.Int
		diskCost.Mul(usage.DiskUsed, machine.DiskPricePerSecond)
		totalPrice.Add(&totalPrice, &diskCost)
	}

	// Add network cost (using upload speed as proxy)
	if usage.NetworkUsed != nil {
		// Assume network cost is 10% of total resource cost
		var networkCost big.Int
		networkCost.Div(&totalPrice, big.NewInt(10))
		totalPrice.Add(&totalPrice, &networkCost)
	}

	return &totalPrice, nil
}

// AdjustPriceForStrategy adjusts price based on bidding strategy
func (p *Engine) AdjustPriceForStrategy(ctx context.Context, basePrice *big.Int, strategy *types.BidStrategy) (*big.Int, error) {
	// Calculate profit margin
	minProfitMargin := strategy.MinProfitMargin
	maxProfitMargin := strategy.MaxProfitMargin

	// Add some randomness to avoid predictable bidding patterns
	profitMargin := minProfitMargin + rand.Float64()*(maxProfitMargin-minProfitMargin)

	// Apply competitive factor
	competitiveAdjustment := 1.0 - strategy.CompetitiveFactor

	// Calculate final price
	profitMultiplier := big.NewFloat(1.0 + profitMargin)
	competitiveMultiplier := big.NewFloat(competitiveAdjustment)

	basePriceFloat := new(big.Float).SetInt(basePrice)
	finalPriceFloat := new(big.Float).Mul(basePriceFloat, profitMultiplier)
	finalPriceFloat = new(big.Float).Mul(finalPriceFloat, competitiveMultiplier)

	adjustedPriceFloat, _ := finalPriceFloat.Int(nil)

	p.logger.Debug("Price strategy adjustment",
		"basePrice", basePrice,
		"profitMargin", profitMargin,
		"competitiveAdjustment", competitiveAdjustment,
		"adjustedPrice", adjustedPriceFloat)

	return adjustedPriceFloat, nil
}

// adjustPriceForMarket adjusts price based on market conditions
func (p *Engine) adjustPriceForMarket(basePrice *big.Int, marketData *types.MarketData) *big.Int {
	if marketData == nil || marketData.AveragePricePerSecond == nil {
		return basePrice
	}

	// If our base price is significantly higher than market average, reduce it
	marketAverage := marketData.AveragePricePerSecond
	basePriceFloat := new(big.Float).SetInt(basePrice)
	marketAverageFloat := new(big.Float).SetInt(marketAverage)

	// Calculate price ratio
	priceRatio := new(big.Float).Quo(basePriceFloat, marketAverageFloat)

	// If our price is more than 20% higher than market average, adjust down
	if priceRatio.Cmp(big.NewFloat(1.2)) > 0 {
		adjustmentFactor := big.NewFloat(0.9) // Reduce by 10%
		adjustedPrice := new(big.Float).Mul(basePriceFloat, adjustmentFactor)
		adjustedPriceInt, _ := adjustedPrice.Int(nil)
		return adjustedPriceInt
	}

	// If our price is significantly lower than market average, adjust up slightly
	if priceRatio.Cmp(big.NewFloat(0.8)) < 0 {
		adjustmentFactor := big.NewFloat(1.05) // Increase by 5%
		adjustedPrice := new(big.Float).Mul(basePriceFloat, adjustmentFactor)
		adjustedPriceInt, _ := adjustedPrice.Int(nil)
		return adjustedPriceInt
	}

	return basePrice
}

// constrainPriceToOrderLimits ensures price is within order constraints
func (p *Engine) constrainPriceToOrderLimits(price *big.Int, order *types.Order) *big.Int {
	// Ensure price is not below minimum
	if order.MinBidPrice != nil && price.Cmp(order.MinBidPrice) < 0 {
		return order.MinBidPrice
	}

	// Ensure price is not above maximum
	if order.MaxBidPrice != nil && price.Cmp(order.MaxBidPrice) > 0 {
		return order.MaxBidPrice
	}

	return price
}
