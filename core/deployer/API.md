# Deployer API Documentation

The deployer package exposes a comprehensive JSON-RPC 2.0 API for deployment operations, along with WebSocket endpoints for real-time functionality.

## JSON-RPC API

All API methods follow the JSON-RPC 2.0 specification and are available via HTTP POST requests.

### Base URL
```
http://localhost:8080
```

### Request Format
```json
{
    "jsonrpc": "2.0",
    "method": "deployments_<methodName>",
    "params": [...],
    "id": 1
}
```

### Response Format
```json
{
    "jsonrpc": "2.0",
    "result": {
        "data": {...},
        "error": null
    },
    "id": 1
}
```

## Core Deployment Methods

### 1. Request Deployment

Creates a new deployment with the specified manifest.

**Method:** `deployments_requestDeployment`

**Parameters:**
- `DeploymentRequest` object

**Example Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "deployments_requestDeployment",
    "params": [
        {
            "order_id": "1",
            "ttl": 50,
            "requester": "0x0000000000000000000000000000000000000001",
            "manifest": {
                "version": "2.0",
                "services": {
                    "web": {
                        "image": "hashicorp/http-echo",
                        "command": ["/http-echo", "-text", "Hello from http-echo!"],
                        "count": 1,
                        "expose": [
                            {
                                "port": 5678,
                                "as": 80,
                                "proto": "TCP",
                                "to": [{"global": true}]
                            }
                        ],
                        "params": {
                            "health": {
                                "readiness": {
                                    "http": {
                                        "path": "/",
                                        "port": 5678
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
                                    "value": 100,
                                    "unit": "m"
                                }
                            },
                            "memory": {
                                "size": {
                                    "value": 128,
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
                                    "request": "100m",
                                    "limit": "200m"
                                },
                                "memory": {
                                    "request": "128Mi",
                                    "limit": "256Mi"
                                }
                            }
                        }
                    }
                },
                "deployment": {
                    "default": {
                        "profile": "default",
                        "count": 1
                    }
                },
                "endpoints": {
                    "echo": {
                        "kind": "http"
                    }
                }
            },
            "signature": "1"
        }
    ],
    "id": 1
}
```

**Example Response:**
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "data": {
            "ID": "1",
            "Requester": "0x0000000000000000000000000000000000000001",
            "Status": {
                "state": "running",
                "services": [
                    {
                        "name": "web",
                        "group": "default",
                        "image": "hashicorp/http-echo",
                        "state": "running",
                        "replicas": 1,
                        "readyReplicas": 1,
                        "availableReplicas": 1,
                        "uris": [
                            "http://localhost:31749"
                        ],
                        "ports": [
                            {
                                "port": 80,
                                "protocol": "TCP",
                                "serviceType": "NodePort",
                                "uri": "http://localhost:31749"
                            }
                        ],
                        "resources": {
                            "cpu": "1m",
                            "memory": "256Mi",
                            "storage": ""
                        },
                        "lastUpdated": "1"
                    }
                ],
                "pods": [
                    {
                        "name": "default-web-93ddaafb-7cfbb6b85b-pmjd6",
                        "serviceName": "web",
                        "group": "default",
                        "image": "hashicorp/http-echo",
                        "state": "running",
                        "phase": "Running",
                        "ready": true,
                        "restartCount": 0,
                        "ip": "10.42.0.29",
                        "hostIP": "192.168.106.2",
                        "resources": {
                            "cpu": "1m",
                            "memory": "256Mi",
                            "storage": ""
                        },
                        "createdAt": "2025-06-26T15:01:36+07:00",
                        "startedAt": "2025-06-26T15:01:36+07:00",
                        "lastRestartTime": "",
                        "containerStatuses": [
                            {
                                "name": "web",
                                "image": "hashicorp/http-echo:latest",
                                "ready": true,
                                "restartCount": 0,
                                "state": "running",
                                "startedAt": "2025-06-26T15:01:37+07:00"
                            }
                        ]
                    }
                ],
                "endpoints": [
                    {
                        "name": "default-echo",
                        "host": "echo.1",
                        "path": "/",
                        "protocol": "http",
                        "uri": "http://localhost:31749/"
                    }
                ],
                "createdAt": "2025-06-26T15:01:47+07:00",
                "updatedAt": "2025-06-26T15:01:47+07:00",
                "deployedAt": "2025-06-26T15:01:36+07:00",
                "ttl": 50,
                "timeLeft": 50
            }
        },
        "error": ""
    }
}
```

