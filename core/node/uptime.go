package node

import (
	"context"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/unicornultrafoundation/subnet-node/config"
	"github.com/unicornultrafoundation/subnet-node/core/account"
	"github.com/unicornultrafoundation/subnet-node/core/apps"
	"github.com/unicornultrafoundation/subnet-node/core/node/uptime"
	"github.com/unicornultrafoundation/subnet-node/p2p"
	"github.com/unicornultrafoundation/subnet-node/repo"
	"go.uber.org/fx"
)

func UptimeService(lc fx.Lifecycle, repo repo.Repo, accountService *account.AccountService, appService *apps.Service, pubSub *pubsub.PubSub, cfg *config.C, id peer.ID, P2P *p2p.P2P) *uptime.UptimeService {
	srv := &uptime.UptimeService{
		Apps:           appService,
		IsVerifier:     cfg.GetBool("uptime.is_verifier", false),
		IsProvider:     cfg.GetBool("provider.enable", false),
		Identity:       id,
		PubSub:         pubSub,
		Datastore:      repo.Datastore(),
		AccountService: accountService,
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return srv.Stop()
		},
		OnStart: func(ctx context.Context) error {
			return srv.Start()
		},
	})
	return srv
}
