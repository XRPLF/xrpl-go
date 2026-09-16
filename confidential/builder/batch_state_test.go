package builder

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// The tests in this file cover how the predicted state advances from one inner to the next.

func TestBuildBatchConvertThenSpend(t *testing.T) {
	const (
		publicBalance = 300
		firstConvert  = 100
		secondConvert = 150
		startSequence = 11
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{
			unregistered: true,
			publicAmount: publicBalance,
		}).
		withSequence(testAccount, startSequence)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			convertOp(fixture, testAccount, testIssuanceID, firstConvert),
			convertOp(fixture, testAccount, testIssuanceID, secondConvert),
		},
	})
	require.NoError(t, err)

	first := innerOf(t, batch, 0)
	second := innerOf(t, batch, 1)
	require.Contains(t, first, "HolderEncryptionKey")
	require.Contains(t, first, "ZKProof")
	require.NotContains(t, second, "HolderEncryptionKey")
	require.NotContains(t, second, "ZKProof")
}

func TestBuildBatchConvertFundsLaterSend(t *testing.T) {
	const (
		senderBalance = 250
		destPublic    = 60
		destConvert   = 40
		sent          = 25
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(senderBalance)}).
		withHolder(testDestination, testIssuanceID, holderFixture{
			unregistered: true,
			publicAmount: destPublic,
		}).
		withSequence(testAccount, 5).
		withSequence(testDestination, 9)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			convertOp(fixture, testDestination, testIssuanceID, destConvert),
			sendOp(fixture, testAccount, testDestination, testIssuanceID, sent),
		},
	})
	require.NoError(t, err)

	destinationKey := fixture.holderKey(testDestination, testIssuanceID)
	send := firstSend(t, batch, 1)
	require.Equal(t, uint64(sent), decryptField(t, send.DestinationEncryptedAmount, destinationKey.PrivKeyHex),
		"the send must encrypt to the key the in-batch convert registers")
}

func TestBuildBatchClawbackSeesEarlierInners(t *testing.T) {
	const (
		holderSpending = 70
		holderPublic   = 200
		converted      = 30
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuerIssuanceID).
		withHolder(testDestination, testIssuerIssuanceID, holderFixture{
			spending:     amountOf(holderSpending),
			publicAmount: holderPublic,
		}).
		withSequence(testAccount, 15).
		withSequence(testDestination, 3)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			convertOp(fixture, testDestination, testIssuerIssuanceID, converted),
			clawbackOp(fixture, testAccount, testDestination, testIssuerIssuanceID),
		},
	})
	require.NoError(t, err)

	clawback := innerOf(t, batch, 1)
	require.Equal(t, types.MPTPlainAmount(holderSpending+converted).String(), clawback["MPTAmount"],
		"the clawback burns the balance the earlier convert leaves")
}

func TestBuildBatchSpendAfterSecondMerge(t *testing.T) {
	const (
		spending = 10
		inbox    = 10
		sent     = 5
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(100), inbox: amountOf(0)}).
		withHolder(testDestination, testIssuanceID, holderFixture{spending: amountOf(spending), inbox: amountOf(inbox)}).
		withSequence(testAccount, 5).
		withSequence(testDestination, 5)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			mergeOp(testDestination, testIssuanceID),
			mergeOp(testDestination, testIssuanceID),
			sendOp(fixture, testDestination, testAccount, testIssuanceID, sent),
		},
	})
	require.NoError(t, err)

	holder := fixture.holderKey(testDestination, testIssuanceID)
	resetInbox := canonicalZeroOf(t, holder.PubKeyHex, testDestination, testIssuanceID)
	afterMerges := addCiphertexts(t,
		spendingCiphertext(t, fixture, testDestination, testIssuanceID),
		fixtureField(t, fixture, testDestination, testIssuanceID, "ConfidentialBalanceInbox"),
		resetInbox,
	)
	require.Equal(t, uint64(spending+inbox), decryptField(t, afterMerges, holder.PrivKeyHex))

	keys := sendKeys{
		sender:   holder.PubKeyHex,
		receiver: fixture.holderKey(testAccount, testIssuanceID).PubKeyHex,
		issuer:   fixture.issuerKey(testIssuanceID).PubKeyHex,
	}
	requireSendBinding(t, firstSend(t, batch, 2), keys, afterMerges, 7, 2)
}

