package node

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/unicornultrafoundation/subnet-node/bidengine"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"go.uber.org/fx"
)

func BidengineService(lc fx.Lifecycle, cfg *config.C, acc *account.AccountService, dataStore datastore.Datastore) (*bidengine.BidEngine, error) {
	if !cfg.GetBool("bidengine.enable", false) {
		return nil, nil
	}

	transactOpts, err := acc.NewKeyedTransactor()
	if err != nil {
		return nil, err
	}

	srv, err := bidengine.NewBidEngineFromConfigC(
		cfg,
		acc.GetClient(),
		transactOpts,
		dataStore,
	)

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return srv.Stop(ctx)
		},
		OnStart: func(ctx context.Context) error {
			return srv.Start(ctx)
		},
	})

	return srv, nil
}
