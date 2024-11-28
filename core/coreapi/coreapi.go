package coreapi

import (
	"context"
	"errors"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	record "github.com/libp2p/go-libp2p-record"
	ci "github.com/libp2p/go-libp2p/core/crypto"
	p2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	pstore "github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/unicornultrafoundation/subnet-node/core"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface/options"
	"github.com/unicornultrafoundation/subnet-node/core/node/resource"
	"github.com/unicornultrafoundation/subnet-node/repo"
)

type CoreAPI struct {
	nctx context.Context

	identity   peer.ID
	privateKey ci.PrivKey

	peerstore       pstore.Peerstore
	peerHost        p2phost.Host
	recordValidator record.Validator

	repo     repo.Repo
	routing  routing.Routing
	pubSub   *pubsub.PubSub
	resource *resource.Service

	checkOnline func(allowOffline bool) error

	// ONLY for re-applying options in WithOptions, DO NOT USE ANYWHERE ELSE
	nd         *core.SubnetNode
	parentOpts options.ApiSettings
}

func NewCoreAPI(n *core.SubnetNode, opts ...options.ApiOption) (coreiface.CoreAPI, error) {
	parentOpts, err := options.ApiOptions()
	if err != nil {
		return nil, err
	}

	return (&CoreAPI{nd: n, parentOpts: *parentOpts}).WithOptions(opts...)
}

func (api *CoreAPI) PubSub() coreiface.PubSubAPI {
	return (*PubSubAPI)(api)
}

func (api *CoreAPI) Identity() peer.ID {
	return api.identity
}

func (api *CoreAPI) Resource() *resource.Service {
	return api.resource
}

func (api *CoreAPI) Routing() coreiface.RoutingAPI {
	return (*RoutingAPI)(api)
}

func (api *CoreAPI) Swarm() coreiface.SwarmAPI {
	return (*SwarmAPI)(api)
}

func (api *CoreAPI) WithOptions(opts ...options.ApiOption) (coreiface.CoreAPI, error) {
	settings := api.parentOpts // make sure to copy
	_, err := options.ApiOptionsTo(&settings, opts...)
	if err != nil {
		return nil, err
	}

	if api.nd == nil {
		return nil, errors.New("cannot apply options to api without node")
	}

	n := api.nd

	subAPI := &CoreAPI{
		nctx: n.Context(),

		identity:   n.Identity,
		privateKey: n.PrivateKey,

		repo: n.Repo,

		peerstore:       n.Peerstore,
		peerHost:        n.PeerHost,
		recordValidator: n.RecordValidator,
		routing:         n.Routing,
		resource:        n.Resource,
		pubSub:          n.PubSub,
	}

	subAPI.checkOnline = func(allowOffline bool) error {
		if !n.IsOnline && !allowOffline {
			return coreiface.ErrOffline
		}
		return nil
	}

	return subAPI, nil
}
