/**
 * Deployment API JavaScript Demo
 * 
 * This file contains practical examples of how to use the Deployment API
 * for managing containerized applications on the subnet network.
 */

// Configuration
const API_BASE_URL = 'http://localhost:8080/api/v1/deployments';
const AUTH_CHAIN = 'your-authchain-string-here'; // Replace with actual authchain
const ORDER_ID = '12345'; // Replace with actual order ID

/**
 * 1. Basic API Client Class
 */
class DeploymentAPIClient {
  constructor(baseUrl = API_BASE_URL, authChain = AUTH_CHAIN) {
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
}

/**
 * 2. WebSocket Log Streamer Class
 */
class DeploymentLogStreamer {
  constructor(orderId, authChain) {
    this.orderId = orderId;
    this.authChain = authChain;
    this.ws = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
  }

  connect() {
    const wsUrl = `ws://localhost:8080/deployments/${this.orderId}/ws/logs`;
    
    this.ws = new WebSocket(wsUrl);
    
    this.ws.onopen = () => {
      console.log('✅ WebSocket connected for log streaming');
      this.isConnected = true;
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      try {
        const logMessage = JSON.parse(event.data);
        console.log(`📝 [${logMessage.name}] ${logMessage.message}`);
        
        // Handle log message
        this.onLogMessage(logMessage);
      } catch (error) {
        console.error('❌ Failed to parse log message:', error);
      }
    };

    this.ws.onclose = () => {
      console.log('🔌 WebSocket connection closed');
      this.isConnected = false;
      
      // Attempt to reconnect
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        console.log(`🔄 Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
        setTimeout(() => this.connect(), 5000);
      }
    };

    this.ws.onerror = (error) => {
      console.error('❌ WebSocket error:', error);
      this.isConnected = false;
    };
  }

  onLogMessage(logMessage) {
    // Override this method to handle log messages
    console.log(`📋 Log from ${logMessage.name}: ${logMessage.message}`);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
      this.isConnected = false;
    }
  }
}

/**
 * 3. WebSocket Command Executor Class
 */
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
      console.log('✅ WebSocket connected for command execution');
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
            console.log('📤 STDOUT:', stdout);
            if (onStdout) onStdout(stdout);
            break;
            
          case 2: // Stderr
            const stderr = new TextDecoder().decode(messageData);
            console.error('📤 STDERR:', stderr);
            if (onStderr) onStderr(stderr);
            break;
            
          case 3: // Result
            try {
              const result = JSON.parse(new TextDecoder().decode(messageData));
              console.log('✅ Command result:', result);
              if (onResult) onResult(result);
            } catch (error) {
              console.error('❌ Failed to parse result:', error);
            }
            break;
            
          case 4: // Failure
            const error = new TextDecoder().decode(messageData);
            console.error('❌ Command failed:', error);
            if (onError) onError(error);
            break;
        }
      }
    };

    this.ws.onclose = () => {
      console.log('🔌 WebSocket connection closed');
      this.isConnected = false;
    };

    this.ws.onerror = (error) => {
      console.error('❌ WebSocket error:', error);
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

/**
 * 4. SDL Manifest Templates
 */
const SDLManifests = {
  // Minimal deployment
  minimal: {
    version: "2.0",
    services: {
      minimal: {
        image: "busybox:latest",
        count: 1
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
    }
  },

  // Simple web server
  webServer: {
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
        params: {
          health: {
            readiness: {
              http: { path: "/", port: 80 },
              initial_delay_seconds: 5,
              period_seconds: 10,
              timeout_seconds: 5,
              success_threshold: 1,
              failure_threshold: 3
            }
          }
        },
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
  },

  // Echo service
  echoService: {
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
        params: {
          health: {
            readiness: {
              http: { path: "/", port: 5678 },
              initial_delay_seconds: 5,
              period_seconds: 10,
              timeout_seconds: 5,
              success_threshold: 1,
              failure_threshold: 3
            }
          }
        },
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
  }
};

/**
 * 5. Demo Functions
 */

// Demo 1: Health Check
async function demoHealthCheck() {
  console.log('🏥 Demo 1: Health Check');
  console.log('========================');
  
  const api = new DeploymentAPIClient();
  
  try {
    const health = await api.health();
    console.log('✅ API Health:', health);
    return health;
  } catch (error) {
    console.error('❌ Health check failed:', error);
    throw error;
  }
}

// Demo 2: Create Deployment
async function demoCreateDeployment(orderId = ORDER_ID) {
  console.log('🚀 Demo 2: Create Deployment');
  console.log('============================');
  
  const api = new DeploymentAPIClient();
  
  try {
    const deployment = await api.createDeployment(orderId, SDLManifests.echoService);
    console.log('✅ Deployment created:', deployment);
    return deployment;
  } catch (error) {
    console.error('❌ Failed to create deployment:', error);
    throw error;
  }
}

// Demo 3: Monitor Deployment Status
async function demoMonitorDeployment(orderId = ORDER_ID) {
  console.log('📊 Demo 3: Monitor Deployment Status');
  console.log('====================================');
  
  const api = new DeploymentAPIClient();
  
  const checkStatus = async () => {
    try {
      const status = await api.getDeployment(orderId);
      console.log(`📈 Deployment status: ${status.status.state}`);
      
      if (status.status.state === 'running') {
        console.log('✅ Deployment is running!');
        console.log('📋 Services:', status.status.services);
        console.log('🔗 Endpoints:', status.status.endpoints);
        return status;
      } else if (status.status.state === 'failed') {
        console.error('❌ Deployment failed');
        return status;
      } else {
        console.log('⏳ Deployment is still starting...');
        // Check again in 5 seconds
        setTimeout(checkStatus, 5000);
      }
    } catch (error) {
      console.error('❌ Failed to get deployment status:', error);
    }
  };
  
  return checkStatus();
}

// Demo 4: Get Service Status
async function demoGetServiceStatus(orderId = ORDER_ID, serviceName = 'echo') {
  console.log('🔍 Demo 4: Get Service Status');
  console.log('==============================');
  
  const api = new DeploymentAPIClient();
  
  try {
    const status = await api.getServiceStatus(orderId, serviceName);
    console.log('✅ Service status:', status);
    return status;
  } catch (error) {
    console.error('❌ Failed to get service status:', error);
    throw error;
  }
}

// Demo 5: Get Deployment Logs
async function demoGetDeploymentLogs(orderId = ORDER_ID) {
  console.log('📝 Demo 5: Get Deployment Logs');
  console.log('==============================');
  
  const api = new DeploymentAPIClient();
  
  try {
    const logs = await api.getDeploymentLogs(orderId);
    console.log('✅ Deployment logs:', logs);
    return logs;
  } catch (error) {
    console.error('❌ Failed to get deployment logs:', error);
    throw error;
  }
}

// Demo 6: Real-time Log Streaming
function demoLogStreaming(orderId = ORDER_ID) {
  console.log('📡 Demo 6: Real-time Log Streaming');
  console.log('==================================');
  
  const logStreamer = new DeploymentLogStreamer(orderId, AUTH_CHAIN);
  
  // Custom log handler
  logStreamer.onLogMessage = (logMessage) => {
    const timestamp = new Date().toISOString();
    console.log(`[${timestamp}] [${logMessage.name}] ${logMessage.message}`);
    
    // You could also add to UI, store in database, etc.
    // document.getElementById('logs').innerHTML += 
    //   `<div>[${logMessage.name}] ${logMessage.message}</div>`;
  };
  
  logStreamer.connect();
  
  // Return the streamer so it can be disconnected later
  return logStreamer;
}

// Demo 7: Command Execution
function demoCommandExecution(orderId = ORDER_ID, podName = 'echo-deployment-abc123', serviceName = 'echo') {
  console.log('💻 Demo 7: Command Execution');
  console.log('============================');
  
  const executor = new DeploymentExecutor(orderId, podName, serviceName, AUTH_CHAIN);
  
  executor.executeCommand(['ls', '-la'], {
    onStdout: (data) => {
      console.log('📤 Command output:', data);
    },
    onStderr: (data) => {
      console.error('📤 Command error:', data);
    },
    onResult: (result) => {
      console.log('✅ Command completed with exit code:', result.exit_code);
    },
    onError: (error) => {
      console.error('❌ Command failed:', error);
    }
  });
  
  // Send some input to the command
  setTimeout(() => {
    executor.sendStdin('echo "Hello from terminal"\n');
  }, 1000);
  
  // Resize terminal
  setTimeout(() => {
    executor.resizeTerminal(80, 24);
  }, 2000);
  
  return executor;
}

// Demo 8: Cleanup Deployment
async function demoCleanupDeployment(orderId = ORDER_ID) {
  console.log('🧹 Demo 8: Cleanup Deployment');
  console.log('=============================');
  
  const api = new DeploymentAPIClient();
  
  try {
    const result = await api.cleanupDeployment(orderId);
    console.log('✅ Deployment cleaned up:', result);
    return result;
  } catch (error) {
    console.error('❌ Failed to cleanup deployment:', error);
    throw error;
  }
}

/**
 * 6. Complete Workflow Demo
 */
async function runCompleteWorkflow() {
  console.log('🎯 Complete Deployment Workflow Demo');
  console.log('====================================');
  
  const orderId = `demo-${Date.now()}`; // Generate unique order ID
  const api = new DeploymentAPIClient();
  let logStreamer = null;
  let executor = null;
  
  try {
    // Step 1: Health check
    await demoHealthCheck();
    
    // Step 2: Create deployment
    await demoCreateDeployment(orderId);
    
    // Step 3: Monitor deployment status
    await demoMonitorDeployment(orderId);
    
    // Step 4: Get service status
    await demoGetServiceStatus(orderId, 'echo');
    
    // Step 5: Get deployment logs
    await demoGetDeploymentLogs(orderId);
    
    // Step 6: Start real-time log streaming
    logStreamer = demoLogStreaming(orderId);
    
    // Step 7: Execute commands (if pod name is known)
    // Note: You'll need to get the actual pod name from the deployment status
    // executor = demoCommandExecution(orderId, 'actual-pod-name', 'echo');
    
    console.log('✅ Complete workflow demo finished successfully!');
    
    // Keep the demo running for a while to see logs
    setTimeout(async () => {
      console.log('🔄 Cleaning up demo deployment...');
      
      // Disconnect WebSocket connections
      if (logStreamer) logStreamer.disconnect();
      if (executor) executor.disconnect();
      
      // Cleanup deployment
      await demoCleanupDeployment(orderId);
      
      console.log('✅ Demo cleanup completed!');
    }, 30000); // Run for 30 seconds
    
  } catch (error) {
    console.error('❌ Workflow demo failed:', error);
    
    // Cleanup on error
    if (logStreamer) logStreamer.disconnect();
    if (executor) executor.disconnect();
  }
}

/**
 * 7. Utility Functions
 */

// Wait for deployment to be ready
async function waitForDeploymentReady(orderId, maxWaitTime = 300000) { // 5 minutes
  const api = new DeploymentAPIClient();
  const startTime = Date.now();
  
  while (Date.now() - startTime < maxWaitTime) {
    try {
      const status = await api.getDeployment(orderId);
      
      if (status.status.state === 'running') {
        console.log('✅ Deployment is ready!');
        return status;
      } else if (status.status.state === 'failed') {
        throw new Error('Deployment failed');
      }
      
      console.log(`⏳ Waiting for deployment to be ready... (${status.status.state})`);
      await new Promise(resolve => setTimeout(resolve, 5000)); // Wait 5 seconds
    } catch (error) {
      console.error('❌ Error checking deployment status:', error);
      await new Promise(resolve => setTimeout(resolve, 5000));
    }
  }
  
  throw new Error('Deployment did not become ready within the timeout period');
}

// Get pod names from deployment
async function getPodNames(orderId) {
  const api = new DeploymentAPIClient();
  const status = await api.getDeployment(orderId);
  
  if (status.status.pods) {
    return status.status.pods.map(pod => pod.name);
  }
  
  return [];
}

/**
 * 8. Export for use in other modules
 */
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    DeploymentAPIClient,
    DeploymentLogStreamer,
    DeploymentExecutor,
    SDLManifests,
    demoHealthCheck,
    demoCreateDeployment,
    demoMonitorDeployment,
    demoGetServiceStatus,
    demoGetDeploymentLogs,
    demoLogStreaming,
    demoCommandExecution,
    demoCleanupDeployment,
    runCompleteWorkflow,
    waitForDeploymentReady,
    getPodNames
  };
}

/**
 * 9. Browser Usage Example
 */
if (typeof window !== 'undefined') {
  // Make available globally for browser usage
  window.DeploymentAPI = {
    DeploymentAPIClient,
    DeploymentLogStreamer,
    DeploymentExecutor,
    SDLManifests,
    demoHealthCheck,
    demoCreateDeployment,
    demoMonitorDeployment,
    demoGetServiceStatus,
    demoGetDeploymentLogs,
    demoLogStreaming,
    demoCommandExecution,
    demoCleanupDeployment,
    runCompleteWorkflow,
    waitForDeploymentReady,
    getPodNames
  };
  
  console.log('🚀 Deployment API Demo loaded!');
  console.log('📖 Usage:');
  console.log('  - Run individual demos: await demoHealthCheck()');
  console.log('  - Run complete workflow: await runCompleteWorkflow()');
  console.log('  - Create API client: const api = new DeploymentAPIClient()');
}

/**
 * 10. Node.js Usage Example
 */
if (typeof module !== 'undefined' && module.exports) {
  console.log('🚀 Deployment API Demo loaded for Node.js!');
  console.log('📖 Usage:');
  console.log('  const { runCompleteWorkflow } = require("./deployment-api-demo.js");');
  console.log('  await runCompleteWorkflow();');
} 