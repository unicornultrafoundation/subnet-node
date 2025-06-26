package deployment

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
