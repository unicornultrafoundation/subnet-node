package reward

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/crypto/merkle"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

// HeartbeatMessage defines the structure of heartbeat messages sent over PubSub
type HeartbeatMessage struct {
	Timestamp int64 `json:"timestamp"` // Time the heartbeat message was generated
}

// UptimeRecord represents a peer's uptime data and proof
type UptimeRecord struct {
	PeerID        peer.ID       `json:"peer_id"`        // Peer ID of the node
	Uptime        int64         `json:"uptime"`         // Total uptime in seconds
	LastTimestamp int64         `json:"last_timestamp"` // Last time the node was seen
	Proof         *merkle.Proof `json:"proof"`          // Merkle proof for this uptime record
}

// OnlineService manages uptime tracking, PubSub communication, and proof generation
type OnlineService struct {
	Identity  peer.ID          // Local node's identity
	PubSub    *pubsub.PubSub   // PubSub instance for communication
	Topic     *pubsub.Topic    // Subscribed PubSub topic
	Peers     map[string]int64 // Tracks peers' last seen timestamps
	cancel    context.CancelFunc
	Datastore repo.Datastore // Datastore for storing uptime records and proofs
}

// Start initializes the OnlineService and starts PubSub-related tasks
func (s *OnlineService) Start(ctx context.Context, topicName string, interval time.Duration) error {
	// Join the PubSub topic
	topic, err := s.PubSub.Join(topicName)
	if err != nil {
		return fmt.Errorf("failed to join topic %s: %w", topicName, err)
	}
	s.Topic = topic
	s.Peers = make(map[string]int64)

	// Create a cancellable context
	ctx, s.cancel = context.WithCancel(ctx)

	// Start publishing, listening, and updating proofs in separate goroutines
	go s.startPublishing(ctx, interval)
	go s.startListening(ctx)
	go s.updateProofs(ctx)

	log.Printf("OnlineService started for topic: %s", topicName)
	return nil
}

// Stop halts all operations and cleans up resources
func (s *OnlineService) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.Topic != nil {
		if err := s.Topic.Close(); err != nil {
			return fmt.Errorf("failed to close topic: %w", err)
		}
	}
	log.Println("OnlineService stopped")
	return nil
}

// startPublishing periodically sends heartbeat messages to the PubSub topic
func (s *OnlineService) startPublishing(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			return
		case <-ticker.C: // On every tick, send a heartbeat message
			message := HeartbeatMessage{
				Timestamp: time.Now().Unix(),
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to marshal heartbeat message: %v", err)
				continue
			}

			if err := s.Topic.Publish(ctx, data); err != nil {
				log.Printf("Failed to publish heartbeat: %v", err)
			} else {
				log.Printf("Published heartbeat: %+v", message)
			}
		}
	}
}

// startListening processes incoming heartbeat messages from PubSub
func (s *OnlineService) startListening(ctx context.Context) {
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

			// Update the peer's uptime based on the received message
			if err := s.updatePeerUptime(ctx, peerId.String(), time.Now().Unix()); err != nil {
				log.Printf("Error updating peer uptime: %v", err)
				continue
			}

			log.Printf("Received heartbeat: %+v", heartbeat)
		}
	}
}

func (s *OnlineService) updatePeerUptime(ctx context.Context, peerID string, currentTimestamp int64) error {
	// Retrieve existing uptime record from the datastore
	var uptime int64
	var lastTimestamp int64
	peerKey := datastore.NewKey("uptime:" + peerID)

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
	if duration >= 1200 {
		uptime += duration
	} else if lastTimestamp > 0 && duration > 0 {
		log.Printf("Peer %s did not meet the continuous online condition: only %d seconds elapsed", peerID, duration)
	}

	// Update the record with the latest timestamp and accumulated uptime
	record := &UptimeRecord{
		PeerID:        peer.ID(peerID),
		Uptime:        uptime,
		LastTimestamp: currentTimestamp,
	}

	// Save the updated record to the datastore
	recordData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal uptime record for peer %s: %v", peerID, err)
	}
	if err := s.Datastore.Put(ctx, peerKey, recordData); err != nil {
		return fmt.Errorf("failed to update uptime record for peer %s: %v", peerID, err)
	}

	log.Printf("Updated uptime for PeerID %s: %d seconds, LastTimestamp: %d", peerID, uptime, currentTimestamp)
	return nil
}

