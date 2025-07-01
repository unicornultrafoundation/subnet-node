# AuthChain Implementation

A comprehensive authentication chain implementation for secure, temporary key delegation in distributed systems.

## Overview

AuthChain provides a three-link authentication chain that enables secure temporary key delegation while maintaining cryptographic proof of ownership. This implementation follows a validator chain pattern similar to JavaScript interfaces, providing a flexible and extensible authentication system.

## Architecture

The AuthChain consists of three cryptographic links that form a complete authentication chain:

1. **SIGNER** - Owner identity verification with personal message signing
2. **ECDSA_PERSONAL_EPHEMERAL** - Ephemeral key delegation with timestamp
3. **ECDSA_PERSONAL_SIGNED_ENTITY** - Entity-specific authorization

### Chain Structure

```
Owner Private Key → Ephemeral Private Key → Entity Authorization
       ↓                    ↓                        ↓
   [SIGNER]    [ECDSA_PERSONAL_EPHEMERAL]  [ECDSA_PERSONAL_SIGNED_ENTITY]
```

### Message Formats

Each link in the chain uses specific message formats:

- **SIGNER**: `"I am the owner of this account"`
- **ECDSA_PERSONAL_EPHEMERAL**: `"I authorize this ephemeral key until {timestamp}"`
- **ECDSA_PERSONAL_SIGNED_ENTITY**: `"I authorize access to {entityID}"`

## Features

- **Secure Key Delegation**: Temporary ephemeral keys for improved security
- **Entity-Specific Authorization**: Granular access control per entity
- **Time-Based Expiry**: Automatic expiration of delegated access
- **Cryptographic Verification**: Full signature validation chain
- **Validator Chain Pattern**: Extensible validation system
- **Comprehensive Utilities**: Helper functions for common operations
- **Type Safety**: Strong typing with Go interfaces

## Installation

```bash
go get github.com/your-repo/authchain
```

## Quick Start

### Creating an AuthChain

```go
package main

import (
    "crypto/ecdsa"
    "crypto/rand"
    "time"
    "fmt"
    
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/your-repo/authchain"
)

func main() {
    // Generate or load your private key
    privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
    if err != nil {
        panic(err)
    }

    // Create authchain service
    service := authchain.NewDefaultService()

    // Build an auth chain for a specific entity
    entityID := "order-123"
    expiry := 1 * time.Hour
    
    authChain, err := service.Build(privateKey, entityID, expiry)
    if err != nil {
        panic(err)
    }

    // Serialize for transmission
    jsonData, err := authchain.SerializeAuthChain(*authChain)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("AuthChain:", jsonData)
}
```

### Validating an AuthChain

```go
func validateAuthChain(jsonData, entityID string) error {
    // Deserialize the auth chain
    authChain, err := authchain.DeserializeAuthChain(jsonData)
    if err != nil {
        return err
    }

    // Create service and validate
    service := authchain.NewDefaultService()
    ctx := context.Background()
    
    authCtx, err := service.Validate(ctx, authChain, entityID)
    if err != nil {
        return err
    }

    fmt.Printf("Validated for user: %s\n", authCtx.UserAddress.Hex())
    fmt.Printf("Expires at: %s\n", authCtx.ExpiresAt)
    
    return nil
}
```

## API Reference

### Core Types

#### AuthChain
```go
type AuthChain []AuthLink
```

#### AuthLink
```go
type AuthLink struct {
    Type      AuthLinkType `json:"type"`
    Payload   string       `json:"payload"`
    Signature string       `json:"signature"`
}
```

#### AuthLinkType
```go
type AuthLinkType string

const (
    AuthLinkTypeSigner                    AuthLinkType = "SIGNER"
    AuthLinkTypeEcdsaPersonalEphemeral    AuthLinkType = "ECDSA_PERSONAL_EPHEMERAL"
    AuthLinkTypeEcdsaPersonalSignedEntity AuthLinkType = "ECDSA_PERSONAL_SIGNED_ENTITY"
)
```

#### Context
```go
type Context struct {
    UserAddress    common.Address
    EphemeralAddr  common.Address
    EntityID       string
    ExpiresAt      time.Time
    SessionID      string
    CreatedAt      time.Time
    ChainValidated bool
}
```

### Service Interface

#### NewDefaultService()
Creates a new authchain service with default validator and builder implementations.

#### Build(privateKey, entityID, expiry)
Creates a new auth chain for the given parameters.

#### Validate(ctx, authChain, entityID)
Validates an auth chain and returns authentication context.

### Validator Interface

The validator system follows a chain pattern:

```go
type Validator interface {
    ValidateAuthChain(ctx context.Context, authChain AuthChain, entityID string) (*Context, error)
    SetNext(next Validator)
}
```

