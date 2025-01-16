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
	puptime "github.com/unicornultrafoundation/subnet-node/proto/subnet/uptime"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

var log = logrus.WithField("module", "uptime")

var UptimeTopic = "topic/uptime"

// HeartbeatMessage defines the structure of heartbeat messages sent over PubSub
type HeartbeatMessage struct {
	Timestamp int64 `json:"timestamp"` // Time the heartbeat message was generated
}

const (
	PublishInterval          = 5 * time.Minute
	UpdateProofsInterval     = 30 * time.Minute
	ReportUptimeInterval     = 2 * time.Hour
	RetrieveVerifierInterval = 5 * time.Minute
)

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
	Datastore      repo.Datastore    // Datastore for storing uptime records and proofs
	cache          map[string]string // Cache for peer-to-subnet mapping
	AccountService *account.AccountService
	verifierPeerID string
	MerkleAPIURL   string
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

	// Retrieve the verifier peer ID once before starting the periodic retrieval
	verifierPeerID, err := s.AccountService.Uptime().VerifierPeerId(nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve verifier peer ID: %w", err)
	}
	s.verifierPeerID = verifierPeerID

	go s.startRetrievingVerifier(ctx)

	if s.IsProvider || s.IsVerifier {
		// Start publishing, listening, and updating proofs in separate goroutines
		go s.startPublishing(ctx)
		go s.startListening(ctx)
	}

	if s.IsVerifier {
		go s.updateProofs(ctx)
	}

	if s.IsProvider {
		go s.startReportingUptime(ctx)
	}

	log.Infof("UptimeService started for topic: %s", UptimeTopic)
	return nil
}

// Stop halts all operations and cleans up resources
func (s *UptimeService) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	log.Println("UptimeService stopped")
	return nil
}

// startPublishing periodically sends heartbeat messages to the PubSub topic
func (s *UptimeService) startPublishing(ctx context.Context) {
	ticker := time.NewTicker(PublishInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			return
		case <-ticker.C: // On every tick, send a heartbeat message
			providerId := s.AccountService.ProviderID()
			if providerId == 0 {
				log.Debugf("subnet ID for peer %s is not registered", s.Identity.String())
				continue
			}

			heartbeatMsg := &puptime.HeartbeatMsg{
				Timestamp:  time.Now().Unix(),
				ProviderId: providerId,
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
			if err := s.updatePeerUptime(ctx, providerId, heartbeatMsg.Timestamp); err != nil {
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
			sub.Cancel()
			return
		default:
			msg, err := sub.Next(ctx)
			if err != nil {
				log.Debugf("Error reading message: %v", err)
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
				s.handleHeartbeatMsg(ctx, payload.Heartbeat)
			case *puptime.Msg_MerkleProof:
				s.handleMerkeProofMsg(ctx, peerId, payload.MerkleProof)
			default:
				log.Debugf("Unknown payload type")
			}
		}
	}
}

// createUptimeKey generates the datastore key for a given provider ID.
func createUptimeKey(providerId int64) datastore.Key {
	return datastore.NewKey(fmt.Sprintf("/uptime/provider/%d", providerId))
}

func (s *UptimeService) handleMerkeProofMsg(ctx context.Context, peerId peer.ID, merkeProof *puptime.MerkleProofMsg) {
	if peerId.String() == s.verifierPeerID {
		log.Debugf("Received MerkleProofs: %+v", merkeProof)
		uptime, err := s.GetUptime(ctx)
		if err != nil {
			return
		}

		for _, proof := range merkeProof.Proofs {
			if proof.ProviderId == uptime.ProviderId {
				uptime.IsClaimed = false
				uptime.Proof = &puptime.UptimeProof{
					Uptime: proof.Uptime,
					Proof:  proof.Proof,
				}
				peerKey := createUptimeKey(uptime.ProviderId)

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

func (s *UptimeService) handleHeartbeatMsg(ctx context.Context, heartbeat *puptime.HeartbeatMsg) {
	// Validate that the heartbeat timestamp is less than or equal to the current time
	currentTime := time.Now().Unix()
	if heartbeat.Timestamp > currentTime+5 {
		log.Errorf("Invalid heartbeat timestamp from provider %d: %d (current time: %d)", heartbeat.ProviderId, heartbeat.Timestamp, currentTime)
		return
	}

	// Validate that the provider ID is not zero
	if heartbeat.ProviderId == 0 {
		log.Debugf("Invalid provider ID: %d", heartbeat.ProviderId)
		return
	}

	// Update the peer's uptime based on the received message
	if err := s.updatePeerUptime(ctx, heartbeat.ProviderId, heartbeat.Timestamp); err != nil {
		log.Debugf("Error updating peer uptime: %v", err)
		return
	}

	log.Debugf("Received Heartbeat: %+v peer: %d", heartbeat, heartbeat.ProviderId)
}

func (s *UptimeService) updatePeerUptime(ctx context.Context, providerId int64, currentTimestamp int64) error {
	// Retrieve existing uptime record from the datastore
	var uptime int64
	var lastTimestamp int64
	peerKey := createUptimeKey(providerId)
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

	if uptime == 0 {
		onchainTotalUptime, err := s.AccountService.Uptime().GetTotalUptime(nil, big.NewInt(providerId))
		if err != nil {
			return err
		}
		uptime = onchainTotalUptime.Int64()
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
		log.Debugf("Provider %d did not meet the continuous online condition: only %d seconds elapsed", providerId, duration)
	}

	// Update the record with the latest timestamp and accumulated uptime
	record := &puptime.UptimeRecord{
		ProviderId:    providerId,
		Uptime:        uptime,
		LastTimestamp: currentTimestamp,
		Proof:         proof,
		IsClaimed:     isClaimed,
	}

	// Save the updated record to the datastore
	recordData, err := proto.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal uptime record for provider %d: %v", providerId, err)
	}
	if err := s.Datastore.Put(ctx, peerKey, recordData); err != nil {
		return fmt.Errorf("failed to update uptime record for provider %d: %v", providerId, err)
	}

	log.Debugf("Updated uptime for ProviderID %d: %d seconds, LastTimestamp: %d", providerId, uptime, currentTimestamp)
	return nil
}

// updateProofs periodically generates and distributes Merkle proofs for all uptimes
func (s *UptimeService) updateProofs(ctx context.Context) {
	ticker := time.NewTicker(UpdateProofsInterval)
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
		Prefix: "/uptime/provider/",
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
		var record puptime.UptimeRecord
		if err := proto.Unmarshal(entry.Value, &record); err != nil {
			log.Debugf("Failed to unmarshal uptime record for key %s: %v", entry.Key, err)
			continue
		}
		records = append(records, &record)
	}
	return records, nil
}

func (s *UptimeService) generateAndDistributeProofs(ctx context.Context) error {
	// Load uptime records from the datastore
	uptimes, err := s.loadUptimes(ctx)
	if err != nil {
		return fmt.Errorf("failed to load uptime records: %v", err)
	}

	if len(uptimes) == 0 {
		log.Debugf("No uptime records found")
		return nil
	}

	// Prepare the payload for the API
	payload, err := json.Marshal(map[string]interface{}{
		"records": prepareRecordsForAPI(uptimes),
	})

	if err != nil {
		return fmt.Errorf("failed to marshal uptime records for API: %v", err)
	}

	// Call the external API
	resp, err := http.Post(s.MerkleAPIURL, "application/json", bytes.NewBuffer(payload))
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
			fmt.Sprintf("%d", record.ProviderId),
			fmt.Sprintf("%d", record.Uptime),
		})
	}
	return records
}

