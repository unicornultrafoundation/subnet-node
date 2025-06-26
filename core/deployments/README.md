# Deployments Module

The Deployments module provides a unified interface for managing Kubernetes deployments in the subnet-node project. It supports Kubernetes-based deployments with comprehensive monitoring, logging, and management capabilities.

## Architecture

```
core/deployments/
├── types.go              # Common types and interfaces
├── interfaces.go          # Core interfaces for deployment management
├── service.go            # Main deployment service
├── factory.go            # Factory for creating deployment managers
├── api.go                # REST API endpoints
├── server.go             # HTTP server implementation
├── api_test.go           # API tests
├── example_server.go     # Example server implementation
├── AUTHORIZATION.md      # Authorization documentation
└── README.md            # This documentation
```

## API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication

All API endpoints require authentication using Ethereum signatures. The authentication header should contain:

```json
{
  "signature": "base64_encoded_signature",
  "address": "0x123456789abcdef123456789abcdef123456789a",
  "message": "deployment_id:action:timestamp",
  "timestamp": 1640995200
}
```

**Signature Format:**
- `message`: `{deployment_id}:{action}:{timestamp}`
- `timestamp`: Unix timestamp (valid for 5 minutes)
- `signature`: SHA256 signature of the message using the private key

### Rate Limiting

- **Limit**: 5 requests per minute per address
- **Headers**: Rate limit information is returned in response headers
- **Status**: 429 Too Many Requests when limit exceeded

---

## Endpoints

### 1. Health Check

#### GET `/health`
Check if the API server is healthy.

**Request:**
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Server is healthy
- `503 Service Unavailable`: Server is unhealthy

---

### 2. Ready Check

#### GET `/ready`
Check if the API server is ready to handle requests.

**Request:**
```http
GET /ready
```

**Response:**
```json
{
  "status": "ready",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Status Codes:**
- `200 OK`: Server is ready
- `503 Service Unavailable`: Server is not ready

---

### 3. Create Deployment

#### POST `/deployments`
Create a new deployment.

**Request:**
```http
POST /deployments
Content-Type: application/json
Authorization: {"signature":"...","address":"0x...","message":"123:create:1640995200","timestamp":1640995200}

