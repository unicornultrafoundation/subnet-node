package discovery

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/api"
)

// HostAdapter implements the api.HostService interface using a libp2p host
type HostAdapter struct {
	host host.Host
}

// NewHostAdapter creates a new host adapter
func NewHostAdapter(host host.Host) *HostAdapter {
	return &HostAdapter{host: host}
}

// ID returns the peer ID of the host
func (a *HostAdapter) ID() peer.ID {
	return a.host.ID()
}

// Peerstore returns the peerstore of the host
func (a *HostAdapter) Peerstore() api.PeerstoreService {
	return &PeerstoreAdapter{privKeyFunc: a.host.Peerstore().PrivKey}
}

// PeerstoreAdapter implements the api.PeerstoreService interface
type PeerstoreAdapter struct {
	privKeyFunc func(peer.ID) crypto.PrivKey
}

// PrivKey returns the private key for a peer ID
func (a *PeerstoreAdapter) PrivKey(p peer.ID) crypto.PrivKey {
	return a.privKeyFunc(p)
}

// DHTAdapter implements the api.DHTService interface using libp2p routing
type DHTAdapter struct {
	routing routing.ValueStore
}

// NewDHTAdapter creates a new DHT adapter
func NewDHTAdapter(routing routing.ValueStore) *DHTAdapter {
	return &DHTAdapter{routing: routing}
}

// GetValue gets a value from the DHT
func (a *DHTAdapter) GetValue(ctx context.Context, key string) ([]byte, error) {
	return a.routing.GetValue(ctx, key)
}

// PutValue puts a value in the DHT
func (a *DHTAdapter) PutValue(ctx context.Context, key string, value []byte) error {
	return a.routing.PutValue(ctx, key, value)
}

// AccountAdapter implements the api.AccountService interface
type AccountAdapter struct {
	accountService *account.AccountService
}

// NewAccountAdapter creates a new account adapter
func NewAccountAdapter(accountService *account.AccountService) *AccountAdapter {
	return &AccountAdapter{accountService: accountService}
}

// IPRegistry returns the IP registry contract
func (a *AccountAdapter) IPRegistry() api.IPRegistry {
	return &IPRegistryAdapter{ipRegistry: a.accountService.IPRegistry()}
}

// IPRegistryAdapter adapts an IP registry contract to the api.IPRegistry interface
type IPRegistryAdapter struct {
	ipRegistry account.IPRegistry
}

// GetPeer gets the peer ID for a token ID
func (a *IPRegistryAdapter) GetPeer(opts any, tokenID *big.Int) (string, error) {
	var callOpts *bind.CallOpts
	if opts != nil {
		var ok bool
		callOpts, ok = opts.(*bind.CallOpts)
		if !ok {
			return "", fmt.Errorf("invalid options type: expected *bind.CallOpts")
		}
	}
	return a.ipRegistry.GetPeer(callOpts, tokenID)
}
