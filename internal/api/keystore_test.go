package api

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

// mockRepoImpl is a simple mock implementation of repo.Repo for testing
type mockRepoImpl struct {
	path string
	cfg  *config.C
}

func (m *mockRepoImpl) Config() *config.C {
	return m.cfg
}

func (m *mockRepoImpl) Path() string {
	return m.path
}

func (m *mockRepoImpl) Datastore() repo.Datastore {
	return nil
}

func (m *mockRepoImpl) SwarmKey() ([]byte, error) {
	return nil, nil
}

func (m *mockRepoImpl) Close() error {
	return nil
}

// mockRepo creates a temporary test repository
func mockRepo(t *testing.T) repo.Repo {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Ensure keystore directory exists
	keystoreDir := filepath.Join(tmpDir, "keystore")
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		t.Fatalf("Failed to create keystore dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return &mockRepoImpl{
		path: tmpDir,
		cfg:  nil,
	}
}

func TestKeystoreAPI_CreateAndList(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Initially should have no accounts
	accounts, err := api.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(accounts) != 0 {
		t.Fatalf("Expected 0 accounts, got %d", len(accounts))
	}

	// Create a new account
	createReq := CreateAccountRequest{
		Password: "test-password-123",
	}

	createResp, err := api.Create(ctx, createReq)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if createResp.Address == "" {
		t.Fatal("Expected non-empty address")
	}

	t.Logf("Created account: %s", createResp.Address)

	// List should now have 1 account
	accounts, err = api.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(accounts))
	}
	if accounts[0].Address != createResp.Address {
		t.Fatalf("Address mismatch: expected %s, got %s", createResp.Address, accounts[0].Address)
	}
}

func TestKeystoreAPI_ImportPrivateKey(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Test private key (DO NOT USE IN PRODUCTION)
	testPrivateKey := "0x4c0883a69102937d6231471b5dbb6204fe512961708279f8c5b0c6e9a6d74f12"

	importReq := ImportPrivateKeyRequest{
		PrivateKeyHex: testPrivateKey,
		Password:      "test-password",
	}

	importResp, err := api.ImportPrivateKey(ctx, importReq)
	if err != nil {
		t.Fatalf("ImportPrivateKey failed: %v", err)
	}

	t.Logf("Imported account: %s", importResp.Address)

	// Verify the account was imported
	accounts, err := api.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(accounts))
	}
}

func TestKeystoreAPI_UpdatePassword(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Create an account
	createReq := CreateAccountRequest{
		Password: "old-password",
	}

	createResp, err := api.Create(ctx, createReq)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	address := createResp.Address

	// Update password
	updateReq := UpdatePasswordRequest{
		Address:     address,
		OldPassword: "old-password",
		NewPassword: "new-password",
	}

	err = api.UpdatePassword(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdatePassword failed: %v", err)
	}

	// Verify new password works by exporting
	exportReq := ExportAccountRequest{
		Address:  address,
		Password: "new-password",
	}

	exportResp, err := api.Export(ctx, exportReq)
	if err != nil {
		t.Fatalf("Export with new password failed: %v", err)
	}

	if exportResp.KeystoreJSON == "" {
		t.Fatal("Expected non-empty keystore JSON")
	}

	// Verify old password doesn't work
	exportReqOld := ExportAccountRequest{
		Address:  address,
		Password: "old-password",
	}

	_, err = api.Export(ctx, exportReqOld)
	if err == nil {
		t.Fatal("Expected export with old password to fail")
	}
}

func TestKeystoreAPI_ExportAndImportKeystore(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Create an account
	createReq := CreateAccountRequest{
		Password: "test-password",
	}

	createResp, err := api.Create(ctx, createReq)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	address := createResp.Address

	// Export the keystore
	exportReq := ExportAccountRequest{
		Address:  address,
		Password: "test-password",
	}

	exportResp, err := api.Export(ctx, exportReq)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	keystoreJSON := exportResp.KeystoreJSON

	// Create a new repo for importing
	repo2 := mockRepo(t)
	api2 := NewKeystoreAPI(repo2)

	// Import the keystore
	importReq := ImportKeystoreRequest{
		KeystoreJSON: keystoreJSON,
		Password:     "test-password",
		NewPassword:  "new-password",
	}

	importResp, err := api2.ImportKeystore(ctx, importReq)
	if err != nil {
		t.Fatalf("ImportKeystore failed: %v", err)
	}

	if importResp.Address != address {
		t.Fatalf("Address mismatch: expected %s, got %s", address, importResp.Address)
	}

	// Verify the account exists in the new repo
	accounts, err := api2.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(accounts))
	}
}

