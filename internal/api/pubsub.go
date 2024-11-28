package api

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface/options"
)

type PubsubAPI struct {
	pubsub coreiface.PubSubAPI
}

func NewPubsubAPI(pubsub coreiface.PubSubAPI) *PubsubAPI {
	return &PubsubAPI{pubsub: pubsub}
}

func (api *PubsubAPI) Peers(ctx context.Context, topic string) ([]peer.ID, error) {
	return api.pubsub.Peers(ctx, options.PubSub.Topic(topic))
}

func (api *PubsubAPI) Ls(ctx context.Context) ([]string, error) {
	return api.pubsub.Ls(ctx)
}
