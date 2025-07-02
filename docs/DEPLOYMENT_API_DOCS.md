# Deployment API Documentation

## Overview

The Deployment API provides REST endpoints and WebSocket connections for managing containerized application deployments on the subnet network. The API uses authentication via authchain headers and supports real-time operations like log streaming and command execution.

## Base URL

```
http://localhost:8080/api/v1/deployments
```

## Authentication

All authenticated endpoints require the `X-AuthChain` header containing a valid authchain for the specific deployment entity.

**Entity ID Format:** `subnet_deployment:{providerId}:{machineId}:{order_id}`

**Example:**
```javascript
const headers = {
  'Content-Type': 'application/json',
  'X-AuthChain': 'your-authchain-string-here'
};
```

## Endpoints

### 1. Health Check

**Endpoint:** `GET /health`

**Description:** Check the health status of the deployment API service.

**Authentication:** None required

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "deployment-api",
  "cache": {
    "total_entries": 5,
    "expired_entries": 1,
    "valid_entries": 4
  }
}
```

**JavaScript Example:**
```javascript
async function checkHealth() {
  try {
    const response = await fetch('http://localhost:8080/api/v1/deployments/health');
    const data = await response.json();
    console.log('API Health:', data);
    return data;
  } catch (error) {
    console.error('Health check failed:', error);
  }
}
```

### 2. Create Deployment

**Endpoint:** `POST /`

**Description:** Create a new deployment with the specified SDL manifest.

**Authentication:** Required

**Request Body:**
```json
{
  "order_id": "12345",
  "manifest": {
    "version": "2.0",
    "services": {
      "web": {
        "image": "nginx:latest",
        "count": 2,
        "expose": [
          {
            "port": 80,
            "as": 80,
            "proto": "TCP",
            "to": [
              {
                "global": true
              }
            ]
          }
        ],
        "params": {
          "health": {
            "readiness": {
              "http": {
                "path": "/",
                "port": 80
              },
              "initial_delay_seconds": 5,
              "period_seconds": 10,
              "timeout_seconds": 5,
              "success_threshold": 1,
              "failure_threshold": 3
            }
          }
        },
        "resources": {
          "cpu": {
            "units": {
              "value": 500,
              "unit": "m"
            }
          },
          "memory": {
            "size": {
              "value": 512,
              "unit": "Mi"
            }
          }
        }
      }
    },
    "profiles": {
      "compute": {
        "default": {
          "resources": {
            "cpu": {
              "request": "500m",
              "limit": "1000m"
            },
            "memory": {
              "request": "512Mi",
              "limit": "1Gi"
            },
            "storage": [
              {
                "name": "data",
                "size": "10Gi",
                "attributes": {
                  "persistent": true,
                  "class": "standard"
                }
              }
            ]
          }
        }
      }
    },
    "deployment": {
      "default": {
        "profile": "default",
        "count": 2
      }
    },
    "endpoints": {
      "web": {
        "kind": "http"
      }
    }
  }
}
```

**Response:**
```json
{
  "id": "deployment-12345",
  "requester": "0x1234567890abcdef...",
  "status": {
    "state": "pending",
    "services": [],
    "pods": [],
    "endpoints": [],
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:30:00Z",
    "deployedAt": null,
    "ttl": 120,
    "timeLeft": 120
  }
}
```

**JavaScript Example:**
```javascript
async function createDeployment(orderId, manifest, authChain) {
  const requestBody = {
    order_id: orderId,
    manifest: manifest
  };

  const headers = {
    'Content-Type': 'application/json',
    'X-AuthChain': authChain
  };

  try {
    const response = await fetch('http://localhost:8080/api/v1/deployments/', {
      method: 'POST',
      headers: headers,
      body: JSON.stringify(requestBody)
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Deployment creation failed: ${error.error}`);
    }

    const deployment = await response.json();
    console.log('Deployment created:', deployment);
    return deployment;
  } catch (error) {
    console.error('Failed to create deployment:', error);
    throw error;
  }
}

// Example usage
const manifest = {
  version: "2.0",
  services: {
    web: {
      image: "nginx:latest",
      count: 1,
      expose: [{
        port: 80,
        as: 80,
        proto: "TCP",
        to: [{ global: true }]
      }],
      resources: {
        cpu: { units: { value: 100, unit: "m" } },
        memory: { size: { value: 128, unit: "Mi" } }
      }
    }
  },
  profiles: {
    compute: {
      default: {
        resources: {
          cpu: { request: "100m", limit: "200m" },
          memory: { request: "128Mi", limit: "256Mi" }
        }
      }
    }
  },
  deployment: {
    default: { profile: "default", count: 1 }
  },
  endpoints: {
    web: { kind: "http" }
  }
};

createDeployment("12345", manifest, "your-authchain-here");
```

### 3. Get Deployment

**Endpoint:** `GET /{orderID}`

**Description:** Retrieve deployment information by order ID.

**Authentication:** Required

**Response:**
```json
{
  "id": "deployment-12345",
  "requester": "0x1234567890abcdef...",
  "status": {
    "state": "running",
    "services": [
      {
        "name": "web",
        "readyReplicas": 2,
        "totalReplicas": 2,
        "availableReplicas": 2,
        "updatedReplicas": 2,
        "conditions": [
          {
            "type": "Available",
            "status": "True",
            "lastUpdateTime": "2024-01-15T10:35:00Z"
          }
        ]
      }
    ],
    "pods": [
      {
        "name": "web-deployment-abc123",
        "phase": "Running",
        "ready": true,
        "restartCount": 0,
        "age": "5m"
      }
    ],
    "endpoints": [
      {
        "name": "web",
        "kind": "http",
        "url": "http://localhost:30080"
      }
    ],
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:35:00Z",
    "deployedAt": "2024-01-15T10:32:00Z",
    "ttl": 120,
    "timeLeft": 115
  }
}
```

**JavaScript Example:**
```javascript
async function getDeployment(orderId, authChain) {
  const headers = {
    'X-AuthChain': authChain
  };

  try {
    const response = await fetch(`http://localhost:8080/api/v1/deployments/${orderId}`, {
      method: 'GET',
      headers: headers
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to get deployment: ${error.error}`);
    }

    const deployment = await response.json();
    console.log('Deployment status:', deployment);
    return deployment;
  } catch (error) {
    console.error('Failed to get deployment:', error);
    throw error;
  }
}
```

### 4. Cleanup Deployment

**Endpoint:** `DELETE /{orderID}`

**Description:** Clean up and delete a deployment.

**Authentication:** Required

**Response:**
```json
{
  "message": "Deployment cleaned up successfully"
}
```

**JavaScript Example:**
```javascript
async function cleanupDeployment(orderId, authChain) {
  const headers = {
    'X-AuthChain': authChain
  };

  try {
    const response = await fetch(`http://localhost:8080/api/v1/deployments/${orderId}`, {
      method: 'DELETE',
      headers: headers
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to cleanup deployment: ${error.error}`);
    }

    const result = await response.json();
    console.log('Deployment cleaned up:', result);
    return result;
  } catch (error) {
    console.error('Failed to cleanup deployment:', error);
    throw error;
  }
}
```

### 5. Get Service Status

**Endpoint:** `GET /{orderID}/services/{serviceName}/status`

**Description:** Get detailed status of a specific service within a deployment.

**Authentication:** Required

**Response:**
```json
{
  "name": "web",
  "readyReplicas": 2,
  "totalReplicas": 2,
  "availableReplicas": 2,
  "updatedReplicas": 2,
  "conditions": [
    {
      "type": "Available",
      "status": "True",
      "lastUpdateTime": "2024-01-15T10:35:00Z",
      "reason": "MinimumReplicasAvailable",
      "message": "Deployment has minimum availability."
    }
  ]
}
```

**JavaScript Example:**
```javascript
async function getServiceStatus(orderId, serviceName, authChain) {
  const headers = {
    'X-AuthChain': authChain
  };

  try {
    const response = await fetch(
      `http://localhost:8080/api/v1/deployments/${orderId}/services/${serviceName}/status`,
      {
        method: 'GET',
        headers: headers
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to get service status: ${error.error}`);
    }

    const status = await response.json();
    console.log('Service status:', status);
    return status;
  } catch (error) {
    console.error('Failed to get service status:', error);
    throw error;
  }
}
```

### 6. Get Deployment Logs

**Endpoint:** `GET /{orderID}/logs`

**Description:** Get deployment logs (non-streaming).

**Authentication:** Required

**Response:**
```json
[
  {
    "name": "web-deployment-abc123",
    "message": "2024-01-15T10:30:00Z [INFO] Starting nginx..."
  },
  {
    "name": "web-deployment-def456",
    "message": "2024-01-15T10:30:01Z [INFO] nginx started successfully"
  }
]
```

**JavaScript Example:**
```javascript
async function getDeploymentLogs(orderId, authChain) {
  const headers = {
    'X-AuthChain': authChain
  };

  try {
    const response = await fetch(`http://localhost:8080/api/v1/deployments/${orderId}/logs`, {
      method: 'GET',
      headers: headers
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Failed to get logs: ${error.error}`);
    }

    const logs = await response.json();
    console.log('Deployment logs:', logs);
    return logs;
  } catch (error) {
    console.error('Failed to get deployment logs:', error);
    throw error;
  }
}
```

## WebSocket Endpoints

### 1. Real-time Log Streaming

**Endpoint:** `GET /{orderID}/ws/logs`

**Description:** Stream real-time logs from deployment containers via WebSocket.

**Authentication:** Required

**Message Format:**
```json
{
  "name": "service-name",
  "message": "log message content"
}
```

**JavaScript Example:**
```javascript
class DeploymentLogStreamer {
  constructor(orderId, authChain) {
    this.orderId = orderId;
    this.authChain = authChain;
    this.ws = null;
    this.isConnected = false;
  }

  connect() {
    const wsUrl = `ws://localhost:8080/deployments/${this.orderId}/ws/logs`;
    
    this.ws = new WebSocket(wsUrl);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected for log streaming');
      this.isConnected = true;
    };

    this.ws.onmessage = (event) => {
      try {
        const logMessage = JSON.parse(event.data);
        console.log(`[${logMessage.name}] ${logMessage.message}`);
        
        // Handle log message
        this.onLogMessage(logMessage);
      } catch (error) {
        console.error('Failed to parse log message:', error);
      }
    };

    this.ws.onclose = () => {
      console.log('WebSocket connection closed');
      this.isConnected = false;
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.isConnected = false;
    };
  }

  onLogMessage(logMessage) {
    // Override this method to handle log messages
    console.log(`Log from ${logMessage.name}: ${logMessage.message}`);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.isConnected = false;
    }
  }
}

// Usage
const logStreamer = new DeploymentLogStreamer('12345', 'your-authchain-here');
logStreamer.connect();

// Custom log handler
logStreamer.onLogMessage = (logMessage) => {
  // Add to UI, store in database, etc.
  document.getElementById('logs').innerHTML += 
    `<div>[${logMessage.name}] ${logMessage.message}</div>`;
};
```

### 2. Command Execution

**Endpoint:** `GET /{orderID}/ws/exec`

**Description:** Execute commands in deployment containers with interactive terminal support.

**Query Parameters:**
- `pod`: Pod name
- `service`: Service name
- `tty`: Enable TTY (1 for true, 0 for false)
- `stdin`: Enable stdin (1 for true, 0 for false)
- `cmd0`, `cmd1`, ...: Command arguments

**Message Types:**
- `0`: Stdin data (client → server)
- `1`: Stdout data (server → client)
- `2`: Stderr data (server → client)
- `3`: Command result (server → client)
- `4`: Command failure (server → client)
- `5`: Terminal resize (client → server)

**JavaScript Example:**
```javascript
class DeploymentExecutor {
  constructor(orderId, podName, serviceName, authChain) {
    this.orderId = orderId;
    this.podName = podName;
    this.serviceName = serviceName;
    this.authChain = authChain;
    this.ws = null;
    this.isConnected = false;
  }

  executeCommand(command, options = {}) {
    const {
      tty = true,
      stdin = true,
      onStdout = null,
      onStderr = null,
      onResult = null,
      onError = null
    } = options;

    // Build query parameters
    const params = new URLSearchParams({
      pod: this.podName,
      service: this.serviceName,
      tty: tty ? '1' : '0',
      stdin: stdin ? '1' : '0'
    });

    // Add command arguments
    const cmdArgs = Array.isArray(command) ? command : [command];
    cmdArgs.forEach((arg, index) => {
      params.append(`cmd${index}`, arg);
    });

    const wsUrl = `ws://localhost:8080/deployments/${this.orderId}/ws/exec?${params}`;
    
    this.ws = new WebSocket(wsUrl);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected for command execution');
      this.isConnected = true;
    };

    this.ws.onmessage = (event) => {
      if (event.data instanceof ArrayBuffer) {
        const data = new Uint8Array(event.data);
        const messageType = data[0];
        const messageData = data.slice(1);
        
        switch (messageType) {
          case 1: // Stdout
            const stdout = new TextDecoder().decode(messageData);
            console.log('STDOUT:', stdout);
            if (onStdout) onStdout(stdout);
            break;
            
          case 2: // Stderr
            const stderr = new TextDecoder().decode(messageData);
            console.error('STDERR:', stderr);
            if (onStderr) onStderr(stderr);
            break;
            
          case 3: // Result
            try {
              const result = JSON.parse(new TextDecoder().decode(messageData));
              console.log('Command result:', result);
              if (onResult) onResult(result);
            } catch (error) {
              console.error('Failed to parse result:', error);
            }
            break;
            
          case 4: // Failure
            const error = new TextDecoder().decode(messageData);
            console.error('Command failed:', error);
            if (onError) onError(error);
            break;
        }
      }
    };

    this.ws.onclose = () => {
      console.log('WebSocket connection closed');
      this.isConnected = false;
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.isConnected = false;
    };
  }

  sendStdin(data) {
    if (this.ws && this.isConnected) {
      const message = new Uint8Array(data.length + 1);
      message[0] = 0; // Stdin message type
      message.set(new TextEncoder().encode(data), 1);
      this.ws.send(message.buffer);
    }
  }

  resizeTerminal(width, height) {
    if (this.ws && this.isConnected) {
      const buffer = new ArrayBuffer(9);
      const view = new DataView(buffer);
      view.setUint8(0, 5); // Resize message type
      view.setUint32(1, width, false); // Big endian
      view.setUint32(5, height, false); // Big endian
      this.ws.send(buffer);
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.isConnected = false;
    }
  }
}