### 2. Get Deployment

Retrieves deployment information and status.

**Method:** `deployments_getDeployment`

**Parameters:**
- `orderID` (string): The deployment order ID

**Example Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "deployments_getDeployment",
    "params": ["1"],
    "id": 1
}
```

**Example Response: Same as successful request deployment response**

### 3. Inspect Deployment

Retrieves detailed deployment information equivalent to `kubectl describe` for comprehensive debugging and monitoring.

**Method:** `deployments_inspectDeployment`

**Parameters:**
- `orderID` (string): The deployment order ID

**Example Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "deployments_inspectDeployment",
    "params": ["1"],
    "id": 1
}
```

**Example Response:**
```json
{
    "jsonrpc": "2.0",
    "result": {
        "data": {
            "namespace": {
                "name": "default",
                "status": "Active",
                "age": "2h"
            },
            "deployments": [
                {
                    "name": "default-web",
                    "namespace": "default",
                    "replicas": "1/1",
                    "updatedReplicas": 1,
                    "readyReplicas": 1,
                    "availableReplicas": 1,
                    "unavailableReplicas": 0,
                    "conditions": [
                        {
                            "type": "Available",
                            "status": "True",
                            "lastUpdateTime": "2025-06-26T15:01:36+07:00",
                            "lastTransitionTime": "2025-06-26T15:01:36+07:00",
                            "reason": "MinimumReplicasAvailable",
                            "message": "Deployment has minimum availability."
                        }
                    ],
                    "selector": "app=web",
                    "strategy": "RollingUpdate",
                    "template": {
                        "metadata": {
                            "labels": {
                                "app": "web"
                            }
                        },
                        "spec": {
                            "containers": [
                                {
                                    "name": "web",
                                    "image": "hashicorp/http-echo:latest",
                                    "command": ["/http-echo", "-text", "Hello from http-echo!"],
                                    "ports": [
                                        {
                                            "containerPort": 5678,
                                            "protocol": "TCP"
                                        }
                                    ],
                                    "resources": {
                                        "requests": {
                                            "cpu": "100m",
                                            "memory": "128Mi"
                                        },
                                        "limits": {
                                            "cpu": "200m",
                                            "memory": "256Mi"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            ],
            "services": [
                {
                    "name": "default-web",
                    "namespace": "default",
                    "type": "NodePort",
                    "clusterIP": "10.43.0.1",
                    "externalIPs": null,
                    "ports": [
                        {
                            "port": 80,
                            "targetPort": 5678,
                            "nodePort": 31749,
                            "protocol": "TCP"
                        }
                    ],
                    "selector": "app=web",
                    "sessionAffinity": "None",
                    "externalTrafficPolicy": "Cluster"
                }
            ],
            "pods": [
                {
                    "name": "default-web-93ddaafb-7cfbb6b85b-pmjd6",
                    "namespace": "default",
                    "phase": "Running",
                    "ready": "1/1",
                    "status": "Running",
                    "restarts": 0,
                    "age": "2h",
                    "ip": "10.42.0.29",
                    "node": "k3s-node-1",
                    "containers": [
                        {
                            "name": "web",
                            "image": "hashicorp/http-echo:latest",
                            "ready": true,
                            "restartCount": 0,
                            "state": "running",
                            "startedAt": "2025-06-26T15:01:37+07:00"
                        }
                    ],
                    "conditions": [
                        {
                            "type": "Ready",
                            "status": "True",
                            "lastProbeTime": null,
                            "lastTransitionTime": "2025-06-26T15:01:37+07:00"
                        }
                    ]
                }
            ],
            "events": [
                {
                    "type": "Normal",
                    "reason": "Scheduled",
                    "object": "Pod/default-web-93ddaafb-7cfbb6b85b-pmjd6",
                    "message": "Successfully assigned default/default-web-93ddaafb-7cfbb6b85b-pmjd6 to k3s-node-1",
                    "firstTimestamp": "2025-06-26T15:01:36+07:00",
                    "lastTimestamp": "2025-06-26T15:01:36+07:00",
                    "count": 1
                },
                {
                    "type": "Normal",
                    "reason": "Pulling",
                    "object": "Pod/default-web-93ddaafb-7cfbb6b85b-pmjd6",
                    "message": "Pulling image \"hashicorp/http-echo:latest\"",
                    "firstTimestamp": "2025-06-26T15:01:36+07:00",
                    "lastTimestamp": "2025-06-26T15:01:36+07:00",
                    "count": 1
                },
                {
                    "type": "Normal",
                    "reason": "Pulled",
                    "object": "Pod/default-web-93ddaafb-7cfbb6b85b-pmjd6",
                    "message": "Successfully pulled image \"hashicorp/http-echo:latest\"",
                    "firstTimestamp": "2025-06-26T15:01:37+07:00",
                    "lastTimestamp": "2025-06-26T15:01:37+07:00",
                    "count": 1
                },
                {
                    "type": "Normal",
                    "reason": "Created",
                    "object": "Pod/default-web-93ddaafb-7cfbb6b85b-pmjd6",
                    "message": "Created container web",
                    "firstTimestamp": "2025-06-26T15:01:37+07:00",
                    "lastTimestamp": "2025-06-26T15:01:37+07:00",
                    "count": 1
                },
                {
                    "type": "Normal",
                    "reason": "Started",
                    "object": "Pod/default-web-93ddaafb-7cfbb6b85b-pmjd6",
                    "message": "Started container web",
                    "firstTimestamp": "2025-06-26T15:01:37+07:00",
                    "lastTimestamp": "2025-06-26T15:01:37+07:00",
                    "count": 1
                }
            ],
            "configmaps": [],
            "secrets": [],
            "persistentVolumes": []
        },
        "error": null
    },
    "id": 1
}
```

