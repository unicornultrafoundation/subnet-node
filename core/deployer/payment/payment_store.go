package payment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/unicornultrafoundation/subnet-node/core/deployer/types"
)

// PaymentStore handles payment and escrow data persistence
type PaymentStore struct {
	storeDir string
	mu       sync.RWMutex
}

// NewPaymentStore creates a new payment store
func NewPaymentStore(storeDir string) (*PaymentStore, error) {
	// Create store directory if it doesn't exist
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	return &PaymentStore{
		storeDir: storeDir,
	}, nil
}

// StorePayment stores a payment transaction
func (s *PaymentStore) StorePayment(payment *types.Payment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create payment directory
	paymentDir := filepath.Join(s.storeDir, payment.DeploymentID, "payments")
	if err := os.MkdirAll(paymentDir, 0755); err != nil {
		return fmt.Errorf("failed to create payment directory: %w", err)
	}

	// Create payment file path
	paymentPath := filepath.Join(paymentDir, fmt.Sprintf("%d.json", payment.Timestamp.Unix()))

	// Marshal payment to JSON
	data, err := json.MarshalIndent(payment, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal payment: %w", err)
	}

	// Write payment to file
	if err := os.WriteFile(paymentPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write payment file: %w", err)
	}

	return nil
}

// GetPayment gets the latest payment for a deployment
func (s *PaymentStore) GetPayment(deploymentID string) (*types.Payment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create payment directory path
	paymentDir := filepath.Join(s.storeDir, deploymentID, "payments")

	// Read payment directory
	files, err := os.ReadDir(paymentDir)
	if err != nil {
		return nil, fmt.Errorf("no payment found")
	}

	// Find latest payment file
	var latestFile string
	var latestTime int64
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Parse timestamp from filename
		var timestamp int64
		if _, err := fmt.Sscanf(file.Name(), "%d.json", &timestamp); err != nil {
			continue
		}

		if timestamp > latestTime {
			latestTime = timestamp
			latestFile = file.Name()
		}
	}

	if latestFile == "" {
		return nil, fmt.Errorf("no payment found")
	}

	// Read latest payment file
	data, err := os.ReadFile(filepath.Join(paymentDir, latestFile))
	if err != nil {
		return nil, fmt.Errorf("no payment found")
	}

	// Unmarshal payment from JSON
	var payment types.Payment
	if err := json.Unmarshal(data, &payment); err != nil {
		return nil, fmt.Errorf("no payment found")
	}

	return &payment, nil
}

// UpdatePayment updates a payment
func (s *PaymentStore) UpdatePayment(payment *types.Payment) error {
	return s.StorePayment(payment)
}

// GetPaymentHistory gets payment history for a deployment
func (s *PaymentStore) GetPaymentHistory(deploymentID string) ([]*types.Payment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create payment directory path
	paymentDir := filepath.Join(s.storeDir, deploymentID, "payments")

	// Read payment directory
	files, err := os.ReadDir(paymentDir)
	if err != nil {
		return nil, fmt.Errorf("no payment found")
	}

	// Read and parse payment files
	var payments []*types.Payment
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Read payment file
		data, err := os.ReadFile(filepath.Join(paymentDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("no payment found")
		}

		// Unmarshal payment from JSON
		var payment types.Payment
		if err := json.Unmarshal(data, &payment); err != nil {
			return nil, fmt.Errorf("no payment found")
		}

		payments = append(payments, &payment)
	}

	return payments, nil
}

// StoreEscrow stores an escrow account
func (s *PaymentStore) StoreEscrow(escrow *types.Escrow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create escrow directory
	escrowDir := filepath.Join(s.storeDir, escrow.DeploymentID, "escrows")
	if err := os.MkdirAll(escrowDir, 0755); err != nil {
		return fmt.Errorf("failed to create escrow directory: %w", err)
	}

	// Create escrow file path
	escrowPath := filepath.Join(escrowDir, fmt.Sprintf("%d.json", escrow.CreatedAt.Unix()))

	// Marshal escrow to JSON
	data, err := json.MarshalIndent(escrow, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal escrow: %w", err)
	}

	// Write escrow to file
	if err := os.WriteFile(escrowPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write escrow file: %w", err)
	}

	return nil
}

