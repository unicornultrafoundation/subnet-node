package v1

import (
	"testing"

	testutil "github.com/unicornultrafoundation/subnet-node/pkg/k8s/testutil"
	types "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/base/v1"
	manifest "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/manifest/v1"
	unit "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/utils"
)

// AppManifestGenerator represents a real-world, deployable configuration.
var AppManifestGenerator Generator = manifestGeneratorApp{}

type manifestGeneratorApp struct{}

func (mg manifestGeneratorApp) Manifest(t testing.TB) manifest.Manifest {
	t.Helper()
	return []manifest.Group{
		mg.Group(t),
	}
}

func (mg manifestGeneratorApp) Group(t testing.TB) manifest.Group {
	t.Helper()
	return manifest.Group{
		Name: testutil.Name(t, "manifest-group"),
		Services: []manifest.Service{
			mg.Service(t),
		},
	}
}

func (mg manifestGeneratorApp) Service(t testing.TB) manifest.Service {
	t.Helper()
	return manifest.Service{
		Name:  "demo",
		Image: "nginx:latest",
		Resources: types.Resources{
			ID: 1,
			CPU: &types.CPU{
				Units: types.NewResourceValue(100),
			},
			Memory: &types.Memory{
				Quantity: types.NewResourceValue(128 * unit.Mi),
			},
			GPU: &types.GPU{
				Units: types.NewResourceValue(0),
			},
			Storage: types.Volumes{
				types.Storage{
					Quantity: types.NewResourceValue(256 * unit.Mi),
				},
			},
		},
		Count: 1,
		Expose: []manifest.ServiceExpose{
			mg.ServiceExpose(t),
		},
	}
}

func (mg manifestGeneratorApp) ServiceExpose(t testing.TB) manifest.ServiceExpose {
	return manifest.ServiceExpose{
		Port:    80,
		Service: "demo",
		Global:  true,
		Proto:   "TCP",
		Hosts: []string{
			testutil.Hostname(t),
		},
	}
}
