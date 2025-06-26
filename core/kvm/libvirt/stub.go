//go:build !libvirt
// +build !libvirt

package libvirt

import (
	"errors"

	"github.com/sirupsen/logrus"
)

var ErrLibvirtNotAvailable = errors.New("libvirt is not available - build with -tags libvirt to enable")

// Client is a stub implementation when libvirt is not available
type Client struct {
	logger *logrus.Entry
}

// NewClient returns an error when libvirt is not available
func NewClient(uri string, logger *logrus.Entry) (*Client, error) {
	return nil, ErrLibvirtNotAvailable
}

// DomainManager is a stub implementation
type DomainManager struct {
	logger *logrus.Entry
}

// NewDomainManager returns a stub domain manager
func NewDomainManager(client *Client, logger *logrus.Entry) *DomainManager {
	return &DomainManager{logger: logger}
}

// StorageManager is a stub implementation
type StorageManager struct {
	logger *logrus.Entry
}

// NewStorageManager returns a stub storage manager
func NewStorageManager(client *Client, logger *logrus.Entry) *StorageManager {
	return &StorageManager{logger: logger}
}

// NetworkManager is a stub implementation
type NetworkManager struct {
	logger *logrus.Entry
}

// NewNetworkManager returns a stub network manager
func NewNetworkManager(client *Client, logger *logrus.Entry) *NetworkManager {
	return &NetworkManager{logger: logger}
}

// CloudInitManager is a stub implementation
type CloudInitManager struct {
	logger *logrus.Entry
}

// NewCloudInitManager returns a stub cloud-init manager
func NewCloudInitManager(logger *logrus.Entry) *CloudInitManager {
	return &CloudInitManager{logger: logger}
}
