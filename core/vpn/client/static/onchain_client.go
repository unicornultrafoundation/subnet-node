package static

import (
	"context"
	"fmt"
	"math/big"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/unicornultrafoundation/subnet-node/core/ethereum"
	"github.com/unicornultrafoundation/subnet-node/core/vpn/utils"
)

type OnchainClient struct {
	ethService *ethereum.EthereumService
	peerID     string
}

var _ StaticIPClient = (*OnchainClient)(nil)

func NewOnchainClient(eth *ethereum.EthereumService, host host.Host) StaticIPClient {
	return &OnchainClient{ethService: eth, peerID: host.ID().String()}
}

func (c *OnchainClient) GetPeerID(ctx context.Context, ip string) (string, error) {
	tokenID := utils.ConvertVirtualIPToNumber(ip)
	if tokenID == 0 {
		return "", fmt.Errorf("invalid IP: %s", ip)
	}

	return c.ethService.IPRegistry().GetPeer(nil, big.NewInt(int64(tokenID)))
}

func (c *OnchainClient) IsIPOwnedByNode(ctx context.Context, ip string) (bool, error) {
	peerID, err := c.GetPeerID(ctx, ip)
	if err != nil {
		return false, err
	}
	return peerID != "" && peerID == c.peerID, nil
}
