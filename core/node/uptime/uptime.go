package uptime

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	puptime "github.com/unicornultrafoundation/subnet-node/proto/subnet/uptime"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

var log = logrus.New().WithField("service", "uptime")

var UptimeTopic = "topic/uptime"

// HeartbeatMessage defines the structure of heartbeat messages sent over PubSub
type HeartbeatMessage struct {
	Timestamp int64 `json:"timestamp"` // Time the heartbeat message was generated
}

const MerkleAPIURL = "http://localhost:8787/generate-proofs" // Replace with the actual API URL

type ProofResponse struct {
	Root   string     `json:"root"`
	Proofs [][]string `json:"proofs"`
}

// UptimeService manages uptime tracking, PubSub communication, and proof generation
type UptimeService struct {
	IsVerifier     bool
	IsProvider     bool
	Identity       peer.ID        // Local node's identity
	PubSub         *pubsub.PubSub // PubSub instance for communication
	Topic          *pubsub.Topic  // Subscribed PubSub topic
	cancel         context.CancelFunc
	Datastore      repo.Datastore // Datastore for storing uptime records and proofs
	Apps           *apps.Service
	cache          map[string]string // Cache for peer-to-subnet mapping
	AccountService *account.AccountService
}

// Start initializes the UptimeService and starts PubSub-related tasks
func (s *UptimeService) Start() error {
	ctx := context.Background()

	s.cache = map[string]string{}

	// Join the PubSub topic
	topic, err := s.PubSub.Join(UptimeTopic)
	if err != nil {
		return fmt.Errorf("failed to join topic %s: %w", UptimeTopic, err)
	}
	s.Topic = topic

	// Create a cancellable context
	ctx, s.cancel = context.WithCancel(ctx)

	if s.IsProvider || s.IsVerifier {
		// Start publishing, listening, and updating proofs in separate goroutines
		go s.startPublishing(ctx)
		go s.startListening(ctx)
	}

	if s.IsVerifier {
		go s.updateProofs(ctx)
	}

	log.Infof("UptimeService started for topic: %s", UptimeTopic)
	return nil
}

// Stop halts all operations and cleans up resources
func (s *UptimeService) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.Topic != nil {
		if err := s.Topic.Close(); err != nil {
			return fmt.Errorf("failed to close topic: %w", err)
		}
	}
	log.Println("UptimeService stopped")
	return nil
}

// startPublishing periodically sends heartbeat messages to the PubSub topic
func (s *UptimeService) startPublishing(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			return
		case <-ticker.C: // On every tick, send a heartbeat message
			subnetID, err := s.Apps.SubnetRegistry().PeerToSubnet(nil, s.Identity.String())
			if err != nil {
				log.Errorf("failed to fetch subnet ID for peer %s: %v", s.Identity.String(), err)
				continue
			}

			if subnetID == nil {
				log.Infof("subnet ID for peer %s is not registered", s.Identity.String())
				continue
			}

			heartbeatMsg := &puptime.HeartbeatMsg{
				Timestamp: time.Now().Unix(),
			}
			msg := &puptime.Msg{
				Payload: &puptime.Msg_Heartbeat{
					Heartbeat: heartbeatMsg,
				},
			}

			data, err := proto.Marshal(msg)
			if err != nil {
				log.Debugf("Failed to marshal heartbeat message: %v", err)
				continue
			}

			// Update the peer's uptime based on the received message
			if err := s.updatePeerUptime(ctx, s.Identity.String(), heartbeatMsg.Timestamp); err != nil {
				log.Debugf("Error updating peer uptime: %v", err)
				continue
			}
			if err := s.Topic.Publish(ctx, data); err != nil {
				log.Errorf("Failed to publish heartbeat: %v", err)
			} else {
				log.Debugf("Published heartbeat: %+v", msg)
			}
		}
	}
}

