package vbox_service

import (
	"context"
	"encoding/json"

	"github.com/ipfs/go-datastore/query"
)

// Load all order mappings from datastore
func (s *VBoxService) loadAllOrderMappingsFromDatastore(ctx context.Context) error {
	q := query.Query{Prefix: "virtualbox/order_mapping/"}
	results, err := s.ds.Query(ctx, q)
	if err != nil {
		return err
	}
	defer results.Close()

	for result := range results.Next() {
		if result.Error != nil {
			continue
		}
		var mapping map[string]string
		if err := json.Unmarshal(result.Value, &mapping); err != nil {
			continue
		}
		orderId := mapping["orderId"]
		vmId := mapping["vmId"]
		if orderId != "" && vmId != "" {
			s.orderToVMMap[orderId] = vmId
		}
	}

	serviceLog.Infof("Loaded %d order mappings from datastore", len(s.orderToVMMap))
	return nil
}

// sync VM to Datastore
func (s *VBoxService) syncVMToDatastore(ctx context.Context) {
	serviceLog.Info("Starting VM synchronization with datastore")

	for {
		select {
		case <-s.syncTicker.C:
			if err := s.performVMSync(ctx); err != nil {
				serviceLog.Errorf("Error during VM synchronization: %v", err)
			}
		case <-s.stopChan:
			serviceLog.Info("Stopping VM synchronization")
			return
		case <-ctx.Done():
			serviceLog.Info("Context cancelled, stopping VM synchronization")
			return
		}
	}
}

// performVMSync does the actual synchronization between VirtualBox and the datastore
func (s *VBoxService) performVMSync(ctx context.Context) error {
	return nil
}
