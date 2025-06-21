package bidengine

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ds "github.com/ipfs/go-datastore"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/repo/snrepo"
)

// ExampleBidEngineUsage demonstrates how to use the BidEngine
func ExampleBidEngineUsage() {
	// Example configuration
	config := &BidEngineConfig{
		BidMarketAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		ProviderAddress:  common.HexToAddress("0x0987654321098765432109876543210987654321"),
		ProviderID:       big.NewInt(1),
		ProviderWallet:   common.HexToAddress("0x1111111111111111111111111111111111111111"),
		BidStrategy: BidStrategy{
			MinProfitMargin:   0.05, // 5% minimum profit margin
			MaxProfitMargin:   0.20, // 20% maximum profit margin
			CompetitiveFactor: 0.10, // 10% competitive factor
			MarketAdjustment:  0.05, // 5% market adjustment
			ResourceWeight: ResourceWeight{
				CPU:     0.25,
				GPU:     0.35,
				Memory:  0.20,
				Disk:    0.15,
				Network: 0.05,
			},
		},
		MaxConcurrentBids: 10,
		BidTimeout:        30 * time.Second,
		OrderSyncInterval: 30 * time.Second,
		BidCheckInterval:  60 * time.Second,
		LogLevel:          "INFO",
		LogFile:           "",
	}

	// Create Ethereum client
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR_PROJECT_ID")
	if err != nil {
		log.Fatal("Failed to connect to Ethereum network:", err)
	}

	// Create transactor (you would load this from a keystore or wallet)
	transactor := &bind.TransactOpts{
		From:   config.ProviderWallet,
		Signer: nil, // You would set up a proper signer here
	}

	// Create datastore (using IPFS datastore)
	datastore := ds.NewMapDatastore()

	// Create BidEngine using the factory function
	bidEngine, err := NewBidEngineFromConfig(config, client, transactor, datastore)
	if err != nil {
		log.Fatal("Failed to create BidEngine:", err)
	}

	// Start the BidEngine
	ctx := context.Background()
	if err := bidEngine.Start(ctx); err != nil {
		log.Fatal("Failed to start BidEngine:", err)
	}

	// Example: Register a machine
	machine := &Machine{
		ID:                   big.NewInt(1),
		Active:               true,
		MachineType:          big.NewInt(1),
		Region:               big.NewInt(1),
		CpuCores:             big.NewInt(4),
		GpuCores:             big.NewInt(1),
		GpuMemory:            big.NewInt(8192),
		MemoryMB:             big.NewInt(16384),
		DiskGB:               big.NewInt(100),
		UploadSpeed:          big.NewInt(100),
		DownloadSpeed:        big.NewInt(100),
		CreatedAt:            big.NewInt(time.Now().Unix()),
		UpdatedAt:            big.NewInt(time.Now().Unix()),
		StakeAmount:          big.NewInt(1000000000000000000), // 1 ETH
		Metadata:             "High-performance GPU machine",
		CpuPricePerSecond:    big.NewInt(1000000000000000), // 0.001 ETH per second
		GpuPricePerSecond:    big.NewInt(5000000000000000), // 0.005 ETH per second
		MemoryPricePerSecond: big.NewInt(500000000000000),  // 0.0005 ETH per second
		DiskPricePerSecond:   big.NewInt(100000000000000),  // 0.0001 ETH per second
	}

	if err := bidEngine.resourceManager.RegisterMachine(ctx, machine); err != nil {
		log.Printf("Failed to register machine: %v", err)
	}

	// Example: Get BidEngine statistics
	stats := bidEngine.GetStats()
	fmt.Printf("BidEngine Stats: %+v\n", stats)

	// Example: Monitor the BidEngine for some time
	fmt.Println("BidEngine is running. Press Ctrl+C to stop...")

	// In a real application, you would run this indefinitely
	// For this example, we'll just wait a bit
	time.Sleep(10 * time.Second)

	// Stop the BidEngine
	if err := bidEngine.Stop(ctx); err != nil {
		log.Printf("Failed to stop BidEngine: %v", err)
	}

	fmt.Println("BidEngine stopped successfully")
}

