package uptime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
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

// UptimeRecord represents a peer's uptime data and proof
type UptimeRecord struct {
	SubnetId      string   `json:"subnet_id"`
	PeerID        string   `json:"peer_id"`        // Peer ID of the node
	Uptime        int64    `json:"uptime"`         // Total uptime in seconds
	LastTimestamp int64    `json:"last_timestamp"` // Last time the node was seen
	Proof         []string `json:"proof"`          // Merkle proof for this uptime record
}

// UptimeService manages uptime tracking, PubSub communication, and proof generation
type UptimeService struct {
	IsVerifier bool
	IsProvider bool
	Identity   peer.ID        // Local node's identity
	PubSub     *pubsub.PubSub // PubSub instance for communication
	Topic      *pubsub.Topic  // Subscribed PubSub topic
	cancel     context.CancelFunc
	Datastore  repo.Datastore // Datastore for storing uptime records and proofs
	Apps       *apps.Service
	cache      map[string]string // Cache for peer-to-subnet mapping
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

	if s.IsProvider {
		// Start publishing, listening, and updating proofs in separate goroutines
		go s.startPublishing(ctx)
	}

	if s.IsVerifier {
		go s.startListening(ctx)
		go s.updateProofs(ctx)

	}

	log.Printf("UptimeService started for topic: %s", UptimeTopic)
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

			message := HeartbeatMessage{
				Timestamp: time.Now().Unix(),
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to marshal heartbeat message: %v", err)
				continue
			}

			// Update the peer's uptime based on the received message
			if err := s.updatePeerUptime(ctx, s.Identity.String(), message.Timestamp); err != nil {
				log.Printf("Error updating peer uptime: %v", err)
				continue
			}
			if err := s.Topic.Publish(ctx, data); err != nil {
				log.Errorf("Failed to publish heartbeat: %v", err)
			} else {
				log.Debugf("Published heartbeat: %+v", message)
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
				log.Printf("Error reading message: %v", err)
				continue
			}
			var heartbeat HeartbeatMessage
			if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
				log.Printf("Failed to unmarshal heartbeat message: %v", err)
				continue
			}

			peerId, err := peer.IDFromBytes(msg.From)
			if err != nil {
				log.Printf("Error parsing peer ID: %v", err)
				continue
			}

			// Validate that the heartbeat timestamp is less than or equal to the current time
			currentTime := time.Now().Unix()
			if heartbeat.Timestamp > currentTime+5 {
				log.Printf("Invalid heartbeat timestamp from %s: %d (current time: %d)", peerId.String(), heartbeat.Timestamp, currentTime)
				continue
			}

			// Update the peer's uptime based on the received message
			if err := s.updatePeerUptime(ctx, peerId.String(), heartbeat.Timestamp); err != nil {
				log.Printf("Error updating peer uptime: %v", err)
				continue
			}

			log.Debugf("Received heartbeat: %+v peerID: %s", heartbeat, peerId)
		}
	}
}

func (s *UptimeService) updatePeerUptime(ctx context.Context, peerID string, currentTimestamp int64) error {
	// Retrieve existing uptime record from the datastore
	var uptime int64
	var lastTimestamp int64
	peerKey := datastore.NewKey(fmt.Sprintln("/uptime/peer:" + peerID))
	data, err := s.Datastore.Get(ctx, peerKey)
	if err == nil { // If the record exists in the datastore
		var record UptimeRecord
		if err := json.Unmarshal(data, &record); err == nil {
			uptime = record.Uptime
			lastTimestamp = record.LastTimestamp
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
	record := &UptimeRecord{
		PeerID:        peerID,
		Uptime:        uptime,
		LastTimestamp: currentTimestamp,
		SubnetId:      subnetId,
	}

	// Save the updated record to the datastore
	recordData, err := json.Marshal(record)
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
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			log.Println("Context cancelled, stopping proof updates.")
			return
		case <-ticker.C: // On every tick, generate and distribute proofs
			go func() {
				if err := s.generateAndDistributeProofs(ctx); err != nil {
					log.Debugf("Error generating and distributing proofs: %v", err)
				} else {
					log.Debugf("Proofs generated and distributed successfully.")
				}
			}()
		}
	}
}

// loadUptimes retrieves all uptime records from the datastore
func (s *UptimeService) loadUptimes(ctx context.Context) ([]*UptimeRecord, error) {
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
	var records []*UptimeRecord
	for _, entry := range entities {
		key := string(entry.Key)
		if len(key) > 13 && key[:13] == "/uptime/peer:" {
			peerID := key[13:]
			var record UptimeRecord
			if err := json.Unmarshal(entry.Value, &record); err != nil {
				log.Printf("Failed to unmarshal uptime record for peer %s: %v", peerID, err)
				continue
			}
			records = append(records, &record)
		}
	}
	return records, nil
}
func (s *UptimeService) generateAndDistributeProofs(ctx context.Context) error {
	batch, err := s.Datastore.Batch(ctx)
	if err != nil {
		return err
	}

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
		uptimes[i].Proof = proof
		uptimeKey := datastore.NewKey(fmt.Sprintln("/uptime/peer:" + uptimes[i].PeerID))
		data, err := json.Marshal(uptimes[i])
		if err != nil {
			return fmt.Errorf("failed to marshal record with proof: %v", err)
		}
		batch.Put(ctx, uptimeKey, data)
	}

	if err := batch.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit datastore batch: %v", err)
	}

	log.Debugf("Merkle root: %s - Proofs successfully stored", response.Root)
	return nil
}

// prepareRecordsForAPI formats uptime records for the Merkle API
func prepareRecordsForAPI(uptimes []*UptimeRecord) [][]string {
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
func (s *UptimeService) GetUptime(ctx context.Context) (*UptimeRecord, error) {
	return s.GetUptimeByPeer(ctx, s.Identity.String())
}

// GetUptime retrieves the uptime for a specific peer from the datastore
func (s *UptimeService) GetUptimeByPeer(ctx context.Context, peerID string) (*UptimeRecord, error) {
	uptimeKey := datastore.NewKey(fmt.Sprintln("/uptime/peer:" + peerID))

	data, err := s.Datastore.Get(ctx, uptimeKey)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("uptime not found for PeerID: %s", peerID)
		}
		return nil, fmt.Errorf("failed to get uptime for PeerID %s: %v", peerID, err)
	}

	var record UptimeRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal uptime for PeerID %s: %v", peerID, err)
	}

	return &record, nil
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

func (s *UptimeService) GetPeerId() (peer.ID, error) {
	return s.Identity, nil
}