### 4. Get Deployment Stats

Retrieves resource usage statistics for a deployment.

**Method:** `deployments_getDeploymentStats`

**Parameters:**
- `orderID` (string): The deployment order ID

**Example Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "deployments_getDeploymentStats",
    "params": ["1"],
    "id": 1
}
```

**Example Response:**
```json
{
    "jsonrpc": "2.0",
    "result": {
        "data": {
            "usedCpu": 150, // millicores (mCPU)
            "usedMemory": 134217728, // bytes
            "usedStorage": 1073741824, // bytes
            "usedUploadBytes": 1024, // bytes
            "usedDownloadBytes": 2048, // bytes
            "usedGpu": 1, // GPU count (number of GPUs)
            "duration": 3600 // seconds
        },
        "error": null
    },
    "id": 1
}
```

### 5. Get User Deployments

Retrieves all deployments for a specific requester.

**Method:** `deployments_getDeployments`

**Parameters:**
- `requester` (string): Ethereum address of the requester

**Example Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "deployments_getDeployments",
    "params": ["0x0000000000000000000000000000000000000000"],
    "id": 1
}
```

### 6. Get Deployment Request

Retrieves the original deployment request for a deployment.

**Method:** `deployments_getDeploymentRequest`

**Parameters:**
- `orderID` (string): The deployment order ID

**Example Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "deployments_getDeploymentRequest",
    "params": ["5"],
    "id": 1
}
```
**Response: Same as body of deployment request**

### 7. Cleanup Deployment

Deletes a deployment and cleans up all associated resources.

**Method:** `deployments_cleanupDeployment`

**Parameters:**
- `orderID` (string): The deployment order ID

**Example Request:**
```json
{
    "jsonrpc": "2.0",
    "method": "deployments_cleanupDeployment",
    "params": ["5"],
    "id": 1
}
```

## WebSocket API

The deployer provides WebSocket endpoints for real-time operations like log streaming and command execution.

### 1. Log Streaming

Stream real-time logs from deployment containers.

**Endpoint:** `ws://localhost:8080/ws/deployments/{orderID}/logs`

**Example:**
```
ws://localhost:8080/ws/deployments/5/logs
```

**Message Format:**
```json
{
    "name": "service-name",
    "message": "log message content"
}
```

### 2. Command Execution

Execute commands in deployment containers with interactive terminal support.

**Endpoint:** `ws://localhost:8080/ws/deployments/{orderID}/exec`

**Query Parameters:**
- `pod`: Pod name
- `service`: Service name
- `tty`: Enable TTY (1 for true, 0 for false)
- `stdin`: Enable stdin (1 for true, 0 for false)
- `cmd0`, `cmd1`, ...: Command arguments

**Example:**
```
ws://localhost:8080/ws/deployments/2/exec?pod=default-proxy-e8b2602e-6b5f45fbb8-zng7p&service=proxy&tty=1&stdin=1&cmd0=sh
```

**Message Types:**
- `100`: Stdout data
- `101`: Stderr data
- `102`: Command result
- `103`: Command failure
- `104`: Stdin data
- `105`: Terminal resize

