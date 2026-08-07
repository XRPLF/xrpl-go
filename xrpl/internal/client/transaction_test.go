package client

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestInspectSignedTransaction(t *testing.T) {
	const (
		account   = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
		publicKey = "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A"
		signature = "AABBCC"
	)
	validSigner := map[string]any{"Signer": map[string]any{
		"Account": account, "SigningPubKey": publicKey, "TxnSignature": signature,
	}}

	tests := []struct {
		name       string
		tx         map[string]any
		allowInner bool
		form       SignedTransactionForm
		wantErr    bool
	}{
		{name: "unsigned", tx: map[string]any{}, form: UnsignedTransaction},
		{name: "single", tx: map[string]any{"SigningPubKey": publicKey, "TxnSignature": signature}, form: SingleSignedTransaction},
		{name: "multisign", tx: map[string]any{"SigningPubKey": "", "Signers": []any{validSigner}}, form: MultiSignedTransaction},
		{name: "multisign missing top-level key", tx: map[string]any{"Signers": []map[string]any{validSigner}}, wantErr: true},
		{name: "multisign nonempty top-level key", tx: map[string]any{"SigningPubKey": publicKey, "Signers": []any{validSigner}}, wantErr: true},
		{name: "multisign wrong top-level key type", tx: map[string]any{"SigningPubKey": 42, "Signers": []any{validSigner}}, wantErr: true},
		{name: "partial single", tx: map[string]any{"SigningPubKey": publicKey}, wantErr: true},
		{name: "empty single public key", tx: map[string]any{"SigningPubKey": "", "TxnSignature": signature}, wantErr: true},
		{name: "empty single signature", tx: map[string]any{"SigningPubKey": publicKey, "TxnSignature": ""}, wantErr: true},
		{name: "wrong single public key type", tx: map[string]any{"SigningPubKey": 42, "TxnSignature": signature}, wantErr: true},
		{name: "wrong single signature type", tx: map[string]any{"SigningPubKey": publicKey, "TxnSignature": 42}, wantErr: true},
		{name: "mixed forms", tx: map[string]any{"SigningPubKey": publicKey, "TxnSignature": signature, "Signers": []any{validSigner}}, wantErr: true},
		{name: "empty signers", tx: map[string]any{"SigningPubKey": "", "Signers": []any{}}, wantErr: true},
		{name: "malformed signers type", tx: map[string]any{"SigningPubKey": "", "Signers": "invalid"}, wantErr: true},
		{name: "malformed signer wrapper", tx: map[string]any{"SigningPubKey": "", "Signers": []any{42}}, wantErr: true},
		{name: "malformed signer data", tx: map[string]any{"SigningPubKey": "", "Signers": []any{map[string]any{"Signer": 42}}}, wantErr: true},
		{name: "signer empty account", tx: map[string]any{"SigningPubKey": "", "Signers": []any{map[string]any{"Signer": map[string]any{"Account": "", "SigningPubKey": publicKey, "TxnSignature": signature}}}}, wantErr: true},
		{name: "signer wrong account type", tx: map[string]any{"SigningPubKey": "", "Signers": []any{map[string]any{"Signer": map[string]any{"Account": 42, "SigningPubKey": publicKey, "TxnSignature": signature}}}}, wantErr: true},
		{name: "signer empty public key", tx: map[string]any{"SigningPubKey": "", "Signers": []any{map[string]any{"Signer": map[string]any{"Account": account, "SigningPubKey": "", "TxnSignature": signature}}}}, wantErr: true},
		{name: "signer wrong public key type", tx: map[string]any{"SigningPubKey": "", "Signers": []any{map[string]any{"Signer": map[string]any{"Account": account, "SigningPubKey": 42, "TxnSignature": signature}}}}, wantErr: true},
		{name: "signer empty signature", tx: map[string]any{"SigningPubKey": "", "Signers": []any{map[string]any{"Signer": map[string]any{"Account": account, "SigningPubKey": publicKey, "TxnSignature": ""}}}}, wantErr: true},
		{name: "signer wrong signature type", tx: map[string]any{"SigningPubKey": "", "Signers": []any{map[string]any{"Signer": map[string]any{"Account": account, "SigningPubKey": publicKey, "TxnSignature": 42}}}}, wantErr: true},
		{name: "inner Batch", tx: map[string]any{"Flags": uint32(types.TfInnerBatchTxn), "SigningPubKey": ""}, allowInner: true, form: InnerBatchTransaction},
		{name: "inner Batch missing key", tx: map[string]any{"Flags": uint32(types.TfInnerBatchTxn)}, allowInner: true, wantErr: true},
		{name: "inner Batch directly submitted", tx: map[string]any{"Flags": uint32(types.TfInnerBatchTxn), "SigningPubKey": ""}, wantErr: true},
		{name: "signed inner Batch", tx: map[string]any{"Flags": uint32(types.TfInnerBatchTxn), "SigningPubKey": publicKey, "TxnSignature": signature}, allowInner: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signingType, err := InspectSignedTransaction(tt.tx, tt.allowInner)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalidSignedTransaction)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.form, signingType)
		})
	}
}

