package libvirt

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

// Client wraps libvirt connection and provides KVM/QEMU management
type Client struct {
	conn   *libvirt.Connect
	uri    string
	logger *logrus.Entry
	mu     sync.RWMutex
}

// NewClient creates a new libvirt client
func NewClient(uri string, logger *logrus.Entry) (*Client, error) {
	if uri == "" {
		uri = "qemu:///system" // Default to system QEMU connection
	}

	client := &Client{
		uri:    uri,
		logger: logger.WithField("component", "libvirt-client"),
	}

	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	return client, nil
}

// Connect establishes connection to libvirt daemon
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		// Check if connection is still alive
		if alive, err := c.conn.IsAlive(); err == nil && alive {
			return nil
		}
		// Close old connection
		c.conn.Close()
	}

	var err error
	c.conn, err = libvirt.NewConnect(c.uri)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt at %s: %w", c.uri, err)
	}

	c.logger.WithField("uri", c.uri).Info("Connected to libvirt")
	return nil
}

// Disconnect closes the libvirt connection
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	_, err := c.conn.Close()
	c.conn = nil

	if err != nil {
		return fmt.Errorf("failed to disconnect from libvirt: %w", err)
	}

	c.logger.Info("Disconnected from libvirt")
	return nil
}

// GetConnection returns the underlying libvirt connection (thread-safe)
func (c *Client) GetConnection() (*libvirt.Connect, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return nil, fmt.Errorf("not connected to libvirt")
	}

	// Check if connection is still alive
	if alive, err := c.conn.IsAlive(); err != nil || !alive {
		c.mu.RUnlock()
		// Reconnect
		if err := c.Connect(); err != nil {
			return nil, fmt.Errorf("failed to reconnect to libvirt: %w", err)
		}
		c.mu.RLock()
	}

	return c.conn, nil
}

// GetVersion returns libvirt version information
func (c *Client) GetVersion() (uint64, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return 0, err
	}

	version, err := conn.GetVersion()
	if err != nil {
		return 0, fmt.Errorf("failed to get libvirt version: %w", err)
	}

	return uint64(version), nil
}

// GetNodeInfo returns information about the hypervisor node
func (c *Client) GetNodeInfo() (*libvirt.NodeInfo, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return nil, err
	}

	nodeInfo, err := conn.GetNodeInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get node info: %w", err)
	}

	return nodeInfo, nil
}

// IsConnected checks if client is connected to libvirt
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return false
	}

	alive, err := c.conn.IsAlive()
	return err == nil && alive
}

// GetHostname returns the hostname of the hypervisor
func (c *Client) GetHostname() (string, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return "", err
	}

	hostname, err := conn.GetHostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}

	return hostname, nil
}

// ListAllDomains returns all domains (VMs)
func (c *Client) ListAllDomains(flags libvirt.ConnectListAllDomainsFlags) ([]libvirt.Domain, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return nil, err
	}

	domains, err := conn.ListAllDomains(flags)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	return domains, nil
}

// GetStoragePoolByName gets a storage pool by name
func (c *Client) GetStoragePoolByName(name string) (*libvirt.StoragePool, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return nil, err
	}

	pool, err := conn.LookupStoragePoolByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage pool %s: %w", name, err)
	}

	return pool, nil
}

// GetNetworkByName gets a network by name
func (c *Client) GetNetworkByName(name string) (*libvirt.Network, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return nil, err
	}

	network, err := conn.LookupNetworkByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get network %s: %w", name, err)
	}

	return network, nil
}

// CreateStoragePool creates a new storage pool
func (c *Client) CreateStoragePool(xml string) (*libvirt.StoragePool, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return nil, err
	}

	pool, err := conn.StoragePoolDefineXML(xml, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to define storage pool: %w", err)
	}

	if err := pool.Create(0); err != nil {
		pool.Undefine()
		return nil, fmt.Errorf("failed to create storage pool: %w", err)
	}

	return pool, nil
}

// CreateNetwork creates a new network
func (c *Client) CreateNetwork(xml string) (*libvirt.Network, error) {
	conn, err := c.GetConnection()
	if err != nil {
		return nil, err
	}

	network, err := conn.NetworkDefineXML(xml)
	if err != nil {
		return nil, fmt.Errorf("failed to define network: %w", err)
	}

	if err := network.Create(); err != nil {
		network.Undefine()
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	return network, nil
}
