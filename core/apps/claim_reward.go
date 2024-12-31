package apps

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/containerd/containerd/namespaces"
	"github.com/ethereum/go-ethereum/common/math"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	papp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
	pusage "github.com/unicornultrafoundation/subnet-node/proto/subnet/usage"
)

func (s *Service) startRewardClaimer(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Đặt thời gian định kỳ là 1 giờ
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

		// Request signature
		var signature string
		for _, peerId := range s.PeerHost.Network().Peers() {
			signature, err = s.RequestSignature(ctx, peerId, PROTOCOL_ID, usage)
			if err != nil {
				log.Errorf("Failed to get signature from peerId %s for container %s: %v", string(peerId), containerId, err)
				continue
			}
			break
		}

		if len(signature) == 0 {
			log.Errorf("Failed to get signature for container %s: %v", containerId, err)
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

func (s *Service) SignResourceUsage(usage *pusage.ResourceUsage) (string, error) {
	typedData, err := s.ConvertUsageToTypedData(usage)
	if err != nil {
		return "", fmt.Errorf("failed to get usage typed data: %v", err)
	}

	typedDataHash, _, err := TypedDataAndHash(*typedData)
	if err != nil {
		return "", fmt.Errorf("failed to hash typed data: %v", err)
	}

	return s.accountService.SignProtoData(typedDataHash)
}

func (s *Service) RequestSignature(ctx context.Context, peerID peer.ID, protoID protocol.ID, usage *ResourceUsage) (string, error) {
	// Open a stream to the remote peer
	stream, err := s.PeerHost.NewStream(ctx, peerID, protoID)
	if err != nil {
		return "", fmt.Errorf("failed to open stream to peer %s: %w", peerID, err)
	}
	defer stream.Close()

	// Send the resource usage data
	if err := SendUsageProto(stream, convertUsageToProto(*usage)); err != nil {
		return "", fmt.Errorf("failed to send resource usage data: %w", err)
	}

	// Receive the signature response
	response, err := ReceiveSignatureProto(stream)
	if err != nil {
		return "", fmt.Errorf("failed to decode signature response: %w", err)
	}

	return response.Signature, nil
}

func SendUsageProto(s network.Stream, usage *pusage.ResourceUsage) error {
	writer := ggio.NewFullWriter(s)
	err := writer.WriteMsg(usage)

	if err != nil {
		s.Reset()
		return err
	}
	return nil
}

func ReceiveUsageProto(s network.Stream) (*pusage.ResourceUsage, error) {
	data := &pusage.ResourceUsage{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		return nil, fmt.Errorf("failed to receive usage proto: %v", err)
	}

	// unmarshal signature
	err = proto.Unmarshal(buf, data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal usage proto: %v", err)
	}

	return data, nil
}

func ReceiveSignatureProto(s network.Stream) (*papp.SignatureResponse, error) {
	data := &papp.SignatureResponse{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		return nil, fmt.Errorf("failed to receive signature proto: %v", err)
	}

	// unmarshal signature
	err = proto.Unmarshal(buf, data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal signature proto: %v", err)
	}

	return data, nil
}

func (s *Service) ConvertUsageToTypedData(protoUsage *pusage.ResourceUsage) (*TypedData, error) {
	var domainType = []Type{
		{Name: "verifyingContract", Type: "address"},
		{Name: "version", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "chainId", Type: "unit256"},
	}

	version, err := s.subnetAppRegistry.Version(nil)
	if err != nil {
		return nil, err
	}
	chainID := math.HexOrDecimal256(*s.accountService.GetChainID())

	usageTypedData := TypedData{
		Types: Types{
			"EIP712Domain": domainType,
			"Usage": []Type{
				{Name: "subnetId", Type: "bytes"},
				{Name: "appId", Type: "bytes"},
				{Name: "usedCpu", Type: "bytes"},
				{Name: "usedGpu", Type: "bytes"},
				{Name: "usedMemory", Type: "bytes"},
				{Name: "usedStorage", Type: "bytes"},
				{Name: "usedUploadBytes", Type: "bytes"},
				{Name: "usedDownloadBytes", Type: "bytes"},
				{Name: "duration", Type: "bytes"},
			},
		},
		Domain: TypedDataDomain{
			Name:              "SubnetAppRegistry",
			VerifyingContract: s.accountService.GetSubnetAppRegistryAddress(),
			ChainId:           &chainID,
			Version:           version,
		},
		PrimaryType: "ResourceUsage",
		Message: TypedDataMessage{
			"AppId":             protoUsage.AppId,
			"SubnetId":          protoUsage.SubnetId,
			"UsedCpu":           protoUsage.UsedCpu,
			"UsedGpu":           protoUsage.UsedGpu,
			"UsedMemory":        protoUsage.UsedMemory,
			"UsedStorage":       protoUsage.UsedStorage,
			"UsedUploadBytes":   protoUsage.UsedUploadBytes,
			"UsedDownloadBytes": protoUsage.UsedDownloadBytes,
			"Duration":          protoUsage.Duration,
		},
	}
	return &usageTypedData, nil
}
