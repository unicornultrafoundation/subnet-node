package builder

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	crd "github.com/unicornultrafoundation/subnet-node/pkg/k8s/apis/subnet.node/v1"
	sdl "github.com/unicornultrafoundation/subnet-node/pkg/k8s/sdl"
	"github.com/unicornultrafoundation/subnet-node/pkg/k8s/testutil"
)

func TestDeploySetsEnvironmentVariables(t *testing.T) {
	log := logrus.New().WithField("test", t.Name()).Logger
	const fakeHostname = "ahostname.dev"
	settings := Settings{
		ClusterPublicHostname: fakeHostname,
	}
	lid := testutil.LeaseID(t)
	sdl, err := sdl.ReadFile("../../../../pkg/k8s/sdl/_testdata/deployment/deployment.yaml")
	require.NoError(t, err)

	mani, err := sdl.Manifest()
	require.NoError(t, err)

	sparams := make([]*crd.SchedulerParams, len(mani.GetGroups()[0].Services))

	cmani, err := crd.NewManifest("lease", lid, &mani.GetGroups()[0], crd.ClusterSettings{SchedulerParams: sparams})
	require.NoError(t, err)

	group, sparams, err := cmani.Spec.Group.FromCRD()
	require.NoError(t, err)

	cdep := &ClusterDeployment{
		Lid:     lid,
		Group:   &group,
		Sparams: crd.ClusterSettings{SchedulerParams: sparams},
	}

	workload, err := NewWorkloadBuilder(log, settings, cdep, cmani, 0)
	require.NoError(t, err)

	deploymentBuilder := NewDeployment(workload)

	require.NotNil(t, deploymentBuilder)

	dbuilder := deploymentBuilder.(*deployment)

	container := dbuilder.container()
	require.NotNil(t, container)

	env := make(map[string]string)
	for _, entry := range container.Env {
		env[entry.Name] = entry.Value
	}

	value, ok := env[envVarSubnetNodeClusterPublicHostname]
	require.True(t, ok)
	require.Equal(t, fakeHostname, value)

	value, ok = env[envVarSubnetNodeDeploymentSequence]
	require.True(t, ok)
	require.Equal(t, fmt.Sprintf("%d", lid.GetDSeq()), value)

	value, ok = env[envVarSubnetNodeGroupSequence]
	require.True(t, ok)
	require.Equal(t, fmt.Sprintf("%d", lid.GetGSeq()), value)

	value, ok = env[envVarSubnetNodeOrderSequence]
	require.True(t, ok)
	require.Equal(t, fmt.Sprintf("%d", lid.GetOSeq()), value)

	value, ok = env[envVarSubnetNodeOwner]
	require.True(t, ok)
	require.Equal(t, lid.Owner, value)

	value, ok = env[envVarSubnetNodeProvider]
	require.True(t, ok)
	require.Equal(t, lid.Provider, value)
}
