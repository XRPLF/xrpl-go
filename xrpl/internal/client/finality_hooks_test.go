package client

import (
	"context"
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/stretchr/testify/require"
)

type responseDecoderFunc func(any) error

func (f responseDecoderFunc) GetResult(v any) error {
	return f(v)
}

func TestTxFinalityHooksRequireValidatedLedgerResponse(t *testing.T) {
	tests := []struct {
		name      string
		response  ledger.Response
		wantIndex uint32
		wantError error
	}{
		{
			name: "validated ledger",
			response: ledger.Response{
				LedgerIndex: common.LedgerIndex(21),
				Validated:   true,
			},
			wantIndex: 21,
		},
		{
			name: "unvalidated ledger",
			response: ledger.Response{
				LedgerIndex: common.LedgerIndex(21),
			},
			wantError: ErrInvalidValidatedLedgerResponse,
		},
		{
			name:      "missing ledger index",
			response:  ledger.Response{Validated: true},
			wantError: ErrInvalidValidatedLedgerResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := TxFinalityHooks(
				func(context.Context) (ResponseDecoder, error) {
					return nil, errors.New("unexpected transaction lookup")
				},
				func(context.Context) (ResponseDecoder, error) {
					return responseDecoderFunc(func(v any) error {
						response := v.(*ledger.Response)
						*response = tt.response
						return nil
					}), nil
				},
				func(error) bool { return false },
			)

			index, err := hooks.GetValidatedLedger(context.Background())
			require.Equal(t, tt.wantIndex, index)
			if tt.wantError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantError)
			}
		})
	}
}

func TestTxFinalityHooksTreatTransactionNotFoundAsPending(t *testing.T) {
	notFound := errors.New("transaction not found")
	hooks := TxFinalityHooks(
		func(context.Context) (ResponseDecoder, error) {
			return nil, notFound
		},
		func(context.Context) (ResponseDecoder, error) {
			return nil, errors.New("unexpected ledger lookup")
		},
		func(err error) bool {
			return errors.Is(err, notFound)
		},
	)

	status, err := hooks.LookupTransaction(context.Background())
	require.NoError(t, err)
	require.False(t, status.Found)
}
