# BidEngine for Subnet Node

The BidEngine is a comprehensive automated bidding and resource management system for the Subnet Node platform. It enables providers to automatically participate in decentralized resource markets by monitoring orders, calculating optimal bid prices, and managing resource allocation with persistent storage and advanced configuration management.

## Features

### Core Functionality
- **Automated Order Monitoring**: Continuously syncs and monitors orders from the blockchain
- **Intelligent Bidding**: Calculates optimal bid prices based on market conditions and resource costs
- **Resource Management**: Tracks and allocates machine resources for accepted bids
- **Event Processing**: Handles blockchain events for real-time updates
- **Metrics Collection**: Comprehensive monitoring and statistics
- **Persistent Storage**: Data persistence using IPFS datastore to survive service restarts
- **Advanced Configuration**: Integration with Subnet Node's config system
- **Structured Logging**: Logrus-based logging with configurable levels

### Key Components

1. **BidEngine Service**: Main orchestrator that coordinates all components
2. **Pricing Engine**: Calculates optimal bid prices using market analysis and strategies
3. **Resource Manager**: Manages resource allocation and availability with persistence
4. **Order Monitor**: Tracks order lifecycle and status changes with storage
5. **Bid Manager**: Handles bid submission and tracking with persistence
6. **Storage Layer**: IPFS datastore integration for data persistence
7. **Contract Interfaces**: Ethereum smart contract bindings for bid market and provider contracts

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   BidEngine     │    │  PricingEngine  │    │ ResourceManager │
│   (Main Service)│◄──►│  (Price Calc)   │◄──►│  (Allocation)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  OrderMonitor   │    │   BidManager    │    │   Storage       │
│  (Order Tracking)│◄──►│  (Bid Handling) │◄──►│ (IPFS Datastore)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌─────────────────┐
                    │ Contract Layer  │
                    │ (Blockchain)    │
                    └─────────────────┘
```

## Usage

### Basic Setup with Configuration

```go
package main

import (
    "context"
    "log"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/bidengine"
    "your-project/config"
    "your-project/repo"
)

