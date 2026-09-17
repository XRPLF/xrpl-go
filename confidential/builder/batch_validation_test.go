package builder

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// The tests in this file cover the input-only rules the assembler enforces before any ledger
// query or proof work.

func TestBuildBatchOperationCount(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(1000)}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)})

	tests := []struct {
		name  string
		count int
	}{
		{name: "none", count: 0},
		{name: "one", count: 1},
		{name: "nine", count: 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := make([]BatchOperation, 0, tt.count)
			for range tt.count {
				operations = append(operations, sendOp(fixture, testAccount, testDestination, testIssuanceID, 1))
			}

			querier := fixture.querier()
			_, err := BuildBatch(querier, BuildBatchParams{Account: testAccount, Operations: operations})
			require.ErrorIs(t, err, ErrBatchOperationCount)
			require.Zero(t, querier.queries(), "an impossible size must cost no ledger reads")
		})
	}
}

func TestBuildBatchOuterValidation(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100)}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)})

	operations := []BatchOperation{
		sendOp(fixture, testAccount, testDestination, testIssuanceID, 10),
		sendOp(fixture, testAccount, testDestination, testIssuanceID, 20),
	}

	tests := []struct {
		name    string
		params  BuildBatchParams
		wantErr error
	}{
		{
			name:    "missing account",
			params:  BuildBatchParams{Operations: operations},
			wantErr: ErrMissingAccount,
		},
		{
			name:    "invalid account",
			params:  BuildBatchParams{Account: "not-an-address", Operations: operations},
			wantErr: ErrInvalidAccount,
		},
		{
			name:    "zero account",
			params:  BuildBatchParams{Account: zeroClassicAccount, Operations: operations},
			wantErr: ErrInvalidAccount,
		},
		{
			name: "conflicting outer nonce",
			params: BuildBatchParams{
				TxOptions:  TxOptions{Sequence: 3, TicketSequence: 4},
				Account:    testAccount,
				Operations: operations,
			},
			wantErr: ErrConflictingNonce,
		},
		{
			name: "delegated batch",
			params: BuildBatchParams{
				TxOptions:  TxOptions{Delegate: testDelegate},
				Account:    testAccount,
				Operations: operations,
			},
			wantErr: ErrDelegateNotAllowed,
		},
		{
			name: "until failure mode",
			params: BuildBatchParams{
				Account:    testAccount,
				Operations: operations,
				Flags:      transaction.TfUntilFailure,
			},
			wantErr: ErrBatchModeNotSupported,
		},
		{
			name: "independent mode",
			params: BuildBatchParams{
				Account:    testAccount,
				Operations: operations,
				Flags:      transaction.TfIndependent,
			},
			wantErr: ErrBatchModeNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			querier := fixture.querier()
			_, err := BuildBatch(querier, tt.params)
			require.ErrorIs(t, err, tt.wantErr)
			require.Zero(t, querier.queries())
		})
	}
}

func TestIsSupportedInnerTransactionType(t *testing.T) {
	tests := []struct {
		txType    transaction.TxType
		supported bool
	}{
		{txType: transaction.AccountSetTx, supported: true},
		{txType: transaction.TicketCreateTx, supported: true},
		{txType: transaction.CredentialAcceptTx, supported: true},
		{txType: transaction.TrustSetTx, supported: true},
		{txType: transaction.PaymentTx},
		{txType: transaction.MPTokenAuthorizeTx},
		{txType: transaction.MPTokenIssuanceSetTx},
		{txType: transaction.ConfidentialMPTSendTx},
		{txType: transaction.BatchTx},
		{txType: transaction.AccountDeleteTx},
	}

	for _, tt := range tests {
		t.Run(tt.txType.String(), func(t *testing.T) {
			require.Equal(t, tt.supported, IsSupportedInnerTransactionType(tt.txType))
		})
	}
}

