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
	IsProvider bool

	resource *ResourceInfo

	stopChan    chan struct{}
	pubsubTopic *pubsub.Topic // Reuse the PubSub topic

}

// Start initializes the service and begins periodic updates
func (s *Service) Start() error {
	if s.UpdateFreq == 0 {
		s.UpdateFreq = 30 * time.Second // Default to 30 seconds
	}
	s.stopChan = make(chan struct{})
	if err := s.updateResourceLoop(); err != nil {
		return err
	}

	if err := s.subscribe(); err != nil {
		return err
	}

	// Launch the periodic update loop
	go s.updateLoop()

	log.Debug("Resource Service started.")
	return nil
}

// Stop halts the service and releases resources
func (s *Service) Stop() error {
	close(s.stopChan)
	if s.pubsubTopic != nil {
		s.pubsubTopic.Close() // Close the topic when stopping
	}
	log.Debug("Service stopped.")
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
			if err := s.updateResourceLoop(); err != nil {
				log.Debugf("Failed to update resource: %v\n", err)
			}

			if s.IsProvider {
				if err := s.updateDHTLoop(); err != nil {
					log.Debugf("Failed to update dht: %v\n", err)
				}
			}
		}
	}
}

func (s *Service) updateResourceLoop() error {
	res, err := GetResource()
	if err != nil {
		return fmt.Errorf("failed to get resource info: %w", err)
	}

	s.resource = res
	return nil
}

func (s *Service) GetResource() *ResourceInfo {
	return s.resource
}

func (s *Service) subscribe() error {
	if !s.IsProvider {
		return nil
	}
	// 2. Publish basic resource information to PubSub
	if s.pubsubTopic == nil {
		data, err := json.Marshal(s.resource)
		if err != nil {
			return fmt.Errorf("failed to marshal resource info: %w", err)
		}
		s.pubsubTopic, err = s.PubSub.Join(s.resource.Topic())
		if err != nil {
			return err
		}
		if err := s.pubsubTopic.Publish(context.Background(), data); err != nil {
			return fmt.Errorf("failed to publish resource info to pubsub: %w", err)
		}

		log.Debug("Published resource info to PubSub.")
		_, err = s.pubsubTopic.Subscribe(pubsub.WithBufferSize(1))
		if err != nil {
			return err
		}

	}
	return nil
}

// Updates resource information to DHT and PubSub
func (s *Service) updateDHTLoop() error {
	ctx := context.Background()

	// Serialize ResourceInfo into JSON
	data, err := json.Marshal(s.resource)
	if err != nil {
		return fmt.Errorf("failed to marshal resource info: %w", err)
	}

	// 1. Store resource information in DHT
	key := fmt.Sprintf("/resource/%s", s.Identity.String())
	if err := s.DHT.PutValue(ctx, key, data); err != nil {
		return fmt.Errorf("failed to store resource in DHT: %w", err)
	}

	log.Debugf("Updated resource in DHT: %s\n", key)
	return nil
}