func TestBuildBatchConvertBackAfterFirstConvert(t *testing.T) {
	const (
		converted = 20
		revealed  = 5
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testDelegate, testIssuanceID, holderFixture{unregistered: true, publicAmount: converted}).
		withSequence(testAccount, 5).
		withSequence(testDelegate, 5)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			convertOp(fixture, testDelegate, testIssuanceID, converted),
			mergeOp(testDelegate, testIssuanceID),
			convertBackOp(fixture, testDelegate, testIssuanceID, revealed),
		},
	})
	require.NoError(t, err)

	holder := fixture.holderKey(testDelegate, testIssuanceID)
	initialSpending := canonicalZeroOf(t, holder.PubKeyHex, testDelegate, testIssuanceID)
	merged := addCiphertexts(t, initialSpending, innerOf(t, batch, 0)["HolderEncryptedAmount"].(string))
	require.Equal(t, uint64(converted), decryptField(t, merged, holder.PrivKeyHex))

	requireConvertBackBinding(t, convertBackInner(t, batch, 2), holder.PubKeyHex, merged, 7, 1)
}

func TestBuildBatchSpendAfterClawback(t *testing.T) {
	const (
		clawedBack = 40
		converted  = 30
		sent       = 10
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuerIssuanceID).
		withHolder(testDestination, testIssuerIssuanceID, holderFixture{spending: amountOf(clawedBack), publicAmount: converted}).
		withHolder(testDelegate, testIssuerIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, 5).
		withSequence(testDestination, 5)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			clawbackOp(fixture, testAccount, testDestination, testIssuerIssuanceID),
			convertOp(fixture, testDestination, testIssuerIssuanceID, converted),
			mergeOp(testDestination, testIssuerIssuanceID),
			sendOp(fixture, testDestination, testDelegate, testIssuerIssuanceID, sent),
		},
	})
	require.NoError(t, err)
	require.Equal(t, types.MPTPlainAmount(clawedBack).String(), innerOf(t, batch, 0)["MPTAmount"])

	holder := fixture.holderKey(testDestination, testIssuerIssuanceID)
	zero := canonicalZeroOf(t, holder.PubKeyHex, testDestination, testIssuerIssuanceID)
	inbox := addCiphertexts(t, zero, innerOf(t, batch, 1)["HolderEncryptedAmount"].(string))
	merged := addCiphertexts(t, zero, inbox)
	require.Equal(t, uint64(converted), decryptField(t, merged, holder.PrivKeyHex))

	keys := sendKeys{
		sender:   holder.PubKeyHex,
		receiver: fixture.holderKey(testDelegate, testIssuerIssuanceID).PubKeyHex,
		issuer:   fixture.issuerKey(testIssuerIssuanceID).PubKeyHex,
	}
	requireSendBinding(t, firstSend(t, batch, 3), keys, merged, 7, 2)
}

func TestBuildBatchClawbackAfterClawbackResetsMirror(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuerIssuanceID).
		withHolder(testDestination, testIssuerIssuanceID, holderFixture{spending: amountOf(40), publicAmount: 25}).
		withSequence(testAccount, 5).
		withSequence(testDestination, 5)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			clawbackOp(fixture, testAccount, testDestination, testIssuerIssuanceID),
			convertOp(fixture, testDestination, testIssuerIssuanceID, 25),
			clawbackOp(fixture, testAccount, testDestination, testIssuerIssuanceID),
		},
	})
	require.NoError(t, err)

	issuer := fixture.issuerKey(testIssuerIssuanceID)
	resetMirror := canonicalZeroOf(t, issuer.PubKeyHex, testDestination, testIssuerIssuanceID)
	mirror := addCiphertexts(t, resetMirror, innerOf(t, batch, 1)["IssuerEncryptedAmount"].(string))
	require.Equal(t, types.MPTPlainAmount(25).String(), innerOf(t, batch, 2)["MPTAmount"])
	requireClawbackBinding(t, clawbackInner(t, batch, 2), issuer.PubKeyHex, mirror, 7)
}