func TestKeystoreAPI_Delete(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Create an account
	createReq := CreateAccountRequest{
		Password: "test-password",
	}

	createResp, err := api.Create(ctx, createReq)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	address := createResp.Address

	// Verify account exists
	accounts, err := api.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("Expected 1 account, got %d", len(accounts))
	}

	// Delete the account
	deleteReq := DeleteAccountRequest{
		Address:  address,
		Password: "test-password",
	}

	err = api.Delete(ctx, deleteReq)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify account is deleted
	accounts, err = api.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(accounts) != 0 {
		t.Fatalf("Expected 0 accounts, got %d", len(accounts))
	}
}

func TestKeystoreAPI_GetPrivateKey(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Test private key
	testPrivateKey := "0x4c0883a69102937d6231471b5dbb6204fe512961708279f8c5b0c6e9a6d74f12"

	// Import the private key
	importReq := ImportPrivateKeyRequest{
		PrivateKeyHex: testPrivateKey,
		Password:      "test-password",
	}

	importResp, err := api.ImportPrivateKey(ctx, importReq)
	if err != nil {
		t.Fatalf("ImportPrivateKey failed: %v", err)
	}

	address := importResp.Address

	// Get the private key back
	exportReq := ExportAccountRequest{
		Address:  address,
		Password: "test-password",
	}

	privateKeyHex, err := api.GetPrivateKey(ctx, exportReq)
	if err != nil {
		t.Fatalf("GetPrivateKey failed: %v", err)
	}

	if privateKeyHex != testPrivateKey {
		t.Fatalf("Private key mismatch: expected %s, got %s", testPrivateKey, privateKeyHex)
	}
}

func TestKeystoreAPI_SignMessage(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Create an account
	createReq := CreateAccountRequest{
		Password: "test-password",
	}

	createResp, err := api.Create(ctx, createReq)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	address := createResp.Address

	// Sign a message
	message := "Hello, World!"
	signature, err := api.SignMessage(ctx, address, message, "test-password")
	if err != nil {
		t.Fatalf("SignMessage failed: %v", err)
	}

	if signature == "" {
		t.Fatal("Expected non-empty signature")
	}

	if len(signature) < 2 || signature[:2] != "0x" {
		t.Fatal("Expected signature to start with 0x")
	}

	t.Logf("Signature: %s", signature)
}

func TestKeystoreAPI_BackupAccount(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Create an account
	createReq := CreateAccountRequest{
		Password: "test-password",
	}

	createResp, err := api.Create(ctx, createReq)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	address := createResp.Address

	// Backup the account
	backupJSON, err := api.BackupAccount(ctx, address)
	if err != nil {
		t.Fatalf("BackupAccount failed: %v", err)
	}

	if backupJSON == "" {
		t.Fatal("Expected non-empty backup JSON")
	}

	t.Logf("Backup JSON length: %d bytes", len(backupJSON))
}

func TestKeystoreAPI_Validate(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Validate the keystore
	info, err := api.Validate(ctx)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if info["keystoreDir"] == "" {
		t.Fatal("Expected non-empty keystoreDir")
	}

	if info["accountCount"] != 0 {
		t.Fatalf("Expected 0 accounts initially, got %v", info["accountCount"])
	}

	t.Logf("Keystore validation: %+v", info)
}

