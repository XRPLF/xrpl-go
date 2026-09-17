package client

import (
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// GetXRPBalance queries the balance and converts drops to an XRP decimal string.
func GetXRPBalance(
	getAccountInfo func(*account.InfoRequest) (*account.InfoResponse, error),
	address types.Address,
	ledgerIndex common.LedgerSpecifier,
) (string, error) {
	balance, err := GetXRPDropsBalance(getAccountInfo, address, ledgerIndex)
	if err != nil {
		return "", err
	}
	return currency.DropsToXrp(balance.String())
}

// GetXRPDropsBalance extracts the balance from account_info. A nil ledgerIndex
// leaves ledger selection to the server.
func GetXRPDropsBalance(
	getAccountInfo func(*account.InfoRequest) (*account.InfoResponse, error),
	address types.Address,
	ledgerIndex common.LedgerSpecifier,
) (types.XRPCurrencyAmount, error) {
	response, err := getAccountInfo(&account.InfoRequest{
		Account:     address,
		LedgerIndex: ledgerIndex,
	})
	if err != nil {
		return 0, err
	}
	return response.AccountData.Balance, nil
}

// GetLedgerIndex requests the validated ledger and extracts its index.
func GetLedgerIndex(getLedger func(*ledger.Request) (*ledger.Response, error)) (common.LedgerIndex, error) {
	response, err := getLedger(&ledger.Request{LedgerIndex: common.Validated})
	if err != nil {
		return 0, err
	}
	return response.LedgerIndex, nil
}

// DecodeServerDefinitions decodes the result and checks it against the request hash.
func DecodeServerDefinitions(
	req *server.DefinitionsRequest,
	result ResponseDecoder,
	requestErr error,
) (*server.DefinitionsResponse, error) {
	response, err := DecodeResult[server.DefinitionsResponse](result, requestErr)
	if err != nil {
		return nil, err
	}
	if err := response.ValidateForRequest(req); err != nil {
		return nil, err
	}
	return response, nil
}

// DecodeSimulate decodes the result and checks its requested output mode.
func DecodeSimulate(
	req *transactions.SimulateRequest,
	result ResponseDecoder,
	requestErr error,
) (*transactions.SimulateResponse, error) {
	response, err := DecodeResult[transactions.SimulateResponse](result, requestErr)
	if err != nil {
		return nil, err
	}
	if err := response.ValidateForRequest(req); err != nil {
		return nil, err
	}
	return response, nil
}
