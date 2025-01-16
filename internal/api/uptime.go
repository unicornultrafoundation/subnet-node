package api

import (
	"context"

	"github.com/unicornultrafoundation/subnet-node/core/node/uptime"
	puptime "github.com/unicornultrafoundation/subnet-node/proto/subnet/uptime"
)

type UptimeAPI struct {
	s *uptime.UptimeService
}

// NewUptimeAPI creates a new instance of UptimeAPI.
func NewUptimeAPI(s *uptime.UptimeService) *UptimeAPI {
	return &UptimeAPI{s: s}
}

func (api *UptimeAPI) GetUptime(ctx context.Context) (*puptime.UptimeRecord, error) {
	return api.s.GetUptime(ctx)
}

func (api *UptimeAPI) GetUptimeByPeerId(ctx context.Context, peerId string) (*puptime.UptimeRecord, error) {
	return api.s.GetUptimeByPeer(ctx, peerId)
}
