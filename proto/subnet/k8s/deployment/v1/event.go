package v1

import (
	"encoding/hex"
	"strconv"
)

const (
	evActionDeploymentCreated = "deployment-created"
	evActionDeploymentUpdated = "deployment-updated"
	evActionDeploymentClosed  = "deployment-closed"
	evActionGroupClosed       = "group-closed"
	evActionGroupPaused       = "group-paused"
	evActionGroupStarted      = "group-started"
	evOwnerKey                = "owner"
	evDSeqKey                 = "dseq"
	evGSeqKey                 = "gseq"
	evVersionKey              = "version"
	encodedVersionHexLen      = 64
)

// EventDeploymentCreated struct
type EventDeploymentCreated struct {
	ID      DeploymentID `json:"id"`
	Version []byte       `json:"version"`
}

// NewEventDeploymentCreated initializes creation event.
func NewEventDeploymentCreated(id DeploymentID, version []byte) EventDeploymentCreated {
	return EventDeploymentCreated{
		ID:      id,
		Version: version,
	}
}

// EventDeploymentUpdated struct
type EventDeploymentUpdated struct {
	ID      DeploymentID `json:"id"`
	Version []byte       `json:"version"`
}

// NewEventDeploymentUpdated initializes SDK type
func NewEventDeploymentUpdated(id DeploymentID, version []byte) EventDeploymentUpdated {
	return EventDeploymentUpdated{
		ID:      id,
		Version: version,
	}
}

// EventDeploymentClosed struct
type EventDeploymentClosed struct {
	ID DeploymentID `json:"id"`
}

func NewEventDeploymentClosed(id DeploymentID) EventDeploymentClosed {
	return EventDeploymentClosed{
		ID: id,
	}
}

// DeploymentIDEVAttributes returns event attribues for given DeploymentID
func DeploymentIDEVAttributes(id DeploymentID) []string {
	return []string{
		id.Owner,
		strconv.FormatUint(id.DSeq, 10),
	}
}

// ParseEVDeploymentID returns deploymentID details for given event attributes
func ParseEVDeploymentID(attrs []string) (DeploymentID, error) {
	owner := attrs[0]
	dseq, err := strconv.ParseUint(attrs[1], 10, 64)
	if err != nil {
		return DeploymentID{}, err
	}

	return DeploymentID{
		Owner: owner,
		DSeq:  dseq,
	}, nil
}

// ParseEVDeploymentVersion returns the Deployment's SDL sha256 sum
func ParseEVDeploymentVersion(attrs []string) ([]byte, error) {
	v := attrs[2]
	return decodeHex([]byte(v))
}

func encodeHex(src []byte) []byte {
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst
}

func decodeHex(src []byte) ([]byte, error) {
	dst := make([]byte, hex.DecodedLen(len(src)))
	if _, err := hex.Decode(dst, src); err != nil {
		return []byte{}, err
	}
	return dst, nil
}

// EventGroupClosed provides SDK event to signal group termination
type EventGroupClosed struct {
	ID GroupID `json:"id"`
}

func NewEventGroupClosed(id GroupID) EventGroupClosed {
	return EventGroupClosed{
		ID: id,
	}
}

// EventGroupPaused provides SDK event to signal group termination
type EventGroupPaused struct {
	ID GroupID `json:"id"`
}

func NewEventGroupPaused(id GroupID) EventGroupPaused {
	return EventGroupPaused{
		ID: id,
	}
}

// EventGroupStarted provides SDK event to signal group termination
type EventGroupStarted struct {
	ID GroupID `json:"id"`
}

func NewEventGroupStarted(id GroupID) EventGroupStarted {
	return EventGroupStarted{
		ID: id,
	}
}

// GroupIDEVAttributes returns event attribues for given GroupID
func GroupIDEVAttributes(id GroupID) []string {
	return append(DeploymentIDEVAttributes(id.DeploymentID()),
		strconv.FormatUint(uint64(id.GSeq), 10))
}

// ParseEVGroupID returns GroupID details for given event attributes
func ParseEVGroupID(attrs []string) (GroupID, error) {
	did, err := ParseEVDeploymentID(attrs)
	if err != nil {
		return GroupID{}, err
	}

	gseq, err := strconv.ParseUint(attrs[3], 10, 64)
	if err != nil {
		return GroupID{}, err
	}

	return GroupID{
		Owner: did.Owner,
		DSeq:  did.DSeq,
		GSeq:  uint32(gseq), // nolint: gosec
	}, nil
}