// GetUptime retrieves the uptime for a specific peer from the datastore
func (s *UptimeService) GetUptime(ctx context.Context) (*puptime.UptimeRecord, error) {
	return s.GetUptimeByProvider(ctx, s.AccountService.ProviderID())
}

// GetUptimeByProvider retrieves the uptime for a specific provider from the datastore
func (s *UptimeService) GetUptimeByProvider(ctx context.Context, providerID int64) (*puptime.UptimeRecord, error) {
	uptimeKey := createUptimeKey(providerID)

	data, err := s.Datastore.Get(ctx, uptimeKey)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("uptime not found for ProviderID: %d", providerID)
		}
		return nil, fmt.Errorf("failed to get uptime for ProviderID %d: %v", providerID, err)
	}

	var record puptime.UptimeRecord
	if err := proto.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal uptime for ProviderID %d: %v", providerID, err)
	}

	return &record, nil
}

func (s *UptimeService) saveUptime(ctx context.Context, data *puptime.UptimeRecord) error {
	uptimeKey := createUptimeKey(data.ProviderId)
	recordData, err := proto.Marshal(data)
	if err != nil {
		return err
	}
	return s.Datastore.Put(ctx, uptimeKey, recordData)
}

func (s *UptimeService) ReportUptime(ctx context.Context) error {
	uptime, err := s.GetUptime(ctx)
	if err != nil {
		return nil
	}

	key, err := s.AccountService.NewKeyedTransactor()
	if err != nil {
		return nil
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

	providerId := big.NewInt(s.AccountService.ProviderID())

	_, err = s.AccountService.Uptime().ReportUptime(key, providerId, big.NewInt(uptime.Proof.Uptime), proofsBytes)
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
			Uptime:     record.Uptime, // Add uptime if needed, or keep as 0
			Proof:      record.Proof.Proof,
			ProviderId: record.ProviderId,
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

	_, err = s.AccountService.Uptime().UpdateMerkleRoot(key, rootBytes)

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

// startReportingUptime periodically calls ReportUptime every 12 hours
func (s *UptimeService) startReportingUptime(ctx context.Context) {
	ticker := time.NewTicker(ReportUptimeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			log.Println("Context cancelled, stopping uptime reporting.")
			return
		case <-ticker.C: // On every tick, report uptime
			if err := s.ReportUptime(ctx); err != nil {
				log.Errorf("Error reporting uptime: %v", err)
			} else {
				log.Debugf("Uptime reported successfully.")
			}
		}
	}
}

// startRetrievingVerifier periodically retrieves the verifier peer ID every minute
func (s *UptimeService) startRetrievingVerifier(ctx context.Context) {
	ticker := time.NewTicker(RetrieveVerifierInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			log.Println("Context cancelled, stopping verifier retrieval.")
			return
		case <-ticker.C: // On every tick, retrieve verifier peer ID
			verifierPeerID, err := s.AccountService.Uptime().VerifierPeerId(nil)
			if err != nil {
				log.Errorf("Error retrieving verifier peer ID: %v", err)
			} else {
				s.verifierPeerID = verifierPeerID
				log.Debugf("Verifier peer ID retrieved successfully: %s", verifierPeerID)
			}
		}
	}
}