// startListening processes incoming heartbeat messages from PubSub
func (s *UptimeService) startListening(ctx context.Context) {
	sub, err := s.Topic.Subscribe()
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}
	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			return
		default:
			msg, err := sub.Next(ctx)
			if err != nil {
				log.Errorf("Error reading message: %v", err)
				continue
			}

			peerId, err := peer.IDFromBytes(msg.From)
			if err != nil {
				log.Debugf("Error parsing peer ID: %v", err)
				continue
			}

			var msgUptime puptime.Msg
			if err := proto.Unmarshal(msg.Data, &msgUptime); err != nil {
				log.Debugf("Failed to unmarshal heartbeat message: %v", err)
				continue
			}

			switch payload := msgUptime.Payload.(type) {
			case *puptime.Msg_Heartbeat:
				s.handleHeartbeatMsg(ctx, peerId, payload.Heartbeat)
			case *puptime.Msg_MerkleProof:
				s.handleMerkeProofMsg(ctx, peerId, payload.MerkleProof)
			default:
				log.Debugf("Unknown payload type")
			}
		}
	}
}

func (s *UptimeService) handleMerkeProofMsg(ctx context.Context, peerId peer.ID, merkeProof *puptime.MerkleProofMsg) {
	if peerId.String() == "12D3KooWGNQYBFWmKgiAgEsQ4u2WznEgR2NmrBbYcfq33yQo4D8a" {
		log.Debugf("Received MerkleProofs: %+v", merkeProof)
		uptime, err := s.GetUptime(ctx)
		if err != nil {
			return
		}

		for _, proof := range merkeProof.Proofs {
			if proof.SubnetId == uptime.SubnetId {
				uptime.IsClaimed = false
				uptime.Proof = &puptime.UptimeProof{
					Uptime: proof.Uptime,
					Proof:  proof.Proof,
				}
				peerKey := datastore.NewKey(fmt.Sprintln("/uptime/peer:" + uptime.PeerId))

				recordData, err := proto.Marshal(uptime)
				if err != nil {
					log.Error(err)
					return
				}

				if err := s.Datastore.Put(ctx, peerKey, recordData); err != nil {
					log.Error(err)
				}
				return
			}
		}
	}
}

func (s *UptimeService) handleHeartbeatMsg(ctx context.Context, peerId peer.ID, heartbeat *puptime.HeartbeatMsg) {
	// Validate that the heartbeat timestamp is less than or equal to the current time
	currentTime := time.Now().Unix()
	if heartbeat.Timestamp > currentTime+5 {
		log.Errorf("Invalid heartbeat timestamp from %s: %d (current time: %d)", peerId.String(), heartbeat.Timestamp, currentTime)
		return
	}

	// Update the peer's uptime based on the received message
	if err := s.updatePeerUptime(ctx, peerId.String(), heartbeat.Timestamp); err != nil {
		log.Errorf("Error updating peer uptime: %v", err)
		return
	}

	log.Debugf("Received Heartbeat: %+v peer: %s", heartbeat, peerId)
}

func (s *UptimeService) updatePeerUptime(ctx context.Context, peerID string, currentTimestamp int64) error {
	// Retrieve existing uptime record from the datastore
	var uptime int64
	var lastTimestamp int64
	peerKey := datastore.NewKey(fmt.Sprintln("/uptime/peer:" + peerID))
	data, err := s.Datastore.Get(ctx, peerKey)
	var proof *puptime.UptimeProof
	isClaimed := false
	if err == nil { // If the record exists in the datastore
		var record puptime.UptimeRecord
		if err := proto.Unmarshal(data, &record); err == nil {
			uptime = record.Uptime
			lastTimestamp = record.LastTimestamp
			proof = record.Proof
			isClaimed = record.IsClaimed
		}
	}

	// Calculate the time elapsed since the last timestamp
	var duration int64
	if lastTimestamp > 0 {
		duration = currentTimestamp - lastTimestamp
	}

	// Only increment uptime if the peer has been online continuously for at least 20 minutes
	if duration <= 1200 {
		uptime += duration
	} else if lastTimestamp > 0 && duration > 0 {
		log.Debugf("Peer %s did not meet the continuous online condition: only %d seconds elapsed", peerID, duration)
	}

	subnetId, err := s.getSubnetID(peerID)

	if err != nil {
		return err
	}

	// Update the record with the latest timestamp and accumulated uptime
	record := &puptime.UptimeRecord{
		PeerId:        peerID,
		Uptime:        uptime,
		LastTimestamp: currentTimestamp,
		SubnetId:      subnetId,
		Proof:         proof,
		IsClaimed:     isClaimed,
	}

	// Save the updated record to the datastore
	recordData, err := proto.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal uptime record for peer %s: %v", peerID, err)
	}
	if err := s.Datastore.Put(ctx, peerKey, recordData); err != nil {
		return fmt.Errorf("failed to update uptime record for peer %s: %v", peerID, err)
	}

	log.Debugf("Updated uptime for PeerID %s: %d seconds, LastTimestamp: %d", peerID, uptime, currentTimestamp)
	return nil
}

