package apps

import (
	"context"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/containerd/containerd/namespaces"
	ctypes "github.com/docker/docker/api/types/container"
	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/core/apps/verifier"
	pvtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/app/verifier"
)

func (s *Service) startReportLoop(ctx context.Context) {
	ticker := time.NewTicker(verifier.ReportTimeThreshold)
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

	verifierPeerIDs := make([]string, 0)

	for _, container := range containers {
		// Get container ID (assuming appID is same as container ID)
		containerId := container.ID()
		appId, err := atypes.GetAppIdFromContainerId(containerId)
		providerId := big.NewInt(s.accountService.ProviderID())

		if err != nil {
			log.Errorf("Failed to get appId from containerId %s: %v", containerId, err)
			continue
		}

		// Get App owner's PeerID
		app, err := s.GetApp(ctx, appId)
		if err != nil {
			log.Errorf("Failed to get app info from appId %s: %v", appId, err)
			continue
		}
		veriferPeerID, err := peer.Decode(app.PeerId)
		if err != nil {
			log.Errorf("Failed to decode peerID %s: %v", app.PeerId, err)
			continue
		}

		verifierPeerIDs = append(verifierPeerIDs, app.PeerId)

		// Fetch resource usage
		err = s.statService.FinalizeStats(containerId)
		if err != nil {
			log.Errorf("Failed to finalize stats from containerId %s: %v", containerId, err)
			continue
		}

		usageEntry, err := s.statService.GetFinalStats(containerId)

		if err != nil {
			log.Errorf("Failed to get final stats from containerId %s: %v", containerId, err)
			continue
		}

		usage := &pvtypes.UsageReport{
			AppId:         appId.Int64(),
			ProviderId:    providerId.Int64(),
			PeerId:        s.peerId.String(),
			Cpu:           int64(usageEntry.UsedCpu),
			Gpu:           int64(usageEntry.UsedGpu),
			Memory:        int64(usageEntry.UsedMemory),
			Storage:       int64(usageEntry.UsedStorage),
			UploadBytes:   int64(usageEntry.UsedUploadBytes),
			DownloadBytes: int64(usageEntry.UsedDownloadBytes),
			Timestamp:     time.Now().Unix(),
		}

		ok := s.sendProtoMessage(veriferPeerID, atypes.ProtocolAppVerifierUsageReport, usage)
		if !ok {
			log.Errorf("Failed to send usage report for app %d", appId)
		}

		s.statService.ClearFinalStats(containerId)
	}

	// Update verifier peers in Pow
	s.pow.UpdateVerifierPeers(verifierPeerIDs)
}

func (s *Service) reportAllRunningContainersDocker(ctx context.Context) {
	// Fetch all running containers
	containers, err := s.dockerClient.ContainerList(ctx, ctypes.ListOptions{})
	if err != nil {
		return
	}

	verifierPeerIDs := make([]string, 0)

	for _, container := range containers {
		// Get container ID (assuming appID is same as container ID)
		containerId := strings.TrimPrefix(container.Names[0], "/")
		appId, err := atypes.GetAppIdFromContainerId(containerId)
		providerId := big.NewInt(s.accountService.ProviderID())

		if err != nil {
			log.Errorf("Failed to get appId from containerId %s: %v", containerId, err)
			continue
		}

		// Get App owner's PeerID
		app, err := s.GetApp(ctx, appId)
		if err != nil {
			log.Errorf("Failed to get app info from appId %s: %v", appId, err)
			continue
		}
		veriferPeerID, err := peer.Decode(app.PeerId)
		if err != nil {
			log.Errorf("Failed to decode peerID %s: %v", app.PeerId, err)
			continue
		}

		verifierPeerIDs = append(verifierPeerIDs, app.PeerId)

		// Fetch resource usage
		err = s.statService.FinalizeStats(containerId)
		if err != nil {
			log.Errorf("Failed to finalize stats from containerId %s: %v", containerId, err)
			continue
		}

		usageEntry, err := s.statService.GetFinalStats(containerId)

		if err != nil {
			log.Errorf("Failed to get final stats from containerId %s: %v", containerId, err)
			continue
		}

		usage := &pvtypes.UsageReport{
			AppId:         appId.Int64(),
			ProviderId:    providerId.Int64(),
			PeerId:        s.peerId.String(),
			Cpu:           int64(usageEntry.UsedCpu),
			Gpu:           int64(usageEntry.UsedGpu),
			Memory:        int64(usageEntry.UsedMemory),
			Storage:       int64(usageEntry.UsedStorage),
			UploadBytes:   int64(usageEntry.UsedUploadBytes),
			DownloadBytes: int64(usageEntry.UsedDownloadBytes),
			Timestamp:     time.Now().Unix(),
		}

		ok := s.sendProtoMessage(veriferPeerID, atypes.ProtocolAppVerifierUsageReport, usage)
		if !ok {
			log.Errorf("Failed to send usage report for app %d", appId)
		}

		s.statService.ClearFinalStats(containerId)
	}

	// Update verifier peers in Pow
	s.pow.UpdateVerifierPeers(verifierPeerIDs)
}

func (s *Service) onSignatureReceive(stream network.Stream) {
	peerID := stream.Conn().RemotePeer().String()
	if !s.pow.IsVerifierPeer(peerID) {
		log.Warnf("Signature response from non-verifier peer %s rejected", peerID)
		stream.Reset()
		return
	}

	msg := &pvtypes.SignatureResponse{}
	buf, err := io.ReadAll(stream)
	if err != nil {
		stream.Reset()
		log.Error(err)
		return
	}
	stream.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, msg)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("%s: Received signature response from %s. Message: %s", stream.Conn().LocalPeer(), stream.Conn().RemotePeer(), msg)

	// Send the message to the channel for sequential processing
	s.signatureResponseChan <- msg
}

func (s *Service) handleSignatureResponses(ctx context.Context) {
	for {
		select {
		case msg := <-s.signatureResponseChan:
			txHash, err := s.ReportUsage(ctx, msg)
			if err != nil {
				log.Errorf("Failed to report for app %d: %v", msg.SignedUsage.AppId, err)
			} else {
				log.Infof("Report successfully for app %d, transaction hash: %s", msg.SignedUsage.AppId, txHash.Hex())
			}
		case <-ctx.Done():
			log.Infof("Context canceled, stopping handleSignatureResponses")
			return
		case <-s.stopChan:
			log.Infof("Stopping handleSignatureResponses")
			return
		}
	}
}

func (s *Service) sendProtoMessage(id peer.ID, p protocol.ID, data proto.Message) bool {
	stream, err := s.PeerHost.NewStream(context.Background(), id, p)
	if err != nil {
		log.Error(err)
		return false
	}
	defer stream.Close()

	writer := ggio.NewFullWriter(stream)
	err = writer.WriteMsg(data)
	if err != nil {
		log.Error(err)
		stream.Reset()
		return false
	}
	return true
}
