package types

import "math/big"

// This describes how the endpoint is implemented when the lease is deployed
type Endpoint_Kind int32

const (
	// Describes an endpoint that becomes a Kubernetes Ingress
	Endpoint_SHARED_HTTP Endpoint_Kind = 0
	// Describes an endpoint that becomes a Kubernetes NodePort
	Endpoint_RANDOM_PORT Endpoint_Kind = 1
	// Describes an endpoint that becomes a leased IP
	Endpoint_LEASED_IP Endpoint_Kind = 2
)

type Resources struct {
	CPU       *CPU      `json:"cpu" yaml:"cpu"`
	Memory    *Memory   `json:"memory" yaml:"memory"`
	GPU       *GPU      `json:"gpu" yaml:"gpu"`
	Volumes   Volumes   `json:"volumes" yaml:"volumes"`
	Endpoints Endpoints `json:"endpoints" yaml:"endpoints"`
}

type CPU struct {
	Units      *big.Int   `json:"units" yaml:"units"`
	Attributes Attributes `json:"attributes" yaml:"attributes"`
}

type Memory struct {
	Size       *big.Int   `json:"size" yaml:"size"`
	Attributes Attributes `json:"attributes" yaml:"attributes"`
}

type GPU struct {
	Units      *big.Int   `json:"units" yaml:"units"`
	Attributes Attributes `json:"attributes" yaml:"attributes"`
}

type Volume struct {
	Name       string     `json:"name" yaml:"name"`
	Size       *big.Int   `json:"size" yaml:"size"`
	Attributes Attributes `json:"attributes" yaml:"attributes"`
}

type Endpoint struct {
	Kind Endpoint_Kind `json:"kind" yaml:"kind"`
}

type Volumes []Volume
type Endpoints []Endpoint
