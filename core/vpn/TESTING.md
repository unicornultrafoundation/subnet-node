# VPN Testing Guide

This document outlines the testing strategy and performance considerations for the VPN package.

## Package Organization

The VPN package uses a mixed approach to testing:

1. **Integration Tests (`package vpn_test`)**:
   - Tests that verify the public API works correctly
   - Tests that simulate real-world usage
   - Tests that verify integration with other packages
   - Files: `integration_table_test.go`, `enhanced_network_table_test.go`, `stress_test.go`, `performance_test.go`, `chaos_test.go`

2. **Unit Tests (`package vpn`)**:
   - Tests that verify internal implementation details
   - Tests that need access to unexported members
   - Files: All other `*_test.go` files

## Benefits of This Approach

### Integration Tests (`package vpn_test`)

- **Tests the Public API**: Verifies that the public API of the package works correctly, which is what clients of the package will use.
- **Enforces Clean API Design**: By only accessing exported members, we ensure that the package has a clean, usable public API.
- **Prevents Accidental Use of Internal Details**: Tests can't accidentally rely on implementation details that might change.
- **Avoids Circular Dependencies**: If we need to import packages that depend on the VPN package for testing, we can do so without circular import issues.

### Unit Tests (`package vpn`)

- **Access to Unexported Members**: Can directly test unexported (private) functions or types.
- **Simpler Setup**: Can directly access internal state without going through the public API.
- **More Focused**: Can focus on testing a single function or type without testing the entire system.

## Test Modes

The VPN tests can be run in different modes to balance between test coverage and execution time:

### 1. Short Mode

Short mode skips long-running tests like stress tests and enhanced network tests.

```bash
go test -short ./core/vpn/...
```

This is ideal for quick verification during development.

### 2. Normal Mode

Normal mode runs all tests but with reduced parameters for faster execution.

```bash
go test ./core/vpn/...
```

This is suitable for regular CI/CD pipelines.

### 3. Full Test Mode

Full test mode runs all tests with comprehensive parameters for thorough testing.

```bash
FULL_TEST=1 STRESS_TEST=1 go test ./core/vpn/...
```

This is recommended for nightly builds or before major releases.

## Test Categories

### Integration Tests

These tests verify the integration between different components using a table-driven approach:

```bash
go test ./core/vpn/integration_table_test.go
```

### Enhanced Network Tests

These tests simulate various network conditions using a table-driven approach:

```bash
go test ./core/vpn/enhanced_network_table_test.go
```

For full network condition testing:

```bash
FULL_TEST=1 go test ./core/vpn/enhanced_network_table_test.go
```

### Stress Tests

These tests verify the system under high load:

```bash
go test ./core/vpn/stress_test.go
```

For comprehensive stress testing:

```bash
STRESS_TEST=1 go test ./core/vpn/stress_test.go
```

### Performance Tests

These tests measure the performance of various components:

```bash
go test ./core/vpn/performance_test.go
```

Performance tests are skipped in short mode. They measure metrics such as:
- Throughput (packets/second, bytes/second)
- Latency (average, P50, P90, P99)
- Success/failure rates

Key performance test types:
- `TestStreamPoolPerformance`: Tests stream pool performance
- `TestPacketDispatcherPerformance`: Tests packet dispatcher performance
- `TestResiliencePerformance`: Tests resilience patterns performance

### Chaos Tests

These tests verify the system's behavior under chaotic conditions:

```bash
go test ./core/vpn/chaos_test.go
```

Chaos tests are skipped in short mode. They simulate various failure scenarios:
- `TestChaosEngineering`: Random fault injections
- `TestNetworkPartition`: Network partitions
- `TestHighLatency`: High latency conditions

These tests use the fault injection capabilities in the testutil package to simulate failures.

## Performance Optimization Tips

1. **Run Tests in Parallel**: Use `-parallel` flag to run tests in parallel.

   ```bash
   go test -parallel 4 ./core/vpn/...
   ```

2. **Run Specific Tests**: Use `-run` flag to run specific tests.

   ```bash
   go test -run TestStreamPool ./core/vpn/...
   ```

