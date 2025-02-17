package apps

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/containerd/containerd/namespaces"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	pbapp "github.com/unicornultrafoundation/subnet-node/proto/subnet/app"
)

func (s *Service) startReportLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Infof("Starting report for all running containers...")
			s.reportAllRunningContainers(ctx)
		case <-s.stopChan:
			log.Infof("Stopping report for all containers")
			return
		case <-ctx.Done():
			log.Infof("Context canceled, stopping report")
			return
		}
	}
}

func (s *Service) reportAllRunningContainers(ctx context.Context) {
	// Fetch all running containers
	containers, err := s.containerdClient.Containers(namespaces.WithNamespace(ctx, NAMESPACE))
	if err != nil {
		log.Errorf("Failed to fetch running containers: %v", err)
		return
	}

	for _, container := range containers {
		// Get container ID (assuming appID is same as container ID)
		containerId := container.ID()
		appId, err := atypes.GetAppIdFromContainerId(containerId)
		providerId := big.NewInt(s.accountService.ProviderID())

		if err != nil {
			log.Errorf("Failed to get appId from containerId %s: %v", containerId, err)
			continue
		}

		// Fetch resource usage
		err = s.statService.FinalizeStats(containerId)
		if err != nil {
			log.Errorf("Failed to finalize stats from containerId %s: %v", containerId, err)
			continue
		}

		usageEntry, err := s.statService.GetFinalStats(containerId)
		usage := atypes.ConvertStatEntryToResourceUsage(usageEntry, appId, providerId)
		usage.PeerId = s.peerId.String()

		if err != nil {
			log.Errorf("Failed to get resource usage for container %s: %v", containerId, err)
			continue
		}
		usageProto := atypes.ConvertUsageToProto(*usage)

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

		signature, err := s.requestSignature(ctx, ownerPeerID, PROTOCOL_ID, usageProto)
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
			s.statService.ClearFinalStats(containerId)
		}
	}
}

func (s *Service) requestSignature(ctx context.Context, peerID peer.ID, protoID protocol.ID, usage *pbapp.ResourceUsage) ([]byte, error) {
	// Request signature from app owner's peer
	// Open a stream to the remote peer
	stream, err := s.PeerHost.NewStream(ctx, peerID, protoID)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to open stream to peer %s: %w", peerID, err)
	}
	defer stream.Close()

	// Send the resource usage data
	if err := atypes.SendSignUsageRequest(stream, usage); err != nil {
		return []byte{}, fmt.Errorf("failed to send sign resource usage request: %w", err)
	}

	// Receive the signature response
	response, err := atypes.ReceiveSignatureResponse(stream)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to receive signature response: %w", err)
	}

	return response.Signature, nil
}