// updateProofs periodically generates and distributes Merkle proofs for all uptimes
func (s *OnlineService) updateProofs(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): // Exit if the context is cancelled
			log.Println("Context cancelled, stopping proof updates.")
			return
		case <-ticker.C: // On every tick, generate and distribute proofs
			go func() {
				if err := s.generateAndDistributeProofs(ctx); err != nil {
					log.Printf("Error generating and distributing proofs: %v", err)
				} else {
					log.Println("Proofs generated and distributed successfully.")
				}
			}()
		}
	}
}

// loadUptimes retrieves all uptime records from the datastore
func (s *OnlineService) loadUptimes(ctx context.Context) ([]*UptimeRecord, error) {
	query := query.Query{
		Prefix: "uptime:",
	}

	iter, err := s.Datastore.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var records []*UptimeRecord
	for {
		result, ok := iter.NextSync()
		if !ok {
			break
		}

		key := string(result.Key)
		if len(key) > 7 && key[:7] == "uptime:" {
			peerID := key[7:]
			var record UptimeRecord
			if err := json.Unmarshal(result.Value, &record); err != nil {
				log.Printf("Failed to unmarshal uptime record for peer %s: %v", peerID, err)
				continue
			}
			records = append(records, &record)
		}
	}
	return records, nil
}

// generateAndDistributeProofs creates Merkle proofs for all peers and stores them in the datastore
func (s *OnlineService) generateAndDistributeProofs(ctx context.Context) error {
	batch, err := s.Datastore.Batch(ctx)
	if err != nil {
		return err
	}

	// Retrieve all uptime records
	uptimes, err := s.loadUptimes(ctx)
	if err != nil {
		return err
	}

	uptimeBytes := make([][]byte, len(uptimes))
	for i, uptime := range uptimes {
		data, err := json.Marshal(uptime)
		if err != nil {
			return fmt.Errorf("failed to marshal uptime record: %v", err)
		}
		uptimeBytes[i] = data
	}

	// Generate Merkle proofs
	rootHash, proofs := merkle.ProofsFromByteSlices(uptimeBytes)
	for i, proof := range proofs {
		uptimes[i].Proof = proof
		data, err := json.Marshal(uptimes[i])
		if err != nil {
			return fmt.Errorf("failed to marshal proof store: %v", err)
		}

		proofKey := datastore.NewKey("proof:" + uptimes[i].PeerID.String())
		batch.Put(ctx, proofKey, data)
	}

	err = batch.Commit(ctx)
	if err != nil {
		return err
	}
	log.Printf("Generated and stored Merkle Root: %s", rootHash)

	return nil
}

// GetProof retrieves a Merkle proof for a specific peer from the datastore
func (s *OnlineService) GetProof(ctx context.Context, peerID string) (*merkle.Proof, error) {
	proofKey := datastore.NewKey("proof:" + peerID)

	data, err := s.Datastore.Get(ctx, proofKey)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, fmt.Errorf("proof not found for PeerID: %s", peerID)
		}
		return nil, fmt.Errorf("failed to get proof for PeerID %s: %v", peerID, err)
	}

	var record UptimeRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proof for PeerID %s: %v", peerID, err)
	}

	return record.Proof, nil
}

// GetUptime retrieves the uptime for a specific peer from the datastore
func (s *OnlineService) GetUptime(ctx context.Context, peerID string) (int64, error) {
	uptimeKey := datastore.NewKey("uptime:" + peerID)

	data, err := s.Datastore.Get(ctx, uptimeKey)
	if err != nil {
		if err == datastore.ErrNotFound {
			return 0, fmt.Errorf("uptime not found for PeerID: %s", peerID)
		}
		return 0, fmt.Errorf("failed to get uptime for PeerID %s: %v", peerID, err)
	}

	var record UptimeRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return 0, fmt.Errorf("failed to unmarshal uptime for PeerID %s: %v", peerID, err)
	}

	return record.Uptime, nil
}
