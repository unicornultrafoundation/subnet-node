package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	atestutil "github.com/unicornultrafoundation/subnet-node/pkg/k8s/testutil"

	mtestutil "github.com/unicornultrafoundation/subnet-node/pkg/k8s/testutil/manifest/v1"
)

func Test_Manifest_encoding(t *testing.T) {
	for _, spec := range mtestutil.Generators {
		// ensure decode(encode(obj)) == obj

		lid := atestutil.LeaseID(t)
		mgroup := spec.Generator.Group(t)
		sparams := make([]*SchedulerParams, len(mgroup.Services))

		kmani, err := NewManifest("foo", lid, &mgroup, ClusterSettings{SchedulerParams: sparams})
		require.NoError(t, err, spec.Name)

		deployment, err := kmani.Deployment()
		require.NoError(t, err, spec.Name)

		assert.Equal(t, lid, deployment.LeaseID(), spec.Name)
		assert.Equal(t, &mgroup, deployment.ManifestGroup(), spec.Name)
	}
}
