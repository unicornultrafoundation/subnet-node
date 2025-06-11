package types

import (
	"github.com/ethereum/go-ethereum/common"
)

// RequesterStatus represents the status of a requester
type RequesterStatus string

const (
	RequesterStatusActive   RequesterStatus = "active"
	RequesterStatusInactive RequesterStatus = "inactive"
	RequesterStatusBlocked  RequesterStatus = "blocked"
)

// Requester represents a service requester
type Requester struct {
	Address common.Address
	Status  RequesterStatus
}
