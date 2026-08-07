package rpc

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	rpctypes "github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	transactiontypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

const rpcSubmitSuccess = `{
	"result": {
		"engine_result": "tesSUCCESS",
		"engine_result_code": 0,
		"engine_result_message": "The transaction was applied."
	},
	"status": "success",
	"type": "response"
}`

func TestClientSubmitOptionsAndCallerOwnership(t *testing.T) {
	signer, err := wallet.FromSeed("sEdSuqBPSQaood2DmNYVkwWTn1oQTj2", "")
	require.NoError(t, err)

	signedTx := func() transaction.FlatTransaction {
		return transaction.FlatTransaction{
			"TransactionType":    "Payment",
			"Account":            signer.ClassicAddress.String(),
			"Destination":        "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
			"Amount":             "1",
			"Fee":                "10",
			"Sequence":           uint32(1),
			"LastLedgerSequence": uint32(20),
			"SigningPubKey":      signer.PublicKey,
			"TxnSignature":       "AABBCC",
		}
	}

	tests := []struct {
		name         string
		tx           transaction.FlatTransaction
		opts         *rpctypes.SubmitOptions
		wantFailHard bool
	}{
		{name: "nil options disable autofill", tx: signedTx(), opts: nil},
		{name: "zero options disable autofill", tx: signedTx(), opts: &rpctypes.SubmitOptions{}},
		{
			name: "enabled autofill signs unsigned transaction",
			tx: transaction.FlatTransaction{
				"TransactionType":    "Payment",
				"Account":            signer.ClassicAddress.String(),
				"Destination":        "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"Amount":             "1",
				"Fee":                "10",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(20),
			},
			opts: &rpctypes.SubmitOptions{Autofill: true, Wallet: &signer},
		},
		{
			name: "Batch autofill does not mutate nested caller maps",
			tx: transaction.FlatTransaction{
				"TransactionType":    "Batch",
				"Account":            signer.ClassicAddress.String(),
				"Flags":              uint32(0x00010000),
				"Fee":                "40",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(20),
				"RawTransactions": []map[string]any{{"RawTransaction": map[string]any{
					"TransactionType": "Payment",
					"Account":         signer.ClassicAddress.String(),
					"Destination":     "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					"Amount":          "1",
					"Flags":           uint32(0x40000000),
					"Sequence":        uint32(2),
				}}},
			},
			opts: &rpctypes.SubmitOptions{Autofill: true, Wallet: &signer},
		},
		{
			name: "AccountDelete forces fail_hard",
			tx: transaction.FlatTransaction{
				"TransactionType":    "AccountDelete",
				"Account":            signer.ClassicAddress.String(),
				"Destination":        "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"Fee":                "2000000",
				"Sequence":           uint32(1),
				"LastLedgerSequence": uint32(20),
				"SigningPubKey":      signer.PublicKey,
				"TxnSignature":       "AABBCC",
			},
			opts:         nil,
			wantFailHard: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := clientinternal.CloneTransaction(tt.tx)
			cl, requestsSeen := setupRPCSubmitCapture(t)

			response, err := cl.SubmitTx(tt.tx, tt.opts)
			require.NoError(t, err)
			require.Equal(t, "tesSUCCESS", response.EngineResult)
			require.Equal(t, original, map[string]any(tt.tx), "submission must not mutate the caller's map")
			require.Len(t, *requestsSeen, 1)
			require.Equal(t, "submit", (*requestsSeen)[0]["method"])
			params := (*requestsSeen)[0]["params"].([]any)
			require.Equal(t, tt.wantFailHard, params[0].(map[string]any)["fail_hard"] == true)
		})
	}
}

func TestClientSubmitRejectsInvalidSignedForms(t *testing.T) {
	cfg, err := NewClientConfig("http://testnode/", WithNetworkIdentity(0, "1.12.0"))
	require.NoError(t, err)
	cl := NewClient(cfg)
	tests := []struct {
		name string
		tx   transaction.FlatTransaction
	}{
		{name: "partial single-sign", tx: transaction.FlatTransaction{"TransactionType": "Payment", "SigningPubKey": "AABB"}},
		{name: "malformed multisign", tx: transaction.FlatTransaction{"TransactionType": "Payment", "Signers": []any{"invalid"}}},
		{name: "inner Batch cannot be submitted", tx: transaction.FlatTransaction{"TransactionType": "Payment", "Flags": uint32(0x40000000), "SigningPubKey": ""}},
		{name: "EnableAmendment cannot be submitted", tx: transaction.FlatTransaction{"TransactionType": transaction.EnableAmendmentTx.String(), "SigningPubKey": ""}},
		{name: "SetFee cannot be submitted", tx: transaction.FlatTransaction{"TransactionType": transaction.SetFeeTx.String(), "SigningPubKey": ""}},
		{name: "UNLModify cannot be submitted", tx: transaction.FlatTransaction{"TransactionType": transaction.UNLModifyTx.String(), "SigningPubKey": ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				_, err := cl.SubmitTx(tt.tx, nil)
				require.ErrorIs(t, err, ErrInvalidSignedTransaction)
			})
		})
	}
}

