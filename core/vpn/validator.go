package vpn

import (
	"errors"
	"fmt"
	"strings"

	record "github.com/libp2p/go-libp2p-record"
)

type VPNValidator struct{}

// Validate checks if the key and value are valid for a resource record
func (vpn VPNValidator) Validate(key string, value []byte) error {
	// Ensure the key starts with the correct prefix
	if !strings.HasPrefix(key, "/vpn/mapping/") {
		return fmt.Errorf("invalid key: must start with '/vpn/mapping/'")
	}

	return nil
}

// Select resolves conflicts by selecting the first valid value
func (vpn VPNValidator) Select(k string, vals [][]byte) (int, error) {
	for i, v := range vals {
		// Attempt to validate each value
		if err := vpn.Validate(k, v); err == nil {
			return i, nil // Return the first valid value
		}
	}
	return -1, errors.New("no valid values found")
}

// Ensure ResourceValidator implements record.Validator
var _ record.Validator = VPNValidator{}
