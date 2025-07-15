package virtualbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHardwareDetector_DetectHardware(t *testing.T) {
	detector := NewHardwareDetector()
	hardware, err := detector.DetectHardware()

	// Hardware detection should not fail
	assert.NoError(t, err)
	assert.NotNil(t, hardware)

	// Basic hardware info should be populated
	assert.NotEmpty(t, hardware.Architecture)
	assert.NotEmpty(t, hardware.OSType)
	assert.Greater(t, hardware.CPUCount, 0)
	assert.Greater(t, hardware.TotalMemoryMB, 0)
	assert.Greater(t, hardware.AvailableMemoryMB, 0)
	assert.NotEmpty(t, hardware.GPUType)
	assert.GreaterOrEqual(t, hardware.GPUCount, 0)

	t.Logf("Detected hardware: %+v", hardware)
}

func TestHardwareDetector_GetVirtualBoxSettings(t *testing.T) {
	detector := NewHardwareDetector()
	hardware, err := detector.DetectHardware()
	assert.NoError(t, err)

	settings := detector.GetVirtualBoxSettings(hardware)
	assert.NotNil(t, settings)

	// Settings should be populated
	assert.NotEmpty(t, settings.OSType)
	assert.NotEmpty(t, settings.Chipset)
	assert.NotEmpty(t, settings.Firmware)
	assert.NotEmpty(t, settings.GraphicsController)
	assert.Greater(t, settings.VRAMMB, 0)
	assert.NotEmpty(t, settings.USBController)
	assert.NotEmpty(t, settings.AudioController)
	assert.NotEmpty(t, settings.AudioOutput)
	assert.NotEmpty(t, settings.AudioInput)

	t.Logf("VirtualBox settings: %+v", settings)
}

func TestHardwareDetector_DetermineOSType(t *testing.T) {
	detector := NewHardwareDetector()

	tests := []struct {
		arch     string
		expected string
	}{
		{"arm64", "Ubuntu_ARM64"},
		{"aarch64", "Ubuntu_ARM64"},
		{"amd64", "Ubuntu_64"},
		{"x86_64", "Ubuntu_64"},
		{"arm", "Ubuntu"},
		{"386", "Ubuntu"},
		{"i386", "Ubuntu"},
		{"unknown", "Ubuntu_64"},
	}

	for _, test := range tests {
		hardware := &HardwareInfo{Architecture: test.arch}
		result := detector.determineOSType(hardware)
		assert.Equal(t, test.expected, result, "Architecture: %s", test.arch)
	}
}

func TestHardwareDetector_DetermineChipset(t *testing.T) {
	detector := NewHardwareDetector()

	tests := []struct {
		arch     string
		expected string
	}{
		{"arm64", "armv8virtual"},
		{"aarch64", "armv8virtual"},
		{"arm", "armv8virtual"},
		{"amd64", "ich9"},
		{"x86_64", "ich9"},
		{"386", "ich9"},
		{"i386", "ich9"},
		{"unknown", "ich9"},
	}

	for _, test := range tests {
		hardware := &HardwareInfo{Architecture: test.arch}
		result := detector.determineChipset(hardware)
		assert.Equal(t, test.expected, result, "Architecture: %s", test.arch)
	}
}

func TestHardwareDetector_DetermineFirmware(t *testing.T) {
	detector := NewHardwareDetector()

	tests := []struct {
		arch     string
		expected string
	}{
		{"arm64", "efi"},
		{"aarch64", "efi"},
		{"arm", "efi"},
		{"amd64", "efi"},
		{"x86_64", "efi"},
		{"386", "bios"},
		{"i386", "bios"},
		{"unknown", "efi"},
	}

	for _, test := range tests {
		hardware := &HardwareInfo{Architecture: test.arch}
		result := detector.determineFirmware(hardware)
		assert.Equal(t, test.expected, result, "Architecture: %s", test.arch)
	}
}

func TestHardwareDetector_DetermineVRAM(t *testing.T) {
	detector := NewHardwareDetector()

	tests := []struct {
		name     string
		hardware *HardwareInfo
		expected int
	}{
		{
			name: "Apple Silicon with 8 cores",
			hardware: &HardwareInfo{
				IsAppleSilicon: true,
				GPUCount:       8,
			},
			expected: 128,
		},
		{
			name: "Apple Silicon with 4 cores",
			hardware: &HardwareInfo{
				IsAppleSilicon: true,
				GPUCount:       4,
			},
			expected: 64,
		},
		{
			name: "NVIDIA GPU",
			hardware: &HardwareInfo{
				GPUType: "NVIDIA RTX 3080",
			},
			expected: 128,
		},
		{
			name: "AMD GPU",
			hardware: &HardwareInfo{
				GPUType: "AMD RX 6800",
			},
			expected: 64,
		},
		{
			name: "Integrated Graphics",
			hardware: &HardwareInfo{
				GPUType: "Intel Integrated Graphics",
			},
			expected: 32,
		},
	}

	for _, test := range tests {
		result := detector.determineVRAM(test.hardware)
		assert.Equal(t, test.expected, result, "Test: %s", test.name)
	}
}

func TestHardwareDetector_DetermineIOAPIC(t *testing.T) {
	detector := NewHardwareDetector()

	tests := []struct {
		arch     string
		expected bool
	}{
		{"arm64", false},
		{"aarch64", false},
		{"arm", false},
		{"amd64", true},
		{"x86_64", true},
		{"386", true},
		{"i386", true},
		{"unknown", true},
	}

	for _, test := range tests {
		hardware := &HardwareInfo{Architecture: test.arch}
		result := detector.determineIOAPIC(hardware)
		assert.Equal(t, test.expected, result, "Architecture: %s", test.arch)
	}
}