func TestClientSubmitAndWaitNilOptions(t *testing.T) {
	cfg, err := NewClientConfig("http://testnode/", WithNetworkIdentity(0, "1.12.0"))
	require.NoError(t, err)
	cl := NewClient(cfg)
	unsigned := transaction.FlatTransaction{"TransactionType": "Payment"}
	require.NotPanics(t, func() {
		_, err := cl.SubmitTxAndWait(unsigned, nil)
		require.ErrorIs(t, err, ErrMissingWallet)
	})
}

func TestClientSubmitNormalizesDeliverMaxOnCopy(t *testing.T) {
	signer, err := wallet.FromSeed("sEdSuqBPSQaood2DmNYVkwWTn1oQTj2", "")
	require.NoError(t, err)

	deliverMax := map[string]any{
		"currency": "USD",
		"issuer":   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"value":    "1",
	}
	tx := transaction.FlatTransaction{
		"TransactionType":    "Payment",
		"Account":            signer.ClassicAddress.String(),
		"Destination":        "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		"DeliverMax":         deliverMax,
		"Fee":                "10",
		"Sequence":           uint32(1),
		"LastLedgerSequence": uint32(20),
	}
	original := clientinternal.CloneTransaction(tx)
	cl, requestsSeen := setupRPCSubmitCapture(t)

	_, err = cl.SubmitTx(tx, &rpctypes.SubmitOptions{Wallet: &signer})
	require.NoError(t, err)
	require.Equal(t, original, map[string]any(tx))

	params := (*requestsSeen)[0]["params"].([]any)
	txBlob := params[0].(map[string]any)["tx_blob"].(string)
	encoded, err := binarycodec.Decode(txBlob)
	require.NoError(t, err)
	require.Contains(t, encoded, "Amount")
	require.NotContains(t, encoded, "DeliverMax")
}

func TestClientPaymentAmountAliasMatrix(t *testing.T) {
	issued := map[string]any{
		"currency": "USD",
		"issuer":   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"value":    "1",
	}
	tests := []struct {
		name        string
		tx          transaction.FlatTransaction
		expectedErr error
	}{
		{
			name: "DeliverMax becomes Amount",
			tx: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"DeliverMax": issued, "Sequence": uint32(1), "Fee": "10", "LastLedgerSequence": uint32(20),
			},
		},
		{
			name: "matching complex values do not panic",
			tx: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"Amount": issued, "DeliverMax": map[string]any{
					"currency": "USD", "issuer": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "value": "1",
				}, "Sequence": uint32(1), "Fee": "10", "LastLedgerSequence": uint32(20),
			},
		},
		{
			name: "conflicting values fail without mutation",
			tx: transaction.FlatTransaction{
				"TransactionType": "Payment", "Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"Amount": "1", "DeliverMax": "2", "Sequence": uint32(1), "Fee": "10", "LastLedgerSequence": uint32(20),
			},
			expectedErr: ErrAmountAndDeliverMaxMustBeIdentical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := clientinternal.CloneTransaction(tt.tx)
			cl := setupTestRPCClientForAutofill(t, nil)
			require.NotPanics(t, func() {
				err := cl.Autofill(&tt.tx)
				if tt.expectedErr != nil {
					require.ErrorIs(t, err, tt.expectedErr)
					require.Equal(t, original, map[string]any(tt.tx))
					return
				}
				require.NoError(t, err)
				require.Equal(t, issued, tt.tx["Amount"])
				require.NotContains(t, tt.tx, "DeliverMax")
			})
		})
	}
}

func TestClientAutofillNormalizesTypedAccount(t *testing.T) {
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	tx := transaction.FlatTransaction{
		"TransactionType":    transaction.AccountSetTx,
		"Account":            transactiontypes.Address(account),
		"Fee":                "10",
		"LastLedgerSequence": uint32(20),
	}
	cl := setupTestRPCClientForAutofill(t, []string{`{
		"result": {
			"account_data": {"Sequence": 42}
		}
	}`})

	err := cl.Autofill(&tx)
	require.NotErrorIs(t, err, ErrMissingAccountInTransaction)
	require.NoError(t, err)
	require.IsType(t, "", tx["Account"])
	require.Equal(t, account, tx["Account"])
	require.Equal(t, uint32(42), tx["Sequence"])
}

func TestClientAutofillAccountDeleteChecksBlockers(t *testing.T) {
	const account = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	tx := transaction.FlatTransaction{
		"TransactionType":    transaction.AccountDeleteTx,
		"Account":            transactiontypes.Address(account),
		"Destination":        "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		"Fee":                "2000000",
		"Sequence":           uint32(1),
		"LastLedgerSequence": uint32(20),
	}
	original := clientinternal.CloneTransaction(tx)
	cl := setupTestRPCClientForAutofill(t, []string{`{
		"result": {
			"account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"account_objects": [{"LedgerEntryType": "Escrow"}],
			"ledger_index": 1,
			"validated": true
		}
	}`})

	err := cl.Autofill(&tx)
	require.NotErrorIs(t, err, ErrMissingAccountInTransaction)
	require.ErrorIs(t, err, ErrAccountCannotBeDeleted)
	require.Equal(t, original, map[string]any(tx), "failed autofill must not mutate the caller's map")
}

