package deployments

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

// DeploymentClient represents a client for the deployment API
type DeploymentClient struct {
	baseURL    string
	privateKey *ecdsa.PrivateKey
	address    string
	client     *http.Client
}

// NewDeploymentClient creates a new deployment client
func NewDeploymentClient(baseURL string, privateKey *ecdsa.PrivateKey) *DeploymentClient {
	return &DeploymentClient{
		baseURL:    baseURL,
		privateKey: privateKey,
		address:    crypto.PubkeyToAddress(privateKey.PublicKey).Hex(),
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

// createAuthorizationHeader creates an authorization header for API requests
func (c *DeploymentClient) createAuthorizationHeader(deploymentID, action string) (string, error) {
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("deployment:%s:%s:%d", deploymentID, action, timestamp)

	messageHash := crypto.Keccak256Hash([]byte(message))
	signature, err := crypto.Sign(messageHash.Bytes(), c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	authData := AuthorizationRequest{
		Signature: "0x" + hex.EncodeToString(signature),
		Address:   c.address,
		Message:   message,
		Timestamp: timestamp,
	}

	authJSON, err := json.Marshal(authData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth data: %w", err)
	}

	return string(authJSON), nil
}

// doRequest performs an HTTP request with authorization
func (c *DeploymentClient) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Extract deployment ID and action from path for authorization
	deploymentID := extractDeploymentID(path)
	action := extractAction(method, path)

	if deploymentID != "" && action != "" {
		authHeader, err := c.createAuthorizationHeader(deploymentID, action)
		if err != nil {
			return nil, fmt.Errorf("failed to create authorization header: %w", err)
		}
		req.Header.Set("Authorization", authHeader)
	}

	req.Header.Set("Content-Type", "application/json")

	return c.client.Do(req)
}

// extractDeploymentID extracts deployment ID from API path
func extractDeploymentID(path string) string {
	// Simple extraction - in a real implementation, you might use regex
	if len(path) > 20 && path[:20] == "/api/v1/deployments/" {
		// Find the next slash after /api/v1/deployments/
		for i := 20; i < len(path); i++ {
			if path[i] == '/' {
				return path[20:i]
			}
		}
		// If no slash found, return the rest
		return path[20:]
	}
	return ""
}

// extractAction extracts the action from HTTP method and path
func extractAction(method, path string) string {
	switch method {
	case "GET":
		if contains(path, "/logs") {
			return "logs"
		}
		if contains(path, "/metrics") {
			return "metrics"
		}
		if contains(path, "/inspect") {
			return "inspect"
		}
		return "get"
	case "POST":
		if contains(path, "/start") {
			return "start"
		}
		if contains(path, "/stop") {
			return "stop"
		}
		if contains(path, "/restart") {
			return "restart"
		}
		if contains(path, "/scale") {
			return "scale"
		}
		if contains(path, "/exec") {
			return "exec"
		}
		return "create"
	case "PUT":
		if contains(path, "/image") {
			return "update_image"
		}
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

// GetDeployment retrieves a deployment by ID
func (c *DeploymentClient) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	path := fmt.Sprintf("/api/v1/deployments/%s", deploymentID)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get deployment: %s - %s", resp.Status, string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Error)
	}

	// Convert response data to Deployment
	deploymentData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal deployment data: %w", err)
	}

	var deployment Deployment
	if err := json.Unmarshal(deploymentData, &deployment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment: %w", err)
	}

	return &deployment, nil
}

// StartDeployment starts a deployment
func (c *DeploymentClient) StartDeployment(ctx context.Context, deploymentID string) error {
	path := fmt.Sprintf("/api/v1/deployments/%s/start", deploymentID)

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return fmt.Errorf("failed to start deployment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start deployment: %s - %s", resp.Status, string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("API error: %s", response.Error)
	}

	return nil
}

// StopDeployment stops a deployment
func (c *DeploymentClient) StopDeployment(ctx context.Context, deploymentID string) error {
	path := fmt.Sprintf("/api/v1/deployments/%s/stop", deploymentID)

	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return fmt.Errorf("failed to stop deployment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to stop deployment: %s - %s", resp.Status, string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("API error: %s", response.Error)
	}

	return nil
}

// GetDeploymentLogs retrieves deployment logs
func (c *DeploymentClient) GetDeploymentLogs(ctx context.Context, deploymentID string, tail int) (string, error) {
	path := fmt.Sprintf("/api/v1/deployments/%s/logs?tail=%d", deploymentID, tail)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get deployment logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get deployment logs: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// UpdateDeploymentImage updates the image of a deployment service
func (c *DeploymentClient) UpdateDeploymentImage(ctx context.Context, deploymentID, serviceName, image string) error {
	path := fmt.Sprintf("/api/v1/deployments/%s/services/%s/image", deploymentID, serviceName)

	imageData := map[string]string{"image": image}
	body, err := json.Marshal(imageData)
	if err != nil {
		return fmt.Errorf("failed to marshal image data: %w", err)
	}

	resp, err := c.doRequest("PUT", path, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to update deployment image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update deployment image: %s - %s", resp.Status, string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("API error: %s", response.Error)
	}

	return nil
}

// Example usage function
func ExampleUsage() {
	// This function demonstrates how to use the DeploymentClient

	// In a real application, you would load the private key from a secure source
	// privateKey, err := crypto.HexToECDSA("your-private-key-here")
	// if err != nil {
	//     log.Fatal(err)
	// }

	// Create client
	// client := NewDeploymentClient("http://localhost:8080", privateKey)

	// Get deployment
	// deployment, err := client.GetDeployment(context.Background(), "123")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// fmt.Printf("Deployment: %+v\n", deployment)

	// Start deployment
	// err = client.StartDeployment(context.Background(), "123")
	// if err != nil {
	//     log.Fatal(err)
	// }

	// Get logs
	// logs, err := client.GetDeploymentLogs(context.Background(), "123", 100)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// fmt.Printf("Logs: %s\n", logs)

	// Update image
	// err = client.UpdateDeploymentImage(context.Background(), "123", "web", "nginx:latest")
	// if err != nil {
	//     log.Fatal(err)
	// }
}