func main() {
    // Load configuration from Subnet Node config system
    cfg := config.C
    
    // Create bid engine from config
    engine, err := bidengine.NewBidEngineFromConfigC(cfg, client, auth)
    if err != nil {
        log.Fatal(err)
    }
    
    // Start the bid engine
    ctx := context.Background()
    if err := engine.Start(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Keep running
    select {}
}
```

### Setup with Repository and Datastore

```go
package main

import (
    "context"
    "log"
    "math/big"
    
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    
    "your-project/bidengine"
    "your-project/repo"
)

func main() {
    // Connect to Ethereum client
    client, err := ethclient.Dial("https://your-rpc-endpoint")
    if err != nil {
        log.Fatal(err)
    }
    
    // Create transactor for transactions
    auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize repository and datastore
    repo, err := repo.NewRepo("path/to/repo")
    if err != nil {
        log.Fatal(err)
    }
    
    datastore := repo.Datastore()
    
    // Contract addresses
    bidMarketAddr := common.HexToAddress("0x...")
    providerAddr := common.HexToAddress("0x...")
    providerID := big.NewInt(1)
    providerWallet := common.HexToAddress("0x...")
    
    // Create default configuration
    config := bidengine.DefaultBidEngineConfig(
        bidMarketAddr,
        providerAddr,
        providerID,
        providerWallet,
    )
    
    // Validate configuration
    if err := bidengine.ValidateBidEngineConfig(config); err != nil {
        log.Fatal(err)
    }
    
    // Create bid engine with datastore
    engine, err := bidengine.NewBidEngineFromConfig(config, client, auth, datastore)
    if err != nil {
        log.Fatal(err)
    }
    
    // Start the bid engine
    ctx := context.Background()
    if err := engine.Start(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Keep running
    select {}
}
```

### Configuration

The BidEngine can be configured through the Subnet Node config system:

```yaml
# config.yaml
bidengine:
  # Contract addresses
  bid_market_address: "0xabcdef123456789abcdef123456789abcdef1234"
  provider_address: "0x123456789abcdef123456789abcdef123456789a"
  
  # Provider information
  provider_id: 1
  provider_wallet: "0x123456789abcdef123456789abcdef123456789a"
  
  # Bidding strategy
  bid_strategy:
    min_profit_margin: 0.08    # 8% minimum profit
    max_profit_margin: 0.25    # 25% maximum profit
    competitive_factor: 0.15   # 15% competitive adjustment
    market_adjustment: 0.05    # 5% market adjustment
    resource_weight:
      cpu: 0.25
      gpu: 0.35
      memory: 0.20
      disk: 0.15
      network: 0.05
  
  # Operational settings
  max_concurrent_bids: 10
  bid_timeout: "30s"
  order_sync_interval: "30s"
  bid_check_interval: "60s"
  
  # Logging
  log_level: "INFO"
  log_file: "logs/bidengine.log"
```

Or programmatically:

```go
config := &bidengine.BidEngineConfig{
    // Contract addresses
    BidMarketAddress: common.HexToAddress("0x..."),
    ProviderAddress:  common.HexToAddress("0x..."),
    
    // Provider information
    ProviderID:     big.NewInt(1),
    ProviderWallet: common.HexToAddress("0x..."),
    
    // Bidding strategy
    BidStrategy: bidengine.BidStrategy{
        MinProfitMargin:   0.05,  // 5% minimum profit
        MaxProfitMargin:   0.20,  // 20% maximum profit
        CompetitiveFactor: 0.10,  // 10% competitive adjustment
        MarketAdjustment:  0.05,  // 5% market adjustment
        ResourceWeight: bidengine.ResourceWeight{
            CPU:     0.25,
            GPU:     0.35,
            Memory:  0.20,
            Disk:    0.15,
            Network: 0.05,
        },
    },
    
    // Operational settings
    MaxConcurrentBids:  10,
    BidTimeout:         30 * time.Second,
    OrderSyncInterval:  30 * time.Second,
    BidCheckInterval:   60 * time.Second,
    
    // Logging
    LogLevel: "INFO",
    LogFile:  "",
}
```

### Persistent Storage

The BidEngine now includes persistent storage using IPFS datastore:

```go
// Data is automatically persisted to the datastore
// Orders, bids, and machine registrations survive service restarts

// Load persisted data on startup
engine.LoadPersistedData()

// Data is automatically saved during operations
// No manual persistence calls needed
```

### Custom Pricing Strategy

You can implement custom pricing strategies by implementing the `PricingEngine` interface:

```go
type CustomPricingEngine struct {
    // Your custom implementation
}

func (p *CustomPricingEngine) CalculateBidPrice(ctx context.Context, order *Order, machine *Machine, marketData *MarketData) (*big.Int, error) {
    // Your custom pricing logic
    return price, nil
}

// ... implement other interface methods
```

## Monitoring and Metrics

The BidEngine provides comprehensive metrics:

```go
// Get current statistics
stats := engine.GetStats()
fmt.Printf("Tracked Orders: %d\n", stats["trackedOrders"])
fmt.Printf("Tracked Bids: %d\n", stats["trackedBids"])

// Get metrics from individual components
metrics := engine.metrics.GetStats()
fmt.Printf("Bids Submitted: %d\n", metrics["bids"].(map[string]interface{})["submitted"])
fmt.Printf("Bids Accepted: %d\n", metrics["bids"].(map[string]interface{})["accepted"])
```

## Logging

The BidEngine uses Logrus for structured logging:

```go
// Log levels: DEBUG, INFO, WARN, ERROR
// Logs include structured fields for easy filtering and analysis

// Example log output:
// time="2024-01-15T10:30:00Z" level=info msg="Calculated bid price" orderID=123 basePrice=1000000 finalPrice=1100000
// time="2024-01-15T10:30:01Z" level=info msg="Bid submitted successfully" orderID=123 bidIndex=0 price=1100000
```

## Error Handling

The BidEngine includes comprehensive error handling:

```go
// Common errors
var (
    ErrMachineNotFound = fmt.Errorf("machine not found")
    ErrInsufficientResources = fmt.Errorf("insufficient resources")
    ErrOrderNotTracked = fmt.Errorf("order not tracked")
    ErrBidNotTracked = fmt.Errorf("bid not tracked")
    ErrStorageError = fmt.Errorf("storage operation failed")
    // ... more error definitions
)
```

## Factory Functions

The BidEngine provides several factory functions for different use cases:

```go
// Create from config struct
engine, err := bidengine.NewBidEngineFromConfig(config, client, auth, datastore)

// Create from Subnet Node config system
engine, err := bidengine.NewBidEngineFromConfigC(cfg, client, auth)

// Create with default configuration
engine, err := bidengine.NewBidEngine(bidMarketAddr, providerAddr, providerID, providerWallet, client, auth, datastore)
```

## Security Considerations

1. **Private Key Management**: Ensure private keys are stored securely
2. **Gas Management**: Monitor gas costs for bid submissions
3. **Resource Limits**: Set appropriate limits to prevent resource exhaustion
4. **Error Recovery**: Implement proper error handling and recovery mechanisms
5. **Data Persistence**: Ensure datastore is properly secured and backed up
6. **Configuration Security**: Validate all configuration parameters

## Recent Improvements

- **Persistent Storage**: Added IPFS datastore integration for data persistence
- **Logrus Integration**: Replaced custom logger with Logrus for better structured logging
- **Config System Integration**: Added support for Subnet Node's config system
- **Factory Functions**: Added multiple factory functions for different initialization patterns
- **Service Renaming**: Removed "Impl" suffixes from service implementations for cleaner naming
- **Enhanced Error Handling**: Improved error definitions and handling throughout the system
- **Thread Safety**: Ensured all components are thread-safe for concurrent operations

## Performance Optimization

1. **Batch Processing**: Process multiple orders efficiently
2. **Caching**: Cache frequently accessed data
3. **Connection Pooling**: Reuse blockchain connections
4. **Rate Limiting**: Implement rate limiting for API calls

## Testing

```go
func TestBidEngine(t *testing.T) {
    // Create test configuration
    config := bidengine.DefaultBidEngineConfig(...)
    
    // Create mock contracts
    mockBidMarket := &MockBidMarketContract{}
    mockProvider := &MockProviderContract{}
    
    // Create bid engine with mocks
    engine := bidengine.NewBidEngine(...)
    
    // Test functionality
    // ...
}
```

## Contributing

1. Follow Go coding standards
2. Add tests for new functionality
3. Update documentation
4. Ensure all interfaces are properly implemented

## License

This project is licensed under the same license as the main Subnet Node project. 