**Message Type Details:**

The WebSocket command execution uses binary messages with a message type prefix. Each message consists of a 1-byte message type ID followed by the message payload.

##### Message Type 100: Stdout Data
**Direction:** Server → Client  
**Format:** Binary message with type 100 prefix + raw stdout data

**Example:**
```javascript
// Client receives stdout data
// Binary message: [100, ...stdout_bytes]
// Example: [100, 72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 10]
// This represents: "Hello World\n"
```

##### Message Type 101: Stderr Data
**Direction:** Server → Client  
**Format:** Binary message with type 101 prefix + raw stderr data

**Example:**
```javascript
// Client receives stderr data
// Binary message: [101, ...stderr_bytes]
// Example: [101, 69, 114, 114, 111, 114, 58, 32, 102, 105, 108, 101, 32, 110, 111, 116, 32, 102, 111, 117, 110, 100, 10]
// This represents: "Error: file not found\n"
```

##### Message Type 102: Command Result
**Direction:** Server → Client  
**Format:** Binary message with type 102 prefix + JSON-encoded result

**Example:**
```javascript
// Client receives command completion result
// Binary message: [102, ...json_bytes]
// JSON payload: {"exit_code": 0}
// Full binary: [102, 123, 34, 101, 120, 105, 116, 95, 99, 111, 100, 101, 34, 58, 32, 48, 125]
```

**JSON Structure:**
```json
{
    "exit_code": 0
}
```

##### Message Type 103: Command Failure
**Direction:** Server → Client  
**Format:** Binary message with type 103 prefix + empty payload

**Example:**
```javascript
// Client receives command failure
// Binary message: [103]
// Empty payload indicates failure occurred
```

##### Message Type 104: Stdin Data
**Direction:** Client → Server  
**Format:** Binary message with type 104 prefix + raw stdin data

**Example:**
```javascript
// Client sends stdin data
// Binary message: [104, ...stdin_bytes]
// Example: [104, 108, 115, 10]  // "ls\n"
// Example: [104, 101, 120, 105, 116, 10]  // "exit\n"
```

##### Message Type 105: Terminal Resize
**Direction:** Client → Server  
**Format:** Binary message with type 105 prefix + 8-byte binary data (width + height)

**Example:**
```javascript
// Client sends terminal resize
// Binary message: [105, 0, 0, 0, 80, 0, 0, 0, 24]
// This represents: width=80, height=24 (big-endian 32-bit integers)
```

**Binary Structure:**
```
[105] + [width (4 bytes, big-endian)] + [height (4 bytes, big-endian)]
```

**Complete Example Session:**

```javascript
// 1. Connection Setup
const ws = new WebSocket('ws://localhost:8080/ws/deployments/2/exec?pod=default-proxy-e8b2602e-6b5f45fbb8-zng7p&service=proxy&tty=1&stdin=1&cmd0=sh');

// 2. Client Sends Stdin (Type 104)
const stdinData = new Uint8Array([104, 108, 115, 10]); // [104] + "ls\n"
ws.send(stdinData);

// 3. Server Responds with Stdout (Type 100)
// Binary: [100, 100, 111, 99, 107, 101, 114, 45, 99, 111, 109, 112, 111, 115, 101, 46, 121, 97, 109, 108, 10, 112, 97, 99, 107, 97, 103, 101, 46, 106, 115, 111, 110, 10]
// This represents: "docker-compose.yaml\npackage.json\n"

// 4. Client Sends Terminal Resize (Type 105)
const width = 120;
const height = 30;
const resizeData = new Uint8Array([
    105, // message type
    (width >> 24) & 0xFF, (width >> 16) & 0xFF, (width >> 8) & 0xFF, width & 0xFF, // width in big-endian
    (height >> 24) & 0xFF, (height >> 16) & 0xFF, (height >> 8) & 0xFF, height & 0xFF // height in big-endian
]);
ws.send(resizeData);

// 5. Client Sends Exit Command (Type 104)
const exitData = new Uint8Array([104, 101, 120, 105, 116, 10]); // [104] + "exit\n"
ws.send(exitData);

// 6. Server Sends Command Result (Type 102)
// Binary: [102, 123, 34, 101, 120, 105, 116, 95, 99, 111, 100, 101, 34, 58, 32, 48, 125]
// JSON: {"exit_code": 0}
```