// ExampleBidEngineWithConfigC demonstrates how to use BidEngine with config.C
func ExampleBidEngineWithConfigC() {
	// Create logger for config
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create config instance
	cfg := config.NewC(logger)

	// Load configuration from file
	err := cfg.Load("./config")
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create Ethereum client using config
	rpcURL := cfg.GetString("blockchain.rpc_url", "https://rpc.u2u.network")
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal("Failed to connect to Ethereum network:", err)
	}

	// Create transactor (you would load this from a keystore or wallet)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Failed to generate private key:", err)
	}

	chainID := big.NewInt(int64(cfg.GetInt("blockchain.chain_id", 1000)))
	transactor, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal("Failed to create transactor:", err)
	}

	// Create datastore (using IPFS datastore)
	datastore := ds.NewMapDatastore()

	// Create BidEngine using config.C
	bidEngine, err := NewBidEngineFromConfigC(cfg, client, transactor, datastore)
	if err != nil {
		log.Fatal("Failed to create BidEngine:", err)
	}

	// Set provider information (these would typically come from your wallet/identity)
	// You would need to set these after creating the BidEngine
	// bidEngine.config.ProviderID = yourProviderID
	// bidEngine.config.ProviderWallet = yourWalletAddress

	// Start the BidEngine
	ctx := context.Background()
	if err := bidEngine.Start(ctx); err != nil {
		log.Fatal("Failed to start BidEngine:", err)
	}

	// Example: Register a machine
	machine := &Machine{
		ID:                   big.NewInt(1),
		Active:               true,
		MachineType:          big.NewInt(1),
		Region:               big.NewInt(1),
		CpuCores:             big.NewInt(4),
		GpuCores:             big.NewInt(1),
		GpuMemory:            big.NewInt(8192),
		MemoryMB:             big.NewInt(16384),
		DiskGB:               big.NewInt(100),
		UploadSpeed:          big.NewInt(100),
		DownloadSpeed:        big.NewInt(100),
		CreatedAt:            big.NewInt(time.Now().Unix()),
		UpdatedAt:            big.NewInt(time.Now().Unix()),
		StakeAmount:          big.NewInt(1000000000000000000), // 1 ETH
		Metadata:             "High-performance GPU machine",
		CpuPricePerSecond:    big.NewInt(1000000000000000), // 0.001 ETH per second
		GpuPricePerSecond:    big.NewInt(5000000000000000), // 0.005 ETH per second
		MemoryPricePerSecond: big.NewInt(500000000000000),  // 0.0005 ETH per second
		DiskPricePerSecond:   big.NewInt(100000000000000),  // 0.0001 ETH per second
	}

	if err := bidEngine.resourceManager.RegisterMachine(ctx, machine); err != nil {
		log.Printf("Failed to register machine: %v", err)
	}

	// Example: Get BidEngine statistics
	stats := bidEngine.GetStats()
	fmt.Printf("BidEngine Stats: %+v\n", stats)

	// Example: Monitor the BidEngine for some time
	fmt.Println("BidEngine is running. Press Ctrl+C to stop...")

	// In a real application, you would run this indefinitely
	// For this example, we'll just wait a bit
	time.Sleep(10 * time.Second)

	// Stop the BidEngine
	if err := bidEngine.Stop(ctx); err != nil {
		log.Printf("Failed to stop BidEngine: %v", err)
	}

	fmt.Println("BidEngine stopped successfully")
}

// ExampleWithCustomComponents demonstrates how to create BidEngine with custom components
func ExampleWithCustomComponents() {
	// Create custom components
	logger, _ := NewLogger("DEBUG", "")
	config := &BidEngineConfig{}
	metrics := NewMetrics()
	datastore := ds.NewMapDatastore()

	pricingEngine := NewPricingEngine(config, logger)
	resourceManager := NewResourceManager(config, nil, logger, metrics, datastore) // nil for provider contract
	orderMonitor := NewOrderMonitor(config, nil, logger, metrics, datastore)       // nil for bidMarket contract
	bidManager := NewBidManager(config, nil, logger, metrics, datastore)           // nil for bidMarket contract
	storage := NewStorage(datastore, logger)

	// Create BidEngine with custom components
	bidEngine := &BidEngine{
		config:          config,
		bidMarket:       nil, // You would pass a real contract here
		provider:        nil, // You would pass a real contract here
		pricingEngine:   pricingEngine,
		resourceManager: resourceManager,
		orderMonitor:    orderMonitor,
		bidManager:      bidManager,
		storage:         storage,
		logger:          logger,
		metrics:         metrics,
		stopChan:        make(chan struct{}),
		orderEventsChan: make(chan *OrderEvent, 100),
		bidEventsChan:   make(chan *BidResult, 100),
		trackedOrders:   make(map[string]*Order),
		trackedBids:     make(map[string]*BidResult),
	}

	fmt.Printf("Created BidEngine with custom components: %+v\n", bidEngine)
}

