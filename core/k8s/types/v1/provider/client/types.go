package rest

import (
	inventoryV1 "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/inventory/v1"
	manifest "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"

	etypes "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/expiry"
)

// ClusterStatus represents the current state of the provider's cluster, including
// the number of active leases and inventory metrics.
type ClusterStatus struct {
	Leases    uint32                       `json:"leases"`
	Inventory inventoryV1.InventoryMetrics `json:"inventory"`
}

// ServiceStatus represents the current state of a service within a lease,
// including availability, total instances, URIs, and various replica counts.
type ServiceStatus struct {
	Name      string   `json:"name"`
	Available int32    `json:"available"`
	Total     int32    `json:"total"`
	URIs      []string `json:"uris"`

	ObservedGeneration int64 `json:"observed_generation"`
	Replicas           int32 `json:"replicas"`
	UpdatedReplicas    int32 `json:"updated_replicas"`
	ReadyReplicas      int32 `json:"ready_replicas"`
	AvailableReplicas  int32 `json:"available_replicas"`
}

// ForwardedPortStatus represents the status of a forwarded port for a service,
// including host, port numbers, protocol, and service name.
type ForwardedPortStatus struct {
	Host         string                   `json:"host,omitempty"`
	Port         uint16                   `json:"port"`
	ExternalPort uint16                   `json:"externalPort"`
	Proto        manifest.ServiceProtocol `json:"proto"`
	Name         string                   `json:"name"`
}

// LeasedIPStatus represents the status of an IP address leased to a service,
// including port numbers, protocol, and the IP address itself.
type LeasedIPStatus struct {
	Port         uint32
	ExternalPort uint32
	Protocol     string
	IP           string
}

// LeaseStatus represents the current state of a lease, including any error messages,
// service statuses, forwarded ports, and IP addresses.
type LeaseStatus struct {
	Messages             []string                            `json:"errors,omitempty"`
	Services             map[string]*ServiceStatus           `json:"services"`
	ForwardedPorts       map[string][]ForwardedPortStatus    `json:"forwarded_ports"` // Container services that are externally accessible
	IPs                  map[string][]LeasedIPStatus         `json:"ips,omitempty"`
	TimeLeft             int64                               `json:"time_left,omitempty"`
	Status               etypes.DeploymentExpiryStatus       `json:"status,omitempty"`
	HostnameVerification map[string]HostnameVerificationInfo `json:"hostname_verification,omitempty"`
}

type HostnameVerificationInfo struct {
	Token    string `json:"token,omitempty"`
	Verified bool   `json:"verified"`
	Message  string `json:"message,omitempty"`
}
