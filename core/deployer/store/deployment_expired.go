package store

import (
	"context"
	"encoding/json"

	"github.com/ipfs/go-datastore"
)

const (
	DeploymentExpiredPrefix = "/deployment_expired/"
)

func (s *Service) GetDeploymentExpiredList(ctx context.Context) ([]string, error) {
	key := datastore.NewKey(DeploymentExpiredPrefix)
	val, err := s.Datastore.Get(ctx, key)
	if err != nil {
		if err == datastore.ErrNotFound {
			return nil, nil
		}
		s.logger.Error("failed to get deployment expired list", err)
		return nil, err
	}
	var expiredList []string
	err = json.Unmarshal(val, &expiredList)
	if err != nil {
		s.logger.Error("failed to unmarshal deployment expired list", err)
		return nil, err
	}
	return expiredList, nil
}

// StoreDeploymentExpiredList stores a deployment expired list in the datastore
func (s *Service) StoreDeploymentExpiredList(ctx context.Context, deploymentExpiredList []string) error {
	key := datastore.NewKey(DeploymentExpiredPrefix)
	val, err := json.Marshal(deploymentExpiredList)
	if err != nil {
		s.logger.Error("failed to marshal deployment expired list", err)
		return err
	}
	if err := s.Datastore.Put(ctx, key, val); err != nil {
		s.logger.Error("failed to store deployment expired list", err)
		return err
	}
	return nil
}

// DeleteDeploymentExpiredList deletes a deployment expired list from the datastore
func (s *Service) DeleteDeploymentExpiredList(ctx context.Context) error {
	key := datastore.NewKey(DeploymentExpiredPrefix)
	if err := s.Datastore.Delete(ctx, key); err != nil {
		s.logger.Error("failed to delete deployment expired list", err)
		return err
	}
	return nil
}