// ExampleBidEngineIntegration shows how to integrate BidEngine with the existing datastore system
func ExampleBidEngineIntegration() {
	// Initialize the repository (this would be done by the main application)
	repo, err := snrepo.Open("./data", nil)
	if err != nil {
		log.Fatal("Failed to open repository:", err)
	}
	defer repo.Close()

	// Get the datastore from the repository
	datastore := repo.Datastore()

	// Create Ethereum client
	client, err := ethclient.Dial("https://rpc.u2u.network")
	if err != nil {
		log.Fatal("Failed to connect to Ethereum network:", err)
	}

	// Create transactor (this would come from your wallet/keystore)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Failed to generate private key:", err)
	}

	transactor, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1000)) // U2U chain ID
	if err != nil {
		log.Fatal("Failed to create transactor:", err)
	}

	// Create BidEngine configuration
	config := DefaultBidEngineConfig(
		common.HexToAddress("0x123456789abcdef123456789abcdef123456789a"), // BidMarket contract address
		common.HexToAddress("0xabcdef123456789abcdef123456789abcdef1234"), // Provider contract address
		big.NewInt(1), // Provider ID
		common.HexToAddress("0x123456789abcdef123456789abcdef123456789a"), // Provider wallet
	)

	// Create BidEngine instance with datastore
	bidEngine, err := NewBidEngineFromConfig(config, client, transactor, datastore)
	if err != nil {
		log.Fatal("Failed to create BidEngine:", err)
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the BidEngine
	if err := bidEngine.Start(ctx); err != nil {
		log.Fatal("Failed to start BidEngine:", err)
	}

	// Example: Register a machine
	machine := &Machine{
		ID:                   big.NewInt(1),
		Active:               true,
		MachineType:          big.NewInt(1),
		Region:               big.NewInt(1),
		CpuCores:             big.NewInt(4),
		GpuCores:             big.NewInt(1),
		GpuMemory:            big.NewInt(8192),
		MemoryMB:             big.NewInt(16384),
		DiskGB:               big.NewInt(100),
		UploadSpeed:          big.NewInt(100),
		DownloadSpeed:        big.NewInt(100),
		CreatedAt:            big.NewInt(time.Now().Unix()),
		UpdatedAt:            big.NewInt(time.Now().Unix()),
		StakeAmount:          big.NewInt(1000000000000000000), // 1 ETH
		Metadata:             "High-performance GPU machine",
		CpuPricePerSecond:    big.NewInt(1000000000000000), // 0.001 ETH per second
		GpuPricePerSecond:    big.NewInt(5000000000000000), // 0.005 ETH per second
		MemoryPricePerSecond: big.NewInt(500000000000000),  // 0.0005 ETH per second
		DiskPricePerSecond:   big.NewInt(100000000000000),  // 0.0001 ETH per second
	}

	if err := bidEngine.resourceManager.RegisterMachine(ctx, machine); err != nil {
		log.Printf("Failed to register machine: %v", err)
	}

	// Example: Get BidEngine statistics
	stats := bidEngine.GetStats()
	fmt.Printf("BidEngine Stats: %+v\n", stats)

	// Example: Monitor the BidEngine for some time
	fmt.Println("BidEngine is running. Press Ctrl+C to stop...")

	// In a real application, you would run this indefinitely
	// For this example, we'll just wait a bit
	time.Sleep(10 * time.Second)

	// Stop the BidEngine
	if err := bidEngine.Stop(ctx); err != nil {
		log.Printf("Failed to stop BidEngine: %v", err)
	}

	fmt.Println("BidEngine stopped successfully")
}

// ExampleWithCustomConfig demonstrates how to use BidEngine with custom configuration
func ExampleWithCustomConfig(datastore ds.Datastore) {
	// Create logger for config
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create config instance
	cfg := config.NewC(logger)

	// Load configuration from string (useful for testing or dynamic config)
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

logging:
  level: "DEBUG"
  file_path: "logs/bidengine.log"
`

	err := cfg.LoadString(configYAML)
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create Ethereum client
	client, err := ethclient.Dial(cfg.GetString("blockchain.rpc_url", ""))
	if err != nil {
		log.Fatal("Failed to connect to Ethereum network:", err)
	}

	// Create transactor
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Failed to generate private key:", err)
	}

	chainID := big.NewInt(int64(cfg.GetInt("blockchain.chain_id", 1000)))
	transactor, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal("Failed to create transactor:", err)
	}

	// Create BidEngine using config.C
	bidEngine, err := NewBidEngineFromConfigC(cfg, client, transactor, datastore)
	if err != nil {
		log.Fatal("Failed to create BidEngine:", err)
	}

	// Set provider information (these would come from your application)
	// bidEngine.config.ProviderID = yourProviderID
	// bidEngine.config.ProviderWallet = yourWalletAddress

	fmt.Printf("Created BidEngine with custom config: %+v\n", bidEngine)
}
