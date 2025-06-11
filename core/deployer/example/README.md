# Kubernetes Cluster Example

This example demonstrates how to deploy a simple HTTP echo service using the deployer package. The example shows the complete lifecycle of a deployment, from creation to cleanup.

## Prerequisites

- Go 1.16 or later
- Kubernetes cluster (local or remote)
- kubectl configured with access to your cluster
- Docker (for building and running containers)

## Directory Structure

```
example/
├── main.go           # Main example program
├── mock_ipfs.go    # Mock IPFS client implementation
└── echo-service.yaml # Service definition file
```

## Service Definition

The example uses a simple HTTP echo service defined in `echo-service.yaml`. The service:
- Uses the `hashicorp/http-echo` image
- Exposes port 5678 internally and 80 externally
- Has resource limits for CPU, memory, and storage
- Includes health check configuration

## Running the Example

1. Make sure you're in the example directory:
```bash
cd core/deployer/example
```

2. Run the example:
```bash
go run .
```

The program will:
1. Parse and validate the SDL file
2. Create a mock IPFS client
3. Start the deployer service
4. Deploy the echo service
5. Monitor the deployment
6. Test the service
7. Clean up resources

## Command Line Options

The example supports the following command line flags:

- `-sdl`: Path to the SDL file (default: "echo-service.yaml")
- `-kubeconfig`: Path to kubeconfig file (default: "~/.kube/config")
- `-verbose`: Enable verbose logging (default: false)

Example with custom options:
```bash
go run . -sdl custom-service.yaml -kubeconfig /path/to/kubeconfig -verbose
```

## Expected Output

When running successfully, you should see:
1. Service deployment logs
2. Pod status updates
3. Service accessibility confirmation
4. Successful HTTP response from the echo service
5. Cleanup confirmation

Example successful output:
```
Service response: "Hello from http-echo!"
Deployment is running successfully!
```

## Troubleshooting

1. **Kubernetes Connection Issues**
   - Verify your kubeconfig is valid
   - Check cluster accessibility
   - Ensure you have necessary permissions

2. **Deployment Failures**
   - Check pod status: `kubectl get pods -n deployment-default`
   - View pod logs: `kubectl logs <pod-name> -n deployment-default`
   - Check service status: `kubectl get svc -n deployment-default`

3. **Service Access Issues**
   - Verify service is running: `kubectl get svc -n deployment-default`
   - Check service endpoints: `kubectl get endpoints -n deployment-default`
   - Test service directly: `curl http://<service-ip>:<port>`

## Cleanup

The example automatically cleans up resources when it exits. If you need to manually clean up:

```bash
kubectl delete namespace deployment-default
```

## Implementation Details

The example demonstrates several key features:

1. **Service Deployment**
   - Namespace creation
   - Manifest creation
   - Pod deployment
   - Service exposure

2. **Health Monitoring**
   - Pod status checking
   - Service accessibility verification
   - Health check implementation

3. **Resource Management**
   - Resource limits and requests
   - Storage configuration
   - Network policy setup

4. **Event Handling**
   - Deployment events
   - Bid submission
   - Provider selection
   - Deployment completion

## Contributing

Feel free to submit issues and enhancement requests! 