package builder

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/require"
)

// The tests in this file verify each built proof against the balance ciphertext, nonce, and
// version the assembler predicted, which is what makes a dependent inner valid on the ledger.

func TestBuildBatchRepeatedSpends(t *testing.T) {
	const (
		initialBalance = 500
		firstAmount    = 120
		secondAmount   = 80
		startSequence  = 7
		startVersion   = 3
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(initialBalance), version: startVersion}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, startSequence)

	senderKey := fixture.holderKey(testAccount, testIssuanceID)
	receiverKey := fixture.holderKey(testDestination, testIssuanceID)
	issuerKey := fixture.issuerKey(testIssuanceID)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			sendOp(fixture, testAccount, testDestination, testIssuanceID, firstAmount),
			sendOp(fixture, testAccount, testDestination, testIssuanceID, secondAmount),
		},
	})
	require.NoError(t, err)

	require.Equal(t, uint32(startSequence), batch.Sequence, "the outer Batch spends the account's current sequence")
	require.Equal(t, transaction.TfAllOrNothing, batch.Flags)
	require.Len(t, batch.RawTransactions, 2)

	first := innerOf(t, batch, 0)
	second := innerOf(t, batch, 1)
	requireInnerShape(t, first)
	requireInnerShape(t, second)
	require.Equal(t, uint32(startSequence+1), first["Sequence"], "the batch account's inners start one past the Batch")
	require.Equal(t, uint32(startSequence+2), second["Sequence"])

	initialCt := spendingCiphertext(t, fixture, testAccount, testIssuanceID)
	firstTx := firstSend(t, batch, 0)
	secondTx := firstSend(t, batch, 1)

	afterFirst, err := elgamal.Subtract(initialCt, firstTx.SenderEncryptedAmount)
	require.NoError(t, err)
	require.Equal(t, uint64(initialBalance-firstAmount), decryptField(t, afterFirst, senderKey.PrivKeyHex))

	keys := sendKeys{sender: senderKey.PubKeyHex, receiver: receiverKey.PubKeyHex, issuer: issuerKey.PubKeyHex}
	requireSendBinding(t, firstTx, keys, initialCt, startSequence+1, startVersion)
	requireSendBinding(t, secondTx, keys, afterFirst, startSequence+2, startVersion+1)
}

func TestBuildBatchReceivedCreditSpentLater(t *testing.T) {
	const (
		senderBalance    = 400
		receiverSpending = 50
		receiverInbox    = 10
		received         = 90
		returned         = 30
		startSequence    = 4
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{
			spending: amountOf(senderBalance),
			inbox:    amountOf(0),
		}).
		withHolder(testDestination, testIssuanceID, holderFixture{
			spending: amountOf(receiverSpending),
			inbox:    amountOf(receiverInbox),
		}).
		withSequence(testAccount, startSequence).
		withSequence(testDestination, 20)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			sendOp(fixture, testAccount, testDestination, testIssuanceID, received),
			mergeOp(testDestination, testIssuanceID),
			sendOp(fixture, testDestination, testAccount, testIssuanceID, returned),
		},
	})
	require.NoError(t, err)
	require.Len(t, batch.RawTransactions, 3)

	require.Equal(t, uint32(startSequence+1), innerOf(t, batch, 0)["Sequence"])
	require.Equal(t, uint32(20), innerOf(t, batch, 1)["Sequence"])
	require.Equal(t, uint32(21), innerOf(t, batch, 2)["Sequence"])

	receiverKey := fixture.holderKey(testDestination, testIssuanceID)
	finalSend := firstSend(t, batch, 2)
	expected := uint64(receiverSpending + receiverInbox + received)
	require.Equal(t, expected, decryptField(t, mergedSpendingOf(t, fixture, batch), receiverKey.PrivKeyHex))

	senderKey := fixture.holderKey(testAccount, testIssuanceID)
	keys := sendKeys{
		sender:   receiverKey.PubKeyHex,
		receiver: senderKey.PubKeyHex,
		issuer:   fixture.issuerKey(testIssuanceID).PubKeyHex,
	}
	requireSendBinding(t, finalSend, keys, mergedSpendingOf(t, fixture, batch), 21, 1)
}

// mergedSpendingOf recomputes the destination's spending balance after the send and merge in
// TestBuildBatchReceivedCreditSpentLater, the way the transactor computes it.
func mergedSpendingOf(t *testing.T, fixture *batchFixture, batch *transaction.Batch) string {
	t.Helper()

	send := firstSend(t, batch, 0)
	challenge, err := sendChallenge(send.ZKProof)
	require.NoError(t, err)

	destinationKey := fixture.holderKey(testDestination, testIssuanceID)
	credited, err := rerandomize(send.DestinationEncryptedAmount, destinationKey.PubKeyHex, challenge)
	require.NoError(t, err)

	index, err := xrplhash.MPToken(testIssuanceID, testDestination)
	require.NoError(t, err)
	entry := fixture.entries[index]

	inbox, err := elgamal.Add(entry["ConfidentialBalanceInbox"].(string), credited)
	require.NoError(t, err)
	merged, err := elgamal.Add(entry["ConfidentialBalanceSpending"].(string), inbox)
	require.NoError(t, err)
	return merged
}

