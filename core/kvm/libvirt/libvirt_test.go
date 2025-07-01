package libvirt

import (
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDetectSystemArchitecture(t *testing.T) {
	// Create a mock client and domain manager for testing
	client := &Client{
		uri:       "qemu:///system",
		logger:    logrus.NewEntry(logrus.New()),
		available: true,
	}

	dm := &DomainManager{
		client: client,
		logger: logrus.NewEntry(logrus.New()),
	}

	// Test architecture detection
	arch := dm.detectSystemArchitecture()

	// Basic validation
	assert.NotNil(t, arch)
	assert.NotEmpty(t, arch.Architecture)
	assert.NotEmpty(t, arch.Machine)

	// Verify that the detected architecture matches runtime.GOARCH
	switch runtime.GOARCH {
	case "amd64":
		assert.Equal(t, "x86_64", arch.Architecture)
		assert.Equal(t, "pc-q35-2.12", arch.Machine)
	case "arm64":
		assert.Equal(t, "aarch64", arch.Architecture)
		assert.Equal(t, "virt-8.2", arch.Machine)
	case "ppc64le":
		assert.Equal(t, "ppc64le", arch.Architecture)
		assert.Equal(t, "pseries", arch.Machine)
	case "s390x":
		assert.Equal(t, "s390x", arch.Architecture)
		assert.Equal(t, "s390-ccw-virtio", arch.Machine)
	}

	t.Logf("Detected architecture: %s", arch.Architecture)
	t.Logf("Detected machine: %s", arch.Machine)
	t.Logf("CPU Model: %s", arch.CPUModel)
	t.Logf("Vendor: %s", arch.Vendor)
	t.Logf("Features count: %d", len(arch.Features))
}

func TestGetSystemArchitectureInfo(t *testing.T) {
	// Create a mock client and domain manager for testing
	client := &Client{
		uri:          "qemu:///system",
		logger:       logrus.NewEntry(logrus.New()),
		available:    true,
		kvmAvailable: true, // Test with KVM available
	}

	dm := &DomainManager{
		client: client,
		logger: logrus.NewEntry(logrus.New()),
	}

	// Test getting architecture info as map
	info := dm.GetSystemArchitectureInfo()

	// Basic validation
	assert.NotNil(t, info)
	assert.Contains(t, info, "architecture")
	assert.Contains(t, info, "machine")
	assert.Contains(t, info, "cpu_model")
	assert.Contains(t, info, "vendor")
	assert.Contains(t, info, "features")
	assert.Contains(t, info, "go_arch")
	assert.Contains(t, info, "go_os")
	assert.Contains(t, info, "kvm_available")
	assert.Contains(t, info, "domain_type")
	assert.Contains(t, info, "libvirt_available")

	// Verify Go runtime information
	assert.Equal(t, runtime.GOARCH, info["go_arch"])
	assert.Equal(t, runtime.GOOS, info["go_os"])

	// Verify KVM information
	assert.Equal(t, true, info["kvm_available"])
	assert.Equal(t, "kvm", info["domain_type"])
	assert.Equal(t, true, info["libvirt_available"])

	t.Logf("Architecture info: %+v", info)
}

func TestEnrichWithLinuxCPUInfo(t *testing.T) {
	// Only run this test on Linux
	if runtime.GOOS != "linux" {
		t.Skip("This test is only for Linux systems")
	}

	dm := &DomainManager{
		logger: logrus.NewEntry(logrus.New()),
	}

	arch := &SystemArchitecture{
		Architecture: "x86_64",
		Machine:      "pc-q35-2.12",
		CPUModel:     "Unknown",
		Vendor:       "Unknown",
		Features:     []string{},
	}

	enriched := dm.enrichWithLinuxCPUInfo(arch)

	assert.NotNil(t, enriched)
	assert.NotEmpty(t, enriched.Architecture)

	t.Logf("Linux CPU info enrichment: %+v", enriched)
}

func TestGetDomainType(t *testing.T) {
	// Test with KVM available
	clientWithKVM := &Client{
		uri:          "qemu:///system",
		logger:       logrus.NewEntry(logrus.New()),
		available:    true,
		kvmAvailable: true,
	}

	assert.Equal(t, "kvm", clientWithKVM.GetDomainType())

	// Test without KVM available
	clientWithoutKVM := &Client{
		uri:          "qemu:///system",
		logger:       logrus.NewEntry(logrus.New()),
		available:    true,
		kvmAvailable: false,
	}

	assert.Equal(t, "qemu", clientWithoutKVM.GetDomainType())
}

func TestIsKVMAvailable(t *testing.T) {
	// Test with KVM available
	clientWithKVM := &Client{
		uri:          "qemu:///system",
		logger:       logrus.NewEntry(logrus.New()),
		available:    true,
		kvmAvailable: true,
	}

	assert.True(t, clientWithKVM.IsKVMAvailable())

	// Test without KVM available
	clientWithoutKVM := &Client{
		uri:          "qemu:///system",
		logger:       logrus.NewEntry(logrus.New()),
		available:    true,
		kvmAvailable: false,
	}

	assert.False(t, clientWithoutKVM.IsKVMAvailable())
}

func TestGenerateDomainXMLArchitectureSpecific(t *testing.T) {
	// Create a mock client and domain manager for testing
	client := &Client{
		uri:          "qemu:///system",
		logger:       logrus.NewEntry(logrus.New()),
		available:    true,
		kvmAvailable: true,
	}

	dm := &DomainManager{
		client: client,
		logger: logrus.NewEntry(logrus.New()),
	}

	// Generate XML for the current architecture
	xml := dm.GenerateDomainXMLWithCPU("test-vm", "test-uuid", 1024, 2, "/path/to/disk", "test-network", "", "")

	// Test that the XML generation handles different architectures correctly
	// by checking the current architecture and verifying appropriate features
	currentArch := runtime.GOARCH
	if currentArch == "amd64" {
		// On x86_64, should include APIC and ACPI
		assert.Contains(t, xml, "<apic/>")
		assert.Contains(t, xml, "<suspend-to-mem enabled='no'/>")
		assert.Contains(t, xml, "<suspend-to-disk enabled='no'/>")
		assert.Contains(t, xml, "machine='pc-q35-2.12'")
		assert.Contains(t, xml, "<domain type='kvm'>")
		t.Logf("Verified x86_64 features are present")
	} else if currentArch == "arm64" {
		// On ARM64, should NOT include APIC and ACPI
		assert.NotContains(t, xml, "<apic/>")
		assert.NotContains(t, xml, "<suspend-to-mem enabled='no'/>")
		assert.NotContains(t, xml, "<suspend-to-disk enabled='no'/>")
		assert.Contains(t, xml, "machine='virt-8.2'")
		assert.Contains(t, xml, "<domain type='kvm'>")
		assert.Contains(t, xml, "<input type='keyboard' bus='virtio'/>")
		assert.Contains(t, xml, "<input type='mouse' bus='virtio'/>")
		t.Logf("Verified ARM64 features are excluded")
	} else {
		// For other architectures, just verify basic structure
		assert.Contains(t, xml, "<domain type='kvm'>")
		assert.Contains(t, xml, "<interface type='network'>")
		t.Logf("Verified basic structure for architecture: %s", currentArch)
	}

	// Always verify common features that should be present on all architectures
	assert.Contains(t, xml, "<interface type='network'>")
	assert.Contains(t, xml, "<model type='virtio'/>")
	assert.Contains(t, xml, "<console type='pty'>")
	assert.Contains(t, xml, "<graphics type='vnc'")
	assert.Contains(t, xml, "<memballoon model='virtio'/>")
	assert.Contains(t, xml, "<rng model='virtio'>")
}

func TestGenerateDomainXMLWithCPUModel(t *testing.T) {
	// Create a mock client and domain manager for testing
	client := &Client{
		uri:          "qemu:///system",
		logger:       logrus.NewEntry(logrus.New()),
		available:    true,
		kvmAvailable: true,
	}

	dm := &DomainManager{
		client: client,
		logger: logrus.NewEntry(logrus.New()),
	}

	// Test ARM64 CPU model selection
	if runtime.GOARCH == "arm64" {
		// Test with specific CPU model
		xml := dm.GenerateDomainXMLWithCPU("test-vm", "test-uuid", 1024, 2, "/path/to/disk", "test-network", "", "cortex-a72")
		assert.Contains(t, xml, "<model fallback='allow'>cortex-a72</model>")
		assert.Contains(t, xml, "<cpu mode='custom' match='exact'>")
		t.Logf("Verified ARM64 CPU model selection works")

		// Test with empty CPU model (should use default)
		xmlDefault := dm.GenerateDomainXMLWithCPU("test-vm", "test-uuid", 1024, 2, "/path/to/disk", "test-network", "", "")
		assert.Contains(t, xmlDefault, "<model fallback='allow'>cortex-a72</model>")
		t.Logf("Verified ARM64 default CPU model works")
	}

	// Test x86_64 CPU mode
	if runtime.GOARCH == "amd64" {
		xml := dm.GenerateDomainXMLWithCPU("test-vm", "test-uuid", 1024, 2, "/path/to/disk", "test-network", "", "")
		assert.Contains(t, xml, "<cpu mode='host-passthrough' check='partial'/>")
		t.Logf("Verified x86_64 CPU mode works")
	}
}

func TestCPUModelGenerationForARM64(t *testing.T) {
	// Create a mock client and domain manager for testing
	client := &Client{
		uri:          "qemu:///system",
		logger:       logrus.NewEntry(logrus.New()),
		available:    true,
		kvmAvailable: true,
	}

	dm := &DomainManager{
		client: client,
		logger: logrus.NewEntry(logrus.New()),
	}

	// Test ARM64 CPU mode generation
	if runtime.GOARCH == "arm64" {
		xml := dm.GenerateDomainXMLWithCPU("test-vm", "test-uuid", 1024, 2, "/path/to/disk", "test-network", "", "")

		// Verify it uses custom mode, not host-model
		assert.Contains(t, xml, "<cpu mode='custom' match='exact'>")
		assert.NotContains(t, xml, "host-model")
		assert.NotContains(t, xml, "host-passthrough")

		// Verify it uses a specific CPU model
		assert.Contains(t, xml, "<model fallback='allow'>cortex-a72</model>")

		t.Logf("✓ ARM64 CPU mode correctly uses custom mode with specific model")
		t.Logf("✓ No host-model or host-passthrough found (correct)")
	} else {
		t.Skip("This test is only for ARM64 systems")
	}
}
