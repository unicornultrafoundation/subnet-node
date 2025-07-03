package deployment

import (
	"context"
	"fmt"
)

func (s *Service) GetDeploymentListCache() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert map keys to slice
	result := make([]string, 0, len(s.deploymentListCache))
	for deploymentID := range s.deploymentListCache {
		result = append(result, deploymentID)
	}
	return result
}

func (s *Service) DeleteDeploymentListCache(deploymentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.deploymentListCache, deploymentID)
}

func (s *Service) AddDeploymentListCache(deploymentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deploymentListCache[deploymentID] = struct{}{}
}

func (s *Service) GetDeploymentExpiryListCache() []string {
	s.muExpiry.Lock()
	defer s.muExpiry.Unlock()

	result := make([]string, 0, len(s.deploymentExpiryListCache))
	for deploymentID := range s.deploymentExpiryListCache {
		result = append(result, deploymentID)
	}
	return result
}

func (s *Service) HasDeploymentExpiryListCache(deploymentID string) bool {
	s.muExpiry.Lock()
	defer s.muExpiry.Unlock()
	_, ok := s.deploymentExpiryListCache[deploymentID]
	return ok
}

func (s *Service) AddDeploymentExpiryListCache(ctx context.Context, deploymentID string) (bool, error) {
	if s.HasDeploymentExpiryListCache(deploymentID) {
		return false, nil
	}

	s.muExpiry.Lock()
	s.deploymentExpiryListCache[deploymentID] = struct{}{}
	s.muExpiry.Unlock()

	if err := s.store.StoreDeploymentExpiredList(ctx, s.GetDeploymentExpiryListCache()); err != nil {
		s.logger.Error("Failed to store deployment expired list", err)
		return false, err
	}

	return true, nil
}

func (s *Service) DeleteDeploymentExpiryListCache(ctx context.Context, deploymentID string) error {
	s.muExpiry.Lock()
	delete(s.deploymentExpiryListCache, deploymentID)
	s.muExpiry.Unlock()

	if err := s.store.StoreDeploymentExpiredList(ctx, s.GetDeploymentExpiryListCache()); err != nil {
		return fmt.Errorf("failed to store deployment expired list: %w", err)
	}

	return nil
}
