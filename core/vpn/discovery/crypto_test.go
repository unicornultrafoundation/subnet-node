package discovery

import (
	"crypto/sha256"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	subnet_vpn "github.com/unicornultrafoundation/subnet-node/proto/subnet/vpn"
)

// TestVirtualIPHashing tests all hash-related functionality
func TestVirtualIPHashing(t *testing.T) {
	// Test 1: Basic SHA-256 hash functionality
	data := []byte("test data")
	hash := sha256.Sum256(data)

	// Verify results
	assert.NotNil(t, hash[:])
	assert.Equal(t, 32, len(hash[:])) // SHA-256 hash is 32 bytes

	// Test 2: createVirtualIPHash method
	// Create a test info
	info := &subnet_vpn.VirtualIPInfo{
		Ip:        "10.1.2.3",
		Timestamp: 123456789,
	}

	// Create the peer discovery service
	peerDiscovery := &PeerDiscovery{}

	// Test creating a hash
	vipHash := peerDiscovery.createVirtualIPHash(info)

	// Verify results
	assert.NotNil(t, vipHash)
	assert.Equal(t, 32, len(vipHash)) // SHA-256 hash is 32 bytes

	// Test 3: Hash changes with different input
	info2 := &subnet_vpn.VirtualIPInfo{
		Ip:        "10.1.2.4", // Different IP
		Timestamp: 123456789,
	}
	hash2 := peerDiscovery.createVirtualIPHash(info2)

	// Verify results
	assert.NotNil(t, hash2)
	assert.NotEqual(t, vipHash, hash2) // Hashes should be different
}

// TestSigningAndRecordCreation tests the signing and record creation functionality
func TestSigningAndRecordCreation(t *testing.T) {

	// Generate a test key pair for signing tests
	privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	assert.NoError(t, err)

	// Create a mock peerstore service
	mockPeerstoreService := new(MockPeerstoreService)
	mockPeerID := peer.ID("peer1")
	mockPeerstoreService.On("PrivKey", mockPeerID).Return(privKey)

	// Create a mock host service
	mockHostService := new(MockHostService)
	mockHostService.On("ID").Return(mockPeerID)
	mockHostService.On("Peerstore").Return(mockPeerstoreService)

	// Create the peer discovery service
	peerDiscovery := &PeerDiscovery{
		host:      mockHostService,
		virtualIP: "10.1.2.3",
	}

	// Test signing a hash
	signatureHash := []byte("test hash")
	signature, err := peerDiscovery.signVirtualIPHash(signatureHash)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, signature)

	// Verify the signature is valid
	valid, err := pubKey.Verify(signatureHash, signature)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test creating a signed record
	record, err := peerDiscovery.createSignedIPRecord("10.1.2.3")

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, "10.1.2.3", record.Info.Ip)
	assert.NotNil(t, record.Signature)

	// Verify the record can be marshaled
	recordData, err := proto.Marshal(record)
	assert.NoError(t, err)
	assert.NotNil(t, recordData)

	// Test error case - no private key
	// Create a new mock with no private key
	mockPeerstoreService2 := new(MockPeerstoreService)
	var nilPrivKey crypto.PrivKey
	mockPeerstoreService2.On("PrivKey", mockPeerID).Return(nilPrivKey)

	mockHostService2 := new(MockHostService)
	mockHostService2.On("ID").Return(mockPeerID)
	mockHostService2.On("Peerstore").Return(mockPeerstoreService2)

	peerDiscovery2 := &PeerDiscovery{
		host: mockHostService2,
	}

	// Test signing with no private key
	signature, err = peerDiscovery2.signVirtualIPHash(signatureHash)

	// Verify results
	assert.Error(t, err)
	assert.Nil(t, signature)
	assert.Contains(t, err.Error(), "private key not found")

	// Verify all expectations were met
	mockHostService.AssertExpectations(t)
	mockPeerstoreService.AssertExpectations(t)
	mockHostService2.AssertExpectations(t)
	mockPeerstoreService2.AssertExpectations(t)
}