func TestBuildBatchAcrossIssuancesAndAccounts(t *testing.T) {
	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withIssuance(secondIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(200), version: 1}).
		withHolder(testAccount, secondIssuanceID, holderFixture{spending: amountOf(90), version: 8}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
		withHolder(testDestination, secondIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, 30).
		withSequence(testDestination, 60)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 50),
			sendOp(fixture, testAccount, testDestination, secondIssuanceID, 40),
		},
	})
	require.NoError(t, err)

	senderKey := fixture.holderKey(testAccount, testIssuanceID)
	otherKey := fixture.holderKey(testAccount, secondIssuanceID)
	keys := sendKeys{
		sender:   senderKey.PubKeyHex,
		receiver: fixture.holderKey(testDestination, testIssuanceID).PubKeyHex,
		issuer:   fixture.issuerKey(testIssuanceID).PubKeyHex,
	}
	otherKeys := sendKeys{
		sender:   otherKey.PubKeyHex,
		receiver: fixture.holderKey(testDestination, secondIssuanceID).PubKeyHex,
		issuer:   fixture.issuerKey(secondIssuanceID).PubKeyHex,
	}
	requireSendBinding(t, firstSend(t, batch, 0), keys, spendingCiphertext(t, fixture, testAccount, testIssuanceID), 31, 1)
	requireSendBinding(t, firstSend(t, batch, 1), otherKeys, spendingCiphertext(t, fixture, testAccount, secondIssuanceID), 32, 8)
}

func TestBuildBatchAuditedIssuance(t *testing.T) {
	fixture := newBatchFixture(t).
		withAuditedIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{spending: amountOf(300)}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(5)}).
		withSequence(testAccount, 2)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 40),
			sendOp(fixture, testAccount, testDestination, testIssuanceID, 60),
		},
	})
	require.NoError(t, err)

	first := firstSend(t, batch, 0)
	require.NotNil(t, first.AuditorEncryptedAmount)

	auditor := fixture.issuances[testIssuanceID].auditor
	require.NotNil(t, auditor)
	require.Equal(t, uint64(40), decryptField(t, *first.AuditorEncryptedAmount, auditor.PrivKeyHex))

	initialCt := spendingCiphertext(t, fixture, testAccount, testIssuanceID)
	afterFirst, err := elgamal.Subtract(initialCt, first.SenderEncryptedAmount)
	require.NoError(t, err)
	keys := sendKeys{
		sender:   fixture.holderKey(testAccount, testIssuanceID).PubKeyHex,
		receiver: fixture.holderKey(testDestination, testIssuanceID).PubKeyHex,
		issuer:   fixture.issuerKey(testIssuanceID).PubKeyHex,
		auditor:  auditor.PubKeyHex,
	}
	requireSendBinding(t, firstSend(t, batch, 1), keys, afterFirst, 4, 1)
}

func TestBuildBatchSendThenConvertBack(t *testing.T) {
	const (
		initialBalance = 300
		publicBalance  = 500
		converted      = 200
		sent           = 120
		revealed       = 90
		startSequence  = 21
		startVersion   = 6
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuanceID).
		withHolder(testAccount, testIssuanceID, holderFixture{
			spending:     amountOf(initialBalance),
			version:      startVersion,
			publicAmount: publicBalance,
		}).
		withHolder(testDestination, testIssuanceID, holderFixture{inbox: amountOf(0)}).
		withSequence(testAccount, startSequence)

	holder := fixture.holderKey(testAccount, testIssuanceID)
	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			convertOp(fixture, testAccount, testIssuanceID, converted),
			sendOp(fixture, testAccount, testDestination, testIssuanceID, sent),
			convertBackOp(fixture, testAccount, testIssuanceID, revealed),
		},
	})
	require.NoError(t, err)

	send := firstSend(t, batch, 1)
	convertBack := convertBackInner(t, batch, 2)

	initialCt := spendingCiphertext(t, fixture, testAccount, testIssuanceID)
	afterSend, err := elgamal.Subtract(initialCt, send.SenderEncryptedAmount)
	require.NoError(t, err)
	require.Equal(t, uint64(initialBalance-sent), decryptField(t, afterSend, holder.PrivKeyHex))

	keys := sendKeys{
		sender:   holder.PubKeyHex,
		receiver: fixture.holderKey(testDestination, testIssuanceID).PubKeyHex,
		issuer:   fixture.issuerKey(testIssuanceID).PubKeyHex,
	}
	requireSendBinding(t, send, keys, initialCt, startSequence+2, startVersion)
	requireConvertBackBinding(t, convertBack, holder.PubKeyHex, afterSend, startSequence+3, startVersion+1)
}

func TestBuildBatchClawbackBinding(t *testing.T) {
	const (
		holderSpending = 80
		holderPublic   = 150
		converted      = 45
		issuerSequence = 17
	)

	fixture := newBatchFixture(t).
		withIssuance(testIssuerIssuanceID).
		withHolder(testDestination, testIssuerIssuanceID, holderFixture{
			spending:     amountOf(holderSpending),
			publicAmount: holderPublic,
		}).
		withSequence(testAccount, issuerSequence).
		withSequence(testDestination, 8)

	batch, err := BuildBatch(fixture.querier(), BuildBatchParams{
		Account: testAccount,
		Operations: []BatchOperation{
			convertOp(fixture, testDestination, testIssuerIssuanceID, converted),
			clawbackOp(fixture, testAccount, testDestination, testIssuerIssuanceID),
		},
	})
	require.NoError(t, err)

	convert := innerOf(t, batch, 0)
	issuerMirror := issuerMirrorOf(t, fixture, testDestination, testIssuerIssuanceID)
	predicted, err := elgamal.Add(issuerMirror, convert["IssuerEncryptedAmount"].(string))
	require.NoError(t, err)

	issuer := fixture.issuerKey(testIssuerIssuanceID)
	require.Equal(t, uint64(holderSpending+converted), decryptField(t, predicted, issuer.PrivKeyHex))
	requireClawbackBinding(t, clawbackInner(t, batch, 1), issuer.PubKeyHex, predicted, issuerSequence+1)
}
