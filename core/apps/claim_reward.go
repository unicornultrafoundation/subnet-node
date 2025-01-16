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
	pbstream "github.com/unicornultrafoundation/subnet-node/common/io"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
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

		// Get App owner's PeerID
		app, err := s.GetApp(ctx, appId)
		if err != nil {
			log.Errorf("Failed to get app info from appId %s: %v", appId, err)
			continue
		}
		ownerPeerID, err := peer.Decode(app.PeerId)
		if err != nil {
			log.Errorf("Failed to decode peerID %s: %v", app.PeerId, err)
			continue
		}

		signature, err := s.RequestSignature(ctx, ownerPeerID, PROTOCOL_ID, usageProto)
		if err != nil {
			log.Errorf("Failed to get signature from peerId %s for container %s: %v", string(app.PeerId), containerId, err)
			continue
		}

		if len(signature) == 0 {
			log.Errorf("Failed to get signature for container %s. Peers num: %d", containerId, len(s.PeerHost.Network().Peers()))
			continue
		}

		// Claim reward
		txHash, err := s.ReportUsage(ctx, appId, usage.UsedCpu, usage.UsedGpu, usage.UsedMemory, usage.UsedStorage, usage.UsedUploadBytes, usage.UsedDownloadBytes, usage.Duration, signature)
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

func (s *Service) RequestSignature(ctx context.Context, peerID peer.ID, protoID protocol.ID, usage *pbapp.ResourceUsageV2) ([]byte, error) {
	if peerID == s.peerId {
		// You are the app owner. Self-sign signature
		return s.SignResourceUsage(convertUsageFromProto(*usage))
	}

	// Request signature from app owner's peer
	// Open a stream to the remote peer
	stream, err := s.PeerHost.NewStream(ctx, peerID, protoID)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to open stream to peer %s: %w", peerID, err)
	}
	defer stream.Close()

	// Send the resource usage data
	if err := SendSignUsageRequest(stream, usage); err != nil {
		return []byte{}, fmt.Errorf("failed to send sign resource usage request: %w", err)
	}

	// Receive the signature response
	response, err := ReceiveSignatureResponse(stream)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to receive signature response: %w", err)
	}

	return response.Signature, nil
}

func SendSignUsageRequest(s network.Stream, usage *pbapp.ResourceUsageV2) error {
	signatureRequest := pbapp.SignatureRequest{
		Data: &pbapp.SignatureRequest_Usage{
			Usage: usage,
		},
	}

	err := pbstream.WriteProtoBuffered(s, &signatureRequest)
	if err != nil {
		s.Reset()
		return fmt.Errorf("failed to send resource usage sign request: %v", err)
	}

	return nil
}

func ReceiveSignRequest(s network.Stream) (*pbapp.SignatureRequest, error) {
	response := &pbapp.SignatureRequest{}

	err := pbstream.ReadProtoBuffered(s, response)
	if err != nil {
		s.Reset()
		return nil, fmt.Errorf("failed to receive signature request: %v", err)
	}

	return response, nil
}

func ReceiveSignatureResponse(s network.Stream) (*pbapp.SignatureResponse, error) {
	response := &pbapp.SignatureResponse{}

	err := pbstream.ReadProtoBuffered(s, response)
	if err != nil {
		s.Reset()
		return nil, fmt.Errorf("failed to receive signature response: %v", err)
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
				{Name: "providerId", Type: "uint256"},
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
			Name:              "SubnetAppStore",
			Version:           "1",
			ChainId:           &chainID,
			VerifyingContract: s.accountService.AppStoreAddr(),
		},
		PrimaryType: "Usage",
		Message: TypedDataMessage{
			"providerId":        usage.ProviderId,
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
