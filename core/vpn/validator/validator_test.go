package validator

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	subnet_vpn "github.com/unicornultrafoundation/subnet-node/proto/subnet/vpn"
)

func TestVPNValidator_Validate(t *testing.T) {
	// Create a test key pair
	privKey, pubKey, err := crypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)

	// Get the peer ID from the public key
	peerID, err := peer.IDFromPublicKey(pubKey)
	require.NoError(t, err)

	// Create a valid record
	validRecord := createTestRecord(t, privKey)
	validValue, err := proto.Marshal(validRecord)
	require.NoError(t, err)

	// Create an invalid record (missing IP)
	invalidRecord := createTestRecord(t, privKey)
	invalidRecord.Info.Ip = ""
	invalidValue, err := proto.Marshal(invalidRecord)
	require.NoError(t, err)

	// Create a record with invalid signature
	badSigRecord := createTestRecord(t, privKey)
	badSigRecord.Signature = []byte("invalid-signature")
	badSigValue, err := proto.Marshal(badSigRecord)
	require.NoError(t, err)

	// Create the validator
	validator := VPNValidator{}

	tests := []struct {
		name    string
		key     string
		value   []byte
		wantErr bool
	}{
		{
			name:    "valid record",
			key:     "/vpn/mapping/" + peerID.String(),
			value:   validValue,
			wantErr: false,
		},
		{
			name:    "invalid key prefix",
			key:     "/invalid/prefix/" + peerID.String(),
			value:   validValue,
			wantErr: true,
		},
		{
			name:    "invalid protobuf",
			key:     "/vpn/mapping/" + peerID.String(),
			value:   []byte("invalid-protobuf"),
			wantErr: true,
		},
		{
			name:    "missing IP",
			key:     "/vpn/mapping/" + peerID.String(),
			value:   invalidValue,
			wantErr: true,
		},
		{
			name:    "invalid signature",
			key:     "/vpn/mapping/" + peerID.String(),
			value:   badSigValue,
			wantErr: true,
		},
		{
			name:    "invalid peer ID",
			key:     "/vpn/mapping/invalid-peer-id",
			value:   validValue,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVPNValidator_Select(t *testing.T) {
	// Create a test key pair
	privKey, pubKey, err := crypto.GenerateEd25519Key(rand.Reader)
	require.NoError(t, err)

	// Get the peer ID from the public key
	peerID, err := peer.IDFromPublicKey(pubKey)
	require.NoError(t, err)

	// Create a valid record
	validRecord := createTestRecord(t, privKey)
	validValue, err := proto.Marshal(validRecord)
	require.NoError(t, err)

	// Create an invalid record
	invalidRecord := createTestRecord(t, privKey)
	invalidRecord.Info.Ip = ""
	invalidValue, err := proto.Marshal(invalidRecord)
	require.NoError(t, err)

	// Create the validator
	validator := VPNValidator{}
	key := "/vpn/mapping/" + peerID.String()

	tests := []struct {
		name    string
		key     string
		values  [][]byte
		want    int
		wantErr bool
	}{
		{
			name:    "empty values",
			key:     key,
			values:  [][]byte{},
			want:    -1,
			wantErr: true,
		},
		{
			name:    "one valid value",
			key:     key,
			values:  [][]byte{validValue},
			want:    0,
			wantErr: false,
		},
		{
			name:    "one invalid value",
			key:     key,
			values:  [][]byte{invalidValue},
			want:    -1,
			wantErr: true,
		},
		{
			name:    "invalid then valid",
			key:     key,
			values:  [][]byte{invalidValue, validValue},
			want:    1,
			wantErr: false,
		},
		{
			name:    "valid then invalid",
			key:     key,
			values:  [][]byte{validValue, invalidValue},
			want:    0,
			wantErr: false,
		},
		{
			name:    "all invalid",
			key:     key,
			values:  [][]byte{invalidValue, invalidValue, invalidValue},
			want:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validator.Select(tt.key, tt.values)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func createTestRecord(t *testing.T, privKey crypto.PrivKey) *subnet_vpn.VirtualIPRecord {
	// Create a VirtualIPInfo
	info := &subnet_vpn.VirtualIPInfo{
		Ip:        "10.0.0.1",
		Timestamp: time.Now().Unix(),
	}

	// Marshal the info to protobuf
	data, err := proto.Marshal(info)
	require.NoError(t, err)

	// Create a SHA-256 hash
	hash := sha256.Sum256(data)

	// Sign the hash
	signature, err := privKey.Sign(hash[:])
	require.NoError(t, err)

	// Create the record
	return &subnet_vpn.VirtualIPRecord{
		Info:      info,
		Signature: signature,
	}
}
