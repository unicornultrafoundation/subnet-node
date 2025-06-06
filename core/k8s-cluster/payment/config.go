package payment

import (
	"context"
	"fmt"

	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

// PricingConfig holds configuration for pricing strategies
type PricingConfig struct {
	MemPriceMin      int64
	MemPriceMax      int64
	BidPriceStrategy string
	BidCPUScale      float64
	BidStorageScale  float64
	BidScriptPath    string
	ProcessLimit     int
	ProcessTimeout   int
}

// PriceCalculator calculates resource prices
type PriceCalculator struct {
	config *PricingConfig
}

// NewPriceCalculator creates a new price calculator
func NewPriceCalculator(config *PricingConfig) *PriceCalculator {
	return &PriceCalculator{
		config: config,
	}
}

// CalculatePrice calculates the price for a deployment request
func (pc *PriceCalculator) CalculatePrice(ctx context.Context, event *types.DeploymentRequestedEvent) (int64, error) {
	switch pc.config.BidPriceStrategy {
	case "fixed":
		return pc.calculateFixedPrice(event)
	case "dynamic":
		return pc.calculateDynamicPrice(event)
	case "script":
		return pc.calculateScriptPrice(ctx, event)
	default:
		return 0, fmt.Errorf("unknown bid price strategy: %s", pc.config.BidPriceStrategy)
	}
}

// calculateFixedPrice calculates price using fixed rates
func (pc *PriceCalculator) calculateFixedPrice(event *types.DeploymentRequestedEvent) (int64, error) {
	// Use minimum memory price as base price
	basePrice := pc.config.MemPriceMin

	// Scale by CPU if specified
	if pc.config.BidCPUScale > 0 {
		basePrice = int64(float64(basePrice) * pc.config.BidCPUScale)
	}

	// Scale by storage if specified
	if pc.config.BidStorageScale > 0 {
		basePrice = int64(float64(basePrice) * pc.config.BidStorageScale)
	}

	return basePrice, nil
}

// calculateDynamicPrice calculates price using dynamic rates
func (pc *PriceCalculator) calculateDynamicPrice(event *types.DeploymentRequestedEvent) (int64, error) {
	// Start with minimum memory price
	price := pc.config.MemPriceMin

	// Apply linear scaling for memory
	price = int64(float64(price) * 1.5) // 50% increase

	// Apply exponential scaling for CPU
	if pc.config.BidCPUScale > 0 {
		price = int64(float64(price) * pc.config.BidCPUScale)
	}

	// Apply storage scaling
	if pc.config.BidStorageScale > 0 {
		price = int64(float64(price) * pc.config.BidStorageScale)
	}

	return price, nil
}

// calculateScriptPrice calculates price using an external script
func (pc *PriceCalculator) calculateScriptPrice(ctx context.Context, event *types.DeploymentRequestedEvent) (int64, error) {
	// TODO: Implement script-based price calculation
	return 0, fmt.Errorf("script-based price calculation not implemented")
}

// ValidatePrice validates that a calculated price is within acceptable ranges
func (pc *PriceCalculator) ValidatePrice(price int64) error {
	if price < pc.config.MemPriceMin {
		return fmt.Errorf("calculated price %d is below minimum %d", price, pc.config.MemPriceMin)
	}
	if price > pc.config.MemPriceMax {
		return fmt.Errorf("calculated price %d is above maximum %d", price, pc.config.MemPriceMax)
	}
	return nil
}
