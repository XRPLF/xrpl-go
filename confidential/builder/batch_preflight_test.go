package builder

import (
	"testing"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	ledgerentries "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// The tests in this file cover the checks the assembler runs against the ledger it reads.

func TestBuildBatchOneValidatedLedger(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100)}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, 1)

	querier := fixture.querier()
	_, err := BuildBatch(querier, BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 10),
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 20),
		},
	})
	require.NoError(t, err)

	validated := 0
	for _, request := range querier.requests {
		if request.LedgerIndex == common.Current {
			continue
		}
		validated++
		if request.LedgerHash != "" {
			require.Equal(t, mockLedgerHash, request.LedgerHash)
			continue
		}
		require.Equal(t, mockLedgerIndex, request.LedgerIndex)
	}
	require.Positive(t, validated)
}

func TestBuildBatchHolderPreflight(t *testing.T) {
	t.Run("locked sender", func(t *testing.T) {
		fixture := newBatchFixture(t).
			withIssuance(testIssuanceID).
			withHolder(testAccount, testIssuanceID, holderFixture{
				spending: amountOf(100),
				flags:    ledgerentries.LsfMPTLocked,
			}).
			withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)})

		_, err := BuildBatch(fixture.querier(), BuildBatchParams{
			Account: testAccount,
			Operations: []BatchOperation{
				sendOp(fixture, testAccount, testDestination, testIssuanceID, 10),
				sendOp(fixture, testAccount, testDestination, testIssuanceID, 20),
			},
		})
		require.ErrorIs(t, err, ErrHolderLocked)
	})

	t.Run("locked clawback holder", func(t *testing.T) {
		fixture := newBatchFixture(t).
			withIssuance(testIssuerIssuanceID).
			withHolder(testDestination, testIssuerIssuanceID, holderFixture{
				spending: amountOf(40),
				flags:    ledgerentries.LsfMPTLocked,
			}).
			withSequence(testAccount, 3)

		batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
			Account: testAccount,
			Operations: []BatchOperation{
				clawbackOp(fixture, testAccount, testDestination, testIssuerIssuanceID),
				TransactionOp{Tx: &transaction.AccountSet{
					BaseTx: transaction.BaseTx{Account: types.Address(testAccount), TransactionType: transaction.AccountSetTx},
				}},
			},
		})
		require.NoError(t, err, "an issuer must be able to claw back from a holder it locked")
		require.Equal(t, types.MPTPlainAmount(40).String(), innerOf(t, batch, 0)["MPTAmount"])
	})

	t.Run("destination never opted in", func(t *testing.T) {
		fixture := newBatchFixture(t).
			withIssuance(testIssuanceID).
			withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100)}).
			withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
			withHolder(testDelegate, testIssuanceID, holderFixture{unregistered: true})

		_, err := BuildBatch(fixture.querier(), BuildBatchParams{
			Account: testAccount,
			Operations: []BatchOperation{
				sendOp(fixture, testAccount, testDelegate, testIssuanceID, 10),
				sendOp(fixture, testAccount, testDestination, testIssuanceID, 20),
			},
		})
		require.ErrorIs(t, err, ErrReceiverNotOptedIn)
	})
}

// TestBuildBatchChecksVersionFreshnessOnlyWhereBound pins that an open-ledger version change
// rejects a Batch only when a proof binds that holder's version, which is the policy the
// standalone builders follow.
func TestBuildBatchChecksVersionFreshnessOnlyWhereBound(t *testing.T) {
	tests := []struct {
		name       string
		stale      string
		issuanceID string
		operations func(f *batchFixture) []BatchOperation
		wantErr    error
	}{
		{
			name:       "repeat converts bind no version",
			stale:      testAccount,
			issuanceID: testIssuanceID,
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					convertOp(f, testAccount, testIssuanceID, 10),
					convertOp(f, testAccount, testIssuanceID, 20),
				}
			},
		},
		{
			name:       "a send binds no version of its destination",
			stale:      testDestination,
			issuanceID: testIssuanceID,
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					sendOp(f, testAccount, testDestination, testIssuanceID, 10),
					mergeOp(testDestination, testIssuanceID),
				}
			},
		},
		{
			name:       "a clawback binds no version of its holder",
			stale:      testDestination,
			issuanceID: testIssuerIssuanceID,
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					clawbackOp(f, testAccount, testDestination, testIssuerIssuanceID),
					plainAccountSet(testAccount, 0),
				}
			},
		},
		{
			name:       "a send binds its sender's version",
			stale:      testAccount,
			issuanceID: testIssuanceID,
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					convertOp(f, testAccount, testIssuanceID, 10),
					sendOp(f, testAccount, testDestination, testIssuanceID, 10),
				}
			},
			wantErr: ErrStaleBalanceVersion,
		},
		{
			name:       "a convert-back binds its holder's version",
			stale:      testAccount,
			issuanceID: testIssuanceID,
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					convertOp(f, testAccount, testIssuanceID, 10),
					convertBackOp(f, testAccount, testIssuanceID, 10),
				}
			},
			wantErr: ErrStaleBalanceVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newBatchFixture(t).
				withIssuance(testIssuanceID).
				withIssuance(testIssuerIssuanceID).
				withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100), inbox: amountOf(0), publicAmount: 100}).
				withHolder(testDestination, testIssuanceID, holderFixture{spending: amountOf(0), inbox: amountOf(0)}).
				withHolder(testDestination, testIssuerIssuanceID, holderFixture{spending: amountOf(40)})

			index, err := xrplhash.MPToken(tt.issuanceID, tt.stale)
			require.NoError(t, err)
			querier := fixture.querier()
			querier.openVersions = map[string]uint32{index: 99}

			_, err = BuildBatch(querier, BuildBatchParams{
				Account:    testAccount,
				Operations: tt.operations(fixture),
			})
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
