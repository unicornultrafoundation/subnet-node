package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAttributes_GetCapabilitiesMap tests the GetCapabilitiesMap function with various scenarios
func TestAttributes_GetCapabilitiesMap(t *testing.T) {
	// Test case 1: Basic capabilities filtering with multiple capability types
	attrs := Attributes{
		"capabilities/gpu/model":   "RTX 3080",
		"capabilities/gpu/memory":  "10GB",
		"capabilities/cpu/cores":   "8",
		"capabilities/cpu/threads": "16",
		"other/attribute":          "value",
		"capabilities/gpu/cuda":    "true",
		"capabilities/memory/size": "16GB",
		"capabilities/memory/type": "DDR4",
	}

	// Test GPU capabilities extraction
	gpuCapabilities := attrs.GetCapabilitiesMap("gpu/")
	expectedGPU := Attributes{
		"model":  "RTX 3080",
		"memory": "10GB",
		"cuda":   "true",
	}
	assert.Equal(t, expectedGPU, gpuCapabilities)

	// Test CPU capabilities extraction
	cpuCapabilities := attrs.GetCapabilitiesMap("cpu/")
	expectedCPU := Attributes{
		"cores":   "8",
		"threads": "16",
	}
	assert.Equal(t, expectedCPU, cpuCapabilities)

	// Test Memory capabilities extraction
	memoryCapabilities := attrs.GetCapabilitiesMap("memory/")
	expectedMemory := Attributes{
		"size": "16GB",
		"type": "DDR4",
	}
	assert.Equal(t, expectedMemory, memoryCapabilities)

	// Test non-existent prefix should return empty map
	nonExistentCapabilities := attrs.GetCapabilitiesMap("nonexistent/")
	assert.Empty(t, nonExistentCapabilities)
}

// TestAttributes_GetCapabilitiesMap_EdgeCases tests edge cases and boundary conditions
func TestAttributes_GetCapabilitiesMap_EdgeCases(t *testing.T) {
	// Test case 2: Empty attributes map
	emptyAttrs := Attributes{}
	result := emptyAttrs.GetCapabilitiesMap("gpu/")
	assert.Empty(t, result)

	// Test case 3: Attributes without capabilities prefix
	attrs := Attributes{
		"gpu/model": "RTX 3080",
		"cpu/cores": "8",
	}
	result = attrs.GetCapabilitiesMap("gpu/")
	assert.Empty(t, result)

	// Test case 4: Exact prefix match (should not include as it needs additional path)
	attrs = Attributes{
		"capabilities/gpu": "value",
	}
	result = attrs.GetCapabilitiesMap("gpu/")
	assert.Empty(t, result)

	// Test case 5: Prefix without trailing slash should still work
	attrs = Attributes{
		"capabilities/gpu/model": "RTX 3080",
	}
	result = attrs.GetCapabilitiesMap("gpu") // without trailing slash
	expected := Attributes{
		"/model": "RTX 3080",
	}
	assert.Equal(t, expected, result)

	// Test case 6: Mixed data types in capabilities
	attrs = Attributes{
		"capabilities/gpu/model":    "RTX 3080",
		"capabilities/gpu/memory":   "10",
		"capabilities/gpu/enabled":  "true",
		"capabilities/gpu/features": "cuda,tensor_cores",
	}
	result = attrs.GetCapabilitiesMap("gpu/")
	expected = Attributes{
		"model":    "RTX 3080",
		"memory":   "10",
		"enabled":  "true",
		"features": "cuda,tensor_cores",
	}
	assert.Equal(t, expected, result)
}

// TestAttributes_AllIn tests the AllIn function with various comparison scenarios
func TestAttributes_AllIn(t *testing.T) {
	// Test data setup
	a := Attributes{
		"a": "1",
		"b": "foo",
		"c": "true",
	}
	b := Attributes{
		"a": "1",
		"b": "foo",
		"c": "true",
		"d": "1",
	}
	c := Attributes{
		"a": "1",
		"b": "foo",
		"c": "false",
	}
	d := Attributes{
		"a": "1",
		"b": "foo",
	}
	e := Attributes{
		"a": "2",
		"b": "foo",
		"c": "true",
	}

	// Test case 1: All key/value pairs in a exist and match in b
	assert.True(t, a.AllIn(b))

	// Test case 2: One value doesn't match
	assert.False(t, a.AllIn(c))

	// Test case 3: One key doesn't exist in target map
	assert.False(t, a.AllIn(d))

	// Test case 4: One value doesn't match
	assert.False(t, a.AllIn(e))

	// Test case 5: Self-comparison should always be true
	assert.True(t, a.AllIn(a))

	// Test case 6: Empty source map should always return true
	assert.True(t, (Attributes{}).AllIn(b))

	// Test case 7: Empty target map should only be true if source is also empty
	assert.False(t, a.AllIn(Attributes{}))
	assert.True(t, (Attributes{}).AllIn(Attributes{}))

	// Test case 8: Comparison with uncomparable types (slices) should be skipped
	f := Attributes{
		"arr": "1,2,3",
	}
	g := Attributes{
		"arr": "1,2,3",
	}
	assert.True(t, f.AllIn(g)) // Slices are not compared, so this returns true
}
