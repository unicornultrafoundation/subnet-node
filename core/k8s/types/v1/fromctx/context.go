package fromctx

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/clients/inventory"
)

type CtxKey string

const (
	CtxKeyClientInventory = CtxKey("client-inventory")
)

func ClientInventoryFromContext(ctx context.Context) inventory.Client {
	var res inventory.Client

	val := ctx.Value(CtxKeyClientInventory)
	if val == nil {
		return res
	}

	res = val.(inventory.Client)
	return res
}
