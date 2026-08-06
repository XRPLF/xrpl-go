package client

import (
	"context"

	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
)

// ResponseDecoder decodes a transport response payload into a typed result.
// Both the RPC and WebSocket client responses satisfy it.
type ResponseDecoder interface {
	GetResult(v any) error
}

// TxFinalityHooks adapts a client's transaction and validated-ledger queries to
// the finality state machine, keeping the response-to-status mapping in one
// place. isNotFound reports transport errors that mean "transaction not found",
// which monitoring treats as an inconclusive poll rather than a failure.
func TxFinalityHooks(
	lookupTx func(context.Context) (ResponseDecoder, error),
	lookupValidatedLedger func(context.Context) (ResponseDecoder, error),
	isNotFound func(error) bool,
) FinalityHooks[transactions.TxResponse] {
	return FinalityHooks[transactions.TxResponse]{
		LookupTransaction: func(ctx context.Context) (TransactionStatus[transactions.TxResponse], error) {
			res, err := lookupTx(ctx)
			if err != nil {
				if isNotFound(err) {
					return TransactionStatus[transactions.TxResponse]{}, nil
				}
				return TransactionStatus[transactions.TxResponse]{}, err
			}

			var txResponse transactions.TxResponse
			if err := res.GetResult(&txResponse); err != nil {
				return TransactionStatus[transactions.TxResponse]{}, err
			}
			return TransactionStatus[transactions.TxResponse]{
				Response:  &txResponse,
				Found:     true,
				Validated: txResponse.Validated,
			}, nil
		},
		GetValidatedLedger: func(ctx context.Context) (uint32, error) {
			res, err := lookupValidatedLedger(ctx)
			if err != nil {
				return 0, err
			}

			var ledgerResponse ledger.Response
			if err := res.GetResult(&ledgerResponse); err != nil {
				return 0, err
			}
			if !ledgerResponse.Validated || ledgerResponse.LedgerIndex.Uint32() == 0 {
				return 0, ErrInvalidValidatedLedgerResponse
			}
			return ledgerResponse.LedgerIndex.Uint32(), nil
		},
	}
}
