package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

const (
	DeploymentRequestPrefix = "/deployment_request/"
)

func (s *Service) GetDeploymentRequest(ctx context.Context, orderID string) (*types.DeploymentRequest, error) {
	key := datastore.NewKey(fmt.Sprintf("%s%s", DeploymentRequestPrefix, orderID))
	val, err := s.Datastore.Get(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, nil
		}
		s.logger.WithField("orderID", orderID).Error("failed to get deployment request", err)
		return nil, err
	}
	var req types.DeploymentRequest
	err = json.Unmarshal(val, &req)
	if err != nil {
		s.logger.WithField("orderID", orderID).Error("failed to unmarshal deployment request", err)
		return nil, err
	}
	return &req, nil
}

// GetDeploymentRequests returns all deployment requests from the datastore
func (s *Service) GetDeploymentRequests(ctx context.Context) ([]*types.DeploymentRequest, error) {
	q := query.Query{Prefix: DeploymentRequestPrefix}
	results, err := s.Datastore.Query(ctx, q)
	if err != nil {
		s.logger.Error("failed to query deployment requests", err)
		return nil, err
	}
	defer results.Close()
	var out []*types.DeploymentRequest
	for result := range results.Next() {
		if result.Error != nil {
			s.logger.Error("error iterating deployment requests", result.Error)
			continue
		}
		var req types.DeploymentRequest
		if err := json.Unmarshal(result.Value, &req); err != nil {
			s.logger.Error("failed to unmarshal deployment request", err)
			continue
		}
		out = append(out, &req)
	}
	return out, nil
}

// StoreDeploymentRequest stores a deployment request in the datastore
func (s *Service) StoreDeploymentRequest(ctx context.Context, deploymentRequest *types.DeploymentRequest) error {
	key := datastore.NewKey(fmt.Sprintf("%s%s", DeploymentRequestPrefix, deploymentRequest.OrderID))
	val, err := json.Marshal(deploymentRequest)
	if err != nil {
		s.logger.WithField("orderID", deploymentRequest.OrderID).Error("failed to marshal deployment request", err)
		return err
	}
	if err := s.Datastore.Put(ctx, key, val); err != nil {
		s.logger.WithField("orderID", deploymentRequest.OrderID).Error("failed to store deployment request", err)
		return err
	}
	return nil
}

// DeleteDeploymentRequest deletes a deployment request from the datastore
func (s *Service) DeleteDeploymentRequest(ctx context.Context, orderID string) error {
	key := datastore.NewKey(fmt.Sprintf("%s%s", DeploymentRequestPrefix, orderID))
	if err := s.Datastore.Delete(ctx, key); err != nil {
		s.logger.WithField("orderID", orderID).Error("failed to delete deployment request", err)
		return err
	}
	return nil
}
