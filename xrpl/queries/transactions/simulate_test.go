package transactions

import (
	"encoding/json"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const simulateTxBlob = "120000240000ACA461400000000000000168400000000000000A730074008114B5F762798A53D543A014CAF8B297CFF8F2F937E88314550FC62003E785DC231A1058A05E56E3F09CF4E6"

func validSimulateTxJSON() transaction.FlatTransaction {
	return transaction.FlatTransaction{
		"TransactionType": "Payment",
		"Account":         types.Address("rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"),
		"Destination":     types.Address("r3kmLJN5D28dHuH8vZNUZpMC43pEHpaocV"),
		"Amount":          "1",
	}
}

func TestSimulateRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request SimulateRequest
		wantErr error
	}{
		{name: "JSON input", request: SimulateRequest{TxJSON: validSimulateTxJSON()}},
		{name: "blob input", request: SimulateRequest{TxBlob: simulateTxBlob, Binary: true}},
		{name: "empty signature fields remain unsigned", request: SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"TxnSignature": "", "SigningPubKey": "", "Signers": []any{},
		}}},
		{name: "neither input", wantErr: ErrInvalidSimulateRequest},
		{name: "both inputs", request: SimulateRequest{TxJSON: validSimulateTxJSON(), TxBlob: simulateTxBlob}, wantErr: ErrInvalidSimulateRequest},
		{name: "empty JSON object", request: SimulateRequest{TxJSON: transaction.FlatTransaction{}}, wantErr: ErrInvalidSimulateTxJSON},
		{name: "missing TransactionType", request: SimulateRequest{TxJSON: transaction.FlatTransaction{"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"}}, wantErr: ErrInvalidSimulateTxJSON},
		{name: "missing Account", request: SimulateRequest{TxJSON: transaction.FlatTransaction{"TransactionType": "Payment"}}, wantErr: ErrInvalidSimulateTxJSON},
		{name: "non-hex blob", request: SimulateRequest{TxBlob: "not-hex"}, wantErr: ErrInvalidSimulateTxBlob},
		{name: "odd-length blob", request: SimulateRequest{TxBlob: "ABC"}, wantErr: ErrInvalidSimulateTxBlob},
		{name: "signed TxnSignature", request: SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "TxnSignature": "DEADBEEF",
		}}, wantErr: ErrSignedSimulateTransaction},
		{name: "signed SigningPubKey", request: SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "SigningPubKey": "ED0123",
		}}, wantErr: ErrSignedSimulateTransaction},
		{name: "signed Signers", request: SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "Signers": []any{map[string]any{"Signer": map[string]any{}}},
		}}, wantErr: ErrSignedSimulateTransaction},
		{name: "malformed signature type", request: SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "TxnSignature": 1,
		}}, wantErr: ErrInvalidSimulateTxJSON},
		{name: "malformed Signers type", request: SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "Signers": "",
		}}, wantErr: ErrInvalidSimulateTxJSON},
		{name: "invalid NetworkID", request: SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "NetworkID": -1,
		}}, wantErr: ErrInvalidSimulateNetworkID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "simulate", tt.request.Method())
			require.Equal(t, version.RippledAPIV2, tt.request.APIVersion())
		})
	}
}

func TestSimulateRequestValidateNetworkID(t *testing.T) {
	tests := []struct {
		name     string
		expected uint32
		network  any
		omit     bool
		blob     bool
		wantErr  error
	}{
		{name: "restricted matching", expected: 2048, network: uint32(2048)},
		{name: "restricted matching alternate numeric representation", expected: 2048, network: json.Number("2048")},
		{name: "restricted missing", expected: 2048, omit: true, wantErr: ErrMissingSimulateNetworkID},
		{name: "restricted mismatch", expected: 2048, network: uint32(2049), wantErr: ErrMismatchedSimulateNetworkID},
		{name: "identified standard network mismatch", expected: 1, network: uint32(2), wantErr: ErrMismatchedSimulateNetworkID},
		{name: "unknown identity accepts valid explicit value", network: uint32(2048)},
		{name: "blob validation remains server authoritative", expected: 2048, blob: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := SimulateRequest{}
			if tt.blob {
				request.TxBlob = simulateTxBlob
			} else {
				request.TxJSON = validSimulateTxJSON()
				if !tt.omit {
					request.TxJSON["NetworkID"] = tt.network
				}
			}

			err := request.ValidateNetworkID(tt.expected)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.blob || tt.omit {
				return
			}
			require.Equal(t, tt.network, request.TxJSON["NetworkID"], "validation must preserve the caller's explicit value")
			encoded, err := json.Marshal(request)
			require.NoError(t, err)
			require.Contains(t, string(encoded), `"NetworkID":2048`)
		})
	}
}

