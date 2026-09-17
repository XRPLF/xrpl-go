package client

import (
	"errors"
	"testing"

	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestGetXRPBalance(t *testing.T) {
	requestErr := errors.New("account lookup failed")
	for _, tt := range []struct {
		name        string
		ledgerIndex common.LedgerSpecifier
		balance     types.XRPCurrencyAmount
		requestErr  error
		want        string
	}{
		{name: "server-selected ledger", balance: 1234567, want: "1.234567"},
		{name: "validated ledger", ledgerIndex: common.Validated, balance: 1000000, want: "1"},
		{name: "zero balance", want: "0"},
		{name: "exact conversion above float64 precision", balance: 9007199254740993, want: "9007199254.740993"},
		{name: "lookup error discards balance", balance: 1000000, requestErr: requestErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			getInfo := func(req *account.InfoRequest) (*account.InfoResponse, error) {
				calls++
				require.Equal(t, &account.InfoRequest{Account: "rAccount", LedgerIndex: tt.ledgerIndex}, req)
				return &account.InfoResponse{AccountData: ledgerentry.AccountRoot{Balance: tt.balance}}, tt.requestErr
			}

			balance, err := GetXRPBalance(getInfo, "rAccount", tt.ledgerIndex)
			require.Equal(t, 1, calls)
			require.Equal(t, tt.want, balance)
			require.ErrorIs(t, err, tt.requestErr)
		})
	}
}

func TestGetXRPDropsBalance(t *testing.T) {
	requestErr := errors.New("account lookup failed")
	for _, tt := range []struct {
		name       string
		requestErr error
		want       types.XRPCurrencyAmount
	}{
		{name: "exact integer drops", want: 9007199254740993},
		{name: "lookup error discards balance", requestErr: requestErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			getInfo := func(req *account.InfoRequest) (*account.InfoResponse, error) {
				require.Equal(t, &account.InfoRequest{Account: "rAccount", LedgerIndex: common.Validated}, req)
				return &account.InfoResponse{AccountData: ledgerentry.AccountRoot{Balance: 9007199254740993}}, tt.requestErr
			}

			balance, err := GetXRPDropsBalance(getInfo, "rAccount", common.Validated)
			require.Equal(t, tt.want, balance)
			require.ErrorIs(t, err, tt.requestErr)
		})
	}
}

func TestGetLedgerIndex(t *testing.T) {
	requestErr := errors.New("ledger lookup failed")
	for _, tt := range []struct {
		name       string
		index      common.LedgerIndex
		requestErr error
		want       common.LedgerIndex
	}{
		{name: "validated ledger index", index: 123, want: 123},
		{name: "zero index is returned unchanged"},
		{name: "lookup error discards index", index: 123, requestErr: requestErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			getLedger := func(req *ledger.Request) (*ledger.Response, error) {
				calls++
				require.Equal(t, &ledger.Request{LedgerIndex: common.Validated}, req)
				return &ledger.Response{LedgerIndex: tt.index}, tt.requestErr
			}

			index, err := GetLedgerIndex(getLedger)
			require.Equal(t, 1, calls)
			require.Equal(t, tt.want, index)
			require.ErrorIs(t, err, tt.requestErr)
		})
	}
}

func TestDecodeServerDefinitions(t *testing.T) {
	const hash = "C685734F5FEB756693B4BB978BBB3A158A65652E71EEB2977068B0D680689213"
	requestErr := errors.New("definitions request failed")
	decodeErr := errors.New("definitions decode failed")
	for _, tt := range []struct {
		name        string
		requestHash string
		requestErr  error
		decodeErr   error
		wantErr     error
	}{
		{name: "matching hash-only response", requestHash: hash},
		{name: "hash-only response without request hash", wantErr: server.ErrInvalidDefinitionsResponse},
		{name: "request error precedes response validation", requestErr: requestErr, wantErr: requestErr},
		{name: "decode error precedes response validation", decodeErr: decodeErr, wantErr: decodeErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := &server.DefinitionsRequest{Hash: tt.requestHash}
			response := server.DefinitionsResponse{Hash: hash}
			decoder := responseDecoderFunc(func(v any) error {
				*v.(*server.DefinitionsResponse) = response
				return tt.decodeErr
			})

			result, err := DecodeServerDefinitions(req, decoder, tt.requestErr)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
				return
			}
			require.NoError(t, err)
			require.Equal(t, &response, result)
		})
	}
}

func TestDecodeSimulate(t *testing.T) {
	jsonResponse := transactions.SimulateResponse{
		EngineResult:        "tesSUCCESS",
		EngineResultMessage: "The simulated transaction succeeded.",
		TxJSON:              transaction.FlatTransaction{"TransactionType": "Payment", "Account": "rAccount"},
	}
	binaryResponse := transactions.SimulateResponse{
		EngineResult:        "tesSUCCESS",
		EngineResultMessage: "The simulated transaction succeeded.",
		TxBlob:              "00",
	}
	requestErr := errors.New("simulate request failed")
	decodeErr := errors.New("simulate decode failed")
	for _, tt := range []struct {
		name       string
		binary     bool
		response   transactions.SimulateResponse
		requestErr error
		decodeErr  error
		wantErr    error
	}{
		{name: "JSON output", response: jsonResponse},
		{name: "binary output", binary: true, response: binaryResponse},
		{name: "wrong output mode", binary: true, response: jsonResponse, wantErr: transactions.ErrInvalidSimulateResponse},
		{name: "request error precedes response validation", requestErr: requestErr, wantErr: requestErr},
		{name: "decode error precedes response validation", decodeErr: decodeErr, wantErr: decodeErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := &transactions.SimulateRequest{TxBlob: "00", Binary: tt.binary}
			decoder := responseDecoderFunc(func(v any) error {
				*v.(*transactions.SimulateResponse) = tt.response
				return tt.decodeErr
			})

			result, err := DecodeSimulate(req, decoder, tt.requestErr)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
				return
			}
			require.NoError(t, err)
			require.Equal(t, &tt.response, result)
		})
	}
}
