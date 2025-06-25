package bidengine

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
)

func TestParseBidEngineConfigFromC(t *testing.T) {
	// Create logger for config
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create config instance
	cfg := config.NewC(logger)

	// Load configuration from string
	configYAML := `
blockchain:
  rpc_url: "https://rpc.u2u.network"
  chain_id: 1000

contracts:
  provider: "0x123456789abcdef123456789abcdef123456789a"
  bid_market: "0xabcdef123456789abcdef123456789abcdef1234"

bidengine:
  min_bid_percent: 75
  max_bid_percent: 90
  max_concurrent_bids: 15
  bid_timeout: "45s"
  order_sync_interval: "45s"
  bid_check_interval: "90s"
  provider_id: 1
  provider_wallet: "0x1111111111111111111111111111111111111111"
  
  strategy:
    min_profit_margin: 0.08
    max_profit_margin: 0.25
    competitive_factor: 0.15
    market_adjustment: 0.08
    
    resource_weights:
      cpu: 0.30
      gpu: 0.40
      memory: 0.20
      disk: 0.08
      network: 0.02

logging:
  level: "DEBUG"
  file_path: "logs/bidengine.log"
`

	err := cfg.LoadString(configYAML)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Parse BidEngine config
	bidEngineConfig, err := ParseBidEngineConfigFromC(cfg)
	if err != nil {
		t.Fatalf("Failed to parse BidEngine config: %v", err)
	}

	// Verify contract addresses
	expectedBidMarket := "0xabcdef123456789abcdef123456789abcdef1234"
	expectedProvider := "0x123456789abcdef123456789abcdef123456789a"

	if !strings.EqualFold(bidEngineConfig.BidMarketAddress.Hex(), expectedBidMarket) {
		t.Errorf("Expected bid market address %s, got %s", expectedBidMarket, bidEngineConfig.BidMarketAddress.Hex())
	}

	if !strings.EqualFold(bidEngineConfig.ProviderAddress.Hex(), expectedProvider) {
		t.Errorf("Expected provider address %s, got %s", expectedProvider, bidEngineConfig.ProviderAddress.Hex())
	}

	// Verify bid strategy
	if bidEngineConfig.BidStrategy.MinProfitMargin != 0.08 {
		t.Errorf("Expected min profit margin 0.08, got %f", bidEngineConfig.BidStrategy.MinProfitMargin)
	}

	if bidEngineConfig.BidStrategy.MaxProfitMargin != 0.25 {
		t.Errorf("Expected max profit margin 0.25, got %f", bidEngineConfig.BidStrategy.MaxProfitMargin)
	}

	if bidEngineConfig.BidStrategy.CompetitiveFactor != 0.15 {
		t.Errorf("Expected competitive factor 0.15, got %f", bidEngineConfig.BidStrategy.CompetitiveFactor)
	}

	if bidEngineConfig.BidStrategy.MarketAdjustment != 0.08 {
		t.Errorf("Expected market adjustment 0.08, got %f", bidEngineConfig.BidStrategy.MarketAdjustment)
	}

	// Verify resource weights
	weights := bidEngineConfig.BidStrategy.ResourceWeight
	if weights.CPU != 0.30 {
		t.Errorf("Expected CPU weight 0.30, got %f", weights.CPU)
	}
	if weights.GPU != 0.40 {
		t.Errorf("Expected GPU weight 0.40, got %f", weights.GPU)
	}
	if weights.Memory != 0.20 {
		t.Errorf("Expected Memory weight 0.20, got %f", weights.Memory)
	}
	if weights.Disk != 0.08 {
		t.Errorf("Expected Disk weight 0.08, got %f", weights.Disk)
	}
	if weights.Network != 0.02 {
		t.Errorf("Expected Network weight 0.02, got %f", weights.Network)
	}

	// Verify operational parameters
	if bidEngineConfig.MaxConcurrentBids != 15 {
		t.Errorf("Expected max concurrent bids 15, got %d", bidEngineConfig.MaxConcurrentBids)
	}

	// Verify logging configuration
	if bidEngineConfig.LogLevel != "DEBUG" {
		t.Errorf("Expected log level DEBUG, got %s", bidEngineConfig.LogLevel)
	}

	if bidEngineConfig.LogFile != "logs/bidengine.log" {
		t.Errorf("Expected log file logs/bidengine.log, got %s", bidEngineConfig.LogFile)
	}
}

func TestParseBidEngineConfigFromCWithDefaults(t *testing.T) {
	// Create logger for config
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create config instance
	cfg := config.NewC(logger)

	// Load minimal configuration
	configYAML := `
contracts:
  provider: "0x123456789abcdef123456789abcdef123456789a"
  bid_market: "0xabcdef123456789abcdef123456789abcdef1234"
bidengine:
  provider_id: 1
  provider_wallet: "0x1111111111111111111111111111111111111111"
`

	err := cfg.LoadString(configYAML)
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Parse BidEngine config
	bidEngineConfig, err := ParseBidEngineConfigFromC(cfg)
	if err != nil {
		t.Fatalf("Failed to parse BidEngine config: %v", err)
	}

	// Verify default values are used
	if bidEngineConfig.MaxConcurrentBids != 10 {
		t.Errorf("Expected default max concurrent bids 10, got %d", bidEngineConfig.MaxConcurrentBids)
	}

	if bidEngineConfig.LogLevel != "INFO" {
		t.Errorf("Expected default log level INFO, got %s", bidEngineConfig.LogLevel)
	}

	// Verify default resource weights
	weights := bidEngineConfig.BidStrategy.ResourceWeight
	if weights.CPU != 0.25 {
		t.Errorf("Expected default CPU weight 0.25, got %f", weights.CPU)
	}
	if weights.GPU != 0.35 {
		t.Errorf("Expected default GPU weight 0.35, got %f", weights.GPU)
	}
}