**Key Points:**
- **Binary Protocol**: Command execution uses binary messages with type prefixes
- **JSON for Results**: Only command results (type 102) use JSON encoding
- **Raw Data**: Stdout, stderr, and stdin use raw binary data
- **Big-Endian**: Terminal resize uses big-endian 32-bit integers
- **Empty Failure**: Command failures (type 103) have empty payloads
- **Thread Safety**: The implementation uses mutex locks for thread-safe message writing

## GPU Deployment Examples

### NVIDIA GPU Deployment

```json
{
    "jsonrpc": "2.0",
    "method": "deployments_requestDeployment",
    "params": [
        {
            "order_id": "4",
            "ttl": 50000,
            "requester": "0x0000000000000000000000000000000000000004",
            "manifest": {
                "version": "2.0",
                "services": {
                    "gpu-test": {
                        "image": "python:3.11-slim",
                        "command": ["python3", "-c", "import time; print('GPU Test'); [print(f'Heartbeat {i}') or time.sleep(30) for i in range(1000)]"],
                        "count": 1,
                        "resources": {
                            "cpu": {
                                "units": {
                                    "value": 1000,
                                    "unit": "m"
                                }
                            },
                            "memory": {
                                "size": {
                                    "value": 1024,
                                    "unit": "Mi"
                                }
                            },
                            "gpu": {
                                "units": 1,
                                "attributes": {
                                    "vendor": {
                                        "nvidia": [
                                            {
                                                "model": "Virtual GPU",
                                                "ram": "8Gi",
                                                "interface": "Virtual"
                                            }
                                        ]
                                    }
                                }
                            }
                        }
                    }
                },
                "profiles": {
                    "compute": {
                        "gpu": {
                            "resources": {
                                "cpu": {
                                    "request": "1000m",
                                    "limit": "2000m"
                                },
                                "memory": {
                                    "request": "1Gi",
                                    "limit": "2Gi"
                                },
                                "gpu": {
                                    "units": 1,
                                    "attributes": {
                                        "vendor": {
                                            "nvidia": [
                                                {
                                                    "model": "Virtual GPU",
                                                    "ram": "8Gi",
                                                    "interface": "Virtual"
                                                }
                                            ]
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                "deployment": {
                    "gpu-test": {
                        "profile": "gpu",
                        "count": 1
                    }
                }
            },
            "signature": "1"
        }
    ],
    "id": 1
}
```

### Apple Silicon GPU Deployment

```json
{
    "jsonrpc": "2.0",
    "method": "deployments_requestDeployment",
    "params": [
        {
            "order_id": "5",
            "ttl": 50000,
            "requester": "0x0000000000000000000000000000000000000005",
            "manifest": {
                "version": "2.0",
                "services": {
                    "gpu-test-apple": {
                        "image": "python:3.11-slim",
                        "command": ["python3", "-c", "import time; import os; print('Apple Silicon GPU Test'); print('Container:', os.environ.get('HOSTNAME', 'Unknown')); print('Architecture:', os.uname().machine); [print(f'Heartbeat {i}') or time.sleep(30) for i in range(1000)]"],
                        "count": 1,
                        "resources": {
                            "cpu": {
                                "units": {
                                    "value": 1000,
                                    "unit": "m"
                                }
                            },
                            "memory": {
                                "size": {
                                    "value": 1024,
                                    "unit": "Mi"
                                }
                            },
                            "gpu": {
                                "units": 1,
                                "attributes": {
                                    "vendor": {
                                        "apple": [
                                            {
                                                "model": "Apple Silicon GPU",
                                                "ram": "2Gi",
                                                "interface": "Metal"
                                            }
                                        ]
                                    }
                                }
                            }
                        }
                    }
                },
                "profiles": {
                    "compute": {
                        "apple-gpu": {
                            "resources": {
                                "cpu": {
                                    "request": "1000m",
                                    "limit": "2000m"
                                },
                                "memory": {
                                    "request": "1Gi",
                                    "limit": "2Gi"
                                },
                                "gpu": {
                                    "units": 1,
                                    "attributes": {
                                        "vendor": {
                                            "apple": [
                                                {
                                                    "model": "Apple Silicon GPU",
                                                    "ram": "2Gi",
                                                    "interface": "Metal"
                                                }
                                            ]
                                        }
                                    }
                                }
                            }
                        }
                    }
                },
                "deployment": {
                    "gpu-test-apple": {
                        "profile": "apple-gpu",
                        "count": 1
                    }
                }
            },
            "signature": "1"
        }
    ],
    "id": 1
}
``` 