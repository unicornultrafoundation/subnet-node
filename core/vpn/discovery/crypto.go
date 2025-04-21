package discovery

import (
	"crypto/sha256"
	"fmt"

	"github.com/gogo/protobuf/proto"
	subnet_vpn "github.com/unicornultrafoundation/subnet-node/proto/subnet/vpn"
)

// createVirtualIPHash generates a hash of the timestamp and IP address
func (p *PeerDiscovery) createVirtualIPHash(info *subnet_vpn.VirtualIPInfo) []byte {
	// Marshal the data to protobuf for the hash
	data, err := proto.Marshal(info)
	if err != nil {
		log.Errorf("Failed to marshal IP info for IP %s: %v", info.Ip, err)
		return nil
	}

	hash := sha256.Sum256(data)
	return hash[:]
}

// signVirtualIPHash signs the hash using the node's private key
func (p *PeerDiscovery) signVirtualIPHash(hash []byte) ([]byte, error) {
	peerID := p.host.ID()

	privKey := p.host.Peerstore().PrivKey(peerID)
	if privKey == nil {
		return nil, fmt.Errorf("private key not found for peer ID %s", peerID)
	}

	signature, err := privKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash for peer ID %s: %w", peerID, err)
	}

	return signature, nil
}