func TestBuildBatchResetBalancesStillUsable(t *testing.T) {
	newFixture := func(t *testing.T) *batchFixture {
		t.Helper()
		return newBatchFixture(t).
			withIssuance(testIssuanceID).
			withHolder(testAccount, testIssuanceID, holderFixture{
				spending:     amountOf(100),
				inbox:        amountOf(0),
				publicAmount: 50,
			}).
			withHolder(testDestination, testIssuanceID, holderFixture{
				spending: amountOf(10),
				inbox:    amountOf(10),
			}).
			withHolder(testDelegate, testIssuanceID, holderFixture{unregistered: true, publicAmount: 20}).
			withSequence(testAccount, 5).
			withSequence(testDestination, 5).
			withSequence(testDelegate, 5)
	}

	tests := []struct {
		name       string
		operations func(f *batchFixture) []BatchOperation
	}{
		{
			name: "receive after a merge reset the inbox",
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					mergeOp(testDestination, testIssuanceID),
					sendOp(f, testAccount, testDestination, testIssuanceID, 10),
				}
			},
		},
		{
			name: "merge twice",
			operations: func(_ *batchFixture) []BatchOperation {
				return []BatchOperation{
					mergeOp(testDestination, testIssuanceID),
					mergeOp(testDestination, testIssuanceID),
				}
			},
		},
		{
			name: "merge after a first convert",
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					convertOp(f, testDelegate, testIssuanceID, 20),
					mergeOp(testDelegate, testIssuanceID),
				}
			},
		},
		{
			name: "send to a holder that converted earlier in the Batch",
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					convertOp(f, testDelegate, testIssuanceID, 20),
					sendOp(f, testAccount, testDelegate, testIssuanceID, 10),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newFixture(t)
			batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
				Account:    testAccount,
				Operations: tt.operations(fixture),
			})
			require.NoError(t, err)
			require.Len(t, batch.RawTransactions, 2)
		})
	}
}

func TestBuildBatchRejectsOverdraws(t *testing.T) {
	tests := []struct {
		name       string
		operations func(f *batchFixture) []BatchOperation
	}{
		{
			name: "second convert overdraws the public balance",
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					convertOp(f, testAccount, testIssuanceID, 60),
					convertOp(f, testAccount, testIssuanceID, 60),
				}
			},
		},
		{
			name: "second send overdraws the chained balance",
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					sendOp(f, testAccount, testDestination, testIssuanceID, 90),
					sendOp(f, testAccount, testDestination, testIssuanceID, 90),
				}
			},
		},
		{
			name: "spend from the zero spending balance a first convert leaves",
			operations: func(f *batchFixture) []BatchOperation {
				return []BatchOperation{
					convertOp(f, testDelegate, testIssuanceID, 20),
					sendOp(f, testDelegate, testAccount, testIssuanceID, 1),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newBatchFixture(t).
				withIssuance(testIssuanceID).
				withHolder(testAccount, testIssuanceID, holderFixture{
					spending:     amountOf(100),
					inbox:        amountOf(0),
					publicAmount: 100,
				}).
				withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
				withHolder(testDelegate, testIssuanceID, holderFixture{unregistered: true, publicAmount: 20}).
				withSequence(testAccount, 5).
				withSequence(testDelegate, 5)

			_, err := BuildBatch(fixture.querier(), BuildBatchParams{
				Account:    testAccount,
				Operations: tt.operations(fixture),
			})
			require.ErrorIs(t, err, ErrInsufficientBalance)
		})
	}
}
