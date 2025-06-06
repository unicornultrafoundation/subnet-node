package types

// HealthStatus represents the health status of a deployment
type HealthStatus string

const (
	// HealthStatusUnknown represents an unknown health status
	HealthStatusUnknown HealthStatus = "unknown"
	// HealthStatusHealthy represents a healthy status
	HealthStatusHealthy HealthStatus = "healthy"
	// HealthStatusUnhealthy represents an unhealthy status
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	// HealthStatusDegraded represents a degraded status
	HealthStatusDegraded HealthStatus = "degraded"
)

// IsHealthy returns true if the health status is healthy
func (h HealthStatus) IsHealthy() bool {
	return h == HealthStatusHealthy
}

// IsUnhealthy returns true if the health status is unhealthy
func (h HealthStatus) IsUnhealthy() bool {
	return h == HealthStatusUnhealthy
}

// IsDegraded returns true if the health status is degraded
func (h HealthStatus) IsDegraded() bool {
	return h == HealthStatusDegraded
}

// IsUnknown returns true if the health status is unknown
func (h HealthStatus) IsUnknown() bool {
	return h == HealthStatusUnknown
}
