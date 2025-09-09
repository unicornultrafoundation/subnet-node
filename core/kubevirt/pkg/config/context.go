package config

import (
	"context"

	corev1 "github.com/rancher/wrangler/v3/pkg/generated/controllers/core"
	"github.com/rancher/wrangler/v3/pkg/generic"
	"github.com/unicornultrafoundation/subnet-node/core/kubevirt/pkg/generated/controllers/kubevirt.io"
	"k8s.io/client-go/rest"
)

type (
	_scaledKey struct{}
)

type Options struct {
	Namespace       string
	Threadiness     int
	HTTPListenPort  int
	HTTPSListenPort int

	RancherEmbedded bool
	RancherURL      string
	HCIMode         bool
}

type Scaled struct {
	Ctx         context.Context
	VirtFactory *kubevirt.Factory
	CoreFactory *corev1.Factory
}

func SetupScaled(ctx context.Context, restConfig *rest.Config, opts *generic.FactoryOptions) (context.Context, *Scaled, error) {
	scaled := &Scaled{
		Ctx: ctx,
	}
	virt, err := kubevirt.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, nil, err
	}
	scaled.VirtFactory = virt
	return ctx, scaled, nil
}
