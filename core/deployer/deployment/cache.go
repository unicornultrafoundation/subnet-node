package deployment

func (s *Service) GetDeploymentListCache() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deploymentListCache
}

func (s *Service) DeleteDeploymentListCache(deploymentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Check if the deploymentID is in the cache
	for i, id := range s.deploymentListCache {
		if id == deploymentID {
			s.deploymentListCache = append(s.deploymentListCache[:i], s.deploymentListCache[i+1:]...)
			return
		}
	}
}

func (s *Service) AddDeploymentListCache(deploymentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Check if the deploymentID is already in the cache
	for _, id := range s.deploymentListCache {
		if id == deploymentID {
			return
		}
	}
	s.deploymentListCache = append(s.deploymentListCache, deploymentID)
}