// Usage example
const executor = new DeploymentExecutor(
  '12345',
  'web-deployment-abc123',
  'web',
  'your-authchain-here'
);

executor.executeCommand(['ls', '-la'], {
  onStdout: (data) => {
    console.log('Command output:', data);
  },
  onStderr: (data) => {
    console.error('Command error:', data);
  },
  onResult: (result) => {
    console.log('Command completed with exit code:', result.exit_code);
  },
  onError: (error) => {
    console.error('Command failed:', error);
  }
});

// Send input to the command
executor.sendStdin('echo "Hello from terminal"\n');

// Resize terminal
executor.resizeTerminal(80, 24);
```

## Error Responses

All endpoints return standardized error responses:

```json
{
  "error": "Error description",
  "status": 400,
  "message": "Bad Request"
}
```

**Common HTTP Status Codes:**
- `200`: Success
- `201`: Created (deployment created)
- `400`: Bad Request (invalid parameters)
- `401`: Unauthorized (missing or invalid authchain)
- `403`: Forbidden (insufficient permissions)
- `404`: Not Found (deployment not found)
- `500`: Internal Server Error

## Complete JavaScript Client Example

```javascript
class DeploymentAPI {
  constructor(baseUrl = 'http://localhost:8080/api/v1/deployments', authChain) {
    this.baseUrl = baseUrl;
    this.authChain = authChain;
  }

