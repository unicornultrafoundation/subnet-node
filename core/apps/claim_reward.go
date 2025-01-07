package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd/namespaces"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	pstream "github.com/unicornultrafoundation/subnet-node/common/io"
	papp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
	pusage "github.com/unicornultrafoundation/subnet-node/proto/subnet/usage"
)

func (s *Service) startRewardClaimer(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Đặt thời gian định kỳ là 1 giờ
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Infof("Starting reward claim for all running containers...")
			s.ClaimRewardsForAllRunningContainers(ctx)
		case <-s.stopChan:
			log.Infof("Stopping reward claimer for all containers")
			return
		case <-ctx.Done():
			log.Infof("Context canceled, stopping reward claimer")
			return
		}
	}
}

func (s *Service) ClaimRewardsForAllRunningContainers(ctx context.Context) {
	// Fetch all running containers
	containers, err := s.containerdClient.Containers(namespaces.WithNamespace(ctx, NAMESPACE))
	if err != nil {
		log.Errorf("Failed to fetch running containers: %v", err)
		return
	}

	for _, container := range containers {
		// Get container ID (assuming appID is same as container ID)
		containerId := container.ID()
		appId, err := getAppIdFromContainerId(containerId)

		if err != nil {
			log.Errorf("Failed to get appId from containerId %s: %v", containerId, err)
			continue
		}

		// Fetch resource usage
		usage, err := s.GetUsage(ctx, appId)
		if err != nil {
			log.Errorf("Failed to get resource usage for container %s: %v", containerId, err)
			continue
		}
		usageProto := convertUsageToProto(*usage)

		// Request signature
		var signature string
		for _, peerId := range s.PeerHost.Network().Peers() {
			signature, err = s.RequestSignature(ctx, peerId, PROTOCOL_ID, usageProto)
			if err != nil {
				log.Errorf("Failed to get signature from peerId %s for container %s: %v", string(peerId), containerId, err)
				continue
			}
			break
		}

		if len(signature) == 0 {
			log.Errorf("Failed to get signature for container %s. Peers num: %d", containerId, len(s.PeerHost.Network().Peers()))
			continue
		}

		// Claim reward
		txHash, err := s.ClaimReward(ctx, appId, usage.UsedCpu, usage.UsedGpu, usage.UsedMemory, usage.UsedStorage, usage.UsedUploadBytes, usage.UsedDownloadBytes, usage.Duration, []byte(signature))
		if err != nil {
			log.Errorf("Failed to claim reward for container %s: %v", containerId, err)
		} else {
			log.Infof("Reward claimed successfully for container %s, transaction hash: %s", containerId, txHash.Hex())
		}
	}
}

func (s *Service) SignResourceUsage(usage *ResourceUsage) ([]byte, error) {
	filledUsage := fillDefaultResourceUsage(usage)
	typedData, err := s.ConvertUsageToTypedData(filledUsage)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to get usage typed data: %v", err)
	}

	typedDataHash, _, err := TypedDataAndHash(*typedData)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to hash typed data: %v", err)
	}

	return s.accountService.Sign(typedDataHash)
}

func (s *Service) RequestSignature(ctx context.Context, peerID peer.ID, protoID protocol.ID, usage *pusage.ResourceUsage) (string, error) {
	// Open a stream to the remote peer
	stream, err := s.PeerHost.NewStream(ctx, peerID, protoID)
	if err != nil {
		return "", fmt.Errorf("failed to open stream to peer %s: %w", peerID, err)
	}
	defer stream.Close()

	// Send the resource usage data
	if err := SendUsage(stream, usage); err != nil {
		return "", fmt.Errorf("failed to send resource usage data: %w", err)
	}

	// Receive the signature response
	response, err := ReceiveSignature(stream)
	if err != nil {
		return "", fmt.Errorf("failed to decode signature response: %w", err)
	}

	return response.Signature, nil
}

func SendUsage(s network.Stream, usage *pusage.ResourceUsage) error {
	err := pstream.WriteProtoBuffered(s, usage)
	if err != nil {
		s.Reset()
		return fmt.Errorf("failed to send resource usage proto: %v", err)
	}

	return nil
}

func ReceiveUsage(s network.Stream) (*pusage.ResourceUsage, error) {
	response := &pusage.ResourceUsage{}

	err := pstream.ReadProtoBuffered(s, response)
	if err != nil {
		s.Reset()
		return nil, fmt.Errorf("failed to receive usage proto: %v", err)
	}

	return response, nil
}

func ReceiveSignature(s network.Stream) (*papp.SignatureResponse, error) {
	response := &papp.SignatureResponse{}

	err := pstream.ReadProtoBuffered(s, response)
	if err != nil {
		s.Reset()
		return nil, fmt.Errorf("failed to receive signature proto: %v", err)
	}

	return response, nil
}

func (s *Service) ConvertUsageToTypedData(usage *ResourceUsage) (*TypedData, error) {
	var domainType = []Type{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	}

	chainID := math.HexOrDecimal256(*s.accountService.GetChainID())

	usageTypedData := TypedData{
		Types: Types{
			"EIP712Domain": domainType,
			"Usage": []Type{
				{Name: "subnetId", Type: "uint256"},
				{Name: "appId", Type: "uint256"},
				{Name: "usedCpu", Type: "uint256"},
				{Name: "usedGpu", Type: "uint256"},
				{Name: "usedMemory", Type: "uint256"},
				{Name: "usedStorage", Type: "uint256"},
				{Name: "usedUploadBytes", Type: "uint256"},
				{Name: "usedDownloadBytes", Type: "uint256"},
				{Name: "duration", Type: "uint256"},
			},
		},
		Domain: TypedDataDomain{
			Name:              "SubnetAppRegistry",
			Version:           "1",
			ChainId:           &chainID,
			VerifyingContract: s.accountService.GetSubnetAppRegistryAddress(),
		},
		PrimaryType: "Usage",
		Message: TypedDataMessage{
			"subnetId":          usage.SubnetId,
			"appId":             usage.AppId,
			"usedCpu":           usage.UsedCpu,
			"usedGpu":           usage.UsedGpu,
			"usedMemory":        usage.UsedMemory,
			"usedStorage":       usage.UsedStorage,
			"usedUploadBytes":   usage.UsedUploadBytes,
			"usedDownloadBytes": usage.UsedDownloadBytes,
			"duration":          usage.Duration,
		},
	}
	return &usageTypedData, nil
}
