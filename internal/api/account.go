package api

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/unicornultrafoundation/subnet-node/core/account"
)

type AccountAPI struct {
	accountService *account.AccountService
}

// NewAccountAPI creates a new instance of AcountAPI.
func NewAccountAPI(accountService *account.AccountService) *AccountAPI {
	return &AccountAPI{accountService: accountService}
}

// GetAddress retrieves the Ethereum address of the account
func (api *AccountAPI) GetAddress(ctx context.Context) (string, error) {
	return api.accountService.GetAddress().Hex(), nil
}

// GetBalance retrieves the Ether balance of the account
func (api *AccountAPI) GetBalance(ctx context.Context) (*hexutil.Big, error) {
	bal, err := api.accountService.GetBalance(api.accountService.GetAddress())
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(bal), nil
}
