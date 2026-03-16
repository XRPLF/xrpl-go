package wallet

import (
	"bytes"
	"maps"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed seeds used across all counterparty tests.
const (
	brokerSeed        = "sEdTCFHBquP36KursdZ17ZiuZenJZHg" // rPZsMhM7jNaixFiiipWUuDPifUXCVNYfb6
	counterpartySeed  = "sEd7HmQFsoyj5TAm6d98gytM9LJA1MF" // rJCxK2hX9tDMzbnn3cg1GU2g19Kfmhzxkp
	counterparty2Seed = "sEdStM1pngFcLQqVfH3RQcg2Qr6ov9e" // rwRNeznwHzdfYeKWpevYmax2NSDioyeEtT
)

// buildBrokerSignedLoanSet returns a minimal LoanSet already signed by the broker (single-sign).
func buildBrokerSignedLoanSet(t *testing.T) transaction.FlatTransaction {
	t.Helper()
	broker, err := FromSeed(brokerSeed, "")
	require.NoError(t, err)

	tx := map[string]any{
		"TransactionType": "LoanSet",
		"Account":         broker.ClassicAddress.String(),
		"Fee":             "12",
		"Sequence":        uint32(1),
		"SigningPubKey":   "",
	}

	blob, _, err := broker.Sign(tx)
	require.NoError(t, err)

	decoded, err := binarycodec.Decode(blob)
	require.NoError(t, err)
	return transaction.FlatTransaction(decoded)
}

// buildCounterpartyMultisigTx signs the broker-signed LoanSet as the given wallet
// in multisign mode and returns the resulting signed FlatTransaction and blob.
func buildCounterpartyMultisigTx(t *testing.T, w Wallet, brokerTx transaction.FlatTransaction) (transaction.FlatTransaction, string) {
	t.Helper()
	flat := make(transaction.FlatTransaction, len(brokerTx))
	maps.Copy(flat, brokerTx)
	blob, _, err := SignLoanSetByCounterparty(w, &flat, &SignLoanSetByCounterpartyOptions{Multisign: true})
	require.NoError(t, err)
	return flat, blob
}

func TestSignLoanSetByCounterparty(t *testing.T) {
	counterparty, err := FromSeed(counterpartySeed, "")
	require.NoError(t, err)

	t.Run("pass - single-sign mode", func(t *testing.T) {
		flat := buildBrokerSignedLoanSet(t)
		blob, txHash, err := SignLoanSetByCounterparty(counterparty, &flat, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, blob)
		assert.NotEmpty(t, txHash)

		cs, ok := flat["CounterpartySignature"].(map[string]any)
		require.True(t, ok, "CounterpartySignature should be a map")
		assert.NotEmpty(t, cs["TxnSignature"])
		assert.NotEmpty(t, cs["SigningPubKey"])
		_, hasSigners := cs["Signers"]
		assert.False(t, hasSigners, "single-sign must not set Signers")
	})

	t.Run("pass - multisign mode", func(t *testing.T) {
		flat := buildBrokerSignedLoanSet(t)
		blob, txHash, err := SignLoanSetByCounterparty(counterparty, &flat, &SignLoanSetByCounterpartyOptions{Multisign: true})
		require.NoError(t, err)
		assert.NotEmpty(t, blob)
		assert.NotEmpty(t, txHash)

		cs, ok := flat["CounterpartySignature"].(map[string]any)
		require.True(t, ok)
		signers, ok := cs["Signers"].([]any)
		require.True(t, ok)
		require.Len(t, signers, 1)

		signer := signers[0].(map[string]any)["Signer"].(map[string]any)
		assert.Equal(t, counterparty.ClassicAddress.String(), signer["Account"])
		assert.NotEmpty(t, signer["TxnSignature"])
	})

	t.Run("pass - multisign with custom MultisignAccount", func(t *testing.T) {
		cp2, err := FromSeed(counterparty2Seed, "")
		require.NoError(t, err)

		flat := buildBrokerSignedLoanSet(t)
		opts := &SignLoanSetByCounterpartyOptions{MultisignAccount: cp2.ClassicAddress.String()}
		blob, _, err := SignLoanSetByCounterparty(counterparty, &flat, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, blob)

		cs := flat["CounterpartySignature"].(map[string]any)
		signers := cs["Signers"].([]any)
		signer := signers[0].(map[string]any)["Signer"].(map[string]any)
		assert.Equal(t, cp2.ClassicAddress.String(), signer["Account"])
	})

	errorTests := []struct {
		name string
		tx   transaction.FlatTransaction
		err  error
	}{
		{
			name: "fail - not a LoanSet",
			tx: transaction.FlatTransaction{
				"TransactionType": "Payment",
				"TxnSignature":    "AABBCC",
				"SigningPubKey":   "DEADBEEF",
			},
			err: ErrTxMustBeLoanSet,
		},
		{
			name: "fail - CounterpartySignature already set",
			tx: transaction.FlatTransaction{
				"TransactionType":       "LoanSet",
				"TxnSignature":          "AABBCC",
				"SigningPubKey":         "DEADBEEF",
				"CounterpartySignature": map[string]any{"TxnSignature": "AA"},
			},
			err: ErrCounterpartyAlreadySigned,
		},
		{
			name: "fail - broker has not signed (no fields)",
			tx: transaction.FlatTransaction{
				"TransactionType": "LoanSet",
				"Account":         "rPZsMhM7jNaixFiiipWUuDPifUXCVNYfb6",
				"Fee":             "12",
				"Sequence":        uint32(1),
			},
			err: ErrBrokerMustSignFirst,
		},
		{
			name: "fail - broker has TxnSignature but no SigningPubKey",
			tx: transaction.FlatTransaction{
				"TransactionType": "LoanSet",
				"TxnSignature":    "AABBCC",
			},
			err: ErrBrokerMustSignFirst,
		},
		{
			name: "fail - broker has SigningPubKey but no TxnSignature",
			tx: transaction.FlatTransaction{
				"TransactionType": "LoanSet",
				"SigningPubKey":   "DEADBEEF",
			},
			err: ErrBrokerMustSignFirst,
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := SignLoanSetByCounterparty(counterparty, &tt.tx, nil)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestSignLoanSetByCounterpartyBlob(t *testing.T) {
	counterparty, err := FromSeed(counterpartySeed, "")
	require.NoError(t, err)

	t.Run("pass - single-sign from blob", func(t *testing.T) {
		flat := buildBrokerSignedLoanSet(t)
		inputBlob, err := binarycodec.Encode(flat)
		require.NoError(t, err)

		tx, blob, txHash, err := SignLoanSetByCounterpartyBlob(counterparty, inputBlob, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, blob)
		assert.NotEmpty(t, txHash)
		assert.NotNil(t, tx)

		cs, ok := tx["CounterpartySignature"].(map[string]any)
		require.True(t, ok)
		assert.NotEmpty(t, cs["TxnSignature"])
		assert.NotEmpty(t, cs["SigningPubKey"])
	})

	t.Run("pass - multisign from blob", func(t *testing.T) {
		flat := buildBrokerSignedLoanSet(t)
		inputBlob, err := binarycodec.Encode(flat)
		require.NoError(t, err)

		tx, blob, _, err := SignLoanSetByCounterpartyBlob(counterparty, inputBlob, &SignLoanSetByCounterpartyOptions{Multisign: true})
		require.NoError(t, err)
		assert.NotEmpty(t, blob)

		cs, ok := tx["CounterpartySignature"].(map[string]any)
		require.True(t, ok)
		signers, ok := cs["Signers"].([]any)
		require.True(t, ok)
		require.Len(t, signers, 1)
	})

	t.Run("fail - invalid blob", func(t *testing.T) {
		_, _, _, err := SignLoanSetByCounterpartyBlob(counterparty, "not-a-valid-hex", nil)
		assert.Error(t, err)
	})
}

func TestCombineLoanSetCounterpartySigners(t *testing.T) {
	cp1, err := FromSeed(counterpartySeed, "")
	require.NoError(t, err)
	cp2, err := FromSeed(counterparty2Seed, "")
	require.NoError(t, err)

	baseTx := buildBrokerSignedLoanSet(t)

	t.Run("pass - combines two transactions and sorts signers", func(t *testing.T) {
		tx1, _ := buildCounterpartyMultisigTx(t, cp1, baseTx)
		tx2, _ := buildCounterpartyMultisigTx(t, cp2, baseTx)

		combined, blob, err := CombineLoanSetCounterpartySigners([]transaction.FlatTransaction{tx1, tx2})
		require.NoError(t, err)
		assert.NotEmpty(t, blob)
		assert.NotNil(t, combined)

		cs := combined["CounterpartySignature"].(map[string]any)
		signers := cs["Signers"].([]any)
		require.Len(t, signers, 2)

		acc0 := signers[0].(map[string]any)["Signer"].(map[string]any)["Account"].(string)
		acc1 := signers[1].(map[string]any)["Signer"].(map[string]any)["Account"].(string)
		_, bytes0, _ := addresscodec.DecodeClassicAddressToAccountID(acc0)
		_, bytes1, _ := addresscodec.DecodeClassicAddressToAccountID(acc1)
		assert.Negative(t, bytes.Compare(bytes0, bytes1), "signers should be sorted ascending by account ID bytes")
	})

	t.Run("fail - empty slice", func(t *testing.T) {
		_, _, err := CombineLoanSetCounterpartySigners([]transaction.FlatTransaction{})
		assert.ErrorIs(t, err, ErrNoTransactionsToSign)
	})

	t.Run("fail - not a LoanSet", func(t *testing.T) {
		tx := transaction.FlatTransaction{
			"TransactionType": "Payment",
			"CounterpartySignature": map[string]any{
				"Signers": []any{map[string]any{"Signer": map[string]any{"Account": "rXXX"}}},
			},
		}
		_, _, err := CombineLoanSetCounterpartySigners([]transaction.FlatTransaction{tx})
		assert.ErrorIs(t, err, ErrTxMustBeLoanSet)
	})

	t.Run("fail - missing CounterpartySignature.Signers (single-sign)", func(t *testing.T) {
		flat := make(transaction.FlatTransaction, len(baseTx))
		maps.Copy(flat, baseTx)
		_, _, err := SignLoanSetByCounterparty(cp1, &flat, nil)
		require.NoError(t, err)

		_, _, err = CombineLoanSetCounterpartySigners([]transaction.FlatTransaction{flat})
		assert.ErrorIs(t, err, ErrTxMustIncludeCounterpartySigners)
	})

	t.Run("fail - different transactions", func(t *testing.T) {
		tx1, _ := buildCounterpartyMultisigTx(t, cp1, baseTx)

		otherTx := transaction.FlatTransaction{
			"TransactionType": "LoanSet",
			"Account":         "rPZsMhM7jNaixFiiipWUuDPifUXCVNYfb6",
			"Fee":             "12",
			"Sequence":        uint32(99),
			"TxnSignature":    "AABBCC",
			"SigningPubKey":   "DEADBEEF",
		}
		tx2, _ := buildCounterpartyMultisigTx(t, cp2, otherTx)

		_, _, err := CombineLoanSetCounterpartySigners([]transaction.FlatTransaction{tx1, tx2})
		assert.ErrorIs(t, err, ErrLoanSetTxNotEqual)
	})
}

func TestCombineLoanSetCounterpartySignersBlob(t *testing.T) {
	cp1, err := FromSeed(counterpartySeed, "")
	require.NoError(t, err)
	cp2, err := FromSeed(counterparty2Seed, "")
	require.NoError(t, err)

	baseTx := buildBrokerSignedLoanSet(t)

	t.Run("pass - combines two blobs", func(t *testing.T) {
		_, blob1 := buildCounterpartyMultisigTx(t, cp1, baseTx)
		_, blob2 := buildCounterpartyMultisigTx(t, cp2, baseTx)

		tx, blob, err := CombineLoanSetCounterpartySignersBlob([]string{blob1, blob2})
		require.NoError(t, err)
		assert.NotEmpty(t, blob)
		assert.NotNil(t, tx)

		cs := tx["CounterpartySignature"].(map[string]any)
		signers := cs["Signers"].([]any)
		require.Len(t, signers, 2)
	})

	t.Run("fail - empty slice", func(t *testing.T) {
		_, _, err := CombineLoanSetCounterpartySignersBlob([]string{})
		assert.ErrorIs(t, err, ErrNoTransactionsToSign)
	})

	t.Run("fail - invalid blob", func(t *testing.T) {
		_, _, err := CombineLoanSetCounterpartySignersBlob([]string{"not-valid-hex"})
		assert.Error(t, err)
	})
}