// updateProofs periodically generates and distributes Merkle proofs for all uptimes
func (s *UptimeService) updateProofs(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			log.Println("Context cancelled, stopping proof updates.")
			return
		case <-ticker.C: // On every tick, generate and distribute proofs
			go func() {
				if err := s.generateAndDistributeProofs(ctx); err != nil {
					log.Errorf("Error generating and distributing proofs: %v", err)
				} else {
					log.Debugf("Proofs generated and distributed successfully.")
				}
			}()
		}
	}
}

// loadUptimes retrieves all uptime records from the datastore
func (s *UptimeService) loadUptimes(ctx context.Context) ([]*puptime.UptimeRecord, error) {
	query := query.Query{
		Prefix: "/uptime",
	}

	iter, err := s.Datastore.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	entities, err := iter.Rest()
	if err != nil {
		return nil, err
	}
	var records []*puptime.UptimeRecord
	for _, entry := range entities {
		key := string(entry.Key)
		if len(key) > 13 && key[:13] == "/uptime/peer:" {
			peerID := key[13:]
			var record puptime.UptimeRecord
			if err := proto.Unmarshal(entry.Value, &record); err != nil {
				log.Debugf("Failed to unmarshal uptime record for peer %s: %v", peerID, err)
				continue
			}
			records = append(records, &record)
		}
	}
	return records, nil
}
func (s *UptimeService) generateAndDistributeProofs(ctx context.Context) error {
	// Load uptime records from the datastore
	uptimes, err := s.loadUptimes(ctx)
	if err != nil {
		return fmt.Errorf("failed to load uptime records: %v", err)
	}

	// Prepare the payload for the API
	payload, err := json.Marshal(map[string]interface{}{
		"records": prepareRecordsForAPI(uptimes),
	})

	if err != nil {
		return fmt.Errorf("failed to marshal uptime records for API: %v", err)
	}

	// Call the external API
	resp, err := http.Post(MerkleAPIURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to call Merkle API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("merkle API error: %s", string(body))
	}

	// Parse the response from the API
	var response ProofResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read API response: %v", err)
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse API response: %v", err)
	}

	// Save updated records with proofs to the datastore
	for i, proof := range response.Proofs {
		uptime := uptimes[i]
		uptime.Proof = &puptime.UptimeProof{
			Uptime: uptime.Uptime,
			Proof:  proof,
		}
	}

	if err := s.publishAllProofs(ctx, response.Root, uptimes); err != nil {
		return err
	}

	log.Debugf("Merkle root: %s - Proofs successfully stored", response.Root)
	return nil
}

// prepareRecordsForAPI formats uptime records for the Merkle API
func prepareRecordsForAPI(uptimes []*puptime.UptimeRecord) [][]string {
	var records [][]string
	for _, record := range uptimes {
		records = append(records, []string{
			record.SubnetId,
			fmt.Sprintf("%d", record.Uptime),
		})
	}
	return records
}

// GetUptime retrieves the uptime for a specific peer from the datastore
func (s *UptimeService) GetUptime(ctx context.Context) (*puptime.UptimeRecord, error) {
	return s.GetUptimeByPeer(ctx, s.Identity.String())
}

// GetUptime retrieves the uptime for a specific peer from the datastore
func (s *UptimeService) GetUptimeByPeer(ctx context.Context, peerID string) (*puptime.UptimeRecord, error) {
	uptimeKey := datastore.NewKey(fmt.Sprintln("/uptime/peer:" + peerID))

	data, err := s.Datastore.Get(ctx, uptimeKey)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("uptime not found for PeerID: %s", peerID)
		}
		return nil, fmt.Errorf("failed to get uptime for PeerID %s: %v", peerID, err)
	}

	var record puptime.UptimeRecord
	if err := proto.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal uptime for PeerID %s: %v", peerID, err)
	}

	return &record, nil
}

