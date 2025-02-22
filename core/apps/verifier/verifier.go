package verifier

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"
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

// Verifier is a struct that provides methods to verify resource usage
type Verifier struct {
	ds              datastore.Datastore
	p2p             *p2p.P2P
	acc             *account.AccountService
	fraudulentNodes map[string]bool
	failureCounts   map[string]int
	ps              p2phost.Host // the network host (server+client)
	previousTimes   *lru.Cache
}

// NewVerifier creates a new instance of Verifier
func NewVerifier(ds datastore.Datastore, ps p2phost.Host, P2P *p2p.P2P, acc *account.AccountService) *Verifier {
	cache, _ := lru.New(128)
	v := &Verifier{
		ds:              ds,
		p2p:             P2P,
		acc:             acc,
		ps:              ps,
		fraudulentNodes: make(map[string]bool),
		failureCounts:   make(map[string]int),
		previousTimes:   cache,
	}
	go v.periodicCheck()
	return v
}

func (v *Verifier) Register() error {
	v.ps.SetStreamHandler(atypes.ProtocolAppVerifierUsageReport, v.onUsageReport)
	v.ps.SetStreamHandler(atypes.ProtocolAppSignatureRequest, v.onSignatureRequest)
	v.ps.SetStreamHandler(atypes.ProtocolAppPoWChallenge, v.onPoWChallenge)
	return nil
}

func (v *Verifier) onSignatureRequest(s network.Stream) {
	msg := &pvtypes.SignatureRequest{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, msg)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("%s: Received signature request from %s. Message: %s", s.Conn().LocalPeer(), s.Conn().RemotePeer(), msg)

	usages, err := v.getSignedUsages(msg.AppId, string(s.Conn().RemotePeer()), 30)

	if err != nil {
		log.Printf("Failed to get usage info from database: %v", err)
		return
	}

	if len(usages) > 0 {
		for _, usage := range usages {
			ok := v.sendProtoMessage(s.Conn().RemotePeer(), atypes.ProtocollAppSignatureReceive, &pvtypes.SignatureResponse{
				SignedUsage: usage,
			})
			if !ok {
				log.Printf("Failed to send signature response to %s", s.Conn().RemotePeer())
			} else {
				log.Printf("Sent signature response to %s", s.Conn().RemotePeer())
			}
		}
	}
}

func (v *Verifier) onUsageReport(s network.Stream) {
	msg := &pvtypes.UsageReport{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, msg)
	if err != nil {
		log.Println(err)
		return
	}

	// Validate timestamp
	currentTime := time.Now().Unix()
	reportTime := msg.Timestamp
	if abs(currentTime-reportTime) > 1 {
		log.Printf("Timestamp validation failed: report time %d is not within 1 second of current time %d", reportTime, currentTime)
		return
	}

	// Check if the remote peer matches the usage report peer
	if s.Conn().RemotePeer().String() != msg.PeerId {
		log.Printf("Peer ID mismatch: remote peer %s does not match usage report peer %s", s.Conn().RemotePeer(), msg.PeerId)
		return
	}

	// Check if the previous report time is greater than 25 seconds
	cacheKey := fmt.Sprintf("%d-%s", msg.AppId, msg.PeerId)
	previousReportTime, ok := v.previousTimes.Get(cacheKey)
	if !ok {
		previousReportTime, err = v.getPreviousTimestampFromDB(msg.AppId, msg.PeerId)
		if err != nil {
			log.Printf("Failed to get previous timestamp from database: %v", err)
			return
		}
		v.previousTimes.Add(cacheKey, previousReportTime)
	}

	if time.Since(time.Unix(previousReportTime.(int64), 0)).Seconds() < 30 {
		log.Printf("Previous timestamp validation failed: previous report time %d is less than 25 seconds", previousReportTime)
		return
	}

	// Log usage report to the database
	err = v.logUsageReportToDB(msg)
	if err != nil {
		log.Printf("Failed to log usage report to database: %v", err)
		return
	}

	// Update the cached previous report time
	v.previousTimes.Add(cacheKey, msg.Timestamp)

	log.Printf("%s: Received usage report from %s. Message: %s", s.Conn().LocalPeer(), s.Conn().RemotePeer(), msg)
}

func (v *Verifier) onPoWChallenge(s network.Stream) {
	msg := &pvtypes.PoWChallenge{}
	buf, err := io.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, msg)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("%s: Received PoW challenge from %s. Message: %s", s.Conn().LocalPeer(), s.Conn().RemotePeer(), msg)

	// Generate a PoW challenge
	challenge := v.generatePoWChallenge()

	// Send the challenge to the peer
	ok := v.sendProtoMessage(s.Conn().RemotePeer(), atypes.ProtocolAppPoWChallenge, &pvtypes.PoWChallenge{
		Challenge: challenge,
	})
	if !ok {
		log.Printf("Failed to send PoW challenge to %s", s.Conn().RemotePeer())
	} else {
		log.Printf("Sent PoW challenge to %s", s.Conn().RemotePeer())
	}
}

