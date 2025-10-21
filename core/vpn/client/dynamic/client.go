package dynamic

import (
	"context"
)

type DynamicIPClient interface {
	GetLeaseByIP(ctx context.Context, ip string) (Lease, error)
	GetLeaseByPeerID(ctx context.Context, peerID string) (Lease, error)
	RequestIP(ctx context.Context) (Lease, error)
	RenewIP(ctx context.Context) (Lease, error)
	ReleaseIP(ctx context.Context) error
	SetIP(ctx context.Context, ip string) error
	GetIP(ctx context.Context) (string, error)
}
