package types

import (
	"github.com/ethereum/go-ethereum/common"
)

// ProviderStatus represents the status of a provider
type ProviderStatus string

const (
	ProviderStatusActive   ProviderStatus = "active"
	ProviderStatusInactive ProviderStatus = "inactive"
	ProviderStatusBlocked  ProviderStatus = "blocked"
)

// Provider represents a service provider
type Provider struct {
	Address common.Address
	Status  ProviderStatus
}
