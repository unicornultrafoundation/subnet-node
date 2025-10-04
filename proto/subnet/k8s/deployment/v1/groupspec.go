package v1

import (
	"fmt"

	atypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/audit/v1"
	types "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/base/v1"
)

type ResourceGroup interface {
	GetName() string
	GetResourceUnits() ResourceUnits
}

var _ ResourceGroup = (*GroupSpec)(nil)

type GroupSpecs []*GroupSpec

func (gspecs GroupSpecs) Dup() GroupSpecs {
	res := make(GroupSpecs, 0, len(gspecs))

	for _, gspec := range gspecs {
		gs := gspec.Dup()
		res = append(res, &gs)
	}
	return res
}

func (g GroupSpec) Dup() GroupSpec {
	res := GroupSpec{
		Name:         g.Name,
		Requirements: g.Requirements.Dup(),
		Resources:    g.Resources.Dup(),
	}

	return res
}

// ValidateBasic asserts non-zero values
func (g GroupSpec) ValidateBasic() error {
	return g.validate()
}

// GetResourceUnits method returns resources list in group
func (g GroupSpec) GetResourceUnits() ResourceUnits {
	resources := make(ResourceUnits, 0, len(g.Resources))

	for _, r := range g.Resources {
		resources = append(resources, r)
	}

	return resources
}

// GetName method returns group name
func (g GroupSpec) GetName() string {
	return g.Name
}

// MatchResourcesRequirements check if resources attributes match provider's capabilities
func (g GroupSpec) MatchResourcesRequirements(pattr types.Attributes) bool {
	for _, rgroup := range g.GetResourceUnits() {
		pgroup := pattr.GetCapabilitiesGroup("storage")
		for _, storage := range rgroup.Storage {
			if len(storage.Attributes) == 0 {
				continue
			}

			if !storage.Attributes.IN(pgroup) {
				return false
			}
		}
		if gpu := rgroup.GPU; gpu.Units.Val.Uint64() > 0 {
			attr := gpu.Attributes
			if len(attr) == 0 {
				continue
			}

			pgroup = pattr.GetCapabilitiesMap("gpu")

			if !gpu.Attributes.AnyIN(pgroup) {
				return false
			}
		}
	}

	return true
}

// MatchRequirements method compares provided attributes with specific group attributes.
// Argument provider is a bit cumbersome. First element is attributes from x/provider store
// in case tenant does not need signed attributes at all
// rest of elements (if any) are attributes signed by various auditors
func (g GroupSpec) MatchRequirements(provider []atypes.Provider) bool {

	return types.AttributesSubsetOf(g.Requirements.Attributes, provider[0].Attributes)
}

// validate does validation for provided deployment group
func (g *GroupSpec) validate() error {
	if g.Name == "" {
		return fmt.Errorf("empty group spec name denomination")
	}

	if err := g.GetResourceUnits().Validate(); err != nil {
		return err
	}

	return nil
}
