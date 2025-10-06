package api

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

// KeystoreAPI provides API for managing Ethereum keystore accounts
type KeystoreAPI struct {
	repo repo.Repo
	ks   *keystore.KeyStore
}

// AccountInfo represents account information returned by the API
type AccountInfo struct {
	Address      string `json:"address"`
	KeystoreFile string `json:"keystoreFile"`
	URL          string `json:"url"`
}

// CreateAccountRequest represents the request to create a new account
type CreateAccountRequest struct {
	Password string `json:"password"`
}

// CreateAccountResponse represents the response after creating an account
type CreateAccountResponse struct {
	Address      string `json:"address"`
	KeystoreFile string `json:"keystoreFile"`
}

// ImportKeystoreRequest represents the request to import a keystore
type ImportKeystoreRequest struct {
	KeystoreJSON string `json:"keystoreJson"`
	Password     string `json:"password"`
	NewPassword  string `json:"newPassword,omitempty"` // Optional, if empty uses same password
}

// ImportPrivateKeyRequest represents the request to import from private key
type ImportPrivateKeyRequest struct {
	PrivateKeyHex string `json:"privateKeyHex"`
	Password      string `json:"password"`
}

// ImportResponse represents the response after importing an account
type ImportResponse struct {
	Address      string `json:"address"`
	KeystoreFile string `json:"keystoreFile"`
}

// UpdatePasswordRequest represents the request to update account password
type UpdatePasswordRequest struct {
	Address     string `json:"address"`
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// ExportAccountRequest represents the request to export account keystore
type ExportAccountRequest struct {
	Address  string `json:"address"`
	Password string `json:"password"`
}

// ExportAccountResponse represents the exported keystore JSON
type ExportAccountResponse struct {
	KeystoreJSON string `json:"keystoreJson"`
}

// DeleteAccountRequest represents the request to delete an account
type DeleteAccountRequest struct {
	Address  string `json:"address"`
	Password string `json:"password"`
}

// NewKeystoreAPI creates a new KeystoreAPI instance
func NewKeystoreAPI(repo repo.Repo) *KeystoreAPI {
	keystoreDir := filepath.Join(repo.Path(), "keystore")

	// Ensure keystore directory exists
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		// Log error but don't fail, will error on actual operations if needed
		fmt.Fprintf(os.Stderr, "Warning: failed to create keystore directory: %v\n", err)
	}

	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	return &KeystoreAPI{
		repo: repo,
		ks:   ks,
	}
}

// List returns all accounts in the keystore
func (api *KeystoreAPI) List(ctx context.Context) ([]AccountInfo, error) {
	accounts := api.ks.Accounts()

	result := make([]AccountInfo, 0, len(accounts))
	keystoreDir := filepath.Join(api.repo.Path(), "keystore")

	for _, acc := range accounts {
		keystorePath := filepath.Join(keystoreDir, filepath.Base(acc.URL.Path))
		result = append(result, AccountInfo{
			Address:      acc.Address.Hex(),
			KeystoreFile: keystorePath,
			URL:          acc.URL.String(),
		})
	}

	return result, nil
}

