package testutil

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

// FaultInjectionConfig contains configuration for fault injection
type FaultInjectionConfig struct {
	// Network faults
	PacketLossRate     float64
	LatencyMean        time.Duration
	LatencyJitter      time.Duration
	ConnectionFailRate float64

	// Service faults
	DiscoveryFailRate float64
	StreamFailRate    float64

	// System faults
	CPULoadPercentage int
	MemoryUsageMB     int

	// Timing
	FaultDuration    time.Duration
	RecoveryDuration time.Duration
}

// DefaultFaultInjectionConfig returns a default fault injection configuration
func DefaultFaultInjectionConfig() *FaultInjectionConfig {
	return &FaultInjectionConfig{
		PacketLossRate:     0.05,
		LatencyMean:        100 * time.Millisecond,
		LatencyJitter:      50 * time.Millisecond,
		ConnectionFailRate: 0.02,
		DiscoveryFailRate:  0.03,
		StreamFailRate:     0.04,
		CPULoadPercentage:  50,
		MemoryUsageMB:      100,
		FaultDuration:      5 * time.Second,
		RecoveryDuration:   5 * time.Second,
	}
}

// FaultInjector injects faults into the system for testing
type FaultInjector struct {
	config            *FaultInjectionConfig
	mockStreamService *MockStreamService
	mockDiscovery     *MockDiscoveryService
	dispatcher        *DispatcherAdapter

	// State
	active         bool
	originalConfig *FaultInjectionConfig
	mutex          sync.Mutex
}

// NewFaultInjector creates a new fault injector
func NewFaultInjector(
	config *FaultInjectionConfig,
	mockStreamService *MockStreamService,
	mockDiscovery *MockDiscoveryService,
	dispatcher *DispatcherAdapter,
) *FaultInjector {
	if config == nil {
		config = DefaultFaultInjectionConfig()
	}

	return &FaultInjector{
		config:            config,
		mockStreamService: mockStreamService,
		mockDiscovery:     mockDiscovery,
		dispatcher:        dispatcher,
		active:            false,
		originalConfig:    nil,
		mutex:             sync.Mutex{},
	}
}

// Start starts fault injection
func (f *FaultInjector) Start() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.active {
		return
	}

	// Save original configuration
	f.originalConfig = &FaultInjectionConfig{
		PacketLossRate:     f.mockStreamService.GetPacketLossRate(),
		LatencyMean:        f.mockStreamService.GetLatency(),
		LatencyJitter:      f.mockStreamService.GetJitter(),
		ConnectionFailRate: f.mockStreamService.GetFailureRate(),
		DiscoveryFailRate:  f.mockDiscovery.GetFailureRate(),
		StreamFailRate:     f.mockStreamService.GetFailureRate(),
	}

	// Apply fault configuration
	f.mockStreamService.SetPacketLossRate(f.config.PacketLossRate)
	f.mockStreamService.SetLatency(f.config.LatencyMean)
	f.mockStreamService.SetJitter(f.config.LatencyJitter)
	f.mockStreamService.SetFailureRate(f.config.StreamFailRate)
	f.mockDiscovery.SetFailureRate(f.config.DiscoveryFailRate)

	f.active = true
}

// Stop stops fault injection
func (f *FaultInjector) Stop() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !f.active || f.originalConfig == nil {
		return
	}

	// Restore original configuration
	f.mockStreamService.SetPacketLossRate(f.originalConfig.PacketLossRate)
	f.mockStreamService.SetLatency(f.originalConfig.LatencyMean)
	f.mockStreamService.SetJitter(f.originalConfig.LatencyJitter)
	f.mockStreamService.SetFailureRate(f.originalConfig.StreamFailRate)
	f.mockDiscovery.SetFailureRate(f.originalConfig.DiscoveryFailRate)

	f.active = false
}

// InjectNetworkPartition simulates a network partition
func (f *FaultInjector) InjectNetworkPartition(duration time.Duration) {
	// Save current failure rate
	originalRate := f.mockStreamService.GetFailureRate()

	// Set 100% failure rate to simulate network partition
	f.mockStreamService.SetFailureRate(1.0)

	// Wait for the specified duration
	time.Sleep(duration)

	// Restore original failure rate
	f.mockStreamService.SetFailureRate(originalRate)
}

// InjectPacketLoss simulates packet loss
func (f *FaultInjector) InjectPacketLoss(rate float64, duration time.Duration) {
	// Save current packet loss rate
	originalRate := f.mockStreamService.GetPacketLossRate()

	// Set specified packet loss rate
	f.mockStreamService.SetPacketLossRate(rate)

	// Wait for the specified duration
	time.Sleep(duration)

	// Restore original packet loss rate
	f.mockStreamService.SetPacketLossRate(originalRate)
}

