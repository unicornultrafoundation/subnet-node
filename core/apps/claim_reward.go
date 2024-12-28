package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/containerd/containerd/namespaces"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
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

func (s *Service) RequestSignature(ctx context.Context, peerID peer.ID, protoID protocol.ID, usage *ResourceUsage) (string, error) {
	// Open a stream to the remote peer
	stream, err := s.PeerHost.NewStream(ctx, peerID, protoID)
	if err != nil {
		return "", fmt.Errorf("failed to open stream to peer %s: %w", peerID, err)
	}
	defer stream.Close()

	// Send the resource usage data
	if err := json.NewEncoder(stream).Encode(usage); err != nil {
		return "", fmt.Errorf("failed to send resource usage data: %w", err)
	}

	// Receive the signature response
	var response SignatureResponse
	if err := json.NewDecoder(stream).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode signature response: %w", err)
	}

	return response.Signature, nil
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
