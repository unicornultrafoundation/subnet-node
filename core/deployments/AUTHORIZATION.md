# Deployment API Authorization

This document describes the authorization mechanism for the Deployment API, which uses Ethereum-based signature verification to ensure only authorized users can perform operations on deployments.

## Overview

The Deployment API implements a signature-based authorization system where:
- Deployment IDs correspond to order IDs in the bid market
- Authorization is validated by verifying the signature of the deployment owner (order owner)
- All protected endpoints require a valid Authorization header

## Message Format

The authorization message follows this format:
```
deployment:{deployment_id}:{action}:{timestamp}
```

Where:
- `deployment_id`: The deployment ID (corresponds to order ID)
- `action`: The action being performed (e.g., "get", "start", "stop", "delete", "update_image")
- `timestamp`: Unix timestamp to prevent replay attacks

## Signature Generation

To generate a valid signature:

1. Create the message string: `deployment:{deployment_id}:{action}:{timestamp}`
2. Hash the message using Keccak256: `hash = keccak256(message)`
3. Sign the hash with your private key: `signature = sign(hash, privateKey)`
4. Format the signature as a hex string with "0x" prefix

## Authorization Header

The Authorization header must contain a JSON object with the following fields:

```json
{
  "signature": "0x...",
  "address": "0x...",
  "message": "deployment:123:get:1640995200",
  "timestamp": 1640995200
}
```

## Validation Process

1. **Header Validation**: Check if Authorization header is present and valid JSON
2. **Timestamp Validation**: Ensure timestamp is within acceptable range (e.g., ±5 minutes)
3. **Deployment ID Validation**: Verify deployment ID is a valid integer
4. **Order Lookup**: Retrieve order information from bid market contract
5. **Ownership Verification**: Verify the signer's address matches the order owner
6. **Signature Verification**: Verify the signature against the message and address

## Protected Endpoints

The following endpoints require authorization:

### Deployment Management
- `POST /api/v1/deployments` - Create deployment
- `GET /api/v1/deployments/{id}` - Get deployment details
- `PUT /api/v1/deployments/{id}` - Update deployment
- `DELETE /api/v1/deployments/{id}` - Delete deployment
- `POST /api/v1/deployments/{id}/start` - Start deployment
- `POST /api/v1/deployments/{id}/stop` - Stop deployment
- `POST /api/v1/deployments/{id}/restart` - Restart deployment

### Monitoring and Inspection
- `GET /api/v1/deployments/{id}/inspect` - Inspect deployment
- `GET /api/v1/deployments/{id}/services/{service}/inspect` - Inspect service
- `GET /api/v1/deployments/{id}/metrics` - Get deployment metrics
- `GET /api/v1/deployments/{id}/services/{service}/metrics` - Get service metrics

### Logs
- `GET /api/v1/deployments/{id}/logs` - Get deployment logs
- `GET /api/v1/deployments/{id}/services/{service}/logs` - Get service logs
- `GET /api/v1/deployments/{id}/logs/stream` - Stream deployment logs
- `GET /api/v1/deployments/{id}/services/{service}/logs/stream` - Stream service logs

### Console Execution
- `POST /api/v1/deployments/{id}/services/{service}/exec` - Execute console command
- `GET /api/v1/deployments/{id}/services/{service}/exec/ws` - WebSocket console

### Image Updates
- `PUT /api/v1/deployments/{id}/services/{service}/image` - Update deployment image

### Events
- `GET /api/v1/deployments/{id}/events` - Get deployment events
- `GET /api/v1/deployments/{id}/events/stream` - Stream deployment events

## Public Endpoints

The following endpoints do not require authorization:

- `GET /api/v1/deployments` - List all deployments (filtered by authorization)

## Error Responses

### Missing Authorization Header
```json
{
  "success": false,
  "error": "Missing authorization header"
}
```

### Invalid Authorization Format
```json
{
  "success": false,
  "error": "Invalid authorization format"
}
```

### Invalid Deployment ID
```json
{
  "success": false,
  "error": "Invalid deployment ID format"
}
```

### Order Not Found
```json
{
  "success": false,
  "error": "Failed to get order information"
}
```

### Not Authorized
```json
{
  "success": false,
  "error": "Not authorized: not the deployment owner"
}
```

### Invalid Signature
```json
{
  "success": false,
  "error": "Invalid signature"
}
```

### Expired Timestamp
```json
{
  "success": false,
  "error": "Authorization expired"
}
```

## Example Usage

### cURL Example