### Builder Interface

```go
type Builder interface {
    BuildAuthChain(privateKey *ecdsa.PrivateKey, entityID string, expiry time.Duration) (*AuthChain, error)
}
```

### Utility Functions

#### Serialization
- `SerializeAuthChain(authChain)` - Convert to JSON string
- `DeserializeAuthChain(data)` - Parse from JSON string

#### Information Extraction
- `GetOwnerAddress(authChain)` - Extract owner address
- `GetEphemeralAddress(authChain)` - Extract ephemeral address
- `GetEntityID(authChain)` - Extract entity ID
- `GetExpiryTime(authChain)` - Extract expiry time
- `IsExpired(authChain)` - Check if expired
- `GetTimeUntilExpiry(authChain)` - Get remaining time

#### Validation
- `ValidateAuthChainFormat(authChain)` - Basic format validation
- `ValidateAuthChainWithService(authChain, entityID)` - Full validation

#### Convenience Functions
- `CreateAuthChainFromPrivateKey(privateKeyHex, entityID, expiry)` - Create from hex key
- `CreateAuthChainForOrder(privateKey, orderID, expiry)` - Create for orders
- `ValidateAuthChainForOrder(authChain, orderID)` - Validate for orders

## Advanced Usage

### Custom Validator Chain

```go
type CustomValidator struct {
    authchain.DefaultValidator
}

func (v *CustomValidator) ValidateAuthChain(ctx context.Context, authChain authchain.AuthChain, entityID string) (*authchain.Context, error) {
    // Add custom validation logic
    if err := v.validateCustomRules(authChain); err != nil {
        return nil, err
    }
    
    // Call next validator in chain
    if v.Next != nil {
        return v.Next.ValidateAuthChain(ctx, authChain, entityID)
    }
    
    return nil, fmt.Errorf("no next validator in chain")
}

func (v *CustomValidator) validateCustomRules(authChain authchain.AuthChain) error {
    // Implement custom validation rules
    return nil
}
```

### Custom Builder

```go
type CustomBuilder struct {
    authchain.DefaultBuilder
}

func (b *CustomBuilder) BuildAuthChain(privateKey *ecdsa.PrivateKey, entityID string, expiry time.Duration) (*authchain.AuthChain, error) {
    // Add custom logic before building
    if err := b.validateEntityID(entityID); err != nil {
        return nil, err
    }
    
    // Call default builder
    return b.DefaultBuilder.BuildAuthChain(privateKey, entityID, expiry)
}
```

### Service with Custom Components

```go
// Create custom validator chain
customValidator := &CustomValidator{}
defaultValidator := authchain.NewDefaultValidator()
customValidator.SetNext(defaultValidator)

// Create custom builder
customBuilder := &CustomBuilder{}

// Create service with custom components
service := authchain.NewService(customValidator, customBuilder)
```

### Message Format Customization

The authchain uses specific message formats that can be customized:

```go
// Default messages
const (
    SignerMessage = "I am the owner of this account"
    EphemeralMessageTemplate = "I authorize this ephemeral key until %d"
    EntityMessageTemplate = "I authorize access to %s"
)
```

## Security Considerations

1. **Private Key Security**: Never expose private keys in logs or error messages
2. **Ephemeral Key Management**: Ephemeral keys should be securely generated and stored
3. **Expiry Validation**: Always validate expiry times on the server side
4. **Signature Verification**: Verify all signatures in the chain
5. **Entity ID Validation**: Ensure entity IDs match expected values
6. **Rate Limiting**: Implement rate limiting for auth chain creation and validation
7. **Message Integrity**: Ensure message formats are consistent across all implementations

## Testing

Run the test suite:

```bash
go test ./authchain
```

Run with coverage:

```bash
go test -cover ./authchain
```

Run specific tests:

```bash
go test -v ./authchain -run TestAuthChain
```

## Examples

The test files provide comprehensive examples of:

- Basic auth chain creation and validation
- Error handling scenarios
- Custom validator and builder implementations
- Utility function usage
- Message format handling
- Timestamp parsing and validation

### Key Test Files

- `authchain_test.go` - Core functionality tests
- `utils_test.go` - Utility function tests

## Error Handling

The authchain provides detailed error types:

- `ErrInvalidAuthChain` - Invalid chain structure
- `ErrInvalidSignature` - Signature verification failed
- `ErrExpiredAuthChain` - Chain has expired
- `ErrInvalidEntityID` - Entity ID mismatch
- `ErrInvalidTimestamp` - Invalid timestamp format

## Performance Considerations

- AuthChain validation is designed to be fast and efficient
- Ephemeral key generation uses secure random sources
- Signature verification leverages optimized cryptographic libraries
- JSON serialization is optimized for network transmission

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Update documentation as needed
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 