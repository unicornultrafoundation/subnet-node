package fromctx

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/hostname"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/inventory"
	"github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/ip"
)

type CtxKey string

const (
	CtxKeyClientHostname  = CtxKey("client-hostname")
	CtxKeyClientInventory = CtxKey("client-inventory")
	CtxKeyClientIP        = CtxKey("client-ip")
)

func ClientIPFromContext(ctx context.Context) ip.Client {
	var res ip.Client

	val := ctx.Value(CtxKeyClientIP)
	if val == nil {
		return nil
	}

	res = val.(ip.Client)
	return res
}

func ClientInventoryFromContext(ctx context.Context) inventory.Client {
	var res inventory.Client

	val := ctx.Value(CtxKeyClientInventory)
	if val == nil {
		return res
	}

	res = val.(inventory.Client)
	return res
}

func ClientHostnameFromContext(ctx context.Context) hostname.Client {
	var res hostname.Client

	val := ctx.Value(CtxKeyClientHostname)
	if val == nil {
		return res
	}

	res = val.(hostname.Client)
	return res
}
