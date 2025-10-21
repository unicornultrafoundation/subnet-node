package static

import "context"

type StaticIPClient interface {
	GetPeerID(ctx context.Context, ip string) (string, error)
	IsIPOwnedByNode(ctx context.Context, ip string) (bool, error)
}