// InjectLatency simulates network latency
func (f *FaultInjector) InjectLatency(latency time.Duration, duration time.Duration) {
	// Save current latency
	originalLatency := f.mockStreamService.GetLatency()

	// Set specified latency
	f.mockStreamService.SetLatency(latency)

	// Wait for the specified duration
	time.Sleep(duration)

	// Restore original latency
	f.mockStreamService.SetLatency(originalLatency)
}

// InjectStreamFailures simulates stream failures
func (f *FaultInjector) InjectStreamFailures(rate float64, duration time.Duration) {
	// Save current failure rate
	originalRate := f.mockStreamService.GetFailureRate()

	// Set specified failure rate
	f.mockStreamService.SetFailureRate(rate)

	// Wait for the specified duration
	time.Sleep(duration)

	// Restore original failure rate
	f.mockStreamService.SetFailureRate(originalRate)
}

// InjectDiscoveryFailures simulates discovery failures
func (f *FaultInjector) InjectDiscoveryFailures(rate float64, duration time.Duration) {
	// Save current failure rate
	originalRate := f.mockDiscovery.GetFailureRate()

	// Set specified failure rate
	f.mockDiscovery.SetFailureRate(rate)

	// Wait for the specified duration
	time.Sleep(duration)

	// Restore original failure rate
	f.mockDiscovery.SetFailureRate(originalRate)
}

// RunChaosTest runs a chaos test with random fault injections
func RunChaosTest(t *testing.T, fixture *TestFixture, duration time.Duration, intensity float64) {
	// Create a fault injector
	injector := NewFaultInjector(
		nil,
		fixture.MockStreamService,
		fixture.MockDiscoveryService,
		fixture.Dispatcher,
	)

	// Start time
	startTime := time.Now()

	// Run until the specified duration has elapsed
	for time.Since(startTime) < duration {
		// Randomly select a fault to inject
		faultType := rand.Intn(5)

		// Calculate fault duration based on intensity
		faultDuration := time.Duration(float64(time.Second) * intensity)

		// Inject the selected fault
		switch faultType {
		case 0:
			// Network partition
			t.Logf("Injecting network partition for %v", faultDuration)
			injector.InjectNetworkPartition(faultDuration)
		case 1:
			// Packet loss
			rate := rand.Float64() * intensity
			t.Logf("Injecting packet loss (rate: %.2f) for %v", rate, faultDuration)
			injector.InjectPacketLoss(rate, faultDuration)
		case 2:
			// Latency
			latency := time.Duration(rand.Float64() * float64(time.Second) * intensity)
			t.Logf("Injecting latency (%v) for %v", latency, faultDuration)
			injector.InjectLatency(latency, faultDuration)
		case 3:
			// Stream failures
			rate := rand.Float64() * intensity
			t.Logf("Injecting stream failures (rate: %.2f) for %v", rate, faultDuration)
			injector.InjectStreamFailures(rate, faultDuration)
		case 4:
			// Discovery failures
			rate := rand.Float64() * intensity
			t.Logf("Injecting discovery failures (rate: %.2f) for %v", rate, faultDuration)
			injector.InjectDiscoveryFailures(rate, faultDuration)
		}

		// Wait for a shorter random period before the next fault
		waitTime := time.Duration(rand.Float64() * float64(time.Second) * 0.5)
		time.Sleep(waitTime)
	}
}

// Add methods to MockStreamService and MockDiscoveryService to support fault injection

// GetPacketLossRate returns the current packet loss rate
func (m *MockStreamService) GetPacketLossRate() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.packetLossRate
}

// SetPacketLossRate sets the packet loss rate
func (m *MockStreamService) SetPacketLossRate(rate float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.packetLossRate = rate
}

// GetLatency returns the current latency
func (m *MockStreamService) GetLatency() time.Duration {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.latency
}

// SetLatency sets the latency
func (m *MockStreamService) SetLatency(latency time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.latency = latency
}

// GetJitter returns the current jitter
func (m *MockStreamService) GetJitter() time.Duration {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.jitter
}

// SetJitter sets the jitter
func (m *MockStreamService) SetJitter(jitter time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.jitter = jitter
}

// GetFailureRate returns the current failure rate
func (m *MockStreamService) GetFailureRate() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.failureRate
}

// SetFailureRate sets the failure rate
func (m *MockStreamService) SetFailureRate(rate float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failureRate = rate
}

// GetFailureRate returns the current failure rate
func (m *MockDiscoveryService) GetFailureRate() float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.failureRate
}

// SetFailureRate sets the failure rate
func (m *MockDiscoveryService) SetFailureRate(rate float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failureRate = rate
}
