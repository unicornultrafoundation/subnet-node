package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Service struct {
	// Self
	Identity peer.ID // the local node's identity
	PubSub   *pubsub.PubSub
	DHT      *ddht.DHT

	UpdateFreq time.Duration

	stopChan    chan struct{}
	pubsubTopic *pubsub.Topic // Reuse the PubSub topic

}

// Start initializes the service and begins periodic updates
func (s *Service) Start() error {
	if s.UpdateFreq == 0 {
		s.UpdateFreq = 30 * time.Second // Default to 30 seconds
	}
	s.stopChan = make(chan struct{})

	// Launch the periodic update loop
	go s.updateLoop()

	log.Println("Resource Service started.")
	return nil
}

// Stop halts the service and releases resources
func (s *Service) Stop() error {
	close(s.stopChan)
	if s.pubsubTopic != nil {
		s.pubsubTopic.Close() // Close the topic when stopping
	}
	log.Println("Service stopped.")
	return nil
}

// Periodically updates resource information
func (s *Service) updateLoop() {
	ticker := time.NewTicker(s.UpdateFreq)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.updateDHTLoop(); err != nil {
				log.Printf("Failed to update resource: %v\n", err)
			}
		}
	}
}

// Updates resource information to DHT and PubSub
func (s *Service) updateDHTLoop() error {
	ctx := context.Background()

	res, err := GetResource()
	fmt.Println(res)
	if err != nil {
		return fmt.Errorf("failed to get resource info: %w", err)
	}

	// Serialize ResourceInfo into JSON
	data, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshal resource info: %w", err)
	}

	// 1. Store resource information in DHT
	key := fmt.Sprintf("/resource/%s", s.Identity.String())
	if err := s.DHT.PutValue(ctx, key, data); err != nil {
		return fmt.Errorf("failed to store resource in DHT: %w", err)
	}

	log.Printf("Updated resource in DHT: %s\n", key)

	// 2. Publish basic resource information to PubSub
	if s.pubsubTopic == nil {
		s.pubsubTopic, err = s.PubSub.Join(res.Topic())
		if err != nil {
			return err
		}
		if err := s.pubsubTopic.Publish(ctx, data); err != nil {
			return fmt.Errorf("failed to publish resource info to pubsub: %w", err)
		}

		log.Println("Published resource info to PubSub.")
	}
	return nil
}