```bash
#!/bin/bash

# Configuration
DEPLOYMENT_ID="123"
ACTION="get"
TIMESTAMP=$(date +%s)
PRIVATE_KEY="your_private_key_here"
API_URL="http://localhost:8080"

# Create message
MESSAGE="deployment:${DEPLOYMENT_ID}:${ACTION}:${TIMESTAMP}"

# Generate signature (requires web3 or similar tool)
SIGNATURE=$(echo -n "$MESSAGE" | keccak256 | sign_with_private_key "$PRIVATE_KEY")

# Get address from private key
ADDRESS=$(get_address_from_private_key "$PRIVATE_KEY")

# Create authorization header
AUTH_HEADER=$(cat <<EOF
{
  "signature": "0x${SIGNATURE}",
  "address": "${ADDRESS}",
  "message": "${MESSAGE}",
  "timestamp": ${TIMESTAMP}
}
EOF
)

# Make API request
curl -X GET \
  "${API_URL}/api/v1/deployments/${DEPLOYMENT_ID}" \
  -H "Authorization: ${AUTH_HEADER}" \
  -H "Content-Type: application/json"
```

### JavaScript Example

```javascript
const ethers = require('ethers');

class DeploymentAPIClient {
  constructor(baseURL, privateKey) {
    this.baseURL = baseURL;
    this.wallet = new ethers.Wallet(privateKey);
  }

  async createAuthHeader(deploymentId, action) {
    const timestamp = Math.floor(Date.now() / 1000);
    const message = `deployment:${deploymentId}:${action}:${timestamp}`;
    
    const messageHash = ethers.utils.keccak256(ethers.utils.toUtf8Bytes(message));
    const signature = await this.wallet.signMessage(ethers.utils.arrayify(messageHash));
    
    return {
      signature: signature,
      address: this.wallet.address,
      message: message,
      timestamp: timestamp
    };
  }

  async getDeployment(deploymentId) {
    const authHeader = await this.createAuthHeader(deploymentId, 'get');
    
    const response = await fetch(`${this.baseURL}/api/v1/deployments/${deploymentId}`, {
      method: 'GET',
      headers: {
        'Authorization': JSON.stringify(authHeader),
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`API request failed: ${response.statusText}`);
    }
    
    return response.json();
  }

  async startDeployment(deploymentId) {
    const authHeader = await this.createAuthHeader(deploymentId, 'start');
    
    const response = await fetch(`${this.baseURL}/api/v1/deployments/${deploymentId}/start`, {
      method: 'POST',
      headers: {
        'Authorization': JSON.stringify(authHeader),
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`API request failed: ${response.statusText}`);
    }
    
    return response.json();
  }

  async updateDeploymentImage(deploymentId, serviceName, image) {
    const authHeader = await this.createAuthHeader(deploymentId, 'update_image');
    
    const response = await fetch(`${this.baseURL}/api/v1/deployments/${deploymentId}/services/${serviceName}/image`, {
      method: 'PUT',
      headers: {
        'Authorization': JSON.stringify(authHeader),
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ image })
    });
    
    if (!response.ok) {
      throw new Error(`API request failed: ${response.statusText}`);
    }
    
    return response.json();
  }
}

// Usage
const client = new DeploymentAPIClient('http://localhost:8080', 'your_private_key_here');

// Get deployment
client.getDeployment('123').then(deployment => {
  console.log('Deployment:', deployment);
});

// Start deployment
client.startDeployment('123').then(result => {
  console.log('Deployment started:', result);
});

// Update image
client.updateDeploymentImage('123', 'web', 'nginx:latest').then(result => {
  console.log('Image updated:', result);
});
```

### Go Example

```go
package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/unicornultrafoundation/subnet-node/core/deployments"
)

func main() {
	// Load private key
	privateKey, err := crypto.HexToECDSA("your_private_key_here")
	if err != nil {
		log.Fatal(err)
	}

	// Create client
	client := deployments.NewDeploymentClient("http://localhost:8080", privateKey)

	ctx := context.Background()

	// Get deployment
	deployment, err := client.GetDeployment(ctx, "123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deployment: %+v\n", deployment)

	// Start deployment
	err = client.StartDeployment(ctx, "123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deployment started successfully")

	// Update image
	err = client.UpdateDeploymentImage(ctx, "123", "web", "nginx:latest")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Image updated successfully")
}
```

## Security Considerations

1. **Private Key Security**: Never expose private keys in client-side code or logs
2. **Timestamp Validation**: Always validate timestamps to prevent replay attacks
3. **Message Format**: Ensure consistent message formatting across all clients
4. **Signature Verification**: Verify signatures on the server side using the correct address
5. **Rate Limiting**: Implement rate limiting to prevent abuse
6. **HTTPS**: Always use HTTPS in production to protect authorization headers
7. **Audit Logging**: Log all authorization attempts for security monitoring

## Implementation Notes

- The authorization system is designed to work with the existing bid market contract
- Deployment IDs must be valid order IDs in the bid market
- Only the order owner can perform operations on their deployments
- The system supports multiple deployment types (Docker, Kubernetes, etc.)
- Authorization is validated for each protected endpoint
- Failed authorization attempts are logged for security monitoring 