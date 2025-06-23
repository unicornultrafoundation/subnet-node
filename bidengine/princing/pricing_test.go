package pricing

import (
	"context"
	"math/big"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/subnet-node/bidengine/types"
)

func testEngine() *Engine {
	cfg := &types.BidEngineConfig{
		BidStrategy: types.BidStrategy{
			MinProfitMargin:   0.05,
			MaxProfitMargin:   0.10,
			CompetitiveFactor: 0.10,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	return NewEngine(cfg, logger)
}

func testMachine() *types.Machine {
	return &types.Machine{
		CpuPricePerSecond:    big.NewInt(10),
		GpuPricePerSecond:    big.NewInt(20),
		MemoryPricePerSecond: big.NewInt(5),
		DiskPricePerSecond:   big.NewInt(2),
	}
}

func testOrder() *types.Order {
	return &types.Order{
		ID:          big.NewInt(1),
		CpuCores:    big.NewInt(2),
		GpuCores:    big.NewInt(1),
		MemoryMB:    big.NewInt(1024),
		DiskGB:      big.NewInt(100),
		UploadMbps:  big.NewInt(10),
		MinBidPrice: big.NewInt(100),
		MaxBidPrice: big.NewInt(10000),
		Status:      types.OrderStatusOpen,
	}
}

func TestCalculateResourcePrice(t *testing.T) {
	engine := testEngine()
	machine := testMachine()
	usage := &types.ResourceUsage{
		CPUUsed:     big.NewInt(2),
		GPUUsed:     big.NewInt(1),
		MemoryUsed:  big.NewInt(1024),
		DiskUsed:    big.NewInt(100),
		NetworkUsed: big.NewInt(10), // ensure network cost is included
	}
	price, err := engine.CalculateResourcePrice(context.Background(), machine, usage)
	assert.NoError(t, err)
	// 2*10 + 1*20 + 1024*5 + 100*2 = 20 + 20 + 5120 + 200 = 5360
	// network cost = 5360 / 10 = 536, total = 5896
	expected := big.NewInt(5896)
	assert.Equal(t, 0, price.Cmp(expected), "resource price should match expected value")
}

func TestAdjustPriceForStrategy(t *testing.T) {
	engine := testEngine()
	base := big.NewInt(1000)
	strategy := &engine.config.BidStrategy
	price, err := engine.AdjustPriceForStrategy(context.Background(), base, strategy)
	assert.NoError(t, err)
	// Theoretical min and max
	min := new(big.Float).Mul(big.NewFloat(1000), big.NewFloat((1+strategy.MinProfitMargin)*(1-strategy.CompetitiveFactor)))
	max := new(big.Float).Mul(big.NewFloat(1000), big.NewFloat((1+strategy.MaxProfitMargin)*(1-strategy.CompetitiveFactor)))
	// priceF := new(big.Float).SetInt(price) // unused
	minInt, _ := min.Int(nil)
	maxInt, _ := max.Int(nil)
	assert.True(t, price.Cmp(minInt) >= 0 && price.Cmp(maxInt) <= 0, "price should be in expected range")
}

func TestAdjustPriceForMarket(t *testing.T) {
	engine := testEngine()
	base := big.NewInt(1000)
	market := &types.MarketData{
		AveragePricePerSecond: big.NewInt(800),
	}
	adjusted := engine.adjustPriceForMarket(base, market)
	assert.Equal(t, int64(900), adjusted.Int64()) // Should reduce price by 10%

	market.AveragePricePerSecond = big.NewInt(1200)
	adjusted = engine.adjustPriceForMarket(base, market)
	assert.Equal(t, int64(1000), adjusted.Int64()) // Should not change

	market.AveragePricePerSecond = big.NewInt(1000)
	adjusted = engine.adjustPriceForMarket(base, market)
	assert.Equal(t, base.Int64(), adjusted.Int64()) // Should not change
}

func TestConstrainPriceToOrderLimits(t *testing.T) {
	engine := testEngine()
	order := testOrder()
	low := big.NewInt(50)
	high := big.NewInt(20000)
	inRange := big.NewInt(500)
	assert.Equal(t, order.MinBidPrice, engine.constrainPriceToOrderLimits(low, order))
	assert.Equal(t, order.MaxBidPrice, engine.constrainPriceToOrderLimits(high, order))
	assert.Equal(t, inRange, engine.constrainPriceToOrderLimits(inRange, order))
}

func TestCalculateBidPrice(t *testing.T) {
	engine := testEngine()
	order := testOrder()
	machine := testMachine()
	market := &types.MarketData{
		AveragePricePerSecond: big.NewInt(1000),
	}
	price, err := engine.CalculateBidPrice(context.Background(), order, machine, market)
	assert.NoError(t, err)
	assert.NotNil(t, price)
	assert.True(t, price.Cmp(order.MinBidPrice) >= 0)
	assert.True(t, price.Cmp(order.MaxBidPrice) <= 0)
}

func TestAnalyzeMarket(t *testing.T) {
	engine := testEngine()
	orders := []*types.Order{
		{
			Status:      types.OrderStatusOpen,
			MinBidPrice: big.NewInt(100),
			MaxBidPrice: big.NewInt(200),
		},
		{
			Status:      types.OrderStatusOpen,
			MinBidPrice: big.NewInt(200),
			MaxBidPrice: big.NewInt(400),
		},
	}
	market, err := engine.AnalyzeMarket(context.Background(), orders)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), market.ActiveOrders.Int64())
	assert.Equal(t, int64(2), market.TotalOrders.Int64())
	assert.Equal(t, int64(150), market.AveragePricePerSecond.Int64())
}
