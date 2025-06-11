// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title DeployerMarketplace
 * @dev Smart contract for managing Kubernetes cluster deployments, bids, and payments
 */
contract DeployerMarketplace is Ownable, ReentrancyGuard {
    // Events
    event DeploymentRequested(
        string indexed deploymentId,
        address indexed requester,
        string sdlHash,
        uint256 minBid,
        uint256 maxBid,
        uint256 duration
    );

    event BidSubmitted(
        string indexed deploymentId,
        address indexed provider,
        uint256 amount,
        uint256 duration
    );

    event BidSelected(
        string indexed deploymentId,
        address indexed provider,
        uint256 amount
    );

    event BidRejected(
        string indexed deploymentId,
        address indexed provider
    );

    event PaymentLocked(
        string indexed deploymentId,
        address indexed requester,
        address indexed provider,
        uint256 amount
    );

    event PaymentReleased(
        string indexed deploymentId,
        address indexed provider,
        uint256 amount
    );

    event PaymentRefunded(
        string indexed deploymentId,
        address indexed requester,
        uint256 amount
    );

    event DisputeInitiated(
        string indexed deploymentId,
        address indexed initiator,
        string reason
    );

    // Structs
    struct Deployment {
        address requester;
        string sdlHash;
        uint256 minBid;
        uint256 maxBid;
        uint256 duration;
        uint256 createdAt;
        DeploymentStatus status;
        address selectedProvider;
        uint256 selectedBidAmount;
    }

    struct Bid {
        address provider;
        uint256 amount;
        uint256 duration;
        uint256 createdAt;
        BidStatus status;
    }

    // Enums
    enum DeploymentStatus {
        Pending,
        Running,
        Completed,
        Failed,
        Terminated
    }

    enum BidStatus {
        Pending,
        Selected,
        Rejected
    }

    // State variables
    mapping(string => Deployment) public deployments;
    mapping(string => Bid[]) public bids;
    mapping(string => uint256) public escrows;
    mapping(address => bool) public providers;
    mapping(address => uint256) public providerBalances;

    IERC20 public paymentToken;
    uint256 public minBidDuration;
    uint256 public maxBidDuration;
    uint256 public platformFee; // in basis points (1% = 100)

    // Modifiers
    modifier onlyProvider() {
        require(providers[msg.sender], "Not a registered provider");
        _;
    }

    modifier deploymentExists(string memory deploymentId) {
        require(deployments[deploymentId].requester != address(0), "Deployment does not exist");
        _;
    }

    modifier deploymentNotStarted(string memory deploymentId) {
        require(deployments[deploymentId].status == DeploymentStatus.Pending, "Deployment already started");
        _;
    }

    modifier deploymentRunning(string memory deploymentId) {
        require(deployments[deploymentId].status == DeploymentStatus.Running, "Deployment not running");
        _;
    }

    // Constructor
    constructor(
        address _paymentToken,
        uint256 _minBidDuration,
        uint256 _maxBidDuration,
        uint256 _platformFee
    ) {
        require(_paymentToken != address(0), "Invalid payment token address");
        require(_minBidDuration > 0, "Invalid min bid duration");
        require(_maxBidDuration > _minBidDuration, "Invalid max bid duration");
        require(_platformFee <= 1000, "Platform fee too high"); // Max 10%

        paymentToken = IERC20(_paymentToken);
        minBidDuration = _minBidDuration;
        maxBidDuration = _maxBidDuration;
        platformFee = _platformFee;
    }

    // External functions
    function registerProvider() external {
        require(!providers[msg.sender], "Already registered");
        providers[msg.sender] = true;
    }

    function requestDeployment(
        string memory deploymentId,
        string memory sdlHash,
        uint256 minBid,
        uint256 maxBid,
        uint256 duration
    ) external {
        require(deployments[deploymentId].requester == address(0), "Deployment ID already exists");
        require(minBid > 0, "Invalid min bid");
        require(maxBid >= minBid, "Invalid max bid");
        require(duration >= minBidDuration && duration <= maxBidDuration, "Invalid duration");

        deployments[deploymentId] = Deployment({
            requester: msg.sender,
            sdlHash: sdlHash,
            minBid: minBid,
            maxBid: maxBid,
            duration: duration,
            createdAt: block.timestamp,
            status: DeploymentStatus.Pending,
            selectedProvider: address(0),
            selectedBidAmount: 0
        });

        emit DeploymentRequested(deploymentId, msg.sender, sdlHash, minBid, maxBid, duration);
    }

    function submitBid(
        string memory deploymentId,
        uint256 amount,
        uint256 duration
    ) external onlyProvider deploymentExists(deploymentId) deploymentNotStarted(deploymentId) {
        Deployment storage deployment = deployments[deploymentId];
        require(amount >= deployment.minBid && amount <= deployment.maxBid, "Invalid bid amount");
        require(duration >= minBidDuration && duration <= maxBidDuration, "Invalid duration");

        bids[deploymentId].push(Bid({
            provider: msg.sender,
            amount: amount,
            duration: duration,
            createdAt: block.timestamp,
            status: BidStatus.Pending
        }));

        emit BidSubmitted(deploymentId, msg.sender, amount, duration);
    }

    function selectBid(
        string memory deploymentId,
        address provider,
        uint256 bidIndex
    ) external nonReentrant deploymentExists(deploymentId) deploymentNotStarted(deploymentId) {
        Deployment storage deployment = deployments[deploymentId];
        require(msg.sender == deployment.requester, "Only requester can select bid");
        require(bids[deploymentId].length > bidIndex, "Invalid bid index");
        
        Bid storage bid = bids[deploymentId][bidIndex];
        require(bid.provider == provider, "Provider mismatch");
        require(bid.status == BidStatus.Pending, "Bid not pending");

        // Lock payment
        uint256 fee = (bid.amount * platformFee) / 10000;
        uint256 providerAmount = bid.amount - fee;
        
        require(
            paymentToken.transferFrom(msg.sender, address(this), bid.amount),
            "Payment transfer failed"
        );

        deployment.status = DeploymentStatus.Running;
        deployment.selectedProvider = provider;
        deployment.selectedBidAmount = bid.amount;
        bid.status = BidStatus.Selected;
        escrows[deploymentId] = providerAmount;
        providerBalances[provider] += providerAmount;

        emit BidSelected(deploymentId, provider, bid.amount);
        emit PaymentLocked(deploymentId, msg.sender, provider, bid.amount);
    }

    function rejectBid(
        string memory deploymentId,
        address provider,
        uint256 bidIndex
    ) external deploymentExists(deploymentId) deploymentNotStarted(deploymentId) {
        Deployment storage deployment = deployments[deploymentId];
        require(msg.sender == deployment.requester, "Only requester can reject bid");
        require(bids[deploymentId].length > bidIndex, "Invalid bid index");
        
        Bid storage bid = bids[deploymentId][bidIndex];
        require(bid.provider == provider, "Provider mismatch");
        require(bid.status == BidStatus.Pending, "Bid not pending");

        bid.status = BidStatus.Rejected;
        emit BidRejected(deploymentId, provider);
    }

    function releasePayment(string memory deploymentId) external nonReentrant deploymentExists(deploymentId) {
        Deployment storage deployment = deployments[deploymentId];
        require(msg.sender == deployment.requester, "Only requester can release payment");
        require(deployment.status == DeploymentStatus.Running, "Deployment not running");

        uint256 amount = escrows[deploymentId];
        require(amount > 0, "No payment to release");

        deployment.status = DeploymentStatus.Completed;
        escrows[deploymentId] = 0;

        require(
            paymentToken.transfer(deployment.selectedProvider, amount),
            "Payment transfer failed"
        );

        emit PaymentReleased(deploymentId, deployment.selectedProvider, amount);
    }

    function refundPayment(string memory deploymentId) external nonReentrant deploymentExists(deploymentId) {
        Deployment storage deployment = deployments[deploymentId];
        require(msg.sender == deployment.requester, "Only requester can refund payment");
        require(deployment.status == DeploymentStatus.Running, "Deployment not running");

        uint256 amount = escrows[deploymentId];
        require(amount > 0, "No payment to refund");

        deployment.status = DeploymentStatus.Failed;
        escrows[deploymentId] = 0;
        providerBalances[deployment.selectedProvider] -= amount;

        require(
            paymentToken.transfer(deployment.requester, amount),
            "Refund transfer failed"
        );

        emit PaymentRefunded(deploymentId, deployment.requester, amount);
    }

    function initiateDispute(
        string memory deploymentId,
        string memory reason
    ) external deploymentExists(deploymentId) deploymentRunning(deploymentId) {
        Deployment storage deployment = deployments[deploymentId];
        require(
            msg.sender == deployment.requester || msg.sender == deployment.selectedProvider,
            "Only requester or provider can initiate dispute"
        );

        emit DisputeInitiated(deploymentId, msg.sender, reason);
    }

    // View functions
    function getDeployment(string memory deploymentId) external view returns (
        address requester,
        string memory sdlHash,
        uint256 minBid,
        uint256 maxBid,
        uint256 duration,
        uint256 createdAt,
        DeploymentStatus status,
        address selectedProvider,
        uint256 selectedBidAmount
    ) {
        Deployment storage deployment = deployments[deploymentId];
        return (
            deployment.requester,
            deployment.sdlHash,
            deployment.minBid,
            deployment.maxBid,
            deployment.duration,
            deployment.createdAt,
            deployment.status,
            deployment.selectedProvider,
            deployment.selectedBidAmount
        );
    }

    function getBids(string memory deploymentId) external view returns (
        address[] memory providers,
        uint256[] memory amounts,
        uint256[] memory durations,
        uint256[] memory createdAt,
        BidStatus[] memory statuses
    ) {
        Bid[] storage deploymentBids = bids[deploymentId];
        uint256 length = deploymentBids.length;

        providers = new address[](length);
        amounts = new uint256[](length);
        durations = new uint256[](length);
        createdAt = new uint256[](length);
        statuses = new BidStatus[](length);

        for (uint256 i = 0; i < length; i++) {
            providers[i] = deploymentBids[i].provider;
            amounts[i] = deploymentBids[i].amount;
            durations[i] = deploymentBids[i].duration;
            createdAt[i] = deploymentBids[i].createdAt;
            statuses[i] = deploymentBids[i].status;
        }

        return (providers, amounts, durations, createdAt, statuses);
    }

    // Admin functions
    function updatePlatformFee(uint256 newFee) external onlyOwner {
        require(newFee <= 1000, "Platform fee too high"); // Max 10%
        platformFee = newFee;
    }

    function updateBidDurations(uint256 newMinDuration, uint256 newMaxDuration) external onlyOwner {
        require(newMinDuration > 0, "Invalid min duration");
        require(newMaxDuration > newMinDuration, "Invalid max duration");
        minBidDuration = newMinDuration;
        maxBidDuration = newMaxDuration;
    }

    function withdrawFees() external onlyOwner {
        uint256 balance = paymentToken.balanceOf(address(this));
        require(balance > 0, "No fees to withdraw");
        require(paymentToken.transfer(owner(), balance), "Fee transfer failed");
    }
} 