{
  "id": "123",
  "name": "web-application",
  "image": "nginx:latest",
  "ports": [
    {
      "containerPort": 80,
      "hostPort": 8080,
      "protocol": "tcp"
    }
  ],
  "environment": {
    "NODE_ENV": "production",
    "DATABASE_URL": "postgresql://user:pass@localhost:5432/db"
  },
  "resources": {
    "cpu": "1000m",
    "memory": "1Gi",
    "storage": "10Gi"
  },
  "replicas": 3
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "web-application",
    "owner": "0x123456789abcdef123456789abcdef123456789a",
    "status": "created",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**Status Codes:**
- `201 Created`: Deployment created successfully
- `400 Bad Request`: Invalid request body or deployment ID format
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the order owner
- `404 Not Found`: Order not found
- `409 Conflict`: Deployment already exists
- `429 Too Many Requests`: Rate limit exceeded

**Validation:**
- Order must exist in the bid market
- Signer must be the order owner
- Order must be in "open" status
- Order must not be expired
- Deployment ID must be unique

---

### 4. List Deployments

#### GET `/deployments`
List all deployments.

**Request:**
```http
GET /deployments
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "123",
      "name": "web-application",
      "owner": "0x123456789abcdef123456789abcdef123456789a",
      "status": "running",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    },
    {
      "id": "124",
      "name": "api-service",
      "owner": "0xabcdef123456789abcdef123456789abcdef1234",
      "status": "stopped",
      "created_at": "2024-01-01T13:00:00Z",
      "updated_at": "2024-01-01T13:00:00Z"
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Deployments retrieved successfully
- `500 Internal Server Error`: Server error

---

### 5. Get Deployment

#### GET `/deployments/{id}`
Get a specific deployment by ID.

**Request:**
```http
GET /deployments/123
Authorization: {"signature":"...","address":"0x...","message":"123:read:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "web-application",
    "owner": "0x123456789abcdef123456789abcdef123456789a",
    "status": "running",
    "image": "nginx:latest",
    "ports": [
      {
        "containerPort": 80,
        "hostPort": 8080,
        "protocol": "tcp"
      }
    ],
    "environment": {
      "NODE_ENV": "production",
      "DATABASE_URL": "postgresql://user:pass@localhost:5432/db"
    },
    "resources": {
      "cpu": "1000m",
      "memory": "1Gi",
      "storage": "10Gi"
    },
    "replicas": 3,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**Status Codes:**
- `200 OK`: Deployment retrieved successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found

---

### 6. Start Deployment

#### POST `/deployments/{id}/start`
Start a deployment.

**Request:**
```http
POST /deployments/123/start
Authorization: {"signature":"...","address":"0x...","message":"123:start:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true
}
```

**Status Codes:**
- `200 OK`: Deployment started successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found
- `500 Internal Server Error`: Failed to start deployment

---

### 7. Stop Deployment

#### POST `/deployments/{id}/stop`
Stop a deployment.

**Request:**
```http
POST /deployments/123/stop
Authorization: {"signature":"...","address":"0x...","message":"123:stop:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true
}
```

**Status Codes:**
- `200 OK`: Deployment stopped successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found
- `500 Internal Server Error`: Failed to stop deployment

---

### 8. Restart Deployment

#### POST `/deployments/{id}/restart`
Restart a deployment.

**Request:**
```http
POST /deployments/123/restart
Authorization: {"signature":"...","address":"0x...","message":"123:restart:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true
}
```

**Status Codes:**
- `200 OK`: Deployment restarted successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found
- `500 Internal Server Error`: Failed to restart deployment

---

### 9. Delete Deployment

#### DELETE `/deployments/{id}`
Delete a deployment.

**Request:**
```http
DELETE /deployments/123
Authorization: {"signature":"...","address":"0x...","message":"123:delete:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true
}
```

**Status Codes:**
- `200 OK`: Deployment deleted successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found
- `500 Internal Server Error`: Failed to delete deployment

---

### 10. Inspect Deployment

#### GET `/deployments/{id}/inspect`
Get detailed information about a deployment.

**Request:**
```http
GET /deployments/123/inspect
Authorization: {"signature":"...","address":"0x...","message":"123:inspect:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "web-application",
    "status": "running",
    "services": {
      "web": {
        "name": "web",
        "image": "nginx:latest",
        "status": "running",
        "ports": [
          {
            "containerPort": 80,
            "hostPort": 8080,
            "protocol": "tcp"
          }
        ],
        "resources": {
          "cpu": "1000m",
          "memory": "1Gi"
        },
        "replicas": {
          "desired": 3,
          "current": 3,
          "ready": 3,
          "available": 3
        }
      }
    },
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

**Status Codes:**
- `200 OK`: Deployment inspection retrieved successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found

---

### 11. Inspect Service

#### GET `/deployments/{id}/services/{service}/inspect`
Get detailed information about a specific service in a deployment.

**Request:**
```http
GET /deployments/123/services/web/inspect
Authorization: {"signature":"...","address":"0x...","message":"123:inspect:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "name": "web",
    "image": "nginx:latest",
    "status": "running",
    "ports": [
      {
        "containerPort": 80,
        "hostPort": 8080,
        "protocol": "tcp"
      }
    ],
    "resources": {
      "cpu": "1000m",
      "memory": "1Gi"
    },
    "replicas": {
      "desired": 3,
      "current": 3,
      "ready": 3,
      "available": 3
    },
    "pods": [
      {
        "name": "web-123-abc123",
        "status": "running",
        "ip": "10.0.0.1",
        "node": "node-1"
      }
    ]
  }
}
```

**Status Codes:**
- `200 OK`: Service inspection retrieved successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment or service not found

---

### 12. Get Deployment Metrics

#### GET `/deployments/{id}/metrics`
Get metrics for a deployment.

**Request:**
```http
GET /deployments/123/metrics?duration=1h
Authorization: {"signature":"...","address":"0x...","message":"123:metrics:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "deployment_id": "123",
    "duration": "1h",
    "total": {
      "cpu_usage": 45.2,
      "memory_usage": 1073741824,
      "disk_usage": 5368709120,
      "network_rx": 1048576,
      "network_tx": 2097152
    },
    "services": {
      "web": {
        "cpu_usage": 45.2,
        "memory_usage": 1073741824,
        "disk_usage": 5368709120,
        "network_rx": 1048576,
        "network_tx": 2097152
      }
    },
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

**Query Parameters:**
- `duration`: Time duration for metrics (e.g., "1h", "30m", "24h")

**Status Codes:**
- `200 OK`: Metrics retrieved successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found

---

### 13. Get Service Metrics

#### GET `/deployments/{id}/services/{service}/metrics`
Get metrics for a specific service.

**Request:**
```http
GET /deployments/123/services/web/metrics?duration=1h
Authorization: {"signature":"...","address":"0x...","message":"123:metrics:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "service_name": "web",
    "duration": "1h",
    "cpu_usage": 45.2,
    "memory_usage": 1073741824,
    "disk_usage": 5368709120,
    "network_rx": 1048576,
    "network_tx": 2097152,
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

**Status Codes:**
- `200 OK`: Service metrics retrieved successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment or service not found

---

### 14. Get Deployment Logs

#### GET `/deployments/{id}/logs`
Get logs for a deployment.

**Request:**
```http
GET /deployments/123/logs?tail=100
Authorization: {"signature":"...","address":"0x...","message":"123:logs:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "timestamp": "2024-01-01T12:00:00Z",
      "service": "web",
      "pod": "web-123-abc123",
      "level": "INFO",
      "message": "Server started on port 80"
    },
    {
      "timestamp": "2024-01-01T12:00:01Z",
      "service": "web",
      "pod": "web-123-def456",
      "level": "INFO",
      "message": "Server started on port 80"
    }
  ]
}
```

**Query Parameters:**
- `tail`: Number of log lines to retrieve (default: 100)

**Status Codes:**
- `200 OK`: Logs retrieved successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found

---

### 15. Get Service Logs

#### GET `/deployments/{id}/services/{service}/logs`
Get logs for a specific service.

**Request:**
```http
GET /deployments/123/services/web/logs?tail=100
Authorization: {"signature":"...","address":"0x...","message":"123:logs:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "timestamp": "2024-01-01T12:00:00Z",
      "service": "web",
      "pod": "web-123-abc123",
      "level": "INFO",
      "message": "Server started on port 80"
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Service logs retrieved successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment or service not found

---

### 16. Stream Deployment Logs

#### GET `/deployments/{id}/logs/stream`
Stream logs for a deployment in real-time.

**Request:**
```http
GET /deployments/123/logs/stream?follow=true
Authorization: {"signature":"...","address":"0x...","message":"123:logs:1640995200","timestamp":1640995200}
```

**Response:** (Server-Sent Events)
```
data: {"timestamp":"2024-01-01T12:00:00Z","service":"web","pod":"web-123-abc123","level":"INFO","message":"Server started on port 80"}

data: {"timestamp":"2024-01-01T12:00:01Z","service":"web","pod":"web-123-def456","level":"INFO","message":"Server started on port 80"}

```

**Query Parameters:**
- `follow`: Whether to follow logs in real-time (default: true)

**Status Codes:**
- `200 OK`: Log stream started successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found

---

### 17. Stream Service Logs

#### GET `/deployments/{id}/services/{service}/logs/stream`
Stream logs for a specific service in real-time.

**Request:**
```http
GET /deployments/123/services/web/logs/stream?follow=true
Authorization: {"signature":"...","address":"0x...","message":"123:logs:1640995200","timestamp":1640995200}
```

**Response:** (Server-Sent Events)
```
data: {"timestamp":"2024-01-01T12:00:00Z","service":"web","pod":"web-123-abc123","level":"INFO","message":"Server started on port 80"}

```

**Status Codes:**
- `200 OK`: Service log stream started successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment or service not found

---

### 18. Execute Console Command

#### POST `/deployments/{id}/services/{service}/exec`
Execute a command in a deployment pod.

**Request:**
```http
POST /deployments/123/services/web/exec
Content-Type: application/json
Authorization: {"signature":"...","address":"0x...","message":"123:exec:1640995200","timestamp":1640995200}

{
  "command": ["ls", "-la"],
  "tty": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "exit_code": 0,
    "stdout": "total 8\ndrwxr-xr-x 2 root root 4096 Jan  1 12:00 .\ndrwxr-xr-x 3 root root 4096 Jan  1 12:00 ..\n",
    "stderr": ""
  }
}
```

**Status Codes:**
- `200 OK`: Command executed successfully
- `400 Bad Request`: Invalid command
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment or service not found
- `500 Internal Server Error`: Command execution failed

---

### 19. WebSocket Console

#### GET `/deployments/{id}/services/{service}/exec/ws`
Interactive console via WebSocket.

**Request:**
```http
GET /deployments/123/services/web/exec/ws
Authorization: {"signature":"...","address":"0x...","message":"123:exec:1640995200","timestamp":1640995200}
```

**WebSocket Messages:**

**Client to Server:**
```json
{
  "type": "command",
  "data": {
    "command": ["ls", "-la"]
  }
}
```

**Server to Client:**
```json
{
  "type": "output",
  "data": {
    "stdout": "total 8\ndrwxr-xr-x 2 root root 4096 Jan  1 12:00 .\n",
    "stderr": ""
  }
}
```

```json
{
  "type": "exit",
  "data": {
    "exit_code": 0
  }
}
```

**Status Codes:**
- `101 Switching Protocols`: WebSocket connection established
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment or service not found

---

### 20. Get Deployment Events

#### GET `/deployments/{id}/events`
Get events for a deployment.

**Request:**
```http
GET /deployments/123/events?limit=100
Authorization: {"signature":"...","address":"0x...","message":"123:events:1640995200","timestamp":1640995200}
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "event-1",
      "type": "deployment_created",
      "timestamp": "2024-01-01T12:00:00Z",
      "data": {
        "deployment_id": "123",
        "name": "web-application"
      }
    },
    {
      "id": "event-2",
      "type": "deployment_started",
      "timestamp": "2024-01-01T12:01:00Z",
      "data": {
        "deployment_id": "123"
      }
    }
  ]
}
```

**Query Parameters:**
- `limit`: Maximum number of events to retrieve (default: 100)

**Status Codes:**
- `200 OK`: Events retrieved successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found

---

### 21. Stream Deployment Events

#### GET `/deployments/{id}/events/stream`
Stream deployment events in real-time.

**Request:**
```http
GET /deployments/123/events/stream
Authorization: {"signature":"...","address":"0x...","message":"123:events:1640995200","timestamp":1640995200}
```

**Response:** (Server-Sent Events)
```
data: {"id":"event-1","type":"deployment_created","timestamp":"2024-01-01T12:00:00Z","data":{"deployment_id":"123","name":"web-application"}}

data: {"id":"event-2","type":"deployment_started","timestamp":"2024-01-01T12:01:00Z","data":{"deployment_id":"123"}}

```

**Status Codes:**
- `200 OK`: Event stream started successfully
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found

---

### 22. Update Deployment

#### PUT `/deployments/{id}`
Update a deployment.

**Request:**
```http
PUT /deployments/123
Content-Type: application/json
Authorization: {"signature":"...","address":"0x...","message":"123:update:1640995200","timestamp":1640995200}

{
  "name": "updated-web-application",
  "image": "nginx:1.21",
  "replicas": 5
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "123",
    "name": "updated-web-application",
    "image": "nginx:1.21",
    "replicas": 5,
    "updated_at": "2024-01-01T12:30:00Z"
  }
}
```

**Status Codes:**
- `200 OK`: Deployment updated successfully
- `400 Bad Request`: Invalid request body
- `401 Unauthorized`: Missing or invalid authorization
- `403 Forbidden`: Not the deployment owner
- `404 Not Found`: Deployment not found
- `500 Internal Server Error`: Failed to update deployment

---

## Error Responses

All endpoints return errors in a consistent format:

```json
{
  "success": false,
  "error": "Error message describing what went wrong",
  "code": 400
}
```

### Common Error Codes

- `400 Bad Request`: Invalid request format or parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource already exists or conflict
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

---

## Client Examples

### JavaScript/Node.js with Axios

```javascript
const axios = require('axios');
const crypto = require('crypto');

// Configuration
const API_BASE_URL = 'http://localhost:8080/api/v1';
const PRIVATE_KEY = 'your_private_key_here';

// Helper function to create signature
function createSignature(message, privateKey) {
    const sign = crypto.createSign('SHA256');
    sign.update(message);
    return sign.sign(privateKey, 'base64');
}

// Helper function to create authorization header
function createAuthHeader(address, deploymentId, action, timestamp) {
    const message = `${deploymentId}:${action}:${timestamp}`;
    const signature = createSignature(message, PRIVATE_KEY);
    
    return JSON.stringify({
        signature: signature,
        address: address,
        message: message,
        timestamp: timestamp
    });
}

// Example: Create deployment
async function createDeployment() {
    const deploymentData = {
        id: "123",
        name: "web-application",
        image: "nginx:latest",
        ports: [{ containerPort: 80, hostPort: 8080, protocol: "tcp" }],
        environment: { NODE_ENV: "production" },
        resources: { cpu: "1000m", memory: "1Gi" },
        replicas: 3
    };

    const timestamp = Math.floor(Date.now() / 1000);
    const authHeader = createAuthHeader(
        '0x123456789abcdef123456789abcdef123456789a',
        "123",
        'create',
        timestamp
    );

    try {
        const response = await axios.post(`${API_BASE_URL}/deployments`, deploymentData, {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': authHeader
            }
        });
        console.log('Deployment created:', response.data);
    } catch (error) {
        console.error('Error:', error.response?.data || error.message);
    }
}

// Example: Get deployment logs
async function getDeploymentLogs(deploymentId) {
    const timestamp = Math.floor(Date.now() / 1000);
    const authHeader = createAuthHeader(
        '0x123456789abcdef123456789abcdef123456789a',
        deploymentId,
        'logs',
        timestamp
    );

    try {
        const response = await axios.get(`${API_BASE_URL}/deployments/${deploymentId}/logs?tail=100`, {
            headers: { 'Authorization': authHeader }
        });
        console.log('Logs:', response.data);
    } catch (error) {
        console.error('Error:', error.response?.data || error.message);
    }
}
```

### Python with requests

```python
import requests
import time
import hashlib
import hmac

# Configuration
API_BASE_URL = "http://localhost:8080/api/v1"
PRIVATE_KEY = b"your_private_key_here"

def create_signature(message, private_key):
    """Create SHA256 signature of message using private key"""
    return hmac.new(private_key, message.encode(), hashlib.sha256).hexdigest()

def create_auth_header(address, deployment_id, action, timestamp):
    """Create authorization header"""
    message = f"{deployment_id}:{action}:{timestamp}"
    signature = create_signature(message, PRIVATE_KEY)
    
    return {
        "signature": signature,
        "address": address,
        "message": message,
        "timestamp": timestamp
    }

# Example: Create deployment
def create_deployment():
    deployment_data = {
        "id": "123",
        "name": "web-application",
        "image": "nginx:latest",
        "ports": [{"containerPort": 80, "hostPort": 8080, "protocol": "tcp"}],
        "environment": {"NODE_ENV": "production"},
        "resources": {"cpu": "1000m", "memory": "1Gi"},
        "replicas": 3
    }
    
    timestamp = int(time.time())
    auth_header = create_auth_header(
        "0x123456789abcdef123456789abcdef123456789a",
        "123",
        "create",
        timestamp
    )
    
    try:
        response = requests.post(
            f"{API_BASE_URL}/deployments",
            json=deployment_data,
            headers={
                "Content-Type": "application/json",
                "Authorization": str(auth_header)
            }
        )
        response.raise_for_status()
        print("Deployment created:", response.json())
    except requests.exceptions.RequestException as e:
        print("Error:", e)

# Example: Get deployment logs
def get_deployment_logs(deployment_id):
    timestamp = int(time.time())
    auth_header = create_auth_header(
        "0x123456789abcdef123456789abcdef123456789a",
        deployment_id,
        "logs",
        timestamp
    )
    
    try:
        response = requests.get(
            f"{API_BASE_URL}/deployments/{deployment_id}/logs?tail=100",
            headers={"Authorization": str(auth_header)}
        )
        response.raise_for_status()
        print("Logs:", response.json())
    except requests.exceptions.RequestException as e:
        print("Error:", e)
```

---

## Supported Deployment Types

### 1. Kubernetes Deployment ✅
- **Status**: Fully implemented
- **Features**: Pod management, service discovery, namespace isolation, resource quotas, comprehensive monitoring
- **Location**: `core/deployments/` (integrated into main service)

## Core Components

### Service
The main `Service` struct provides a unified interface for managing Kubernetes deployments:

```go
service := deployments.NewService(cfg, logger)
service.SetManagers(eventManager, storageManager, resourceManager, factory)
service.Start(ctx)
```

### Factory
The `Factory` creates deployment managers for Kubernetes deployments:

```go
factory := deployments.NewFactory(logger)
factory.RegisterManagerCreator(deployments.DeploymentTypeKubernetes, kubernetes.CreateManager)
```

### Interfaces
Common interfaces that the deployment system implements:

- `DeploymentManager`: Core deployment operations
- `ManifestManager`: Manifest processing
- `PortManager`: Port allocation
- `NetworkManager`: Network management
- `ResourceManager`: Resource tracking
- `EventManager`: Event handling
- `StorageManager`: Data persistence

## Features

### 1. Deployment Management
- Create, update, delete Kubernetes deployments
- Start, stop, restart deployments
- Scale deployments up/down
- Update deployment images

### 2. Monitoring & Inspection
- **Deployment Inspection**: Get detailed information about deployments
- **Service Inspection**: Get detailed information about specific services
- **Resource Monitoring**: Track CPU, memory, disk, and network usage
- **Health Checks**: Monitor service health and status
- **Metrics Collection**: Collect performance metrics over time

### 3. Logs Management
- **Log Retrieval**: Get logs for deployments and services
- **Real-time Log Streaming**: Stream logs in real-time with follow mode
- **Log Filtering**: Filter logs by service, time range, log level
- **Log Aggregation**: Aggregate logs across multiple services

### 4. Console Execution
- **Command Execution**: Execute commands in deployment pods
- **Interactive Console**: Interactive terminal access to pods
- **WebSocket Support**: Real-time interactive console via WebSocket
- **Session Management**: Manage multiple execution sessions

### 5. Resource Management
- **Resource Limits**: Set and enforce resource limits via Kubernetes resource quotas
- **Port Management**: Automatic port allocation and management
- **Network Management**: Network configuration and management

## Usage Examples

### Creating a Kubernetes Deployment

```go
import (
    "github.com/unicornultrafoundation/subnet-node/core/deployments"
)

// Create deployment
deployment := &deployments.Deployment{
    ID:       "deploy-123",
    Name:     "web-app",
    Type:     deployments.DeploymentTypeKubernetes,
    Manifest: &kubernetes.Manifest{
        // Kubernetes manifest configuration
    },
}

err := service.CreateDeployment(ctx, deployment)
```

### Getting Deployment Logs

```go
// Get logs for a deployment
logs, err := service.GetDeploymentLogs(ctx, "deploy-123", "", 100)
if err != nil {
    log.Fatal(err)
}
defer logs.Close()

// Copy logs to stdout
io.Copy(os.Stdout, logs)
```

### Streaming Real-time Logs

```go
// Stream logs in real-time
logChan, err := service.StreamDeploymentLogs(ctx, "deploy-123", "web", true)
if err != nil {
    log.Fatal(err)
}

for logEntry := range logChan {
    fmt.Printf("[%s] %s: %s\n", logEntry.Timestamp, logEntry.Service, logEntry.Message)
}
```

### Executing Commands

```go
// Execute a command in a pod
session, err := service.ExecConsole(ctx, "deploy-123", "web", []string{"ls", "-la"}, false)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

result, err := session.Execute(ctx, []string{"ls", "-la"})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Exit code: %d\n", result.ExitCode)
fmt.Printf("Output: %s\n", result.Stdout)
```

### Inspecting Deployments

```go
// Get detailed deployment information
inspection, err := service.InspectDeployment(ctx, "deploy-123")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Deployment: %s\n", inspection.Name)
fmt.Printf("Status: %s\n", inspection.Status)
fmt.Printf("Services: %d\n", len(inspection.Services))

for name, service := range inspection.Services {
    fmt.Printf("  - %s: %s (%s)\n", name, service.Image, service.Status)
}
```

### Getting Metrics

```go
// Get deployment metrics for the last hour
metrics, err := service.GetDeploymentMetrics(ctx, "deploy-123", time.Hour)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("CPU Usage: %.2f%%\n", metrics.Total.CPUUsage)
fmt.Printf("Memory Usage: %d MB\n", metrics.Total.MemoryUsage/1024/1024)
```

## Configuration

The deployment module can be configured in the main configuration file:

```yaml
deployments:
  enabled: true
  
  # Kubernetes configuration
  kubernetes:
    enabled: true
``` 