# Kubernetes Cluster Orchestration Service

## Overview
This service provides automated orchestration of Kubernetes deployments based on SDL (Service Definition Language) configurations and integrates with an Ethereum-based marketplace for resource rental. The service acts as a provider in the marketplace, responding to deployment requests and managing the lifecycle of deployments.

## Core Components

### 1. Service (`service.go`)
- Main service orchestrator
- Manages deployment lifecycle
- Handles marketplace events
- Provides health monitoring
- Manages resource allocation
- Integrates with payment system

### 2. Deployment Manager (`deployment/deployment_manager.go`)
- Manages Kubernetes deployments
- Handles deployment lifecycle
- Provides deployment status updates
- Manages deployment resources
- Handles deployment scaling

### 3. Resource Monitor (`monitor/resource_monitor.go`)
- Tracks resource usage (CPU, Memory, Storage, GPU)
- Generates resource alerts
- Monitors resource trends
- Provides resource reports
- Handles resource scaling

### 4. Security Manager (`security/`)
- Network policy management
- Service mesh configuration
- Pod security policies
- Access control (RBAC)
- Security auditing

### 5. Payment Manager (`payment/`)
- Handles escrow management
- Manages payment processing
- Tracks deployment status
- Handles payment releases
- Manages payment history

### 6. Event System (`events/`)
- Handles marketplace events
- Manages deployment events
- Processes bid events
- Manages provider selection
- Handles deployment status events

## Implementation Details

### Event Flow
1. **Deployment Request Phase**
   - Receive `DeploymentRequestedEvent`
   - Parse SDL configuration
   - Calculate resource requirements
   - Submit bid to marketplace
   - Publish `DeploymentRequestReceivedEvent`

2. **Bidding Phase**
   - Handle `BidSubmittedEvent`
   - Validate bid
   - Check escrow status
   - Track bids
   - Select best provider
   - Publish `ProviderSelectedEvent`

3. **Deployment Phase**
   - Handle `DeploymentApprovedEvent`
   - Create Kubernetes namespace
   - Configure networking and security
   - Deploy services
   - Set up monitoring
   - Publish `DeploymentCompletedEvent`

4. **Monitoring Phase**
   - Track resource usage
   - Monitor health status
   - Generate alerts
   - Update deployment status
   - Handle escrow status

5. **Termination Phase**
   - Handle `DeploymentTerminatedEvent`
   - Stop deployment
   - Clean up resources
   - Delete namespace
   - Update marketplace status
   - Handle escrow release/refund

### Service and Ingress Flow
1. **Service Creation**
   - Services are created during deployment phase
   - Each service is configured with:
     - Selector labels to target pods
     - Port mappings (internal and external)
     - Service type (ClusterIP, NodePort, or LoadBalancer)
   - Services are created in the deployment's namespace
   - Network policies are applied to control service access

2. **Ingress Configuration**
   - Ingress resources are created for services that need external access
   - Ingress configuration includes:
     - Hostname routing
     - Path-based routing
     - TLS termination
     - Load balancing rules
   - Ingress controller (nginx) handles external traffic
   - Ingress class is set to "subnet-node-ingress-class"

3. **Network Policies**
   - Default deny policy for all pods
   - Allow ingress traffic to exposed ports
   - Allow egress traffic based on configuration
   - DNS policy for name resolution
   - Kubernetes API access policy
   - Ingress and egress policies for service mesh

4. **Service Mesh Integration** (Optional)
   - Service mesh configuration for traffic management
   - mTLS for service-to-service communication
   - Traffic splitting and routing rules
   - Load balancing configuration
   - Circuit breaking and retry policies

5. **Hostname Management**
   - Hostname to service mapping
   - DNS record management
   - TLS certificate management
   - Custom domain support
   - Hostname validation and verification

### Resource Management
- CPU and Memory limits from pod specifications
- Storage volume management
- Network bandwidth monitoring
- GPU resource allocation
- Resource usage alerts
- Resource quotas per namespace

### Security Features
- Network policies (default deny)
- Service mesh integration
- Pod security policies
- RBAC configuration
- Security context constraints
- Audit logging

## Technical Requirements

### Backend
- Go 1.23+
- Kubernetes client-go v0.32.0
- Ethereum client (go-ethereum v1.13.5)
- Zap logger
- Event bus system

### Dependencies
- k8s.io/api v0.32.0
- k8s.io/apimachinery v0.32.0
- k8s.io/client-go v0.32.0
- k8s.io/metrics v0.32.0
- go.uber.org/zap v1.26.0
- github.com/ethereum/go-ethereum

## API Endpoints

### Deployment Management
```http
POST /api/v1/deploy
GET /api/v1/deployments
GET /api/v1/deployments/{id}
DELETE /api/v1/deployments/{id}
```

### Resource Monitoring
```http
GET /api/v1/resources
GET /api/v1/metrics/{deployment_id}
GET /api/v1/alerts
```

### Health Status
```http
GET /api/v1/health
GET /api/v1/status
```

## Event Types

### Deployment Events
- `DeploymentRequestedEvent`
- `DeploymentRequestReceivedEvent`
- `DeploymentApprovedEvent`
- `DeploymentCompletedEvent`
- `DeploymentTerminatedEvent`

### Bid Events
- `BidSubmittedEvent`
- `BidClosedEvent`
- `ProviderSelectedEvent`

### Payment Events
- `PaymentCreatedEvent`
- `PaymentReleasedEvent`

## Next Steps
1. Enhance service mesh configuration
2. Implement advanced resource scheduling
3. Add cost optimization features
4. Improve monitoring capabilities
5. Add multi-cluster support
6. Enhance security features

## Future Enhancements
1. Multi-cluster support
2. Advanced resource scheduling
3. Automated scaling
4. Cost optimization
5. Enhanced marketplace features
6. Resource verification
7. Performance optimization 