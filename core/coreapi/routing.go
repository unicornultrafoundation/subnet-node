package coreapi

import (
	"context"
	"errors"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/core/coreiface"
	caopts "github.com/unicornultrafoundation/subnet-node/core/coreiface/options"
	"github.com/unicornultrafoundation/subnet-node/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type RoutingAPI CoreAPI

func (api *RoutingAPI) Get(ctx context.Context, key string) ([]byte, error) {
	if !api.nd.IsOnline {
		return nil, coreiface.ErrOffline
	}

	dhtKey, err := normalizeKey(key)
	if err != nil {
		return nil, err
	}

	return api.routing.GetValue(ctx, dhtKey)
}

func (api *RoutingAPI) Put(ctx context.Context, key string, value []byte, opts ...caopts.RoutingPutOption) error {
	options, err := caopts.RoutingPutOptions(opts...)
	if err != nil {
		return err
	}

	err = api.checkOnline(options.AllowOffline)
	if err != nil {
		return err
	}

	dhtKey, err := normalizeKey(key)
	if err != nil {
		return err
	}

	return api.routing.PutValue(ctx, dhtKey, value)
}

func normalizeKey(s string) (string, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 3 ||
		parts[0] != "" ||
		!(parts[1] == "ipns" || parts[1] == "pk") {
		return "", errors.New("invalid key")
	}

	k, err := peer.Decode(parts[2])
	if err != nil {
		return "", err
	}
	return strings.Join(append(parts[:2], string(k)), "/"), nil
}

func (api *RoutingAPI) FindPeer(ctx context.Context, p peer.ID) (peer.AddrInfo, error) {
	ctx, span := tracing.Span(ctx, "CoreAPI.DhtAPI", "FindPeer", trace.WithAttributes(attribute.String("peer", p.String())))
	defer span.End()
	err := api.checkOnline(false)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	pi, err := api.routing.FindPeer(ctx, peer.ID(p))
	if err != nil {
		return peer.AddrInfo{}, err
	}

	return pi, nil
}
