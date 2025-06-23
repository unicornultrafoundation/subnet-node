# BidEngine

## Overview

**BidEngine** is a core component responsible for managing the lifecycle of bids and resources in a decentralized compute marketplace. It is designed to be event-driven, robust, and easy to extend. The engine listens to order events and manages bids, resource allocation, and cleanup automatically.

## Architecture

```
OrderMonitor (event source)
   │
   ├──> BidManager (handles events, manages bids/resources)
   │        ├── ResourceManager (allocates/deallocates resources)
   │        ├── PricingEngine (calculates bid price)
   │        └── Storage (persists bids, orders, machines)
   │
   └──> Metrics (collects bid/order/resource metrics)
```

### Main Components
- **OrderMonitor**: Listens to blockchain events and emits order lifecycle events (new, closed, expired, accepted, etc).
- **BidManager**: Handles all bid logic, including submitting, tracking, cancelling, and cleaning up bids based on events.
- **ResourceManager**: Allocates and deallocates compute resources for bids.
- **PricingEngine**: Calculates optimal bid prices based on order and market data.
- **Storage**: Persists all relevant data (bids, orders, machines, etc).
- **Metrics**: Collects and exposes metrics for monitoring and analysis.

## Event-Driven Flow

1. **Order Created**: `OrderMonitor` emits a new order event → `BidManager` attempts to find a suitable machine and submit a bid.
2. **Order Accepted**: If our bid is accepted, resources are kept; if another provider's bid is accepted, our resources are deallocated and bid is cleaned up.
3. **Order Closed/Expired**: All resources and bids related to the order are cleaned up.
4. **Bid Cancelled**: BidManager cancels the bid and deallocates resources.

All cleanup is automatic and event-driven—no manual expiry checks are needed.

## Usage

### Initialization
```go
import (
    "github.com/unicornultrafoundation/subnet-node/bidengine/manager"
    "github.com/unicornultrafoundation/subnet-node/bidengine/types"
    // ... other imports
)

// Create and start the BidManager
bm := manager.NewManager(
    config,         // *types.BidEngineConfig
    bidMarket,      // types.BidMarketContract
    logger,         // *logrus.Logger
    metrics,        // types.Metrics
    datastore,      // ds.Datastore
    resourceMgr,    // types.ResourceManager
    pricingEngine,  // types.PricingEngine
    orderMonitor,   // types.OrderMonitor
)

bm.Start(ctx)
```

### Event Handling
You do not need to manually poll or check for expiry. All bid/resource cleanup is handled via event handlers registered in the manager:
- `OrderEventNew` → handleOrderCreate
- `OrderEventAccepted` → handleOrderAccepted
- `OrderEventClosed` → handleOrderClosed
- `OrderEventExpired` → handleOrderExpired

### Bid Lifecycle Example
```go
// Submit a bid (usually triggered by an event)
bidResult, err := bm.SubmitBid(ctx, orderID, pricePerSecond, machineID)
if err != nil {
    // handle error
}

// Cancel a bid
err := bm.CancelBid(ctx, orderID, bidIndex)
```

## Contributing
- Please open issues or pull requests for bugs, improvements, or questions.
- See code comments for extension points and architectural notes.

## License
MIT 