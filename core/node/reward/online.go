package reward

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

// HeartbeatMessage defines the structure of heartbeat messages
type HeartbeatMessage struct {
	PeerID    string `json:"peer_id"`
	Timestamp int64  `json:"timestamp"`
}

// OnlineService manages online tracking and PubSub communication
type OnlineService struct {
	Identity peer.ID          // Local node's identity
	PubSub   *pubsub.PubSub   // PubSub instance
	Topic    *pubsub.Topic    // Subscribed PubSub topic
	Peers    map[string]int64 // Tracks peers' last seen timestamps
	cancel   context.CancelFunc
}

// Start initializes PubSub communication, starts publishing and listening
func (s *OnlineService) Start(ctx context.Context, topicName string, interval time.Duration) error {
	// Join the PubSub topic
	topic, err := s.PubSub.Join(topicName)
	if err != nil {
		return fmt.Errorf("failed to join topic %s: %w", topicName, err)
	}
	s.Topic = topic
	s.Peers = make(map[string]int64)

	ctx, s.cancel = context.WithCancel(ctx)

	// Start publishing and listening in separate goroutines
	go s.startPublishing(ctx, interval)
	go s.startListening(ctx)

	log.Printf("OnlineService started for topic: %s", topicName)
	return nil
}

// Stop stops all PubSub operations and cleans up resources
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

// startPublishing sends periodic heartbeat messages
func (s *OnlineService) startPublishing(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			message := HeartbeatMessage{
				PeerID:    s.Identity.String(),
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

// startListening processes incoming PubSub messages
func (s *OnlineService) startListening(ctx context.Context) {
	sub, err := s.Topic.Subscribe()
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
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

			// Update peer's last seen timestamp
			s.Peers[heartbeat.PeerID] = heartbeat.Timestamp
			log.Printf("Received heartbeat: %+v", heartbeat)
		}
	}
}
