package payment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/subnet-node/core/k8s-cluster/types"
)

func TestNewPriceCalculator(t *testing.T) {
	config := &PricingConfig{
		MemPriceMin:      100,
		MemPriceMax:      1000,
		BidPriceStrategy: "fixed",
		BidCPUScale:      1.5,
		BidStorageScale:  2.0,
	}

	pc := NewPriceCalculator(config)
	assert.NotNil(t, pc)
	assert.Equal(t, config, pc.config)
}

func TestCalculateFixedPrice(t *testing.T) {
	config := &PricingConfig{
		MemPriceMin:      100,
		MemPriceMax:      1000,
		BidPriceStrategy: "fixed",
		BidCPUScale:      1.5,
		BidStorageScale:  2.0,
	}

	pc := NewPriceCalculator(config)

	event := &types.DeploymentRequestedEvent{
		DeploymentID: "test-deployment",
		SDLHash:      "test-hash",
	}

	price, err := pc.calculateFixedPrice(event)
	require.NoError(t, err)
	assert.Equal(t, int64(300), price) // 100 * 1.5 * 2.0
}

func TestCalculateDynamicPrice(t *testing.T) {
	config := &PricingConfig{
		MemPriceMin:      100,
		MemPriceMax:      1000,
		BidPriceStrategy: "dynamic",
		BidCPUScale:      1.5,
		BidStorageScale:  2.0,
	}

	pc := NewPriceCalculator(config)

	event := &types.DeploymentRequestedEvent{
		DeploymentID: "test-deployment",
		SDLHash:      "test-hash",
	}

	price, err := pc.calculateDynamicPrice(event)
	require.NoError(t, err)
	assert.Equal(t, int64(450), price) // (100 * 1.5) * 1.5 * 2.0
}

func TestCalculateScriptPrice(t *testing.T) {
	config := &PricingConfig{
		MemPriceMin:      100,
		MemPriceMax:      1000,
		BidPriceStrategy: "script",
		BidScriptPath:    "/path/to/script",
		ProcessLimit:     10,
		ProcessTimeout:   5,
	}

	pc := NewPriceCalculator(config)

	event := &types.DeploymentRequestedEvent{
		DeploymentID: "test-deployment",
		SDLHash:      "test-hash",
	}

	price, err := pc.calculateScriptPrice(context.Background(), event)
	require.Error(t, err)
	assert.Equal(t, int64(0), price)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestValidatePrice(t *testing.T) {
	config := &PricingConfig{
		MemPriceMin:      100,
		MemPriceMax:      1000,
		BidPriceStrategy: "fixed",
		BidCPUScale:      1.5,
		BidStorageScale:  2.0,
	}

	pc := NewPriceCalculator(config)

	// Test valid price
	err := pc.ValidatePrice(500)
	assert.NoError(t, err)

	// Test price below minimum
	err = pc.ValidatePrice(50)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "below minimum")

	// Test price above maximum
	err = pc.ValidatePrice(1500)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "above maximum")
}

func TestCalculatePriceWithInvalidStrategy(t *testing.T) {
	config := &PricingConfig{
		MemPriceMin:      100,
		MemPriceMax:      1000,
		BidPriceStrategy: "invalid",
		BidCPUScale:      1.5,
		BidStorageScale:  2.0,
	}

	pc := NewPriceCalculator(config)

	event := &types.DeploymentRequestedEvent{
		DeploymentID: "test-deployment",
		SDLHash:      "test-hash",
	}

	price, err := pc.CalculatePrice(context.Background(), event)
	assert.Error(t, err)
	assert.Equal(t, int64(0), price)
	assert.Contains(t, err.Error(), "unknown bid price strategy")
}
