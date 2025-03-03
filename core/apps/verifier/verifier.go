package verifier

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	ggio "github.com/gogo/protobuf/io"
	proto "github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	atypes "github.com/unicornultrafoundation/subnet-node/core/apps/types"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	pvtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/app/verifier"
)

var log = logrus.WithField("service", "app-verifier")

const (
	ReportTimeThreshold  = 1 * time.Minute
	DefaultCheckInterval = 1 * time.Hour // Default check interval is 1 hour
)

// Verifier is a struct that provides methods to verify resource usage
type Verifier struct {
	ds            datastore.Datastore
	p2p           *p2p.P2P
	acc           *account.AccountService
	ps            p2phost.Host // the network host (server+client)
	previousTimes *lru.Cache
	pow           *Pow
}

// NewVerifier creates a new instance of Verifier
func NewVerifier(ds datastore.Datastore, ps p2phost.Host, P2P *p2p.P2P, acc *account.AccountService) *Verifier {
	cache, _ := lru.New(128)
	v := &Verifier{
		ds:            ds,
		p2p:           P2P,
		acc:           acc,
		ps:            ps,
		previousTimes: cache,
		pow:           NewPow(NodeVerifier, ps, P2P),
	}
	go v.periodicCheck(DefaultCheckInterval) // Pass the default check interval
	return v
}

func (v *Verifier) Register() error {
	v.ps.SetStreamHandler(atypes.ProtocolAppVerifierUsageReport, v.onUsageReport)
	v.ps.SetStreamHandler(atypes.ProtocolAppSignatureRequest, v.onSignatureRequest)
	return nil
}

func (v *Verifier) onSignatureRequest(s network.Stream) {
	msg := &pvtypes.SignatureRequest{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Error(err)
		return
	}
	s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, msg)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("%s: Received signature request from %s. Message: %s", s.Conn().LocalPeer(), s.Conn().RemotePeer(), msg)

	usages, err := v.getSignedUsages(msg.AppId, string(s.Conn().RemotePeer()), 30)

	if err != nil {
		log.Errorf("Failed to get usage info from database: %v", err)
		return
	}

	if len(usages) > 0 {
		for _, usage := range usages {
			ok := v.sendProtoMessage(s.Conn().RemotePeer(), atypes.ProtocollAppSignatureReceive, &pvtypes.SignatureResponse{
				SignedUsage: usage,
			})
			if !ok {
				log.Warnf("Failed to send signature response to %s", s.Conn().RemotePeer())
			} else {
				log.Infof("Sent signature response to %s", s.Conn().RemotePeer())
			}
		}
	}
}

func (v *Verifier) onUsageReport(s network.Stream) {
	currentTime := time.Now().Unix()
	msg := &pvtypes.UsageReport{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Error(err)
		return
	}
	s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, msg)
	if err != nil {
		log.Error(err)
		return
	}

	// Validate timestamp
	reportTime := msg.Timestamp
	if abs(currentTime-reportTime) > 2 {
		log.Warnf("Timestamp validation failed: report time %d is not within 2 second of current time %d", reportTime, currentTime)
		return
	}

	// Check if the remote peer matches the usage report peer
	if s.Conn().RemotePeer().String() != msg.PeerId {
		log.Warnf("Peer ID mismatch: remote peer %s does not match usage report peer %s", s.Conn().RemotePeer(), msg.PeerId)
		return
	}

	// Check if the previous report time is greater than 25 seconds
	cacheKey := fmt.Sprintf("%d-%s", msg.AppId, msg.PeerId)
	previousReportTime, ok := v.previousTimes.Get(cacheKey)
	if !ok {
		previousReportTime, err = v.getPreviousTimestampFromDB(msg.AppId, msg.PeerId)
		if err != nil {
			log.Errorf("Failed to get previous timestamp from database: %v", err)
			return
		}
		v.previousTimes.Add(cacheKey, previousReportTime)
	}

	if msg.Timestamp-previousReportTime.(int64) < int64(ReportTimeThreshold.Seconds()) {
		log.Warnf("Previous timestamp validation failed: previous report time %d is less than %s. Current report time: %d", previousReportTime, ReportTimeThreshold, msg.Timestamp)
		return
	}

	// Log usage report to the database
	err = v.logUsageReportToDB(msg)
	if err != nil {
		log.Errorf("Failed to log usage report to database: %v", err)
		return
	}

	// Update the cached previous report time
	v.previousTimes.Add(cacheKey, msg.Timestamp)

	log.Infof("%s: Received usage report from %s. Message: %s", s.Conn().LocalPeer(), s.Conn().RemotePeer(), msg)
}

