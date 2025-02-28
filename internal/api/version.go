package api

import (
	"context"

	version "github.com/unicornultrafoundation/subnet-node"
)

type VersionAPI struct{}

func NewVersionAPI() *VersionAPI {
	return &VersionAPI{}
}

func (api *VersionAPI) GetVersion(ctx context.Context) (*version.VersionInfo, error) {
	return version.GetVersionInfo(), nil
}
