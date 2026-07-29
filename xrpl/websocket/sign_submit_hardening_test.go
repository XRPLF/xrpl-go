package websocket

import (
	"encoding/json"
	"testing"
	"time"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	wstypes "github.com/Peersyst/xrpl-go/xrpl/websocket/types"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

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
		opts         *wstypes.SubmitOptions
		wantFailHard bool
	}{
		{name: "nil options disable autofill", tx: signedTx(), opts: nil},
		{name: "zero options disable autofill", tx: signedTx(), opts: &wstypes.SubmitOptions{}},
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
			opts: &wstypes.SubmitOptions{Autofill: true, Wallet: &signer},
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
			opts: &wstypes.SubmitOptions{Autofill: true, Wallet: &signer},
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
			cl, seen, cleanup := setupWSSubmitCapture(t)
			defer cleanup()

			response, err := cl.SubmitTx(tt.tx, tt.opts)
			require.NoError(t, err)
			require.Equal(t, "tesSUCCESS", response.EngineResult)
			require.Equal(t, original, map[string]any(tt.tx), "submission must not mutate the caller's map")
			request := <-seen
			require.Equal(t, "submit", request["command"])
			require.Equal(t, tt.wantFailHard, request["fail_hard"] == true)
		})
	}
}

func TestClientSubmitRejectsInvalidSignedForms(t *testing.T) {
	cl := NewClient(*NewClientConfig())
	tests := []struct {
		name string
		tx   transaction.FlatTransaction
	}{
		{name: "partial single-sign", tx: transaction.FlatTransaction{"TransactionType": "Payment", "SigningPubKey": "AABB"}},
		{name: "malformed multisign", tx: transaction.FlatTransaction{"TransactionType": "Payment", "Signers": []any{"invalid"}}},
		{name: "inner Batch cannot be submitted", tx: transaction.FlatTransaction{"TransactionType": "Payment", "Flags": uint32(0x40000000), "SigningPubKey": ""}},
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
	cl := NewClient(*NewClientConfig())
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
	cl, seen, cleanup := setupWSSubmitCapture(t)
	defer cleanup()

	_, err = cl.SubmitTx(tx, &wstypes.SubmitOptions{Wallet: &signer})
	require.NoError(t, err)
	require.Equal(t, original, map[string]any(tx))

	request := <-seen
	txBlob := request["tx_blob"].(string)
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
			cl, cleanup := setupTestClientForAutofill(t, nil)
			defer cleanup()
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
			cl, cleanup := setupTestClientForAutofill(t, nil)
			defer cleanup()
			require.ErrorIs(t, cl.autofillRawTransactions(&tx), tt.expectedErr)
		})
	}
}

func TestClientFeeParity(t *testing.T) {
	serverInfo := map[string]any{
		"id": 1,
		"result": map[string]any{
			"info": map[string]any{
				"validated_ledger": map[string]any{"base_fee_xrp": float32(0.00001)},
				"load_factor":      float32(1),
			},
		},
	}
	reserve := map[string]any{
		"id": 2,
		"result": map[string]any{
			"state": map[string]any{
				"validated_ledger": map[string]any{"reserve_inc": 2000000},
			},
		},
	}
	tests := []struct {
		name      string
		txType    string
		nSigners  uint64
		responses []map[string]any
		expected  string
	}{
		{name: "single sign base fee", txType: "Payment", responses: []map[string]any{serverInfo}, expected: "10"},
		{name: "one multisigner", txType: "Payment", nSigners: 1, responses: []map[string]any{serverInfo}, expected: "20"},
		{name: "two multisigners", txType: "Payment", nSigners: 2, responses: []map[string]any{serverInfo}, expected: "30"},
		{name: "VaultCreate owner reserve", txType: "VaultCreate", responses: []map[string]any{serverInfo, reserve}, expected: "2000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cleanup := setupTestClient(t, tt.responses)
			defer cleanup()
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
	cl := NewClient(*NewClientConfig())
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
	cl, seen, cleanup := setupWSSubmitCapture(t)
	defer cleanup()
	_, err = cl.SubmitTxBlob(multisignedBlob, false)
	require.NoError(t, err)
	<-seen
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
	cl, seen, cleanup := setupWSSubmitCapture(t)
	defer cleanup()

	response, err := cl.SubmitMultisigned(blob, false)
	require.NoError(t, err)
	require.Equal(t, "tesSUCCESS", response.EngineResult)
	<-seen

	require.NotPanics(t, func() {
		_, err := cl.SubmitMultisigned("malformed", false)
		require.Error(t, err)
	})
}

func setupWSSubmitCapture(t *testing.T) (*Client, <-chan map[string]any, func()) {
	t.Helper()
	seen := make(chan map[string]any, 4)
	mock := &testutil.MockWebSocketServer{}
	server := mock.TestWebSocketServer(func(conn *websocket.Conn) {
		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var request map[string]any
			if err := json.Unmarshal(payload, &request); err != nil {
				return
			}
			seen <- request
			response := map[string]any{
				"id": request["id"],
				"result": map[string]any{
					"engine_result":         "tesSUCCESS",
					"engine_result_code":    0,
					"engine_result_message": "The transaction was applied.",
				},
				"status": "success",
				"type":   "response",
			}
			if err := conn.WriteJSON(response); err != nil {
				return
			}
		}
	})

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	cl := NewClient(NewClientConfig().WithHost(url).WithTimeout(time.Second))
	setTrustedTestNetworkIdentity(cl, 0)
	require.NoError(t, cl.Connect())

	cleanup := func() {
		cl.Disconnect()
		server.Close()
	}
	return cl, seen, cleanup
}
