package v1

import (
	"errors"
	"strconv"

	dtypes "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/deployment/v1"
)

const (
	evActionOrderCreated = "order-created"
	evActionOrderClosed  = "order-closed"
	evActionBidCreated   = "bid-created"
	evActionBidClosed    = "bid-closed"
	evActionLeaseCreated = "lease-created"
	evActionLeaseClosed  = "lease-closed"

	evOSeqKey        = "oseq"
	evProviderKey    = "provider"
	evPriceDenomKey  = "price-denom"
	evPriceAmountKey = "price-amount"
)

var (
	ErrParsingPrice = errors.New("error parsing price")
)

// EventOrderCreated struct
type EventOrderCreated struct {
	ID OrderID `json:"id"`
}

func NewEventOrderCreated(id OrderID) EventOrderCreated {
	return EventOrderCreated{
		ID: id,
	}
}

// EventOrderClosed struct
type EventOrderClosed struct {
	ID OrderID `json:"id"`
}

func NewEventOrderClosed(id OrderID) EventOrderClosed {
	return EventOrderClosed{
		ID: id,
	}
}

// EventBidCreated struct
type EventBidCreated struct {
	ID BidID `json:"id"`
}

func NewEventBidCreated(id BidID) EventBidCreated {
	return EventBidCreated{
		ID: id,
	}
}

// EventBidClosed struct
type EventBidClosed struct {
	ID BidID `json:"id"`
}

func NewEventBidClosed(id BidID) EventBidClosed {
	return EventBidClosed{
		ID: id,
	}
}

// EventLeaseCreated struct
type EventLeaseCreated struct {
	ID LeaseID `json:"id"`
}

func NewEventLeaseCreated(id LeaseID) EventLeaseCreated {
	return EventLeaseCreated{
		ID: id,
	}
}

// EventLeaseClosed struct
type EventLeaseClosed struct {
	ID LeaseID `json:"id"`
}

func NewEventLeaseClosed(id LeaseID) EventLeaseClosed {
	return EventLeaseClosed{
		ID: id,
	}
}

// orderIDEVAttributes returns event attribues for given orderID
func orderIDEVAttributes(id OrderID) []string {
	return append(dtypes.GroupIDEVAttributes(id.GroupID()),
		strconv.FormatUint(uint64(id.OSeq), 10))
}

// parseEVOrderID returns orderID for given event attributes
func parseEVOrderID(attrs []string) (OrderID, error) {
	gid, err := dtypes.ParseEVGroupID(attrs)
	if err != nil {
		return OrderID{}, err
	}
	oseq, err := strconv.ParseUint(attrs[2], 10, 64)
	if err != nil {
		return OrderID{}, err
	}

	return OrderID{
		Owner: gid.Owner,
		DSeq:  gid.DSeq,
		GSeq:  gid.GSeq,
		OSeq:  uint32(oseq), // nolint: gosec
	}, nil

}

// bidIDEVAttributes returns event attribues for given bidID
func bidIDEVAttributes(id BidID) []string {
	return append(orderIDEVAttributes(id.OrderID()),
		id.Provider)
}

// parseEVBidID returns bidID for given event attributes
func parseEVBidID(attrs []string) (BidID, error) {
	oid, err := parseEVOrderID(attrs)
	if err != nil {
		return BidID{}, err
	}

	provider := attrs[4]
	if err != nil {
		return BidID{}, err
	}

	return BidID{
		Owner:    oid.Owner,
		DSeq:     oid.DSeq,
		GSeq:     oid.GSeq,
		OSeq:     oid.OSeq,
		Provider: provider,
	}, nil
}

// leaseIDEVAttributes returns event attribues for given LeaseID
func leaseIDEVAttributes(id LeaseID) []string {
	return append(orderIDEVAttributes(id.OrderID()),
		id.Provider)
}

// parseEVLeaseID returns leaseID for given event attributes
func parseEVLeaseID(attrs []string) (LeaseID, error) {
	bid, err := parseEVBidID(attrs)
	if err != nil {
		return LeaseID{}, err
	}
	return LeaseID(bid), nil
}