func (s *UptimeService) saveUptime(ctx context.Context, data *puptime.UptimeRecord) error {
	uptimeKey := datastore.NewKey(fmt.Sprintln("/uptime/peer:" + data.PeerId))
	recordData, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return s.Datastore.Put(ctx, uptimeKey, recordData)
}

func (s *UptimeService) getSubnetID(peerID string) (string, error) {
	// Check the cache first
	if subnetID, found := s.cache[peerID]; found {
		return subnetID, nil
	}

	// If not in the cache, fetch from SubnetRegistry
	subnetID, err := s.Apps.SubnetRegistry().PeerToSubnet(nil, peerID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch subnet ID for peer %s: %v", peerID, err)
	}

	// Store the result in the cache
	if subnetID != nil {
		s.cache[peerID] = subnetID.String()
	} else {
		s.cache[peerID] = ""
	}
	return subnetID.String(), nil
}

func (s *UptimeService) GetSubnetID() (*big.Int, error) {
	return s.Apps.SubnetRegistry().PeerToSubnet(nil, s.Identity.String())
}

func (s *UptimeService) ClaimReward(ctx context.Context) error {
	uptime, err := s.GetUptime(ctx)
	if err != nil {
		return nil
	}

	key, err := s.AccountService.NewKeyedTransactor()

	if err != nil {
		return nil
	}

	subnetId, err := s.GetSubnetID()

	if err != nil {
		return err
	}

	proofsBytes := make([][32]byte, 0)

	for _, p := range uptime.Proof.Proof {
		proofBytes, err := hexStringToByte32(p)
		if err != nil {
			return err
		}
		proofsBytes = append(proofsBytes, proofBytes)
	}

	uptime.IsClaimed = true

	_, err = s.Apps.SubnetRegistry().ClaimReward(key, subnetId, big.NewInt(uptime.Proof.Uptime), proofsBytes)
	if err != nil {
		return err
	}

	return s.saveUptime(ctx, uptime)
}

func (s *UptimeService) GetPeerId() (peer.ID, error) {
	return s.Identity, nil
}

func (s *UptimeService) publishAllProofs(ctx context.Context, root string, records []*puptime.UptimeRecord) error {
	// Create the MerkleProof message
	merkleProofMsg := &puptime.MerkleProofMsg{
		Root:   root,
		Proofs: make([]*puptime.Proof, len(records)),
	}

	// Map proofs from API response to protobuf format
	for i, record := range records {
		merkleProofMsg.Proofs[i] = &puptime.Proof{
			Uptime:   record.Uptime, // Add uptime if needed, or keep as 0
			Proof:    record.Proof.Proof,
			SubnetId: record.SubnetId,
		}
	}

	// Wrap the message
	msg := &puptime.Msg{
		Payload: &puptime.Msg_MerkleProof{
			MerkleProof: merkleProofMsg,
		},
	}

	// Serialize the message
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal Merkle proof message: %v", err)
	}

	// Publish the message to the topic
	if err := s.Topic.Publish(ctx, data); err != nil {
		return fmt.Errorf("failed to publish Merkle proof message: %v", err)
	}

	log.Debugf("Published Merkle proof message with root: %s", root)
	rootBytes, err := hexStringToByte32(root)

	if err != nil {
		return err
	}

	key, err := s.AccountService.NewKeyedTransactor()

	if err != nil {
		return err
	}

	_, err = s.AccountService.SubnetRegistry().UpdateMerkleRoot(key, rootBytes)

	if err != nil {
		return err
	}

	return nil
}

func hexStringToByte32(hexStr string) ([32]byte, error) {
	var result [32]byte

	// Loại bỏ prefix "0x" nếu có
	if len(hexStr) >= 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	// Decode hex string thành slice byte
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return result, fmt.Errorf("failed to decode hex string: %v", err)
	}

	// Kiểm tra độ dài (phải chính xác 32 bytes)
	if len(bytes) != 32 {
		return result, fmt.Errorf("invalid length: got %d bytes, expected 32", len(bytes))
	}

	// Copy vào mảng [32]byte
	copy(result[:], bytes)

	return result, nil
}