3. **Skip Long-Running Tests**: Use `-short` flag to skip long-running tests.

   ```bash
   go test -short ./core/vpn/...
   ```

4. **Profile Tests**: Use `-cpuprofile` and `-memprofile` flags to profile tests.

   ```bash
   go test -cpuprofile cpu.prof -memprofile mem.prof ./core/vpn/...
   ```

## CI/CD Integration

For CI/CD pipelines, we recommend:

1. **Pull Requests**: Run short mode tests for quick feedback.

   ```bash
   go test -short ./core/vpn/...
   ```

2. **Main Branch**: Run normal mode tests for regular verification.

   ```bash
   go test ./core/vpn/...
   ```

3. **Nightly Builds**: Run full test mode for comprehensive testing.

   ```bash
   FULL_TEST=1 STRESS_TEST=1 go test ./core/vpn/...
   ```

## Adding New Tests

When adding new tests, consider whether they should be integration tests or unit tests:

- If the test verifies the public API or simulates real-world usage, it should be an integration test (`package vpn_test`).
- If the test verifies internal implementation details or needs access to unexported members, it should be a unit test (`package vpn`).

### Table-Driven Test Approach

We use a table-driven approach for most integration tests. This approach has several benefits:

- **Maintainability**: Makes it easier to add new test cases without duplicating code
- **Readability**: Clearly separates test setup, execution, and verification
- **Consistency**: Ensures all test cases follow the same pattern
- **Comprehensive**: Makes it easier to test multiple scenarios

Example of a table-driven test:

```go
func TestTableDrivenExample(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "hello",
            expected: "HELLO",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test setup
            // ...

            // Execute the function being tested
            result, err := functionBeingTested(tc.input)

            // Verify results
            if tc.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.expected, result)
            }
        })
    }
}
```

### Using the TestUtil Package

The `testutil` package provides a comprehensive set of utilities for testing the VPN package. Key features include:

1. **Test Fixtures**: Complete test environments for VPN tests
2. **Mock Implementations**: Mock implementations of interfaces
3. **Test Helpers**: Utilities for creating test data and verifying results
4. **Network Simulation**: Utilities for simulating various network conditions
5. **Performance Testing**: Utilities for measuring performance
6. **Fault Injection**: Utilities for injecting faults to test resilience

Example of using the TestFixture:

```go
// Create a test fixture with specific network conditions
fixture := testutil.NewTestFixture(t, &testutil.NetworkCondition{
    Name:       "poor_network",
    Latency:    200 * time.Millisecond,
    Jitter:     50 * time.Millisecond,
    PacketLoss: 0.1,  // 10% packet loss
    Bandwidth:  1000, // 1 Mbps
}, 3)
defer fixture.Cleanup()

// Use the fixture to test your code
testutil.VerifyPacketDelivery(t, fixture.Dispatcher, "192.168.1.1:80", "192.168.1.1", testPacket)
```

### Performance Testing Best Practices

1. **Establish Baselines**: Create baseline performance metrics to detect regressions
2. **Test Under Various Loads**: Test with light, moderate, and heavy loads
3. **Use Realistic Data**: Test with realistic packet sizes and patterns
4. **Measure Multiple Metrics**: Track throughput, latency, and resource usage
5. **Use Percentiles**: Look at P50, P90, and P99 latencies, not just averages
6. **Include Warm-up**: Allow the system to stabilize before taking measurements
7. **Automate**: Run performance tests regularly in CI/CD pipelines

### Chaos Testing Best Practices

1. **Start Small**: Begin with mild chaos and gradually increase intensity
2. **Test Realistic Scenarios**: Focus on failures that could occur in production
3. **Verify Recovery**: Ensure the system recovers automatically after failures
4. **Test Circuit Breakers**: Verify circuit breakers open and close as expected
5. **Test Retries**: Ensure retry mechanisms work correctly under adverse conditions
6. **Test Resource Cleanup**: Verify resources are properly cleaned up after failures
7. **Document Findings**: Record any issues discovered during chaos testing

For more details, see the documentation in the `testutil` package.
