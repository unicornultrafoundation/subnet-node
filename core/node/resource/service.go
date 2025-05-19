package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	ddht "github.com/libp2p/go-libp2p-kad-dht/dual"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Service struct {
	// Self
	Identity peer.ID // the local node's identity
	PubSub   *pubsub.PubSub
	DHT      *ddht.DHT

	UpdateFreq  time.Duration
	stopChan    chan struct{}
	pubsubTopic *pubsub.Topic // Reuse the PubSub topic
	DS          datastore.Datastore
	mu          sync.Mutex // Add a mutex for thread-safe access
}

// Start initializes the service and begins periodic updates
func (s *Service) Start() error {
	s.UpdateFreq = 24 * time.Hour * 30 // Default to 30 days
	s.stopChan = make(chan struct{})

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

func (s *Service) PeerId() peer.ID {
	return s.Identity
}

// Periodically updates resource information
func (s *Service) updateLoop() {
	ticker := time.NewTicker(s.UpdateFreq)
	defer ticker.Stop()

	if err := s.updateDHTLoop(); err != nil {
		log.Debugf("Failed to update resource: %v\n", err)
	}

	// if err := s.subscribe(); err != nil {
	// 	log.Debugf("Failed to subscribe: %v\n", err)
	// }

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.updateDHTLoop(); err != nil {
				log.Debugf("Failed to update dht: %v\n", err)
			}
		}
	}
}

// Updates resource information to DHT and PubSub
func (s *Service) updateDHTLoop() error {
	ctx := context.Background()

	res, err := s.GetResource()
	if err != nil {
		return fmt.Errorf("failed to get resource: %w", err)
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

	log.Debugf("Updated resource in DHT: %s\n", key)

	// if err := s.pubsubTopic.Publish(context.Background(), data); err != nil {
	// 	return fmt.Errorf("failed to publish resource info to pubsub: %w", err)
	// }

	// log.Debug("Published resource info to PubSub.")

	return nil
}

func (s *Service) saveResource(resource *ResourceInfo) error {
	// Serialize ResourceInfo into JSON
	data, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal resource info: %w", err)
	}

	// Create a datastore key using the peer ID
	key := datastore.NewKey(fmt.Sprintf("resource/%s", s.Identity.String()))

	// Save the serialized data to the datastore
	if err := s.DS.Put(context.Background(), key, data); err != nil {
		return fmt.Errorf("failed to save resource to datastore: %w", err)
	}

	log.Debugf("Saved resource to datastore with key: %s", key.String())
	return nil
}

func (s *Service) GetResource() (*ResourceInfo, error) {
	s.mu.Lock()         // Lock before accessing the datastore
	defer s.mu.Unlock() // Unlock after the operation is complete

	// Create a datastore key using the peer ID
	key := datastore.NewKey(fmt.Sprintf("resource/%s", s.Identity.String()))

	// Check if the resource exists in the datastore
	data, err := s.DS.Get(context.Background(), key)
	if err == datastore.ErrNotFound {
		log.Debugf("Resource not found in datastore, fetching new resource.")
		return s.fetchAndSaveResource()
	} else if err != nil {
		return nil, fmt.Errorf("failed to get resource from datastore: %w", err)
	}

	// Deserialize the resource
	var resource ResourceInfo
	if err := json.Unmarshal(data, &resource); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource info: %w", err)
	}

	// Check the last update timestamp
	lastUpdate := resource.LastUpdate
	if time.Since(lastUpdate) >= 30*24*time.Hour {
		log.Debugf("Resource is older than 30 days, deleting and fetching new resource.")
		if err := s.DS.Delete(context.Background(), key); err != nil {
			return nil, fmt.Errorf("failed to delete old resource from datastore: %w", err)
		}
		return s.fetchAndSaveResource()
	}

	log.Debugf("Resource retrieved from datastore.")
	return &resource, nil
}

func (s *Service) fetchAndSaveResource() (*ResourceInfo, error) {
	// Fetch the resource
	res, err := GetResource()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource: %w", err)
	}

	// Save the resource to the datastore
	if err := s.saveResource(res); err != nil {
		return nil, fmt.Errorf("failed to save resource to datastore: %w", err)
	}

	return res, nil
}