func TestClientAutofillRawTransactionsRejectsNullSigningFields(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		expectedErr error
	}{
		{name: "null SigningPubKey", field: "SigningPubKey", expectedErr: ErrSigningPubKeyFieldMustBeEmpty},
		{name: "null TxnSignature", field: "TxnSignature", expectedErr: ErrTxnSignatureFieldMustBeEmpty},
		{name: "null Signers", field: "Signers", expectedErr: ErrSignersFieldMustBeEmpty},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inner := map[string]any{"Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH", tt.field: nil}
			tx := transaction.FlatTransaction{
				"TransactionType": "Batch",
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"RawTransactions": []map[string]any{{"RawTransaction": inner}},
			}
			cl := setupTestRPCClientForAutofill(t, nil)
			require.ErrorIs(t, cl.autofillRawTransactions(&tx), tt.expectedErr)
		})
	}
}

func TestClientFeeParity(t *testing.T) {
	serverInfo := `{"result":{"info":{"validated_ledger":{"base_fee_xrp":0.00001},"load_factor":1}}}`
	reserve := `{"result":{"state":{"validated_ledger":{"reserve_inc":2000000}}}}`
	tests := []struct {
		name      string
		txType    string
		nSigners  uint64
		responses []string
		expected  string
	}{
		{name: "single sign base fee", txType: "Payment", responses: []string{serverInfo}, expected: "10"},
		{name: "one multisigner", txType: "Payment", nSigners: 1, responses: []string{serverInfo}, expected: "20"},
		{name: "two multisigners", txType: "Payment", nSigners: 2, responses: []string{serverInfo}, expected: "30"},
		{name: "VaultCreate owner reserve", txType: "VaultCreate", responses: []string{serverInfo, reserve}, expected: "2000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := setupTestRPCClientForAutofill(t, tt.responses)
			cl.cfg.feeCushion = 1
			tx := transaction.FlatTransaction{"TransactionType": tt.txType}
			require.NoError(t, cl.calculateFeePerTransactionType(&tx, tt.nSigners))
			require.Equal(t, tt.expected, tx["Fee"])
		})
	}
}

func TestClientSubmitTxBlobStructuralPreflight(t *testing.T) {
	partialBlob, err := binarycodec.Encode(map[string]any{
		"TransactionType": "Payment",
		"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
		"SigningPubKey":   "AABB",
	})
	require.NoError(t, err)
	cfg, err := NewClientConfig("http://testnode/", WithNetworkIdentity(0, "1.12.0"))
	require.NoError(t, err)
	cl := NewClient(cfg)
	require.NotPanics(t, func() {
		_, err := cl.SubmitTxBlob(partialBlob, false)
		require.ErrorIs(t, err, ErrInvalidSignedTransaction)
	})

	signer, err := wallet.FromSeed("sEdSuqBPSQaood2DmNYVkwWTn1oQTj2", "")
	require.NoError(t, err)
	multisignedBlob, _, err := signer.Multisign(map[string]any{
		"TransactionType": "Payment",
		"Account":         signer.ClassicAddress.String(),
		"Destination":     "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		"Amount":          "1",
		"Fee":             "20",
		"Sequence":        uint32(1),
	})
	require.NoError(t, err)
	cl, _ = setupRPCSubmitCapture(t)
	_, err = cl.SubmitTxBlob(multisignedBlob, false)
	require.NoError(t, err)
}

func TestClientSubmitMultisignedStructuralPreflight(t *testing.T) {
	signer, err := wallet.FromSeed("sEdSuqBPSQaood2DmNYVkwWTn1oQTj2", "")
	require.NoError(t, err)
	blob, _, err := signer.Multisign(map[string]any{
		"TransactionType": "Payment",
		"Account":         signer.ClassicAddress.String(),
		"Destination":     "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		"Amount":          "1",
		"Fee":             "20",
		"Sequence":        uint32(1),
	})
	require.NoError(t, err)
	cl, _ := setupRPCSubmitCapture(t)

	response, err := cl.SubmitMultisigned(blob, false)
	require.NoError(t, err)
	require.Equal(t, "tesSUCCESS", response.EngineResult)

	require.NotPanics(t, func() {
		_, err := cl.SubmitMultisigned("malformed", false)
		require.Error(t, err)
	})
}

func setupRPCSubmitCapture(t *testing.T) (*Client, *[]map[string]any) {
	t.Helper()
	seen := make([]map[string]any, 0, 1)
	mock := &testutil.JSONRPCMockClient{}
	mock.DoFunc = func(req *http.Request) (*http.Response, error) {
		payload, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		var decoded map[string]any
		require.NoError(t, json.Unmarshal(payload, &decoded))
		seen = append(seen, decoded)
		return testutil.MockResponse(rpcSubmitSuccess, http.StatusOK, mock)(req)
	}
	cfg, err := NewClientConfig(
		"http://testnode/",
		WithHTTPClient(mock),
		WithNetworkIdentity(0, "1.12.0"),
	)
	require.NoError(t, err)
	return NewClient(cfg), &seen
}