// Create creates a new account with the given password
func (api *KeystoreAPI) Create(ctx context.Context, req CreateAccountRequest) (*CreateAccountResponse, error) {
	if req.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	account, err := api.ks.NewAccount(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	keystoreDir := filepath.Join(api.repo.Path(), "keystore")
	keystorePath := filepath.Join(keystoreDir, filepath.Base(account.URL.Path))

	return &CreateAccountResponse{
		Address:      account.Address.Hex(),
		KeystoreFile: keystorePath,
	}, nil
}

// ImportKeystore imports a keystore JSON file
func (api *KeystoreAPI) ImportKeystore(ctx context.Context, req ImportKeystoreRequest) (*ImportResponse, error) {
	if req.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if req.KeystoreJSON == "" {
		return nil, fmt.Errorf("keystore JSON cannot be empty")
	}

	// Use the same password for import and encryption if newPassword is not provided
	newPassword := req.NewPassword
	if newPassword == "" {
		newPassword = req.Password
	}

	keyjson := []byte(req.KeystoreJSON)
	account, err := api.ks.Import(keyjson, req.Password, newPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to import keystore: %w", err)
	}

	keystoreDir := filepath.Join(api.repo.Path(), "keystore")
	keystorePath := filepath.Join(keystoreDir, filepath.Base(account.URL.Path))

	return &ImportResponse{
		Address:      account.Address.Hex(),
		KeystoreFile: keystorePath,
	}, nil
}

// ImportPrivateKey imports an account from a private key hex string
func (api *KeystoreAPI) ImportPrivateKey(ctx context.Context, req ImportPrivateKeyRequest) (*ImportResponse, error) {
	if req.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if req.PrivateKeyHex == "" {
		return nil, fmt.Errorf("private key cannot be empty")
	}

	// Decode hex key
	privBytes, err := decodeHexKey(req.PrivateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}

	// Convert to ECDSA private key
	privKey, err := ethcrypto.ToECDSA(privBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Import into keystore
	account, err := api.ks.ImportECDSA(privKey, req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to import private key: %w", err)
	}

	keystoreDir := filepath.Join(api.repo.Path(), "keystore")
	keystorePath := filepath.Join(keystoreDir, filepath.Base(account.URL.Path))

	return &ImportResponse{
		Address:      account.Address.Hex(),
		KeystoreFile: keystorePath,
	}, nil
}

// UpdatePassword changes the password for an account
func (api *KeystoreAPI) UpdatePassword(ctx context.Context, req UpdatePasswordRequest) error {
	if req.OldPassword == "" || req.NewPassword == "" {
		return fmt.Errorf("old and new passwords cannot be empty")
	}

	if req.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// Find the account
	account, err := api.findAccount(req.Address)
	if err != nil {
		return err
	}

	// Update the password
	err = api.ks.Update(account, req.OldPassword, req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Export exports the keystore JSON for an account
func (api *KeystoreAPI) Export(ctx context.Context, req ExportAccountRequest) (*ExportAccountResponse, error) {
	if req.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if req.Address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	// Find the account
	account, err := api.findAccount(req.Address)
	if err != nil {
		return nil, err
	}

	// Read the keystore file
	keystoreDir := filepath.Join(api.repo.Path(), "keystore")
	keystorePath := filepath.Join(keystoreDir, filepath.Base(account.URL.Path))

	keyjson, err := os.ReadFile(keystorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Verify the password is correct by attempting to decrypt
	_, err = keystore.DecryptKey(keyjson, req.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	return &ExportAccountResponse{
		KeystoreJSON: string(keyjson),
	}, nil
}

// Delete removes an account from the keystore
func (api *KeystoreAPI) Delete(ctx context.Context, req DeleteAccountRequest) error {
	if req.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if req.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}

	// Find the account
	account, err := api.findAccount(req.Address)
	if err != nil {
		return err
	}

	// Verify password before deletion
	keystoreDir := filepath.Join(api.repo.Path(), "keystore")
	keystorePath := filepath.Join(keystoreDir, filepath.Base(account.URL.Path))

	keyjson, err := os.ReadFile(keystorePath)
	if err != nil {
		return fmt.Errorf("failed to read keystore file: %w", err)
	}

	_, err = keystore.DecryptKey(keyjson, req.Password)
	if err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	// Delete the account
	err = api.ks.Delete(account, req.Password)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

// GetPrivateKey exports the private key for an account (use with caution)
func (api *KeystoreAPI) GetPrivateKey(ctx context.Context, req ExportAccountRequest) (string, error) {
	if req.Password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if req.Address == "" {
		return "", fmt.Errorf("address cannot be empty")
	}

	// Find the account
	account, err := api.findAccount(req.Address)
	if err != nil {
		return "", err
	}

	// Read and decrypt the keystore file
	keystoreDir := filepath.Join(api.repo.Path(), "keystore")
	keystorePath := filepath.Join(keystoreDir, filepath.Base(account.URL.Path))

	keyjson, err := os.ReadFile(keystorePath)
	if err != nil {
		return "", fmt.Errorf("failed to read keystore file: %w", err)
	}

	key, err := keystore.DecryptKey(keyjson, req.Password)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt key: %w", err)
	}

	privateKeyBytes := ethcrypto.FromECDSA(key.PrivateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	return "0x" + privateKeyHex, nil
}

// findAccount finds an account by address
func (api *KeystoreAPI) findAccount(address string) (accounts.Account, error) {
	accs := api.ks.Accounts()

	for _, acc := range accs {
		if acc.Address.Hex() == address || acc.Address.Hex() == normalizeAddress(address) {
			return acc, nil
		}
	}

	return accounts.Account{}, fmt.Errorf("account not found: %s", address)
}

// normalizeAddress ensures the address has the 0x prefix
func normalizeAddress(address string) string {
	if len(address) >= 2 && address[:2] == "0x" {
		return address
	}
	return "0x" + address
}

// decodeHexKey decodes a hex-encoded private key
func decodeHexKey(hexkey string) ([]byte, error) {
	if len(hexkey) >= 2 && hexkey[:2] == "0x" {
		hexkey = hexkey[2:]
	}
	return hex.DecodeString(hexkey)
}

// toECDSA converts bytes to an ECDSA private key
func toECDSA(privBytes []byte) (*ecdsa.PrivateKey, error) {
	return ethcrypto.ToECDSA(privBytes)
}

// Validate validates that the keystore system is working correctly
func (api *KeystoreAPI) Validate(ctx context.Context) (map[string]interface{}, error) {
	keystoreDir := filepath.Join(api.repo.Path(), "keystore")

	// Check if keystore directory exists
	info, err := os.Stat(keystoreDir)
	if err != nil {
		return nil, fmt.Errorf("keystore directory error: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("keystore path is not a directory")
	}

	// Check permissions
	mode := info.Mode()

	// Count accounts
	accounts := api.ks.Accounts()

	return map[string]interface{}{
		"keystoreDir":  keystoreDir,
		"permissions":  mode.String(),
		"accountCount": len(accounts),
		"wallets":      len(api.ks.Wallets()),
	}, nil
}

// SignMessage signs a message with the account's private key
func (api *KeystoreAPI) SignMessage(ctx context.Context, address string, message string, password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if address == "" {
		return "", fmt.Errorf("address cannot be empty")
	}

	if message == "" {
		return "", fmt.Errorf("message cannot be empty")
	}

	// Find the account
	account, err := api.findAccount(address)
	if err != nil {
		return "", err
	}

	// Read and decrypt the keystore file
	keystoreDir := filepath.Join(api.repo.Path(), "keystore")
	keystorePath := filepath.Join(keystoreDir, filepath.Base(account.URL.Path))

	keyjson, err := os.ReadFile(keystorePath)
	if err != nil {
		return "", fmt.Errorf("failed to read keystore file: %w", err)
	}

	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt key: %w", err)
	}

	// Hash the message
	messageHash := ethcrypto.Keccak256Hash([]byte(message))

	// Sign the hash
	signature, err := ethcrypto.Sign(messageHash.Bytes(), key.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	return "0x" + hex.EncodeToString(signature), nil
}

// BackupAccount creates a backup of the keystore file
func (api *KeystoreAPI) BackupAccount(ctx context.Context, address string) (string, error) {
	if address == "" {
		return "", fmt.Errorf("address cannot be empty")
	}

	// Find the account
	account, err := api.findAccount(address)
	if err != nil {
		return "", err
	}

	// Read the keystore file
	keystoreDir := filepath.Join(api.repo.Path(), "keystore")
	keystorePath := filepath.Join(keystoreDir, filepath.Base(account.URL.Path))

	keyjson, err := os.ReadFile(keystorePath)
	if err != nil {
		return "", fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Parse to pretty JSON
	var keystoreData map[string]interface{}
	if err := json.Unmarshal(keyjson, &keystoreData); err != nil {
		return "", fmt.Errorf("failed to parse keystore JSON: %w", err)
	}

	prettyJSON, err := json.MarshalIndent(keystoreData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}

	return string(prettyJSON), nil
}
