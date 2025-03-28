package coreapi

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface"
	caopts "github.com/unicornultrafoundation/subnet-node/core/coreiface/options"
	"github.com/unicornultrafoundation/subnet-node/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type PubSubAPI CoreAPI

type pubSubSubscription struct {
	subscription *pubsub.Subscription
}

type pubSubMessage struct {
	msg *pubsub.Message
}

func (api *PubSubAPI) Ls(ctx context.Context) ([]string, error) {
	_, span := tracing.Span(ctx, "CoreAPI.PubSubAPI", "Ls")
	defer span.End()

	return api.pubSub.GetTopics(), nil
}

func (api *PubSubAPI) Peers(ctx context.Context, opts ...caopts.PubSubPeersOption) ([]peer.ID, error) {
	_, span := tracing.Span(ctx, "CoreAPI.PubSubAPI", "Peers")
	defer span.End()

	settings, err := caopts.PubSubPeersOptions(opts...)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.String("topic", settings.Topic))

	return api.pubSub.ListPeers(settings.Topic), nil
}

func (api *PubSubAPI) Publish(ctx context.Context, topic string, data []byte) error {
	_, span := tracing.Span(ctx, "CoreAPI.PubSubAPI", "Publish", trace.WithAttributes(attribute.String("topic", topic)))
	defer span.End()

	ptopic, err := api.pubSub.Join(topic)
	if err != nil {
		return err
	}
	return ptopic.Publish(ctx, data)
}

func (api *PubSubAPI) Subscribe(ctx context.Context, topic string, opts ...caopts.PubSubSubscribeOption) (coreiface.PubSubSubscription, error) {
	_, span := tracing.Span(ctx, "CoreAPI.PubSubAPI", "Subscribe", trace.WithAttributes(attribute.String("topic", topic)))
	defer span.End()

	// Parse the options to avoid introducing silent failures for invalid
	// options. However, we don't currently have any use for them. The only
	// subscription option, discovery, is now a no-op as it's handled by
	// pubsub itself.
	_, err := caopts.PubSubSubscribeOptions(opts...)
	if err != nil {
		return nil, err
	}

	//nolint deprecated
	ptopic, err := api.pubSub.Join(topic)
	if err != nil {
		return nil, err
	}
	sub, err := ptopic.Subscribe()
	if err != nil {
		return nil, err
	}
	return &pubSubSubscription{sub}, nil
}

func (sub *pubSubSubscription) Close() error {
	sub.subscription.Cancel()
	return nil
}

func (sub *pubSubSubscription) Next(ctx context.Context) (coreiface.PubSubMessage, error) {
	ctx, span := tracing.Span(ctx, "CoreAPI.PubSubSubscription", "Next")
	defer span.End()

	msg, err := sub.subscription.Next(ctx)
	if err != nil {
		return nil, err
	}

	return &pubSubMessage{msg}, nil
}

func (msg *pubSubMessage) From() peer.ID {
	return peer.ID(msg.msg.From)
}

func (msg *pubSubMessage) Data() []byte {
	return msg.msg.Data
}

func (msg *pubSubMessage) Seq() []byte {
	return msg.msg.Seqno
}

func (msg *pubSubMessage) Topics() []string {
	// TODO: handle breaking downstream changes by returning a single string.
	if msg.msg.Topic == nil {
		return nil
	}
	return []string{*msg.msg.Topic}
}
