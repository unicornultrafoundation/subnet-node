package verifier

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPerformPoW(t *testing.T) {
	pow := &Pow{}
	data := "test"
	peerID := uuid.New().String() // Generate test peerID using UUID
	difficulty := 2

	hash, nonce := pow.performPoW(data, peerID, difficulty)

	assert.NotEmpty(t, hash, "Hash should not be empty")
	assert.GreaterOrEqual(t, nonce, 0, "Nonce should be greater than or equal to 0")
	assert.True(t, pow.verifyPoW(data, peerID, nonce, hash, difficulty), "PoW verification should pass")
}

func TestVerifyPoW(t *testing.T) {
	pow := &Pow{}
	data := uuid.New().String()   // Generate test data using UUID
	peerID := uuid.New().String() // Generate test peerID using UUID
	difficulty := 2

	hash, nonce := pow.performPoW(data, peerID, difficulty)

	assert.True(t, pow.verifyPoW(data, peerID, nonce, hash, difficulty), "PoW verification should pass")

	// Test with incorrect nonce
	assert.False(t, pow.verifyPoW(data, peerID, nonce+1, hash, difficulty), "PoW verification should fail with incorrect nonce")

	// Test with incorrect hash
	assert.False(t, pow.verifyPoW(data, peerID, nonce, "incorrecthash", difficulty), "PoW verification should fail with incorrect hash")
}