  getHeaders() {
    return {
      'Content-Type': 'application/json',
      'X-AuthChain': this.authChain
    };
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    const config = {
      headers: this.getHeaders(),
      ...options
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || `HTTP ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error(`API request failed: ${error.message}`);
      throw error;
    }
  }

  // Health check
  async health() {
    return this.request('/health', { headers: {} }); // No auth required
  }

  // Create deployment
  async createDeployment(orderId, manifest) {
    return this.request('/', {
      method: 'POST',
      body: JSON.stringify({
        order_id: orderId,
        manifest: manifest
      })
    });
  }

  // Get deployment
  async getDeployment(orderId) {
    return this.request(`/${orderId}`);
  }

  // Cleanup deployment
  async cleanupDeployment(orderId) {
    return this.request(`/${orderId}`, { method: 'DELETE' });
  }

  // Get service status
  async getServiceStatus(orderId, serviceName) {
    return this.request(`/${orderId}/services/${serviceName}/status`);
  }

  // Get deployment logs
  async getDeploymentLogs(orderId) {
    return this.request(`/${orderId}/logs`);
  }

  // Create log streamer
  createLogStreamer(orderId) {
    return new DeploymentLogStreamer(orderId, this.authChain);
  }

  // Create executor
  createExecutor(orderId, podName, serviceName) {
    return new DeploymentExecutor(orderId, podName, serviceName, this.authChain);
  }
}

// Usage example
const api = new DeploymentAPI('http://localhost:8080/api/v1/deployments', 'your-authchain-here');

// Check health
api.health().then(health => console.log('API Health:', health));

// Create a simple deployment
const simpleManifest = {
  version: "2.0",
  services: {
    echo: {
      image: "hashicorp/http-echo",
      command: ["/http-echo", "-text", "Hello from deployment!"],
      count: 1,
      expose: [{
        port: 5678,
        as: 80,
        proto: "TCP",
        to: [{ global: true }]
      }],
      resources: {
        cpu: { units: { value: 100, unit: "m" } },
        memory: { size: { value: 128, unit: "Mi" } }
      }
    }
  },
  profiles: {
    compute: {
      default: {
        resources: {
          cpu: { request: "100m", limit: "200m" },
          memory: { request: "128Mi", limit: "256Mi" }
        }
      }
    }
  },
  deployment: {
    default: { profile: "default", count: 1 }
  },
  endpoints: {
    echo: { kind: "http" }
  }
};

// Create deployment
api.createDeployment('12345', simpleManifest)
  .then(deployment => {
    console.log('Deployment created:', deployment);
    
    // Monitor deployment status
    const checkStatus = () => {
      api.getDeployment('12345').then(status => {
        console.log('Deployment status:', status.status.state);
        if (status.status.state === 'running') {
          console.log('Deployment is running!');
          
          // Start log streaming
          const logStreamer = api.createLogStreamer('12345');
          logStreamer.onLogMessage = (log) => {
            console.log(`[${log.name}] ${log.message}`);
          };
          logStreamer.connect();
        } else if (status.status.state === 'failed') {
          console.error('Deployment failed');
        } else {
          // Check again in 5 seconds
          setTimeout(checkStatus, 5000);
        }
      });
    };
    
    checkStatus();
  })
  .catch(error => {
    console.error('Failed to create deployment:', error);
  });
```

## SDL Manifest Examples

### Minimal Deployment
```yaml
version: "2.0"
services:
  minimal:
    image: "busybox:latest"
    count: 1
profiles:
  compute:
    default:
      resources:
        cpu:
          request: "100m"
          limit: "200m"
        memory:
          request: "128Mi"
          limit: "256Mi"
deployment:
  default:
    profile: "default"
    count: 1
```

### Web Application
```yaml
version: "2.0"
services:
  web:
    image: "nginx:latest"
    count: 2
    expose:
      - port: 80
        as: 80
        proto: "TCP"
        to:
          - global: true
    params:
      health:
        readiness:
          http:
            path: "/"
            port: 80
          initial_delay_seconds: 5
          period_seconds: 10
          timeout_seconds: 5
          success_threshold: 1
          failure_threshold: 3
    resources:
      cpu:
        units:
          value: 500
          unit: "m"
      memory:
        size:
          value: 512
          unit: "Mi"
profiles:
  compute:
    default:
      resources:
        cpu:
          request: "500m"
          limit: "1000m"
        memory:
          request: "512Mi"
          limit: "1Gi"
        storage:
          - name: "data"
            size: "10Gi"
            attributes:
              persistent: true
              class: "standard"
deployment:
  default:
    profile: "default"
    count: 2
endpoints:
  web:
    kind: "http"
```

### Multi-Service Application
```yaml
version: "2.0"
services:
  web:
    image: "nginx:latest"
    count: 2
    depends-on:
      - "api"
    expose:
      - port: 80
        as: 80
        proto: "TCP"
        to:
          - global: true
    resources:
      cpu:
        units:
          value: 200
          unit: "m"
      memory:
        size:
          value: 256
          unit: "Mi"
  
  api:
    image: "myapp/api:latest"
    count: 3
    depends-on:
      - "db"
    expose:
      - port: 8080
        as: 80
        proto: "TCP"
        to:
          - global: true
    env:
      - "DATABASE_URL=postgresql://user:pass@db:5432/mydb"
    resources:
      cpu:
        units:
          value: 500
          unit: "m"
      memory:
        size:
          value: 512
          unit: "Mi"
  
  db:
    image: "postgres:13"
    count: 1
    env:
      - "POSTGRES_DB=mydb"
      - "POSTGRES_USER=user"
      - "POSTGRES_PASSWORD=pass"
    resources:
      cpu:
        units:
          value: 300
          unit: "m"
      memory:
        size:
          value: 1
          unit: "Gi"
profiles:
  compute:
    default:
      resources:
        cpu:
          request: "100m"
          limit: "1000m"
        memory:
          request: "128Mi"
          limit: "2Gi"
        storage:
          - name: "data"
            size: "20Gi"
            attributes:
              persistent: true
              class: "standard"
deployment:
  default:
    profile: "default"
    count: 1
endpoints:
  web:
    kind: "http"
  api:
    kind: "http"
```

This documentation provides a comprehensive guide to using the Deployment API with practical JavaScript examples for all endpoints and WebSocket connections. 