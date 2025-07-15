package market

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	types "github.com/unicornultrafoundation/subnet-node/proto/subnet/k8s/market/v1"

	deployment "github.com/unicornultrafoundation/subnet-node/core/k8s/types/v1/query/deployment"
)

const (
	ordersPath = "orders"
	orderPath  = "order"
	bidsPath   = "bids"
	bidPath    = "bid"
	leasesPath = "leases"
	leasePath  = "lease"
)

var (
	ErrInvalidPath = errors.New("query: invalid path")
)

// getOrdersPath returns orders path for queries
func getOrdersPath(ofilters OrderFilters) string {
	return fmt.Sprintf("%s/%s/%v", ordersPath, ofilters.Owner, ofilters.StateFlagVal)
}

// OrderPath return order path of given order id for queries
func OrderPath(id types.OrderID) string {
	return fmt.Sprintf("%s/%s", orderPath, orderParts(id))
}

// getBidsPath returns bids path for queries
func getBidsPath(bfilters BidFilters) string {
	return fmt.Sprintf("%s/%s/%v", bidsPath, bfilters.Owner, bfilters.StateFlagVal)
}

// getBidPath return bid path of given bid id for queries
func getBidPath(id types.BidID) string {
	return fmt.Sprintf("%s/%s/%s", bidPath, orderParts(id.OrderID()), id.Provider)
}

// getLeasesPath returns leases path for queries
func getLeasesPath(lfilters LeaseFilters) string {
	return fmt.Sprintf("%s/%s/%v", leasesPath, lfilters.Owner, lfilters.StateFlagVal)
}

// LeasePath return lease path of given lease id for queries
func LeasePath(id types.LeaseID) string {
	return fmt.Sprintf("%s/%s/%s", leasePath, orderParts(id.OrderID()), id.Provider)
}

func orderParts(id types.OrderID) string {
	return fmt.Sprintf("%s/%v/%v/%v", id.Owner, id.DSeq, id.GSeq, id.OSeq)
}

// parseOrderPath returns orderID details with provided queries, and return
// error if occurred due to wrong query
func parseOrderPath(parts []string) (types.OrderID, error) {
	if len(parts) < 4 {
		return types.OrderID{}, ErrInvalidPath
	}

	did, err := deployment.ParseGroupPath(parts[0:3])
	if err != nil {
		return types.OrderID{}, err
	}

	oseq, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		return types.OrderID{}, err
	}

	return types.MakeOrderID(did, uint32(oseq)), nil
}

// parseBidPath returns bidID details with provided queries, and return
// error if occurred due to wrong query
func parseBidPath(parts []string) (types.BidID, error) {
	if len(parts) < 5 {
		return types.BidID{}, ErrInvalidPath
	}

	oid, err := parseOrderPath(parts[0:4])
	if err != nil {
		return types.BidID{}, err
	}

	if !common.IsHexAddress(parts[4]) {
		return types.BidID{}, ErrInvalidPath
	}

	provider := common.HexToAddress(parts[4])

	return types.MakeBidID(oid, provider), nil
}

// ParseLeasePath returns leaseID details with provided queries, and return
// error if occurred due to wrong query
func ParseLeasePath(parts []string) (types.LeaseID, error) {
	bid, err := parseBidPath(parts)
	if err != nil {
		return types.LeaseID{}, err
	}

	return types.MakeLeaseID(bid), nil
}
