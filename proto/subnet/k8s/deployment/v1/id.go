package v1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// Equals method compares specific deployment with provided deployment
func (id DeploymentID) Equals(other DeploymentID) bool {
	return id.Owner == other.Owner && id.DSeq == other.DSeq
}

// Validate method for DeploymentID and returns nil
func (id DeploymentID) Validate() error {
	// Check if the owner is a valid Ethereum address
	if !common.IsHexAddress(id.Owner) {
		return fmt.Errorf("DeploymentID: Invalid Owner Address")
	}

	if id.DSeq == 0 {
		return fmt.Errorf("DeploymentID: Invalid Deployment Sequence")
	}

	return nil
}

// String method for deployment IDs
func (id DeploymentID) String() string {
	return fmt.Sprintf("%s/%d", strings.ToLower(id.Owner), id.DSeq)
}

func (id DeploymentID) GetOwnerAddress() (common.Address, error) {
	if !common.IsHexAddress(id.Owner) {
		return common.Address{}, fmt.Errorf("DeploymentID: Invalid Owner Address")
	}
	return common.HexToAddress(id.Owner), nil
}

func ParseDeploymentID(val string) (DeploymentID, error) {
	parts := strings.Split(val, "/")
	return ParseDeploymentPath(parts)
}

// ParseDeploymentPath returns DeploymentID details with provided queries, and return
// error if occurred due to wrong query
func ParseDeploymentPath(parts []string) (DeploymentID, error) {
	if len(parts) != 2 {
		return DeploymentID{}, ErrInvalidIDPath
	}

	if !common.IsHexAddress(parts[0]) {
		return DeploymentID{}, fmt.Errorf("DeploymentID: Invalid Owner Address")
	}

	owner := common.HexToAddress(parts[0])

	dseq, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return DeploymentID{}, err
	}

	return DeploymentID{
		Owner: strings.ToLower(owner.String()),
		DSeq:  dseq,
	}, nil
}

// MakeGroupID returns GroupID instance with provided deployment details
// and group sequence number.
func MakeGroupID(id DeploymentID, gseq uint32) GroupID {
	return GroupID{
		Owner: id.Owner,
		DSeq:  id.DSeq,
		GSeq:  gseq,
	}
}

// DeploymentID method returns DeploymentID details with specific group details
func (id GroupID) DeploymentID() DeploymentID {
	return DeploymentID{
		Owner: id.Owner,
		DSeq:  id.DSeq,
	}
}

// Equals method compares specific group with provided group
func (id GroupID) Equals(other GroupID) bool {
	return id.DeploymentID().Equals(other.DeploymentID()) && id.GSeq == other.GSeq
}

// Validate method for GroupID and returns nil
func (id GroupID) Validate() error {
	if err := id.DeploymentID().Validate(); err != nil {
		return fmt.Errorf("GroupID: Invalid DeploymentID: %w", err)
	}
	if id.GSeq == 0 {
		return fmt.Errorf("GroupID: Invalid Group Sequence")
	}
	return nil
}

// String method provides human readable representation of GroupID.
func (id GroupID) String() string {
	return fmt.Sprintf("%s/%d", id.DeploymentID(), id.GSeq)
}
