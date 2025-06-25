package managers

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	deployment "github.com/unicornultrafoundation/subnet-node/core/deployments/docker"
)

// PortManagerImpl implements PortManager
type PortManagerImpl struct {
	mu       sync.RWMutex
	logger   *logrus.Logger
	portPool *deployment.PortPool
}

// NewPortManager creates a new port manager
func NewPortManager(logger *logrus.Logger, startPort, endPort int) deployment.PortManager {
	return &PortManagerImpl{
		logger: logger,
		portPool: &deployment.PortPool{
			StartPort: startPort,
			EndPort:   endPort,
			UsedPorts: make(map[int]string),
		},
	}
}

// AllocatePort allocates a port for a tenant
func (pm *PortManagerImpl) AllocatePort(ctx context.Context, tenantID string, preferredPort int) (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// If preferred port is specified and available, use it
	if preferredPort != 0 {
		if _, used := pm.portPool.UsedPorts[preferredPort]; !used {
			if preferredPort >= pm.portPool.StartPort && preferredPort <= pm.portPool.EndPort {
				pm.portPool.UsedPorts[preferredPort] = tenantID
				pm.logger.WithFields(logrus.Fields{
					"port":      preferredPort,
					"tenant_id": tenantID,
				}).Debug("Allocated preferred port")
				return preferredPort, nil
			}
		}
	}

	// Find an available port
	availablePorts := make([]int, 0)
	for port := pm.portPool.StartPort; port <= pm.portPool.EndPort; port++ {
		if _, used := pm.portPool.UsedPorts[port]; !used {
			availablePorts = append(availablePorts, port)
		}
	}

	if len(availablePorts) == 0 {
		return 0, fmt.Errorf("no available ports in range %d-%d", pm.portPool.StartPort, pm.portPool.EndPort)
	}

	// Randomly select an available port
	rand.Seed(time.Now().UnixNano())
	selectedPort := availablePorts[rand.Intn(len(availablePorts))]

	pm.portPool.UsedPorts[selectedPort] = tenantID

	pm.logger.WithFields(logrus.Fields{
		"port":      selectedPort,
		"tenant_id": tenantID,
	}).Debug("Allocated port")

	return selectedPort, nil
}

// ReleasePort releases a port
func (pm *PortManagerImpl) ReleasePort(ctx context.Context, port int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, used := pm.portPool.UsedPorts[port]; !used {
		return fmt.Errorf("port %d is not allocated", port)
	}

	delete(pm.portPool.UsedPorts, port)

	pm.logger.WithField("port", port).Debug("Released port")
	return nil
}

// GetPortPool gets the current port pool status
func (pm *PortManagerImpl) GetPortPool(ctx context.Context) *deployment.PortPool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Create a copy to avoid race conditions
	pool := &deployment.PortPool{
		StartPort: pm.portPool.StartPort,
		EndPort:   pm.portPool.EndPort,
		UsedPorts: make(map[int]string),
	}

	for port, tenantID := range pm.portPool.UsedPorts {
		pool.UsedPorts[port] = tenantID
	}

	return pool
}

// ReservePortRange reserves a range of ports for a tenant
func (pm *PortManagerImpl) ReservePortRange(ctx context.Context, tenantID string, startPort, endPort int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Validate port range
	if startPort < pm.portPool.StartPort || endPort > pm.portPool.EndPort {
		return fmt.Errorf("port range %d-%d is outside the allowed range %d-%d", startPort, endPort, pm.portPool.StartPort, pm.portPool.EndPort)
	}

	if startPort > endPort {
		return fmt.Errorf("invalid port range: start port %d is greater than end port %d", startPort, endPort)
	}

	// Check if any port in the range is already used
	for port := startPort; port <= endPort; port++ {
		if _, used := pm.portPool.UsedPorts[port]; used {
			return fmt.Errorf("port %d is already allocated", port)
		}
	}

	// Reserve all ports in the range
	for port := startPort; port <= endPort; port++ {
		pm.portPool.UsedPorts[port] = tenantID
	}

	pm.logger.WithFields(logrus.Fields{
		"start_port": startPort,
		"end_port":   endPort,
		"tenant_id":  tenantID,
	}).Debug("Reserved port range")

	return nil
}

// ReleasePortRange releases a range of ports
func (pm *PortManagerImpl) ReleasePortRange(ctx context.Context, startPort, endPort int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Validate port range
	if startPort > endPort {
		return fmt.Errorf("invalid port range: start port %d is greater than end port %d", startPort, endPort)
	}

	// Release all ports in the range
	for port := startPort; port <= endPort; port++ {
		delete(pm.portPool.UsedPorts, port)
	}

	pm.logger.WithFields(logrus.Fields{
		"start_port": startPort,
		"end_port":   endPort,
	}).Debug("Released port range")

	return nil
}

// GetTenantPorts gets all ports allocated to a tenant
func (pm *PortManagerImpl) GetTenantPorts(ctx context.Context, tenantID string) ([]int, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var ports []int
	for port, allocatedTenantID := range pm.portPool.UsedPorts {
		if allocatedTenantID == tenantID {
			ports = append(ports, port)
		}
	}

	return ports, nil
}
