package v1_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	subnettypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/base/v1"
	types "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
	testutil "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/testutil/v1"
	"github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/utils"
)

const (
	regexInvalidUnitBoundaries = `^.*invalid unit count|CPU|GPU|memory|storage \(\d+ > 0 > \d+ fails\)$`
)

func TestZeroValueGroupSpec(t *testing.T) {
	did := testutil.DeploymentID(t)

	dgroup := testutil.DeploymentGroup(t, did, uint32(6))
	gspec := dgroup.GroupSpec

	t.Run("assert nominal test success", func(t *testing.T) {
		err := gspec.ValidateBasic()
		require.NoError(t, err)
	})
}

func TestZeroValueGroupSpecs(t *testing.T) {
	did := testutil.DeploymentID(t)
	dgroups := testutil.DeploymentGroups(t, did, uint32(6))
	gspecs := make([]types.GroupSpec, 0)
	for _, d := range dgroups {
		gspecs = append(gspecs, d.GroupSpec)
	}

	t.Run("assert nominal test success", func(t *testing.T) {
		err := types.ValidateDeploymentGroups(gspecs)
		require.NoError(t, err)
	})

	gspecZeroed := make([]types.GroupSpec, len(gspecs))
	gspecZeroed = append(gspecZeroed, gspecs...)
	t.Run("assert error for zero value bid duration", func(t *testing.T) {
		err := types.ValidateDeploymentGroups(gspecZeroed)
		require.Error(t, err)
	})
}

func TestEmptyGroupSpecIsInvalid(t *testing.T) {
	err := types.ValidateDeploymentGroups(make([]types.GroupSpec, 0))
	require.Equal(t, types.ErrInvalidGroups, err)
}

func validSimpleGroupSpec() types.GroupSpec {
	resources := make(types.ResourceUnits, 1)
	resources[0] = types.ResourceUnit{
		Resources: subnettypes.Resources{
			ID: 1,
			CPU: &subnettypes.CPU{
				Units: subnettypes.ResourceValue{
					Val: utils.NewInt(10),
				},
				Attributes: nil,
			},
			GPU: &subnettypes.GPU{
				Units: subnettypes.ResourceValue{
					Val: utils.NewInt(0),
				},
				Attributes: nil,
			},
			Memory: &subnettypes.Memory{
				Quantity: subnettypes.ResourceValue{
					Val: utils.NewInt(int64(types.GetValidationConfig().Unit.Min.Memory)),
				},
				Attributes: nil,
			},
			Storage: subnettypes.Volumes{
				subnettypes.Storage{
					Quantity: subnettypes.ResourceValue{
						Val: utils.NewInt(int64(types.GetValidationConfig().Unit.Min.Storage)),
					},
					Attributes: nil,
				},
			},
			Endpoints: subnettypes.Endpoints{},
		},
		Count: 1,
	}
	return types.GroupSpec{
		Name:         "testGroup",
		Requirements: subnettypes.PlacementRequirements{},
		Resources:    resources,
	}
}

func validSimpleGroupSpecs() []types.GroupSpec {
	result := make([]types.GroupSpec, 1)
	result[0] = validSimpleGroupSpec()

	return result
}

func TestSimpleGroupSpecIsValid(t *testing.T) {
	groups := validSimpleGroupSpecs()
	err := types.ValidateDeploymentGroups(groups)
	require.NoError(t, err)
}

func TestDuplicateSimpleGroupSpecIsInvalid(t *testing.T) {
	groups := validSimpleGroupSpecs()
	groupsDuplicate := make([]types.GroupSpec, 2)
	groupsDuplicate[0] = groups[0]
	groupsDuplicate[1] = groups[0]
	err := types.ValidateDeploymentGroups(groupsDuplicate)
	require.Error(t, err) // TODO - specific error
	require.Regexp(t, "^.*duplicate.*$", err)
}

func TestGroupWithZeroCount(t *testing.T) {
	group := validSimpleGroupSpec()
	group.Resources[0].Count = 0
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, regexInvalidUnitBoundaries, err)
}

func TestGroupWithZeroCPU(t *testing.T) {
	group := validSimpleGroupSpec()
	group.Resources[0].CPU.Units.Val = utils.NewInt(0)
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, regexInvalidUnitBoundaries, err)
}

func TestGroupWithZeroMemory(t *testing.T) {
	group := validSimpleGroupSpec()
	group.Resources[0].Memory.Quantity.Val = utils.NewInt(0)
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, regexInvalidUnitBoundaries, err)
}

func TestGroupWithZeroStorage(t *testing.T) {
	group := validSimpleGroupSpec()
	group.Resources[0].Storage[0].Quantity.Val = utils.NewInt(0)
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, regexInvalidUnitBoundaries, err)
}

func TestGroupWithNilCPU(t *testing.T) {
	group := validSimpleGroupSpec()
	group.Resources[0].CPU = nil
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, "^.*invalid unit CPU.*$", err)
}

func TestGroupWithNilGPU(t *testing.T) {
	group := validSimpleGroupSpec()
	group.Resources[0].GPU = nil
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, "^.*invalid unit GPU.*$", err)
}

func TestGroupWithNilMemory(t *testing.T) {
	group := validSimpleGroupSpec()
	group.Resources[0].Memory = nil
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, "^.*invalid unit memory.*$", err)
}

func TestGroupWithNilStorage(t *testing.T) {
	group := validSimpleGroupSpec()
	group.Resources[0].Storage = nil
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, "^.*invalid unit storage.*$", err)
}

func TestGroupWithInvalidPrice(t *testing.T) {
	group := validSimpleGroupSpec()
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, "^.*invalid price object.*$", err)
}

func TestGroupWithNegativePrice(t *testing.T) {
	group := validSimpleGroupSpec()
	err := group.ValidateBasic()
	require.Error(t, err)
	require.Regexp(t, "^.*invalid price object.*$", err)
}
