package v1

import (
	"github.com/pkg/errors"
)

var (
	defaultOrderMaxBids uint32 = 20
	maxOrderMaxBids     uint32 = 500
)

const (
	keyBidMinDeposit = "BidMinDeposit"
	keyOrderMaxBids  = "OrderMaxBids"
)

func DefaultParams() Params {
	return Params{
		OrderMaxBids: defaultOrderMaxBids,
	}
}

func (p Params) Validate() error {
	if err := validateOrderMaxBids(p.OrderMaxBids); err != nil {
		return err
	}
	return nil
}

func validateOrderMaxBids(i interface{}) error {
	val, ok := i.(uint32)

	if !ok {
		return errors.Wrapf(ErrInvalidParam, "invalid type %T", i)
	}

	if val == 0 {
		return errors.Wrap(ErrInvalidParam, "order max bids too low")
	}

	if val > maxOrderMaxBids {
		return errors.Wrap(ErrInvalidParam, "order max bids too high")
	}

	return nil
}