func TestKeystoreAPI_ErrorHandling(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	t.Run("CreateWithEmptyPassword", func(t *testing.T) {
		createReq := CreateAccountRequest{
			Password: "",
		}
		_, err := api.Create(ctx, createReq)
		if err == nil {
			t.Fatal("Expected error with empty password")
		}
	})

	t.Run("ExportNonExistentAccount", func(t *testing.T) {
		exportReq := ExportAccountRequest{
			Address:  "0x0000000000000000000000000000000000000000",
			Password: "test-password",
		}
		_, err := api.Export(ctx, exportReq)
		if err == nil {
			t.Fatal("Expected error for non-existent account")
		}
	})

	t.Run("ImportInvalidPrivateKey", func(t *testing.T) {
		importReq := ImportPrivateKeyRequest{
			PrivateKeyHex: "invalid-key",
			Password:      "test-password",
		}
		_, err := api.ImportPrivateKey(ctx, importReq)
		if err == nil {
			t.Fatal("Expected error for invalid private key")
		}
	})

	t.Run("UpdatePasswordWithWrongOldPassword", func(t *testing.T) {
		// Create an account
		createReq := CreateAccountRequest{
			Password: "correct-password",
		}
		createResp, err := api.Create(ctx, createReq)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		// Try to update with wrong old password
		updateReq := UpdatePasswordRequest{
			Address:     createResp.Address,
			OldPassword: "wrong-password",
			NewPassword: "new-password",
		}
		err = api.UpdatePassword(ctx, updateReq)
		if err == nil {
			t.Fatal("Expected error with wrong old password")
		}
	})
}

func TestKeystoreAPI_MultipleAccounts(t *testing.T) {
	repo := mockRepo(t)
	api := NewKeystoreAPI(repo)
	ctx := context.Background()

	// Create multiple accounts
	numAccounts := 5
	addresses := make([]string, numAccounts)

	for i := 0; i < numAccounts; i++ {
		createReq := CreateAccountRequest{
			Password: "test-password",
		}
		createResp, err := api.Create(ctx, createReq)
		if err != nil {
			t.Fatalf("Create account %d failed: %v", i, err)
		}
		addresses[i] = createResp.Address
	}

	// List all accounts
	accounts, err := api.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(accounts) != numAccounts {
		t.Fatalf("Expected %d accounts, got %d", numAccounts, len(accounts))
	}

	// Verify all addresses are present
	addressSet := make(map[string]bool)
	for _, acc := range accounts {
		addressSet[acc.Address] = true
	}

	for _, addr := range addresses {
		if !addressSet[addr] {
			t.Fatalf("Address %s not found in list", addr)
		}
	}

	t.Logf("Successfully created and listed %d accounts", numAccounts)
}

// BenchmarkKeystoreAPI_Create benchmarks account creation
func BenchmarkKeystoreAPI_Create(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "keystore-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keystoreDir := filepath.Join(tmpDir, "keystore")
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		b.Fatalf("Failed to create keystore dir: %v", err)
	}

	mockRepo := &mockRepoImpl{
		path: tmpDir,
		cfg:  nil,
	}

	api := NewKeystoreAPI(mockRepo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createReq := CreateAccountRequest{
			Password: "test-password",
		}
		_, err := api.Create(ctx, createReq)
		if err != nil {
			b.Fatalf("Create failed: %v", err)
		}
	}
}

// BenchmarkKeystoreAPI_SignMessage benchmarks message signing
func BenchmarkKeystoreAPI_SignMessage(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "keystore-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	keystoreDir := filepath.Join(tmpDir, "keystore")
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		b.Fatalf("Failed to create keystore dir: %v", err)
	}

	mockRepo := &mockRepoImpl{
		path: tmpDir,
		cfg:  nil,
	}

	api := NewKeystoreAPI(mockRepo)
	ctx := context.Background()

	// Create an account
	createReq := CreateAccountRequest{
		Password: "test-password",
	}
	createResp, err := api.Create(ctx, createReq)
	if err != nil {
		b.Fatalf("Create failed: %v", err)
	}

	address := createResp.Address
	message := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := api.SignMessage(ctx, address, message, "test-password")
		if err != nil {
			b.Fatalf("SignMessage failed: %v", err)
		}
	}
}