func (v *Verifier) generatePoWChallenge() string {
	// Generate a random challenge
	challenge := fmt.Sprintf("%d", time.Now().UnixNano())
	return challenge
}

func (v *Verifier) verifyPoWResponse(challenge, response string, difficulty int) bool {
	// Verify the PoW response
	hash := sha256.Sum256([]byte(challenge + response))
	hashInt := new(big.Int).SetBytes(hash[:])
	target := new(big.Int).Lsh(big.NewInt(1), uint(256-difficulty))
	return hashInt.Cmp(target) == -1
}

func (v *Verifier) periodicCheck() {
	for {
		time.Sleep(1 * time.Minute) // Run the check every hour

		usagesByAppId, usageReportIds, err := v.queryUsageReports()
		if err != nil {
			log.Errorf("Failed to query usage reports: %v", err)
			continue
		}

		signedUsages, err := v.processUsageReports(usagesByAppId)
		if err != nil {
			log.Errorf("Failed to process usage reports: %v", err)
			continue
		}

		err = v.saveAndSendSignedUsages(signedUsages, usageReportIds)
		if err != nil {
			log.Errorf("Failed to save and send signed usages: %v", err)
		}
	}
}

func (v *Verifier) queryUsageReports() (map[int64][]*pvtypes.UsageReport, []string, error) {
	query := query.Query{
		Prefix: "/usage_reports/",
	}

	results, err := v.ds.Query(context.Background(), query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query usage info from datastore: %v", err)
	}

	usagesByAppId := make(map[int64][]*pvtypes.UsageReport)
	usageReportIds := make([]string, 0)
	for result := range results.Next() {
		if result.Error != nil {
			return nil, nil, fmt.Errorf("error iterating results: %v", result.Error)
		}

		usage := &pvtypes.UsageReport{}
		if err := proto.Unmarshal(result.Entry.Value, usage); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal usage report: %v", err)
		}

		usageReportIds = append(usageReportIds, result.Entry.Key)

		if len(usagesByAppId[usage.AppId]) == 0 {
			usagesByAppId[usage.AppId] = make([]*pvtypes.UsageReport, 0)
		}
		usagesByAppId[usage.AppId] = append(usagesByAppId[usage.AppId], usage)
	}
	log.Printf("Usage reports: %+v\n", usagesByAppId)
	return usagesByAppId, usageReportIds, nil
}

func (v *Verifier) processUsageReports(usagesByAppId map[int64][]*pvtypes.UsageReport) ([]*pvtypes.SignedUsage, error) {
	signedUsages := make([]*pvtypes.SignedUsage, 0)
	for _, logs := range usagesByAppId {
		// Initialize detector with threshold 2
		detector := AnomalyDetector{Logs: logs, Threshold: 2}
		peerScores := detector.detect()

		usagesByPeer := make(map[string][]*pvtypes.UsageReport)
		for _, log := range logs {
			if peerScores[log.PeerId] > 0 {
				if len(usagesByPeer[log.PeerId]) == 0 {
					usagesByPeer[log.PeerId] = make([]*pvtypes.UsageReport, 0)
				}
				usagesByPeer[log.PeerId] = append(usagesByPeer[log.PeerId], log)
			}
		}
		for _, peerLogs := range usagesByPeer {
			signedUsage := &pvtypes.SignedUsage{}
			for _, peerlog := range peerLogs {
				signedUsage.Cpu += peerlog.Cpu
				signedUsage.Gpu += peerlog.Gpu
				signedUsage.Memory = +peerlog.Memory
				signedUsage.UploadBytes = +peerlog.UploadBytes
				signedUsage.DownloadBytes = +peerlog.DownloadBytes
				signedUsage.Storage = +peerlog.Storage
				signedUsage.PeerId = peerlog.PeerId
				signedUsage.AppId = peerlog.AppId
				signedUsage.ProviderId = peerlog.ProviderId
				signedUsage.Duration += 30
			}
			peerlognum := int64(len(peerLogs))
			signedUsage.Cpu /= peerlognum
			signedUsage.Gpu /= peerlognum
			signedUsage.Memory /= peerlognum
			signedUsage.UploadBytes /= peerlognum
			signedUsage.DownloadBytes /= peerlognum
			signedUsage.Storage /= peerlognum
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

	for _, signedUsage := range signedUsages {
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

	return nil
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func (s *Verifier) signResourceUsage(usage *pvtypes.SignedUsage) error {
	log.Printf("Signing usage: %+v\n", usage)
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
		log.Println(err)
		return false
	}
	defer s.Close()

	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.Println(err)
		s.Reset()
		return false
	}
	return true
}
