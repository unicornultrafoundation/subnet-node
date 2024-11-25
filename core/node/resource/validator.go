package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	record "github.com/libp2p/go-libp2p-record"
)

type ResourceValidator struct{}

// Validate checks if the key and value are valid for a resource record
func (rs ResourceValidator) Validate(key string, value []byte) error {
	// Ensure the key starts with the correct prefix
	if !strings.HasPrefix(key, "/resource/") {
		return fmt.Errorf("invalid key: must start with '/resource/'")
	}

	// Parse the value as ResourceInfo
	var resource ResourceInfo
	if err := json.Unmarshal(value, &resource); err != nil {
		return fmt.Errorf("invalid value: failed to unmarshal JSON: %w", err)
	}

	// Additional validation for ResourceInfo fields
	if resource.CPU.Count <= 0 || resource.Memory.Total <= 0 {
		return errors.New("invalid resource: CPU and Memory must be greater than zero")
	}

	return nil
}

// Select resolves conflicts by selecting the first valid value
func (rs ResourceValidator) Select(k string, vals [][]byte) (int, error) {
	for i, v := range vals {
		// Attempt to validate each value
		if err := rs.Validate(k, v); err == nil {
			return i, nil // Return the first valid value
		}
	}
	return -1, errors.New("no valid values found")
}

// Ensure ResourceValidator implements record.Validator
var _ record.Validator = ResourceValidator{}
