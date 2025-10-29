package deployment

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	types "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
)

// DeploymentFilters defines flags for deployment list filter
type DeploymentFilters struct {
	Owner common.Address
	// State flag value given
	StateFlagVal string
	// Actual state value decoded from DeploymentStateMap
	State types.Deployment_State
}

// Accept returns whether deployment filters valid or not
func (filters DeploymentFilters) Accept(obj types.Deployment, isValidState bool) bool {
	if (filters.Owner.String() == "" && !isValidState) ||
		(filters.Owner.String() == "" && (obj.State == filters.State)) ||
		(!isValidState && (obj.DeploymentID.Owner == filters.Owner.String())) ||
		(obj.DeploymentID.Owner == filters.Owner.String() && obj.State == filters.State) {
		return true
	}

	return false
}

// Deployment stores deployment and groups details
type Deployment struct {
	types.Deployment `json:"deployment"`
	Groups           []types.Group `json:"groups"`
}

func (d Deployment) String() string {
	return fmt.Sprintf(`Deployment
	Owner:   %s
	DSeq:    %d
	State:   %v
	Version: %s
	Num Groups: %d
	`, d.Deployment.DeploymentID.Owner, d.Deployment.DeploymentID.DSeq,
		d.Deployment.State, d.Deployment.Version, len(d.Groups))
}

// Deployments represents slice of deployment struct
type Deployments []Deployment

func (ds Deployments) String() string {
	var buf bytes.Buffer

	const sep = "\n\n"

	for _, d := range ds {
		buf.WriteString(d.String())
		buf.WriteString(sep)
	}

	if len(ds) > 0 {
		buf.Truncate(buf.Len() - len(sep))
	}

	return buf.String()
}

// Group stores group ID, state and other specifications
type Group types.Group

// GroupFilters defines flags for group list filter
type GroupFilters struct {
	Owner common.Address
	// State flag value given
	StateFlagVal string
	// Actual state value decoded from GroupStateMap
	State types.Group_State
}
