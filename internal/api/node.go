package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"

	ic "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
)

type NodeAPI struct {
	resource *resource.Service
	node     *core.SubnetNode
	app      *apps.Service
}

// NewNodeAPI creates a new instance of NodeAPI.
func NewNodeAPI(resource *resource.Service, appS *apps.Service, node *core.SubnetNode) *NodeAPI {
	return &NodeAPI{resource: resource, node: node, app: appS}
}

func (api *NodeAPI) GetResource(ctx context.Context) (*resource.ResourceInfo, error) {
	if api.resource == nil {
		return nil, fmt.Errorf("api.resource is not supported")
	}
	return api.resource.GetResource()
}

func (api *NodeAPI) Restart(ctx context.Context) error {
	return api.node.Close()
}

type IdOutput struct { // nolint
	ID           string
	PublicKey    string
	Addresses    []string
	AgentVersion string
	Protocols    []protocol.ID
}

func (api *NodeAPI) Id(ctx context.Context) (*IdOutput, error) {
	p := api.resource.PeerId()
	info := new(IdOutput)
	info.ID = p.String()
	ps := api.node.Peerstore
	if pk := ps.PubKey(p); pk != nil {
		pkb, err := ic.MarshalPublicKey(pk)
		if err != nil {
			return nil, err
		}
		info.PublicKey = base64.StdEncoding.EncodeToString(pkb)
	}

	addrInfo := ps.PeerInfo(p)
	addrs, err := peer.AddrInfoToP2pAddrs(&addrInfo)
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		info.Addresses = append(info.Addresses, a.String())
	}
	sort.Strings(info.Addresses)

	protocols, _ := ps.GetProtocols(p) // don't care about errors here.
	info.Protocols = append(info.Protocols, protocols...)
	sort.Slice(info.Protocols, func(i, j int) bool { return info.Protocols[i] < info.Protocols[j] })

	if v, err := ps.Get(p, "AgentVersion"); err == nil {
		if vs, ok := v.(string); ok {
			info.AgentVersion = vs
		}
	}

	return info, nil
}
