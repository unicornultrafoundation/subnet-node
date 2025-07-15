package types

import "reflect"

// Attributes represents a map of string keys to any values, commonly used for storing
// configuration attributes, capabilities, and metadata.
type Attributes map[string]string

// GetCapabilitiesMap extracts capabilities from the attributes map that match the given prefix.
// It filters keys that start with "capabilities/" + prefix and returns a new map with the prefix removed.
// For example, if prefix is "gpu", it will extract keys like "capabilities/gpu/memory" -> "memory".
func (a Attributes) GetCapabilitiesMap(prefix string) Attributes {
	capabilities := make(Attributes)
	capabilityPrefix := "capabilities/" + prefix

	for key, value := range a {
		if len(key) > len(capabilityPrefix) && key[:len(capabilityPrefix)] == capabilityPrefix {
			// Remove the prefix and add to capabilities map
			capabilityKey := key[len(capabilityPrefix):]
			capabilities[capabilityKey] = value
		}
	}
	return capabilities
}

// AllIn checks if all key/value pairs in the current attributes map exist and match in the target map.
// It returns true only if every key in the current map exists in the target map with an equal value.
// Uses reflection to safely compare values and skips uncomparable types.
func (a Attributes) AllIn(b Attributes) bool {
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		// Only compare if value is comparable
		if reflect.TypeOf(v) == nil || !reflect.TypeOf(v).Comparable() || reflect.TypeOf(bv) == nil || !reflect.TypeOf(bv).Comparable() {
			continue // Skip uncomparable types
		}
		if bv != v {
			return false
		}
	}
	return true
}