func (v *Verifier) periodicCheck(interval time.Duration) {
	for {
		time.Sleep(interval) // Use the provided interval for the check

		usagesByAppId, usageReportIds, uniquePeerIds, err := v.queryUsageReports()
		if err != nil {
			log.Errorf("Failed to query usage reports: %v", err)
			continue
		}

		v.pow.RequestPoWFromPeers(uniquePeerIds, 10)
		time.Sleep(1 * time.Minute)

		signedUsages, err := v.processUsageReports(usagesByAppId)
		if err != nil {
			log.Errorf("Failed to process usage reports: %v", err)
			continue
		}

		v.pow.Clear()

		err = v.saveAndSendSignedUsages(signedUsages, usageReportIds)
		if err != nil {
			log.Errorf("Failed to save and send signed usages: %v", err)
		}

		log.Infof("Unique peer IDs: %v", uniquePeerIds)
	}
}

func (v *Verifier) queryUsageReports() (map[int64][]*pvtypes.UsageReport, []string, []peer.ID, error) {
	query := query.Query{
		Prefix: "/usage_reports/",
	}

	results, err := v.ds.Query(context.Background(), query)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to query usage info from datastore: %v", err)
	}

	usagesByAppId := make(map[int64][]*pvtypes.UsageReport)
	usageReportIds := make([]string, 0)
	uniquePeerIds := make(map[string]struct{})
	for result := range results.Next() {
		if result.Error != nil {
			return nil, nil, nil, fmt.Errorf("error iterating results: %v", result.Error)
		}

		usage := &pvtypes.UsageReport{}
		if err := proto.Unmarshal(result.Entry.Value, usage); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to unmarshal usage report: %v", err)
		}

		usageReportIds = append(usageReportIds, result.Entry.Key)
		uniquePeerIds[usage.PeerId] = struct{}{}

		if len(usagesByAppId[usage.AppId]) == 0 {
			usagesByAppId[usage.AppId] = make([]*pvtypes.UsageReport, 0)
		}
		usagesByAppId[usage.AppId] = append(usagesByAppId[usage.AppId], usage)
	}
	log.Debugf("Usage reports: %+v\n", usagesByAppId)

	peerIds := make([]peer.ID, 0, len(uniquePeerIds))
	for peerId := range uniquePeerIds {
		parsedPeerId, _ := peer.Decode(peerId)
		peerIds = append(peerIds, parsedPeerId)
	}

	return usagesByAppId, usageReportIds, peerIds, nil
}