// GetEscrow gets the latest escrow for a deployment
func (s *PaymentStore) GetEscrow(deploymentID string) (*types.Escrow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create escrow directory path
	escrowDir := filepath.Join(s.storeDir, deploymentID, "escrows")

	// Read escrow directory
	files, err := os.ReadDir(escrowDir)
	if err != nil {
		return nil, fmt.Errorf("no escrow found")
	}

	// Find latest escrow file
	var latestFile string
	var latestTime int64
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Parse timestamp from filename
		var timestamp int64
		if _, err := fmt.Sscanf(file.Name(), "%d.json", &timestamp); err != nil {
			continue
		}

		if timestamp > latestTime {
			latestTime = timestamp
			latestFile = file.Name()
		}
	}

	if latestFile == "" {
		return nil, fmt.Errorf("no escrow found")
	}

	// Read latest escrow file
	data, err := os.ReadFile(filepath.Join(escrowDir, latestFile))
	if err != nil {
		return nil, fmt.Errorf("no escrow found")
	}

	// Unmarshal escrow from JSON
	var escrow types.Escrow
	if err := json.Unmarshal(data, &escrow); err != nil {
		return nil, fmt.Errorf("no escrow found")
	}

	return &escrow, nil
}

// UpdateEscrow updates an escrow
func (s *PaymentStore) UpdateEscrow(escrow *types.Escrow) error {
	return s.StoreEscrow(escrow)
}

// GetEscrowHistory gets escrow history for a deployment
func (s *PaymentStore) GetEscrowHistory(deploymentID string) ([]*types.Escrow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create escrow directory path
	escrowDir := filepath.Join(s.storeDir, deploymentID, "escrows")

	// Read escrow directory
	files, err := os.ReadDir(escrowDir)
	if err != nil {
		return nil, fmt.Errorf("no escrow found")
	}

	// Read and parse escrow files
	var escrows []*types.Escrow
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Read escrow file
		data, err := os.ReadFile(filepath.Join(escrowDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("no escrow found")
		}

		// Unmarshal escrow from JSON
		var escrow types.Escrow
		if err := json.Unmarshal(data, &escrow); err != nil {
			return nil, fmt.Errorf("no escrow found")
		}

		escrows = append(escrows, &escrow)
	}

	return escrows, nil
}

// GetActiveEscrows gets all active escrows
func (s *PaymentStore) GetActiveEscrows() ([]*types.Escrow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Read store directory
	deployments, err := os.ReadDir(s.storeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read store directory: %w", err)
	}

	var activeEscrows []*types.Escrow
	for _, deployment := range deployments {
		if !deployment.IsDir() {
			continue
		}

		// Get latest escrow
		escrow, err := s.GetEscrow(deployment.Name())
		if err != nil {
			continue
		}

		// Check if escrow is active
		if escrow.Status == "active" {
			activeEscrows = append(activeEscrows, escrow)
		}
	}

	return activeEscrows, nil
}

// GetDeployment gets a deployment
func (s *PaymentStore) GetDeployment(deploymentID string) (*types.ManagedDeployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create deployment file path
	depPath := filepath.Join(s.storeDir, deploymentID, "deployment.json")

	// Read deployment file
	data, err := os.ReadFile(depPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployment file: %w", err)
	}

	// Unmarshal deployment from JSON
	var dep types.ManagedDeployment
	if err := json.Unmarshal(data, &dep); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment: %w", err)
	}

	return &dep, nil
}

// UpdateDeployment updates a deployment
func (s *PaymentStore) UpdateDeployment(dep *types.ManagedDeployment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create deployment directory
	depDir := filepath.Join(s.storeDir, dep.ID)
	if err := os.MkdirAll(depDir, 0755); err != nil {
		return fmt.Errorf("failed to create deployment directory: %w", err)
	}

	// Create deployment file path
	depPath := filepath.Join(depDir, "deployment.json")

	// Marshal deployment to JSON
	data, err := json.MarshalIndent(dep, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal deployment: %w", err)
	}

	// Write deployment to file
	if err := os.WriteFile(depPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write deployment file: %w", err)
	}

	return nil
}
