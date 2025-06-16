// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "./ISubnetProvider.sol";


contract SubnetBidMarketplace is Initializable, OwnableUpgradeable {
    using SafeERC20 for IERC20;

    enum OrderStatus { Open, Matched, Closed }
    enum BidStatus { Pending, Accepted, Cancelled }

    struct Order {
        uint256 machineType;
        address owner;
        OrderStatus status;
        uint256 createdAt;
        uint256 duration; 
        uint256 minBidPrice;
        uint256 maxBidPrice;
        uint256 acceptedBidPricePerSecond;
        uint256 parentOrderId;
        address paymentToken;
        uint256 cpuCores;    // Number of CPU cores
        uint256 gpuCores;    // GPU cores (0 if no GPU)
        uint256 gpuMemory;   // GPU memory in MB
        uint256 memoryMB;    // RAM in MB
        uint256 diskGB;      // Storage in GB
        uint256 uploadMbps; // Upload speed in Mbps
        uint256 downloadMbps; // Download speed in Mbps
        uint256 region; // Region ID (0 if not specified)
        string specs; // Resource specifications
        uint256 acceptedProviderId; // Provider NFT ID that was selected
        uint256 acceptedMachineId; // Machine ID that was selected
        uint256 startAt; // Add start time
        uint256 expiredAt; // Add expiration time
        uint256 lastPaidAt; // Add last paid time
    }

    struct Bid {
        address provider;
        uint256 pricePerSecond;
        BidStatus status;
        uint256 createdAt;
        uint256 providerId;    // Provider NFT ID
        uint256 machineId;     // Machine ID from the provider
    }

    // Constants
    uint256 public bidTimeLimit; // Time limit for submitting bids (default: 5 minutes)

    // orderId counter
    uint256 public orderCount;
    // orderId => Order
    mapping(uint256 => Order) public orders;
    // orderId => array of bids
    mapping(uint256 => Bid[]) public orderBids;

    address public paymentToken;
    address public subnetProviderContract;

    // Platform fee variables
    uint256 public platformFeePercentage; // Fee percentage in basis points (e.g., 100 = 1%)
    address public platformWallet; // Address where fees will be sent
    uint256 public totalAccumulatedFees; // Total fees collected by platform

    // Events
    event OrderCreated(uint256 indexed orderId, address owner, uint256 duration);
    event BidSubmitted(uint256 indexed orderId, address indexed provider, uint256 price, uint256 providerId, uint256 machineId);
    event BidAccepted(uint256 indexed orderId, address indexed provider, uint256 price);
    event OrderCancelled(uint256 indexed orderId, address owner);
    event BidCancelled(uint256 indexed orderId, address indexed provider, uint256 bidIndex);
    event OrderExtended(uint256 indexed orderId, uint256 additionalDuration, uint256 newExpiry);
    event SpecsUpdated(uint256 indexed orderId, string newSpecs);
    event BidTimeExpired(uint256 indexed orderId);
    event BidTimeLimitUpdated(uint256 oldLimit, uint256 newLimit);
    event OrderClosed(uint256 indexed orderId, uint256 refundAmount, string reason);
    event PlatformFeeUpdated(uint256 oldFee, uint256 newFee);
    event PlatformWalletUpdated(address oldWallet, address newWallet);
    event FeesWithdrawn(address wallet, uint256 amount);

    /**
     * @dev Initialize the contract (only callable once).
     * @param _paymentToken The ERC20 token address.
     * @param _subnetProviderContract The SubnetProvider contract address.
     */
    function initialize(address owner, address _paymentToken, address _subnetProviderContract) external initializer {
        __Ownable_init(owner);
        paymentToken = _paymentToken;
        subnetProviderContract = _subnetProviderContract;
        bidTimeLimit = 5 minutes; // Set default to 5 minutes
        platformFeePercentage = 100; // Default 1%
        platformWallet = owner; // Default to contract owner
    }

    /**
     * @dev Update the bid time limit (owner only)
     * @param newBidTimeLimit New time limit in seconds
     */
    function setBidTimeLimit(uint256 newBidTimeLimit) external onlyOwner {
        require(newBidTimeLimit > 0, "Time limit must be positive");
        uint256 oldLimit = bidTimeLimit;
        bidTimeLimit = newBidTimeLimit;
        emit BidTimeLimitUpdated(oldLimit, newBidTimeLimit);
    }

    /**
     * @dev Set payment token (admin only, can be extended as needed).
     * @param _paymentToken The ERC20 token address.
     */
    function setPaymentConfig(address _paymentToken) external onlyOwner {
        require(_paymentToken != address(0), "Invalid token");
        paymentToken = _paymentToken;
    }

    /**
     * @dev Update SubnetProvider contract address (owner only)
     * @param _subnetProviderContract New SubnetProvider contract address
     */
    function setSubnetProviderContract(address _subnetProviderContract) external onlyOwner {
        require(_subnetProviderContract != address(0), "Invalid provider contract");
        subnetProviderContract = _subnetProviderContract;
    }
    
    /**
     * @dev Renter creates an order.
     * @param duration Duration (in seconds) for the order.
     * @param minBidPrice Minimum bid price allowed.
     * @param maxBidPrice Maximum bid price allowed.
     * @param specs Resource specifications.
     */
    function createOrder(
        uint256 machineType,
        uint256 duration,
        uint256 minBidPrice,
        uint256 maxBidPrice,
        uint256 region,
        uint256 cpuCores,
        uint256 gpuCores,
        uint256 gpuMemory,
        uint256 memoryMB,
        uint256 diskGB,
        uint256 uploadMbps,
        uint256 downloadMbps,
        string memory specs
    ) external returns (uint256) {
        require(paymentToken != address(0), "Payment token not set");
        require(duration > 0, "Duration must be positive");
        require(minBidPrice <= maxBidPrice, "minBidPrice must be <= maxBidPrice");

        orderCount++;
        orders[orderCount] = Order({
            machineType: machineType,
            owner: msg.sender,
            status: OrderStatus.Open,
            createdAt: block.timestamp,
            duration: duration,
            minBidPrice: minBidPrice,
            maxBidPrice: maxBidPrice,
            acceptedBidPricePerSecond: 0,
            parentOrderId: 0,
            paymentToken: paymentToken,
            specs: specs,
            acceptedProviderId: 0,
            acceptedMachineId: 0,
            startAt: 0,
            expiredAt: 0,
            lastPaidAt: 0,
            cpuCores: cpuCores,
            gpuCores: gpuCores,
            gpuMemory: gpuMemory,
            memoryMB: memoryMB,
            diskGB: diskGB,
            uploadMbps: uploadMbps,
            downloadMbps: downloadMbps,
            region: region
        });
        emit OrderCreated(orderCount, msg.sender, duration);
        return orderCount;
    }

    /**
     * @dev Provider submits a bid for an order.
     * @param orderId The order ID.
     * @param pricePerSecond The bid price.
     * @param providerId The provider NFT ID.
     * @param machineId The machine ID from the provider.
     */
    function submitBid(
        uint256 orderId,
        uint256 pricePerSecond,
        uint256 providerId,
        uint256 machineId
    ) external {
        Order storage order = orders[orderId];
        require(order.status == OrderStatus.Open, "Order not open");
        require(block.timestamp <= order.createdAt + bidTimeLimit, "Bidding time expired");
        require(pricePerSecond >= order.minBidPrice, "Bid price below minimum");
        require(pricePerSecond <= order.maxBidPrice, "Bid price above maximum");
        
        ISubnetProvider providerContract = ISubnetProvider(subnetProviderContract);

        // Check if the caller is the provider owner or operator
        require(providerContract.isProviderOperatorOrOwner(providerId, msg.sender),
            "Not authorized to bid for this provider"
        );

        require(
            providerContract.validateMachineRequirements(
                order.machineType,
                providerId,
                machineId,
                order.cpuCores,
                order.memoryMB,
                order.diskGB,
                order.gpuCores,
                order.uploadMbps,
                order.downloadMbps
            ),
            "Machine does not meet requirements"
        );

        orderBids[orderId].push(Bid({
            provider: msg.sender,
            pricePerSecond: pricePerSecond,
            status: BidStatus.Pending,
            createdAt: block.timestamp,
            providerId: providerId,
            machineId: machineId
        }));
        emit BidSubmitted(orderId, msg.sender, pricePerSecond, providerId, machineId);
    }

    /**
     * @dev Renter accepts a bid for an order.
     * @param orderId The order ID.
     * @param bidIndex The index of the bid in the order's bids array.
     */
    function acceptBid(
        uint256 orderId,
        uint256 bidIndex
    ) external {
        Order storage order = orders[orderId];
        require(order.owner == msg.sender, "Only order owner can accept");
        require(order.status == OrderStatus.Open, "Order not open");
        require(bidIndex < orderBids[orderId].length, "Invalid bid index");
        Bid storage bid = orderBids[orderId][bidIndex];
        require(bid.status == BidStatus.Pending, "Bid not pending");

        // Check if provider and machine are valid
        ISubnetProvider providerContract = ISubnetProvider(subnetProviderContract);
        require(providerContract.isMachineActive(bid.providerId, bid.machineId), "Machine does not exist");

        // Payment logic: renter pays (pricePerSecond * duration) to contract
        require(order.paymentToken != address(0), "Payment token not set");
        uint256 totalCost = bid.pricePerSecond * order.duration;
        
        // Transfer full amount from user
        IERC20(order.paymentToken).safeTransferFrom(msg.sender, address(this), totalCost);

        // Update bid and order status
        bid.status = BidStatus.Accepted;
        order.status = OrderStatus.Matched;
        order.acceptedBidPricePerSecond = bid.pricePerSecond;
        order.acceptedProviderId = bid.providerId;
        order.acceptedMachineId = bid.machineId;
        order.startAt = block.timestamp;
        order.expiredAt = block.timestamp + order.duration;
        order.lastPaidAt = block.timestamp;

        emit BidAccepted(orderId, bid.provider, bid.pricePerSecond);
    }

    /**
     * @dev Check if bidding is still open for an order
     * @param orderId The order ID
     * @return true if bidding is still open, false otherwise
     */
    function isBiddingOpen(uint256 orderId) public view returns (bool) {
        Order storage order = orders[orderId];
        return (order.status == OrderStatus.Open && 
                block.timestamp <= order.createdAt + bidTimeLimit);
    }

    /**
     * @dev Get remaining bidding time for an order in seconds
     * @param orderId The order ID
     * @return Remaining time in seconds, 0 if expired
     */
    function getRemainingBidTime(uint256 orderId) public view returns (uint256) {
        Order storage order = orders[orderId];
        uint256 endTime = order.createdAt + bidTimeLimit;
        
        if (block.timestamp >= endTime) {
            return 0;
        }
        
        return endTime - block.timestamp;
    }

    /**
     * @dev Get all bids for an order.
     * @param orderId The order ID.
     * @return Array of Bid structs.
     */
    function getBids(uint256 orderId) external view returns (Bid[] memory) {
        return orderBids[orderId];
    }

    /**
     * @dev Extend the duration of an order.
     * Only the order owner can extend, and only if not expired.
     * @param orderId The order ID.
     */
    function extend(uint256 orderId) external {
        Order storage order = orders[orderId];
        require(order.owner == msg.sender, "Only order owner can extend");
        require(order.status == OrderStatus.Matched, "Order not matched");
        require(block.timestamp < order.expiredAt, "Order already expired");

        uint256 pricePerSecond = order.acceptedBidPricePerSecond;
        require(pricePerSecond > 0, "Price not set");

        uint256 totalCost = order.duration * pricePerSecond;
        
        IERC20(order.paymentToken).safeTransferFrom(msg.sender, address(this), totalCost);

        // Update expiration
        order.expiredAt += order.duration;

        emit OrderExtended(orderId, order.duration, order.expiredAt);
    }


    /**
     * @dev Cancel an order if it is still pending (Open). Only owner can cancel.
     * @param orderId The order ID.
     */
    function cancelOrder(uint256 orderId) external {
        Order storage order = orders[orderId];
        require(order.owner == msg.sender, "Only order owner can cancel");
        require(order.status == OrderStatus.Open, "Order not open");
        order.status = OrderStatus.Closed;
        emit OrderCancelled(orderId, msg.sender);
    }

    /**
     * @dev Close an active (Matched) order and refund remaining payment
     * @param orderId The order ID
     * @param reason Reason for closing the order
     */
    function closeOrder(uint256 orderId, string memory reason) external {
        Order storage order = orders[orderId];
        require(order.status == OrderStatus.Matched, "Order not active");
        require(order.owner == msg.sender, "Not authorized");
        
        // Calculate time used and remaining time
        uint256 timeUsed = block.timestamp > order.lastPaidAt ? 
            block.timestamp - order.lastPaidAt : 0;
        uint256 remainingTime = order.expiredAt > block.timestamp ?
            order.expiredAt - block.timestamp : 0;
            
        // Calculate payment for provider and refund for order owner
        uint256 paymentForProvider = timeUsed * order.acceptedBidPricePerSecond;
        uint256 refundAmount = remainingTime * order.acceptedBidPricePerSecond;
        
      
        
        // Update order status
        order.status = OrderStatus.Closed;
        order.lastPaidAt = block.timestamp;

        bool isMachineActive = ISubnetProvider(subnetProviderContract).isMachineActive(order.acceptedProviderId, order.acceptedMachineId);
        
        // Process payments if needed
        if (paymentForProvider > 0 && isMachineActive) {
            // Apply platform fee to provider payment
            uint256 platformFee = 0;
            if (paymentForProvider > 0) {
                platformFee = calculateFee(paymentForProvider);
                paymentForProvider -= platformFee;
                totalAccumulatedFees += platformFee;
            }

            address providerOwner = IERC721(subnetProviderContract).ownerOf(order.acceptedProviderId);
            IERC20(order.paymentToken).safeTransfer(providerOwner, paymentForProvider);
        } else {
           refundAmount += paymentForProvider; // If provider is inactive, refund the full amount
        }
        
        if (refundAmount > 0) {
            IERC20(order.paymentToken).safeTransfer(order.owner, refundAmount);
        }
        
        emit OrderClosed(orderId, refundAmount, reason);
    }

    /**
     * @dev Cancel a bid if order is still open. Only bid provider can cancel.
     * @param orderId The order ID.
     * @param bidIndex The index of the bid in the order's bids array.
     */
    function cancelBid(uint256 orderId, uint256 bidIndex) external {
        Order storage order = orders[orderId];
        require(order.status == OrderStatus.Open, "Order not open");
        require(bidIndex < orderBids[orderId].length, "Invalid bid index");
        Bid storage bid = orderBids[orderId][bidIndex];
        require(bid.provider == msg.sender, "Only bid provider can cancel");
        require(bid.status == BidStatus.Pending, "Cannot cancel non-pending bid");

        bid.status = BidStatus.Cancelled;
        emit BidCancelled(orderId, msg.sender, bidIndex);
    }

    /**
     * @dev Provider claims payment for time used.
     * Payment is based on actual time used since last payment.
     * @param orderId The order ID.
     */
    function claimPayment(uint256 orderId) external {
        Order storage order = orders[orderId];
        require(order.status == OrderStatus.Matched, "Order not matched");
        ISubnetProvider providerContract = ISubnetProvider(subnetProviderContract);
        require(providerContract.isMachineActive(order.acceptedProviderId, order.acceptedMachineId), "Machine does not exist");
        // Calculate time used since last payment
        uint256 endTime = block.timestamp < order.expiredAt ? block.timestamp : order.expiredAt;
        require(endTime > order.lastPaidAt, "No new time to pay for");
        
        uint256 actualTimeUsed = endTime - order.lastPaidAt;
        
        // Calculate payment based on actual time used
        uint256 totalPayment = actualTimeUsed * order.acceptedBidPricePerSecond;
        require(totalPayment > 0, "Nothing to claim");
        
        // Apply platform fee
        uint256 platformFee = calculateFee(totalPayment);
        uint256 providerPayment = totalPayment - platformFee;
        totalAccumulatedFees += platformFee;
        
        // Update lastPaidAt to prevent double payment
        order.lastPaidAt = endTime;

        address providerOwner = IERC721(subnetProviderContract).ownerOf(order.acceptedProviderId);
        require(providerOwner != address(0), "Provider does not exist");
        
        // Send payment to the provider (after fee)
        IERC20(order.paymentToken).safeTransfer(providerOwner, providerPayment);
    }

    /**
     * @dev Set platform fee percentage (admin only)
     * @param newFeePercentage New fee percentage in basis points (100 = 1%)
     */
    function setPlatformFee(uint256 newFeePercentage) external onlyOwner {
        require(newFeePercentage <= 2000, "Fee cannot exceed 20%");
        uint256 oldFee = platformFeePercentage;
        platformFeePercentage = newFeePercentage;
        emit PlatformFeeUpdated(oldFee, newFeePercentage);
    }

    /**
     * @dev Set platform wallet address (admin only)
     * @param newWallet New wallet address
     */
    function setPlatformWallet(address newWallet) external onlyOwner {
        require(newWallet != address(0), "Cannot set zero address");
        address oldWallet = platformWallet;
        platformWallet = newWallet;
        emit PlatformWalletUpdated(oldWallet, newWallet);
    }

    /**
     * @dev Withdraw accumulated fees (admin only)
     * @param amount Amount to withdraw
     */
    function withdrawFees(uint256 amount) external onlyOwner {
        require(amount <= totalAccumulatedFees, "Insufficient fee balance");
        
        totalAccumulatedFees -= amount;
        IERC20(paymentToken).safeTransfer(platformWallet, amount);
        
        emit FeesWithdrawn(platformWallet, amount);
    }

    /**
     * @dev Calculate fee amount based on platform fee percentage
     * @param amount Base amount
     * @return Fee amount
     */
    function calculateFee(uint256 amount) internal view returns (uint256) {
        return (amount * platformFeePercentage) / 10000;
    }
}