func (v *Verifier) processUsageReports(usagesByAppId map[int64][]*pvtypes.UsageReport) ([]*pvtypes.SignedUsage, error) {
	signedUsages := make([]*pvtypes.SignedUsage, 0)

	for _, logs := range usagesByAppId {
		// Filter logs by qualified peers
		filteredLogs := make([]*pvtypes.UsageReport, 0)
		for _, log := range logs {
			if v.pow.IsPeerQualified(log.PeerId) {
				filteredLogs = append(filteredLogs, log)
			}
		}

		// Initialize detector with threshold 2
		detector := AnomalyDetector{Logs: filteredLogs, Threshold: 2}
		peerScores := detector.detect()

		usagesByPeer := make(map[string][]*pvtypes.UsageReport)
		for _, log := range filteredLogs {
			if peerScores[log.PeerId] > 0 {
				if len(usagesByPeer[log.PeerId]) == 0 {
					usagesByPeer[log.PeerId] = make([]*pvtypes.UsageReport, 0)
				}
				usagesByPeer[log.PeerId] = append(usagesByPeer[log.PeerId], log)
			}
		}
		for peerId, peerLogs := range usagesByPeer {
			if len(peerLogs) == 0 {
				continue
			}

			// Check if peerLogs come from different providers
			providerSet := make(map[int64]struct{})
			for _, peerlog := range peerLogs {
				providerSet[peerlog.ProviderId] = struct{}{}
			}
			if len(providerSet) > 1 {
				log.Warnf("Skipping signing for peer %s as logs come from multiple providers", peerId)
				continue
			}

			signedUsage := &pvtypes.SignedUsage{}
			for _, peerlog := range peerLogs {
				signedUsage.Cpu += peerlog.Cpu
				signedUsage.Gpu += peerlog.Gpu
				signedUsage.Memory += peerlog.Memory
				signedUsage.UploadBytes += peerlog.UploadBytes
				signedUsage.DownloadBytes += peerlog.DownloadBytes
				signedUsage.Storage += peerlog.Storage
				signedUsage.PeerId = peerlog.PeerId
				signedUsage.AppId = peerlog.AppId
				signedUsage.ProviderId = peerlog.ProviderId
				signedUsage.Duration += int64(ReportTimeThreshold.Seconds())
			}
			peerlognum := int64(len(peerLogs))

			if peerlognum > 0 {
				signedUsage.Cpu /= peerlognum
				signedUsage.Gpu /= peerlognum
				signedUsage.Memory /= peerlognum
				signedUsage.UploadBytes /= peerlognum
				signedUsage.DownloadBytes /= peerlognum
				signedUsage.Storage /= peerlognum
			}

			signedUsage.Timestamp = time.Now().Unix()
			if err := v.signResourceUsage(signedUsage); err != nil {
				return nil, fmt.Errorf("failed to sign resource usage: %v", err)
			}

			signedUsages = append(signedUsages, signedUsage)
		}
	}
	return signedUsages, nil
}

func (v *Verifier) saveAndSendSignedUsages(signedUsages []*pvtypes.SignedUsage, usageReportIds []string) error {
	workerCount := runtime.NumCPU() // Use the number of CPUs available on the machine

	b := datastore.NewBasicBatch(v.ds)

	err := v.saveSignedUsages(b, signedUsages)
	if err != nil {
		return fmt.Errorf("failed to save signed usages: %v", err)
	}

	for _, id := range usageReportIds {
		b.Delete(context.Background(), datastore.NewKey(id))
	}

	err = b.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("failed to commit batch: %v", err)
	}

	return v.sendSignedUsages(signedUsages, workerCount)
}

func (v *Verifier) sendSignedUsages(signedUsages []*pvtypes.SignedUsage, workerCount int) error {
	var wg sync.WaitGroup
	jobs := make(chan *pvtypes.SignedUsage, len(signedUsages))

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for signedUsage := range jobs {
				peerID, err := peer.Decode(signedUsage.PeerId)
				if err != nil {
					log.Errorf("Failed to decode peerID %s: %v", peerID, err)
					continue
				}

				ok := v.sendProtoMessage(peerID, atypes.ProtocollAppSignatureReceive, &pvtypes.SignatureResponse{
					SignedUsage: signedUsage,
				})
				if !ok {
					log.Errorf("failed to send signed usage to peer %s", signedUsage.PeerId)
				} else {
					log.Infof("Sent signed usage to peer %s", signedUsage.PeerId)
				}
			}
		}()
	}

	for _, signedUsage := range signedUsages {
		jobs <- signedUsage
	}
	close(jobs)

	wg.Wait()
	return nil
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func (s *Verifier) signResourceUsage(usage *pvtypes.SignedUsage) error {
	log.Debugf("Signing usage: %+v\n", usage)
	typedData, err := atypes.ConvertUsageToTypedData(usage, s.acc.GetChainID(), s.acc.AppStoreAddr())
	if err != nil {
		return fmt.Errorf("failed to get usage typed data: %v", err)
	}
	hash, signature, err := s.acc.SignTypedData(typedData)
	if err != nil {
		return fmt.Errorf("failed to sign usage: %v", err)
	}

	usage.Signature = signature
	usage.Hash = hash

	return nil
}

func (v *Verifier) sendProtoMessage(id peer.ID, p protocol.ID, data proto.Message) bool {
	s, err := v.ps.NewStream(context.Background(), id, p)
	if err != nil {
		log.Error(err)
		return false
	}
	defer s.Close()

	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.Error(err)
		s.Reset()
		return false
	}
	return true
}
