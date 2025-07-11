package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttributes_GetCapabilitiesMap(t *testing.T) {
	// Test case 1: Basic capabilities filtering
	attrs := Attributes{
		"capabilities/gpu/model":   "RTX 3080",
		"capabilities/gpu/memory":  "10GB",
		"capabilities/cpu/cores":   "8",
		"capabilities/cpu/threads": "16",
		"other/attribute":          "value",
		"capabilities/gpu/cuda":    true,
		"capabilities/memory/size": "16GB",
		"capabilities/memory/type": "DDR4",
	}

	// Test GPU capabilities
	gpuCapabilities := attrs.GetCapabilitiesMap("gpu/")
	expectedGPU := Attributes{
		"model":  "RTX 3080",
		"memory": "10GB",
		"cuda":   true,
	}
	assert.Equal(t, expectedGPU, gpuCapabilities)

	// Test CPU capabilities
	cpuCapabilities := attrs.GetCapabilitiesMap("cpu/")
	expectedCPU := Attributes{
		"cores":   "8",
		"threads": "16",
	}
	assert.Equal(t, expectedCPU, cpuCapabilities)

	// Test Memory capabilities
	memoryCapabilities := attrs.GetCapabilitiesMap("memory/")
	expectedMemory := Attributes{
		"size": "16GB",
		"type": "DDR4",
	}
	assert.Equal(t, expectedMemory, memoryCapabilities)

	// Test non-existent prefix
	nonExistentCapabilities := attrs.GetCapabilitiesMap("nonexistent/")
	assert.Empty(t, nonExistentCapabilities)
}

func TestAttributes_GetCapabilitiesMap_EdgeCases(t *testing.T) {
	// Test case 2: Empty attributes
	emptyAttrs := Attributes{}
	result := emptyAttrs.GetCapabilitiesMap("gpu/")
	assert.Empty(t, result)

	// Test case 3: No capabilities prefix
	attrs := Attributes{
		"gpu/model": "RTX 3080",
		"cpu/cores": "8",
	}
	result = attrs.GetCapabilitiesMap("gpu/")
	assert.Empty(t, result)

	// Test case 4: Exact prefix match (should not include)
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

	// Test case 6: Mixed types in capabilities
	attrs = Attributes{
		"capabilities/gpu/model":    "RTX 3080",
		"capabilities/gpu/memory":   10,
		"capabilities/gpu/enabled":  true,
		"capabilities/gpu/features": []string{"cuda", "tensor_cores"},
	}
	result = attrs.GetCapabilitiesMap("gpu/")
	expected = Attributes{
		"model":    "RTX 3080",
		"memory":   10,
		"enabled":  true,
		"features": []string{"cuda", "tensor_cores"},
	}
	assert.Equal(t, expected, result)
}

func TestAttributes_AllIn(t *testing.T) {
	a := Attributes{
		"a": 1,
		"b": "foo",
		"c": true,
	}
	b := Attributes{
		"a": 1,
		"b": "foo",
		"c": true,
		"d": 1,
	}
	c := Attributes{
		"a": 1,
		"b": "foo",
		"c": false,
	}
	d := Attributes{
		"a": 1,
		"b": "foo",
	}
	e := Attributes{
		"a": 2,
		"b": "foo",
		"c": true,
	}
	// Tất cả key/value của a đều có trong b
	assert.True(t, a.AllIn(b))
	// Một value không khớp
	assert.False(t, a.AllIn(c))
	// Một key không có trong b
	assert.False(t, a.AllIn(d))
	// Một value không khớp
	assert.False(t, a.AllIn(e))
	// So sánh với chính nó
	assert.True(t, a.AllIn(a))
	// a rỗng thì luôn true
	assert.True(t, (Attributes{}).AllIn(b))
	// b rỗng thì chỉ true nếu a cũng rỗng
	assert.False(t, a.AllIn(Attributes{}))
	assert.True(t, (Attributes{}).AllIn(Attributes{}))
	// So sánh với value là slice (không so sánh, bỏ qua)
	f := Attributes{
		"arr": []int{1, 2, 3},
	}
	g := Attributes{
		"arr": []int{1, 2, 3},
	}
	assert.True(t, f.AllIn(g)) // vì slice không so sánh, sẽ bỏ qua
}
