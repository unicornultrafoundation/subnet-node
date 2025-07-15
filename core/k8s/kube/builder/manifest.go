package builder

import (
	"github.com/sirupsen/logrus"
	crd "github.com/unicornultrafoundation/subnet-node/pkg/k8s/apis/subnet.node/v1"
)

type Manifest interface {
	builderBase
	Create() (*crd.Manifest, error)
	Update(obj *crd.Manifest) (*crd.Manifest, error)
	Name() string
}

// manifest composes the k8s subnetv1.Manifest type from LeaseID and
// manifest.Group data.
type manifest struct {
	builder
	mns string
}

var _ Manifest = (*manifest)(nil)

func BuildManifest(log *logrus.Logger, settings Settings, ns string, deployment IClusterDeployment) Manifest {
	return &manifest{
		builder: builder{
			log:        log.WithField("module", "kube-builder").Logger,
			settings:   settings,
			deployment: deployment,
		},
		mns: ns,
	}
}

func (b *manifest) labels() map[string]string {
	return AppendLeaseLabels(b.deployment.LeaseID(), b.builder.labels())
}

func (b *manifest) Create() (*crd.Manifest, error) {
	obj, err := crd.NewManifest(b.mns, b.deployment.LeaseID(), b.deployment.ManifestGroup(), b.deployment.ClusterParams())

	if err != nil {
		return nil, err
	}
	obj.Labels = b.labels()
	return obj, nil
}

func (b *manifest) Update(obj *crd.Manifest) (*crd.Manifest, error) {
	m, err := crd.NewManifest(b.mns, b.deployment.LeaseID(), b.deployment.ManifestGroup(), b.deployment.ClusterParams())
	if err != nil {
		return nil, err
	}

	uobj := obj.DeepCopy()

	uobj.Spec = m.Spec
	uobj.Labels = b.labels()

	return uobj, nil
}

func (b *manifest) NS() string {
	return b.mns
}
