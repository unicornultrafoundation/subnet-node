/*
Package testutil provides testing utilities for the VPN package.

This package contains a comprehensive set of testing utilities for the VPN package, including:

1. Test Fixtures:
  - TestFixture: A complete test environment for VPN tests
  - NetworkCondition: Simulated network conditions for testing

2. Mock Implementations:
  - MockStream: A mock implementation of the Stream interface
  - MockStreamService: A mock implementation of the StreamService interface (for testing)
  - MockDiscoveryService: A mock implementation of the DiscoveryService interface
  - MockPoolService: A mock implementation of the PoolService interface

3. Test Helpers:
  - CreateTestPacket: Creates a test packet of a specified size
  - VerifyPacketDelivery: Verifies that a packet was delivered successfully
  - VerifyMetrics: Verifies metrics
  - TestContext: Creates a context with timeout for testing
  - TestWithRetry: Runs a test with retries
  - SortLatencies: Sorts a slice of latency measurements

4. Setup Helpers:
  - SetupMockStream: Sets up a mock stream with specified configuration
  - SetupMockStreamService: Sets up a mock stream service with specified configuration
  - SetupDiscoveryService: Sets up a mock discovery service with specified configuration
  - SetupResilienceService: Sets up a resilience service with specified configuration
  - SetupTestStreamPool: Sets up a test stream pool
  - SetupTestDispatcher: Sets up a test packet dispatcher

5. Performance Testing:
  - PerformanceConfig: Configuration for performance tests
  - PerformanceResult: Results of a performance test
  - RunStreamPerformanceTest: Runs a performance test on a stream pool
  - RunDispatcherPerformanceTest: Runs a performance test on a packet dispatcher
  - VerifyPerformanceRequirements: Verifies performance requirements
  - DefaultPerformanceConfig: Returns a default performance test configuration

6. Fault Injection:
  - FaultInjectionConfig: Configuration for fault injection
  - FaultInjector: Injects faults into the system for testing
  - InjectNetworkPartition: Simulates a network partition
  - InjectPacketLoss: Simulates packet loss
  - InjectLatency: Simulates network latency
  - InjectJitter: Simulates network jitter
  - InjectStreamFailures: Simulates stream failures
  - InjectDiscoveryFailures: Simulates discovery failures

7. Chaos Testing:
  - RunChaosTest: Runs a chaos test with random fault injections

Usage Examples:

 1. Basic Test Fixture:
    ```go
    // Create a test fixture with good network conditions
    fixture := testutil.NewTestFixture(t, nil, 3)
    defer fixture.Cleanup()

    // Run a basic test
    fixture.RunBasicTest(t)
    ```

 2. Network Conditions Testing:
    ```go
    // Create a test fixture with poor network conditions
    fixture := testutil.NewTestFixture(t, &testutil.NetworkCondition{
    Name:       "poor_network",
    Latency:    200 * time.Millisecond,
    Jitter:     50 * time.Millisecond,
    PacketLoss: 0.1,  // 10% packet loss
    Bandwidth:  1000, // 1 Mbps
    }, 3)
    defer fixture.Cleanup()

    // Test with poor network conditions
    testutil.VerifyPacketDelivery(t, fixture.Dispatcher, "192.168.1.1:80", "192.168.1.1", testPacket)
    ```

 3. Performance Testing:
    ```go
    // Create a performance test configuration
    config := &testutil.PerformanceConfig{
    PacketSize:       1024,
    PacketsPerSecond: 1000,
    Duration:         5 * time.Second,
    Concurrency:      4,
    WarmupDuration:   1 * time.Second,
    CooldownDuration: 1 * time.Second,
    }

    // Run a performance test
    result := testutil.RunDispatcherPerformanceTest(t, fixture.Dispatcher, "192.168.1.1", config)

    // Log performance metrics
    t.Logf("Performance: %d packets/sec, %v latency", int(result.PacketsPerSecond), result.AverageLatency)
    ```

 4. Fault Injection:
    ```go
    // Create a fault injector
    injector := testutil.NewFaultInjector(
    nil,
    fixture.MockStreamService,
    fixture.MockDiscoveryService,
    fixture.Dispatcher,
    fixture.StreamService,
    )

    // Inject a network partition
    injector.InjectNetworkPartition(5 * time.Second)

    // Test system behavior during the partition
    // ...
    ```

 5. Chaos Testing:
    ```go
    // Run a chaos test
    testutil.RunChaosTest(t, fixture, 30 * time.Second, 0.5) // 30 seconds, 50% intensity
    ```

Best Practices:
1. Always use defer fixture.Cleanup() to ensure resources are cleaned up
2. Use table-driven tests for testing multiple scenarios
3. Use realistic network conditions for testing
4. Test both happy and error paths
5. Use the resilience patterns for handling failures
6. Use performance tests to establish baselines and detect regressions
7. Use chaos tests to verify system behavior under adverse conditions
8. Use fault injection to test specific failure scenarios
*/
package testutil