// TestBuildBatchRejectsInvalidOperationsBeforeLedgerAccess pins that every operation, including
// a later one, is validated before the first ledger query, and reports the sentinel its
// standalone builder reports for the same input.
func TestBuildBatchRejectsInvalidOperationsBeforeLedgerAccess(t *testing.T) {
	tests := []struct {
		name      string
		operation func(f *batchFixture) BatchOperation
		wantErr   error
	}{
		{
			name:      "nil operation",
			operation: func(_ *batchFixture) BatchOperation { return nil },
			wantErr:   ErrBatchMissingOperation,
		},
		{
			name:      "nil operation pointer",
			operation: func(_ *batchFixture) BatchOperation { return (*SendOp)(nil) },
			wantErr:   ErrBatchMissingOperation,
		},
		{
			name:      "nil ready-made inner pointer",
			operation: func(_ *batchFixture) BatchOperation { return (*TransactionOp)(nil) },
			wantErr:   ErrBatchMissingOperation,
		},
		{
			name: "typed nil ready-made transaction",
			operation: func(_ *batchFixture) BatchOperation {
				return TransactionOp{Tx: (*transaction.AccountSet)(nil)}
			},
			wantErr: ErrBatchMissingOperation,
		},
		{
			name: "unsupported ready-made inner",
			operation: func(_ *batchFixture) BatchOperation {
				holder := types.Address(testDestination)
				return TransactionOp{Tx: &transaction.MPTokenAuthorize{
					BaseTx:            transaction.BaseTx{Account: types.Address(testAccount), TransactionType: transaction.MPTokenAuthorizeTx},
					MPTokenIssuanceID: testIssuanceID,
					Holder:            &holder,
				}}
			},
			wantErr: ErrBatchInnerNotSupported,
		},
		{
			name: "ready-made inner with both nonces",
			operation: func(_ *batchFixture) BatchOperation {
				return TransactionOp{Tx: &transaction.AccountSet{BaseTx: transaction.BaseTx{
					Account:         types.Address(testAccount),
					TransactionType: transaction.AccountSetTx,
					Sequence:        11,
					TicketSequence:  900,
				}}}
			},
			wantErr: ErrConflictingNonce,
		},
		{
			name: "ready-made inner with a fractional TicketCount",
			operation: func(_ *batchFixture) BatchOperation {
				return TransactionOp{Tx: flatInner{flat: transaction.FlatTransaction{
					"Account":         testAccount,
					"TransactionType": transaction.TicketCreateTx.String(),
					"TicketCount":     1.5,
				}}}
			},
			wantErr: ErrInvalidTransaction,
		},
		{
			name: "ready-made inner with a negative Sequence",
			operation: func(_ *batchFixture) BatchOperation {
				return TransactionOp{Tx: flatInner{flat: transaction.FlatTransaction{
					"Account":         testAccount,
					"TransactionType": transaction.AccountSetTx.String(),
					"Sequence":        -1,
				}}}
			},
			wantErr: ErrInvalidTransaction,
		},
		{
			name: "ready-made inner with non-numeric Flags",
			operation: func(_ *batchFixture) BatchOperation {
				return TransactionOp{Tx: flatInner{flat: transaction.FlatTransaction{
					"Account":         testAccount,
					"TransactionType": transaction.AccountSetTx.String(),
					"Flags":           "3",
				}}}
			},
			wantErr: ErrInvalidTransaction,
		},
		{
			name: "confidential operation with its own sequence",
			operation: func(f *batchFixture) BatchOperation {
				op := sendOp(f, testAccount, testDestination, testIssuanceID, 10)
				op.Sequence = 99
				return op
			},
			wantErr: ErrBatchInnerSequenceSet,
		},
		{
			name: "confidential operation with both nonces",
			operation: func(f *batchFixture) BatchOperation {
				op := sendOp(f, testAccount, testDestination, testIssuanceID, 10)
				op.Sequence = 99
				op.TicketSequence = 900
				return op
			},
			wantErr: ErrConflictingNonce,
		},
		{
			name: "delegated convert",
			operation: func(f *batchFixture) BatchOperation {
				op := convertOp(f, testAccount, testIssuanceID, 10)
				op.Delegate = testDelegate
				return op
			},
			wantErr: ErrDelegateNotAllowed,
		},
		{
			name: "send to a malformed destination",
			operation: func(f *batchFixture) BatchOperation {
				op := sendOp(f, testAccount, testDestination, testIssuanceID, 10)
				op.Destination = "not-an-address"
				return op
			},
			wantErr: ErrInvalidDestination,
		},
		{
			name: "send of zero",
			operation: func(f *batchFixture) BatchOperation {
				return sendOp(f, testAccount, testDestination, testIssuanceID, 0)
			},
			wantErr: ErrZeroAmount,
		},
		{
			name: "convert-back with an inverted balance range",
			operation: func(f *batchFixture) BatchOperation {
				op := convertBackOp(f, testAccount, testIssuanceID, 10)
				op.BalanceRange = elgamal.AmountRange{Low: 10, High: 1}
				return op
			},
			wantErr: elgamal.ErrInvalidAmountRange,
		},
		{
			name: "clawback without the issuer key",
			operation: func(f *batchFixture) BatchOperation {
				op := clawbackOp(f, testAccount, testDestination, testIssuerIssuanceID)
				op.IssuerPrivKey = ""
				return op
			},
			wantErr: ErrMissingIssuerKey,
		},
		{
			name: "merge with a malformed issuance",
			operation: func(_ *batchFixture) BatchOperation {
				return mergeOp(testAccount, "not-an-issuance")
			},
			wantErr: ErrInvalidIssuanceID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newBatchFixture(t).
				withIssuance(testIssuanceID).
				withIssuance(testIssuerIssuanceID).
				withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100), publicAmount: 100}).
				withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
				withHolder(testDestination, testIssuerIssuanceID, holderFixture{spending: amountOf(10)})

			querier := fixture.querier()
			_, err := BuildBatch(querier, BuildBatchParams{
				Account: testAccount,
				Operations: []BatchOperation{
					sendOp(fixture, testAccount, testDestination, testIssuanceID, 10),
					tt.operation(fixture),
				},
			})
			require.ErrorIs(t, err, tt.wantErr)
			require.Zero(t, querier.queries(), "an invalid operation must cost no ledger reads")
		})
	}
}

func TestBuildBatchAcceptsOperationPointers(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100)}).
		withHolder(testDestination, testIssuanceID, holderFixture{spending: amountOf(0), inbox: amountOf(0)}).
		withSequence(testAccount, 40)

	send := sendOp(fixture, testAccount, testDestination, testIssuanceID, 10)
	plain := plainAccountSet(testAccount, 0)
	merge := mergeOp(testDestination, testIssuanceID)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account:    testAccount,
		Operations: []BatchOperation{&send, &plain, &merge},
	})
	require.NoError(t, err)
	require.Len(t, batch.RawTransactions, 3)
	require.Equal(t, transaction.ConfidentialMPTSendTx.String(), innerOf(t, batch, 0)["TransactionType"])
	require.Equal(t, uint32(42), innerOf(t, batch, 1)["Sequence"])
	require.Equal(t, transaction.ConfidentialMPTMergeInboxTx.String(), innerOf(t, batch, 2)["TransactionType"])
}
