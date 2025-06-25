package deployments

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Example demonstrates how to use the deployment service with monitoring, logs, and exec features
func Example() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create deployment service
	service := NewService(nil, logger)

	// Set up managers (these would be actual implementations in production)
	// service.SetManagers(tenantManager, eventManager, storageManager, resourceManager, factory)

	// Start the service
	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		log.Fatal("Failed to start deployment service:", err)
	}
	defer service.Stop(ctx)

	// Example 1: Create a deployment
	fmt.Println("=== Example 1: Creating a deployment ===")
	deployment := &Deployment{
		ID:        "example-deploy-123",
		TenantID:  "tenant-456",
		Name:      "web-application",
		Type:      DeploymentTypeDocker,
		Status:    string(DeploymentStatusPending),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// In a real implementation, you would create the deployment
	// err := service.CreateDeployment(ctx, deployment)
	// if err != nil {
	//     log.Fatal("Failed to create deployment:", err)
	// }

	fmt.Printf("Created deployment: %s (%s)\n", deployment.Name, deployment.ID)

	// Example 2: Get deployment logs
	fmt.Println("\n=== Example 2: Getting deployment logs ===")
	logs, err := service.GetDeploymentLogs(ctx, "example-deploy-123", "", 100)
	if err != nil {
		fmt.Printf("Error getting logs: %v\n", err)
	} else {
		defer logs.Close()
		fmt.Println("Logs retrieved successfully")
		// Copy logs to stdout
		io.Copy(os.Stdout, logs)
	}

	// Example 3: Stream real-time logs
	fmt.Println("\n=== Example 3: Streaming real-time logs ===")
	logChan, err := service.StreamDeploymentLogs(ctx, "example-deploy-123", "web", true)
	if err != nil {
		fmt.Printf("Error streaming logs: %v\n", err)
	} else {
		// In a real application, you would process the log channel
		fmt.Println("Log streaming started")
		// Simulate receiving some logs
		go func() {
			for logEntry := range logChan {
				fmt.Printf("[%s] %s: %s\n", logEntry.Timestamp, logEntry.Service, logEntry.Message)
			}
		}()
	}

	// Example 4: Execute commands in container
	fmt.Println("\n=== Example 4: Executing commands ===")
	session, err := service.ExecConsole(ctx, "example-deploy-123", "web", []string{"ls", "-la"}, false)
	if err != nil {
		fmt.Printf("Error creating exec session: %v\n", err)
	} else {
		defer session.Close()
		fmt.Println("Exec session created successfully")

		// Execute a command
		result, err := session.Execute(ctx, []string{"ls", "-la"})
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
		} else {
			fmt.Printf("Command executed successfully\n")
			fmt.Printf("Exit code: %d\n", result.ExitCode)
			fmt.Printf("Output: %s\n", result.Stdout)
			if result.Stderr != "" {
				fmt.Printf("Error: %s\n", result.Stderr)
			}
		}
	}

	// Example 5: Inspect deployment
	fmt.Println("\n=== Example 5: Inspecting deployment ===")
	inspection, err := service.InspectDeployment(ctx, "example-deploy-123")
	if err != nil {
		fmt.Printf("Error inspecting deployment: %v\n", err)
	} else {
		fmt.Printf("Deployment: %s\n", inspection.Name)
		fmt.Printf("Status: %s\n", inspection.Status)
		fmt.Printf("Services: %d\n", len(inspection.Services))

		for name, service := range inspection.Services {
			fmt.Printf("  - %s: %s (%s)\n", name, service.Image, service.Status)
		}

		if inspection.Resources != nil {
			fmt.Printf("CPU Usage: %.2f%%\n", inspection.Resources.CPUUsage)
			fmt.Printf("Memory Usage: %d MB\n", inspection.Resources.MemoryUsage/1024/1024)
		}
	}

	// Example 6: Inspect specific service
	fmt.Println("\n=== Example 6: Inspecting service ===")
	serviceInspection, err := service.InspectService(ctx, "example-deploy-123", "web")
	if err != nil {
		fmt.Printf("Error inspecting service: %v\n", err)
	} else {
		fmt.Printf("Service: %s\n", serviceInspection.Name)
		fmt.Printf("Image: %s\n", serviceInspection.Image)
		fmt.Printf("Status: %s\n", serviceInspection.Status)
		fmt.Printf("Replicas: %d\n", serviceInspection.Replicas)

		if serviceInspection.Resources != nil {
			fmt.Printf("CPU Usage: %.2f%%\n", serviceInspection.Resources.CPUUsage)
			fmt.Printf("Memory Usage: %d MB\n", serviceInspection.Resources.MemoryUsage/1024/1024)
		}

		fmt.Printf("Ports: %d\n", len(serviceInspection.Ports))
		for _, port := range serviceInspection.Ports {
			fmt.Printf("  - %d:%d (%s)\n", port.HostPort, port.ContainerPort, port.Protocol)
		}
	}

	// Example 7: Get deployment metrics
	fmt.Println("\n=== Example 7: Getting deployment metrics ===")
	metrics, err := service.GetDeploymentMetrics(ctx, "example-deploy-123", time.Hour)
	if err != nil {
		fmt.Printf("Error getting metrics: %v\n", err)
	} else {
		fmt.Printf("Deployment Metrics:\n")
		fmt.Printf("  Duration: %v\n", metrics.Duration)
		fmt.Printf("  Services: %d\n", len(metrics.Services))

		if metrics.Total != nil {
			fmt.Printf("  Total CPU Usage: %.2f%%\n", metrics.Total.CPUUsage)
			fmt.Printf("  Total Memory Usage: %d MB\n", metrics.Total.MemoryUsage/1024/1024)
			fmt.Printf("  Total Disk Usage: %d MB\n", metrics.Total.DiskUsage/1024/1024)
		}

		for serviceName, serviceMetrics := range metrics.Services {
			fmt.Printf("  Service %s:\n", serviceName)
			if serviceMetrics.Resources != nil {
				fmt.Printf("    CPU: %.2f%%\n", serviceMetrics.Resources.CPUUsage)
				fmt.Printf("    Memory: %d MB\n", serviceMetrics.Resources.MemoryUsage/1024/1024)
			}
			if serviceMetrics.Requests != nil {
				fmt.Printf("    Requests: %d total, %d successful, %d failed\n",
					serviceMetrics.Requests.Total,
					serviceMetrics.Requests.Successful,
					serviceMetrics.Requests.Failed)
			}
		}
	}

	// Example 8: Get service metrics
	fmt.Println("\n=== Example 8: Getting service metrics ===")
	serviceMetrics, err := service.GetServiceMetrics(ctx, "example-deploy-123", "web", time.Hour)
	if err != nil {
		fmt.Printf("Error getting service metrics: %v\n", err)
	} else {
		fmt.Printf("Service Metrics for 'web':\n")
		fmt.Printf("  Duration: %v\n", serviceMetrics.Duration)

		if serviceMetrics.Resources != nil {
			fmt.Printf("  CPU Usage: %.2f%%\n", serviceMetrics.Resources.CPUUsage)
			fmt.Printf("  Memory Usage: %d MB\n", serviceMetrics.Resources.MemoryUsage/1024/1024)
		}

		if serviceMetrics.Requests != nil {
			fmt.Printf("  Requests:\n")
			fmt.Printf("    Total: %d\n", serviceMetrics.Requests.Total)
			fmt.Printf("    Successful: %d\n", serviceMetrics.Requests.Successful)
			fmt.Printf("    Failed: %d\n", serviceMetrics.Requests.Failed)
			fmt.Printf("    Avg Response: %.2fms\n", serviceMetrics.Requests.AvgResponse)
		}

		if serviceMetrics.Errors != nil {
			fmt.Printf("  Errors:\n")
			fmt.Printf("    Total: %d\n", serviceMetrics.Errors.Total)
			if serviceMetrics.Errors.LastError != nil {
				fmt.Printf("    Last Error: %s (%s)\n",
					serviceMetrics.Errors.LastError.Message,
					serviceMetrics.Errors.LastError.Type)
			}
		}
	}

	// Example 9: Scale deployment
	fmt.Println("\n=== Example 9: Scaling deployment ===")
	err = service.ScaleDeployment(ctx, "example-deploy-123", "web", 3)
	if err != nil {
		fmt.Printf("Error scaling deployment: %v\n", err)
	} else {
		fmt.Println("Deployment scaled successfully to 3 replicas")
	}

	// Example 10: Update deployment image
	fmt.Println("\n=== Example 10: Updating deployment image ===")
	err = service.UpdateDeploymentImage(ctx, "example-deploy-123", "web", "nginx:1.21")
	if err != nil {
		fmt.Printf("Error updating image: %v\n", err)
	} else {
		fmt.Println("Deployment image updated successfully")
	}

	// Example 11: Get tenant resource usage
	fmt.Println("\n=== Example 11: Getting tenant resource usage ===")
	usage, err := service.GetTenantResourceUsage(ctx, "tenant-456")
	if err != nil {
		fmt.Printf("Error getting tenant resource usage: %v\n", err)
	} else {
		fmt.Printf("Tenant Resource Usage:\n")
		fmt.Printf("  CPU Usage: %.2f%%\n", usage.CPUUsage)
		fmt.Printf("  Memory Usage: %d MB\n", usage.MemoryUsage/1024/1024)
		fmt.Printf("  Disk Usage: %d MB\n", usage.DiskUsage/1024/1024)
		fmt.Printf("  Network Usage: %d MB\n", usage.NetworkUsage/1024/1024)
		fmt.Printf("  Port Count: %d\n", usage.PortCount)
		fmt.Printf("  Container Count: %d\n", usage.ContainerCount)
	}

	fmt.Println("\n=== Examples completed ===")
}

// ExampleWebSocket demonstrates WebSocket-based interactive console
func ExampleWebSocket() {
	fmt.Println("=== WebSocket Interactive Console Example ===")

	// This would be implemented with a WebSocket library like gorilla/websocket
	// The implementation would allow real-time interactive terminal access

	fmt.Println("WebSocket interactive console would provide:")
	fmt.Println("- Real-time bidirectional communication")
	fmt.Println("- Interactive terminal access to containers")
	fmt.Println("- Support for all terminal features (colors, cursor movement, etc.)")
	fmt.Println("- Session management and authentication")
	fmt.Println("- Multi-user support with proper isolation")
}

// ExampleEventStreaming demonstrates event streaming capabilities
func ExampleEventStreaming() {
	fmt.Println("=== Event Streaming Example ===")

	// This would demonstrate how to stream deployment events in real-time
	// Events would include deployment lifecycle events, health checks, metrics, etc.

	fmt.Println("Event streaming would provide:")
	fmt.Println("- Real-time deployment lifecycle events")
	fmt.Println("- Health check status updates")
	fmt.Println("- Resource usage alerts")
	fmt.Println("- Performance metrics")
	fmt.Println("- Error notifications")
	fmt.Println("- Integration with external monitoring systems")
}