func TestInspectSignedBatchInners(t *testing.T) {
	const (
		publicKey = "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A"
		signature = "AABBCC"
	)
	validInner := map[string]any{"RawTransaction": map[string]any{"Flags": uint32(types.TfInnerBatchTxn), "SigningPubKey": ""}}

	tests := []struct {
		name    string
		tx      map[string]any
		wantErr bool
	}{
		{name: "non-Batch is ignored", tx: map[string]any{"TransactionType": "Payment", "SigningPubKey": publicKey}},
		{name: "valid inner form", tx: map[string]any{"TransactionType": "Batch", "RawTransactions": []any{validInner}}},
		{name: "flattened inner slice", tx: map[string]any{"TransactionType": "Batch", "RawTransactions": []map[string]any{validInner}}},
		{name: "inner carries a signature", tx: map[string]any{"TransactionType": "Batch", "RawTransactions": []any{map[string]any{"RawTransaction": map[string]any{"Flags": uint32(types.TfInnerBatchTxn), "SigningPubKey": "", "TxnSignature": signature}}}}, wantErr: true},
		{name: "inner missing inner-batch flag", tx: map[string]any{"TransactionType": "Batch", "RawTransactions": []any{map[string]any{"RawTransaction": map[string]any{"SigningPubKey": ""}}}}, wantErr: true},
		{name: "malformed RawTransactions", tx: map[string]any{"TransactionType": "Batch", "RawTransactions": "invalid"}, wantErr: true},
		{name: "malformed inner wrapper", tx: map[string]any{"TransactionType": "Batch", "RawTransactions": []any{42}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InspectSignedBatchInners(tt.tx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransactionHelpers(t *testing.T) {
	t.Run("clone is deep for JSON-like values", func(t *testing.T) {
		original := map[string]any{
			"RawTransactions": []map[string]any{{
				"RawTransaction": map[string]any{"Fee": "10"},
			}},
		}
		cloned := CloneTransaction(original)
		clonedRaw := cloned["RawTransactions"].([]map[string]any)
		clonedRaw[0]["RawTransaction"].(map[string]any)["Fee"] = "0"

		originalRaw := original["RawTransactions"].([]map[string]any)
		require.Equal(t, "10", originalRaw[0]["RawTransaction"].(map[string]any)["Fee"])
	})

	t.Run("clone is deep for flattened concrete slice types", func(t *testing.T) {
		original := map[string]any{
			"NFTokenOffers": []string{"offer-1", "offer-2"},
			"Paths":         [][]any{{map[string]any{"issuer": "rIssuer"}}},
		}
		cloned := CloneTransaction(original)
		cloned["NFTokenOffers"].([]string)[0] = "changed"
		cloned["Paths"].([][]any)[0][0].(map[string]any)["issuer"] = "rChanged"

		require.Equal(t, []string{"offer-1", "offer-2"}, original["NFTokenOffers"])
		require.Equal(t, "rIssuer", original["Paths"].([][]any)[0][0].(map[string]any)["issuer"])
	})

	t.Run("replace preserves map aliases", func(t *testing.T) {
		destination := map[string]any{"old": true}
		alias := destination
		ReplaceTransactionContents(destination, map[string]any{"new": true})
		require.Equal(t, map[string]any{"new": true}, alias)
	})

	t.Run("DeliverMax conversion", func(t *testing.T) {
		deliverMax := map[string]any{"currency": "USD", "issuer": "rIssuer", "value": "1"}
		tx := map[string]any{"DeliverMax": deliverMax}
		require.True(t, NormalizeDeliverMax(tx))
		require.Equal(t, deliverMax, tx["Amount"])
		require.NotContains(t, tx, "DeliverMax")
	})

	t.Run("DeliverMax conflict is non-mutating", func(t *testing.T) {
		tx := map[string]any{"Amount": "1", "DeliverMax": "2"}
		require.False(t, NormalizeDeliverMax(tx))
		require.Equal(t, map[string]any{"Amount": "1", "DeliverMax": "2"}, tx)
	})

	t.Run("AccountDelete forces fail hard", func(t *testing.T) {
		require.True(t, SubmissionFailHard(map[string]any{"TransactionType": "AccountDelete"}, false))
		require.False(t, SubmissionFailHard(map[string]any{"TransactionType": "Payment"}, false))
		require.True(t, SubmissionFailHard(map[string]any{"TransactionType": "Payment"}, true))
	})

	t.Run("malformed blob never panics", func(t *testing.T) {
		require.NotPanics(t, func() {
			_, err := DecodeTransactionBlob("not-hex")
			require.Error(t, err)
			require.NotErrorIs(t, err, ErrInvalidSignedTransaction)
		})
	})
}