func TestSimulateResponseJSONVariants(t *testing.T) {
	const jsonSuccess = `{
		"applied": false,
		"engine_result": "tesSUCCESS",
		"engine_result_code": 0,
		"engine_result_message": "The simulated transaction would have been applied.",
		"ledger_index": 105935704,
		"meta": {
			"AffectedNodes": [],
			"TransactionIndex": 59,
			"TransactionResult": "tesSUCCESS",
			"delivered_amount": "1"
		},
		"tx_json": {
			"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"Amount": "1",
			"Destination": "r3kmLJN5D28dHuH8vZNUZpMC43pEHpaocV",
			"Fee": "10",
			"Sequence": 44196,
			"SigningPubKey": "",
			"TransactionType": "Payment",
			"TxnSignature": ""
		}
	}`
	const binarySuccess = `{
		"applied": false,
		"engine_result": "tesSUCCESS",
		"engine_result_code": 0,
		"engine_result_message": "The simulated transaction would have been applied.",
		"ledger_index": 105935704,
		"meta_blob": "201C0000003BF8E5110061250644",
		"tx_blob": "` + simulateTxBlob + `"
	}`
	const nonTecWithoutMetadata = `{
		"applied": false,
		"engine_result": "temREDUNDANT",
		"engine_result_code": -275,
		"engine_result_message": "The transaction is redundant.",
		"ledger_index": 105935694,
		"tx_json": {
			"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"TransactionType": "Payment"
		}
	}`

	tests := []struct {
		name       string
		fixture    string
		wantErr    error
		checkValue func(*testing.T, SimulateResponse)
	}{
		{
			name:    "JSON output with metadata",
			fixture: jsonSuccess,
			checkValue: func(t *testing.T, got SimulateResponse) {
				require.NotNil(t, got.TxJSON)
				require.Empty(t, got.MetaBlob)
				require.NotNil(t, got.Meta)
				require.Equal(t, "tesSUCCESS", got.Meta.TransactionResult)
				require.Equal(t, "1", got.Meta.DeliveredAmount)
			},
		},
		{
			name:    "binary output with metadata",
			fixture: binarySuccess,
			checkValue: func(t *testing.T, got SimulateResponse) {
				require.Equal(t, simulateTxBlob, got.TxBlob)
				require.NotEmpty(t, got.MetaBlob)
				require.Nil(t, got.Meta)
			},
		},
		{
			name:    "non-tec JSON output without metadata",
			fixture: nonTecWithoutMetadata,
			checkValue: func(t *testing.T, got SimulateResponse) {
				require.Equal(t, "temREDUNDANT", got.EngineResult)
				require.Nil(t, got.Meta)
			},
		},
		{name: "reject both transaction output variants", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"tx_blob":"1200"
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject neither transaction output variant", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject JSON output with meta_blob", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"meta_blob":"1200"
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject binary output with meta", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"1200","meta":{}
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject malformed binary transaction", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"XYZ"
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject malformed binary metadata", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"1200","meta_blob":"XYZ"
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject missing applied flag", fixture: `{
			"engine_result":"tesSUCCESS","engine_result_code":0,"engine_result_message":"ok",
			"ledger_index":1,"tx_blob":"1200"
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject applied true", fixture: `{
			"applied":true,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"1200"
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject null meta_blob in JSON response", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"meta_blob":null
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject empty meta_blob in JSON response", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"meta_blob":""
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject null meta_blob in binary response", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"1200","meta_blob":null
		}`, wantErr: ErrInvalidSimulateResponse},
		{name: "reject null metadata", fixture: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"meta":null
		}`, wantErr: ErrInvalidSimulateResponse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got SimulateResponse
			err := json.Unmarshal([]byte(tt.fixture), &got)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, got.Validate())
			tt.checkValue(t, got)

			encoded, err := json.Marshal(got)
			require.NoError(t, err)
			var roundTrip SimulateResponse
			require.NoError(t, json.Unmarshal(encoded, &roundTrip))
			require.Equal(t, got, roundTrip)
		})
	}
}
