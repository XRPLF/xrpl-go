package transactions_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
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
		request transactions.SimulateRequest
		wantErr error
	}{
		{name: "JSON input", request: transactions.SimulateRequest{TxJSON: validSimulateTxJSON()}},
		{name: "blob input", request: transactions.SimulateRequest{TxBlob: simulateTxBlob, Binary: true}},
		{name: "empty JSON signature fields remain unsigned", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"TxnSignature": "", "SigningPubKey": "", "Signers": []any{},
		}}},
		{name: "neither input", wantErr: transactions.ErrInvalidSimulateRequest},
		{name: "both inputs", request: transactions.SimulateRequest{TxJSON: validSimulateTxJSON(), TxBlob: simulateTxBlob}, wantErr: transactions.ErrInvalidSimulateRequest},
		{name: "empty JSON object", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{}}, wantErr: transactions.ErrInvalidSimulateTxJSON},
		{name: "missing TransactionType", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"}}, wantErr: transactions.ErrInvalidSimulateTxJSON},
		{name: "missing Account", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{"TransactionType": "Payment"}}, wantErr: transactions.ErrInvalidSimulateTxJSON},
		{name: "non-hex blob", request: transactions.SimulateRequest{TxBlob: "not-hex"}, wantErr: transactions.ErrInvalidSimulateTxBlob},
		{name: "odd-length blob", request: transactions.SimulateRequest{TxBlob: "ABC"}, wantErr: transactions.ErrInvalidSimulateTxBlob},
		{name: "opaque end-marker blob is server-validated", request: transactions.SimulateRequest{TxBlob: "E1"}},
		{name: "opaque serialized blob is server-validated", request: transactions.SimulateRequest{TxBlob: "DEADBEEF"}},
		{name: "signed JSON TxnSignature", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "TxnSignature": "DEADBEEF",
		}}, wantErr: transactions.ErrSignedSimulateTransaction},
		{name: "JSON SigningPubKey remains unsigned", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "SigningPubKey": "ED0123",
		}}},
		{name: "unsigned JSON Signers remain unsigned", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"SigningPubKey": "", "Signers": []any{map[string]any{"Signer": map[string]any{
				"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "SigningPubKey": "ED0123", "TxnSignature": "",
			}}},
		}}},
		{name: "signed JSON Signers", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"SigningPubKey": "", "Signers": []any{map[string]any{"Signer": map[string]any{
				"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "SigningPubKey": "ED0123", "TxnSignature": "3045022100AB",
			}}},
		}}, wantErr: transactions.ErrSignedSimulateTransaction},
		{name: "unsigned JSON BatchSigners remain unsigned", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Batch", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"BatchSigners": []any{map[string]any{"BatchSigner": map[string]any{
				"Account": "rLs1MzkFWCxTbuAHgjeTZK4fcCDDnf2KRv", "SigningPubKey": "ED0123", "TxnSignature": "",
			}}},
		}}},
		{name: "signed JSON BatchSigners", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Batch", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"BatchSigners": []any{map[string]any{"BatchSigner": map[string]any{
				"Account": "rLs1MzkFWCxTbuAHgjeTZK4fcCDDnf2KRv", "SigningPubKey": "ED0123", "TxnSignature": "3045022100AB",
			}}},
		}}, wantErr: transactions.ErrSignedSimulateTransaction},
		{name: "signed nested JSON BatchSigners", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Batch", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"BatchSigners": []any{map[string]any{"BatchSigner": map[string]any{
				"Account": "rLs1MzkFWCxTbuAHgjeTZK4fcCDDnf2KRv", "Signers": []any{map[string]any{"Signer": map[string]any{
					"Account": "rK5VzeCz2zAYvfni1fN6sC2CaqZiXYvS3N", "SigningPubKey": "ED0456", "TxnSignature": "3045022100CD",
				}}},
			}}},
		}}, wantErr: transactions.ErrSignedSimulateTransaction},
		{name: "malformed batch signer signature type", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Batch", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"BatchSigners": []any{map[string]any{"BatchSigner": map[string]any{"TxnSignature": 1}}},
		}}, wantErr: transactions.ErrInvalidSimulateTxJSON},
		{name: "malformed signature type", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "TxnSignature": 1,
		}}, wantErr: transactions.ErrInvalidSimulateTxJSON},
		{name: "malformed signer signature type", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"Signers": []any{map[string]any{"Signer": map[string]any{"TxnSignature": 1}}},
		}}, wantErr: transactions.ErrInvalidSimulateTxJSON},
		{name: "malformed signer public key type", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
			"Signers": []any{map[string]any{"Signer": map[string]any{"SigningPubKey": 1}}},
		}}, wantErr: transactions.ErrInvalidSimulateTxJSON},
		{name: "malformed Signers type", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "Signers": "",
		}}, wantErr: transactions.ErrInvalidSimulateTxJSON},
		{name: "invalid NetworkID", request: transactions.SimulateRequest{TxJSON: transaction.FlatTransaction{
			"TransactionType": "Payment", "Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh", "NetworkID": -1,
		}}, wantErr: transactions.ErrInvalidSimulateNetworkID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalBlob := tt.request.TxBlob
			err := tt.request.Validate()
			require.Equal(t, originalBlob, tt.request.TxBlob, "validation must not mutate tx_blob")
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

func TestSimulateRequestValidateNil(t *testing.T) {
	var request *transactions.SimulateRequest
	require.ErrorIs(t, request.Validate(), transactions.ErrInvalidSimulateRequest)
}

func TestSimulateRequestValidateNetworkID(t *testing.T) {
	knownMainnet := uint32(0)
	knownStandard := uint32(1)
	knownRestricted := uint32(2048)
	tests := []struct {
		name     string
		expected *uint32
		network  any
		omit     bool
		blob     string
		wantErr  error
	}{
		{name: "restricted JSON matching", expected: &knownRestricted, network: uint32(2048)},
		{name: "restricted JSON matching alternate numeric representation", expected: &knownRestricted, network: json.Number("2048")},
		{name: "restricted JSON missing is server-autofilled", expected: &knownRestricted, omit: true},
		{name: "restricted JSON mismatch", expected: &knownRestricted, network: uint32(2049), wantErr: transactions.ErrMismatchedSimulateNetworkID},
		{name: "identified standard JSON mismatch", expected: &knownStandard, network: uint32(2), wantErr: transactions.ErrMismatchedSimulateNetworkID},
		{name: "known Mainnet JSON matching", expected: &knownMainnet, network: uint32(0)},
		{name: "known Mainnet JSON mismatch", expected: &knownMainnet, network: uint32(2048), wantErr: transactions.ErrMismatchedSimulateNetworkID},
		{name: "unknown identity accepts valid explicit JSON value", network: uint32(2048)},
		{name: "opaque blob skips local NetworkID validation", expected: &knownRestricted, blob: "E1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := transactions.SimulateRequest{}
			if tt.blob != "" {
				request.TxBlob = tt.blob
			} else {
				request.TxJSON = validSimulateTxJSON()
				if !tt.omit {
					request.TxJSON["NetworkID"] = tt.network
				}
			}

			err := request.ValidateNetworkID(tt.expected)
			require.Equal(t, tt.blob, request.TxBlob, "validation must not mutate tx_blob")
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.blob != "" || tt.omit {
				return
			}
			require.Equal(t, tt.network, request.TxJSON["NetworkID"], "validation must preserve the caller's explicit value")
			encoded, err := json.Marshal(request)
			require.NoError(t, err)
			require.Contains(t, string(encoded), fmt.Sprintf(`"NetworkID":%v`, tt.network))
		})
	}
}

func TestSimulateResponseMarshalIncompleteValue(t *testing.T) {
	type auditRecord struct {
		RequestID string                        `json:"request_id"`
		Operator  string                        `json:"operator"`
		Response  transactions.SimulateResponse `json:"response"`
	}

	response := transactions.SimulateResponse{}
	require.ErrorIs(t, response.Validate(), transactions.ErrInvalidSimulateResponse)

	encoded, err := json.Marshal(auditRecord{
		RequestID: "request-1",
		Operator:  "alice",
		Response:  response,
	})
	require.NoError(t, err)
	require.JSONEq(t, `{
		"request_id":"request-1",
		"operator":"alice",
		"response":{
			"applied":false,
			"engine_result":"",
			"engine_result_code":0,
			"engine_result_message":"",
			"ledger_index":0
		}
	}`, string(encoded))
}

type simulateResponseFixture struct {
	name string
	want transactions.SimulateResponse
	json string
}

func simulateResponseFixtures() []simulateResponseFixture {
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
			"NetworkID": 2048,
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
	return []simulateResponseFixture{
		{
			name: "JSON output with metadata",
			want: transactions.SimulateResponse{
				EngineResult:        "tesSUCCESS",
				EngineResultMessage: "The simulated transaction would have been applied.",
				LedgerIndex:         105935704,
				Meta: &transaction.TxMetadataBuilder{
					AffectedNodes:     []transaction.AffectedNode{},
					TransactionIndex:  59,
					TransactionResult: "tesSUCCESS",
					DeliveredAmount:   "1",
				},
				TxJSON: transaction.FlatTransaction{
					"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"Amount":          "1",
					"Destination":     "r3kmLJN5D28dHuH8vZNUZpMC43pEHpaocV",
					"Fee":             "10",
					"NetworkID":       float64(2048),
					"Sequence":        float64(44196),
					"SigningPubKey":   "",
					"TransactionType": "Payment",
					"TxnSignature":    "",
				},
			},
			json: jsonSuccess,
		},
		{
			name: "binary output with metadata",
			want: transactions.SimulateResponse{
				EngineResult:        "tesSUCCESS",
				EngineResultMessage: "The simulated transaction would have been applied.",
				LedgerIndex:         105935704,
				TxBlob:              simulateTxBlob,
				MetaBlob:            "201C0000003BF8E5110061250644",
			},
			json: binarySuccess,
		},
		{
			name: "non-tec JSON output without metadata",
			want: transactions.SimulateResponse{
				EngineResult:        "temREDUNDANT",
				EngineResultCode:    -275,
				EngineResultMessage: "The transaction is redundant.",
				LedgerIndex:         105935694,
				TxJSON: transaction.FlatTransaction{
					"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"TransactionType": "Payment",
				},
			},
			json: nonTecWithoutMetadata,
		},
		{
			name: "non-tec binary output without metadata",
			want: transactions.SimulateResponse{
				EngineResult:        "temREDUNDANT",
				EngineResultCode:    -275,
				EngineResultMessage: "The transaction is redundant.",
				LedgerIndex:         105935694,
				TxBlob:              simulateTxBlob,
			},
			json: `{
				"applied": false,
				"engine_result": "temREDUNDANT",
				"engine_result_code": -275,
				"engine_result_message": "The transaction is redundant.",
				"ledger_index": 105935694,
				"tx_blob": "` + simulateTxBlob + `"
			}`,
		},
	}
}

func TestSimulateResponseSerialize(t *testing.T) {
	for _, tt := range simulateResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.want)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(encoded))
		})
	}
}

func TestSimulateResponseJSONDecode(t *testing.T) {
	for _, tt := range simulateResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var got transactions.SimulateResponse
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSimulateResponseClientDecode(t *testing.T) {
	for _, tt := range simulateResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&data))
			var got transactions.SimulateResponse
			require.NoError(t, clientinternal.DecodeResultInto(data, &got))
			require.Equal(t, tt.want, got)
		})
	}
}

type simulateResponseInvalidFixture struct {
	name    string
	json    string
	wantErr error
}

func simulateResponseInvalidFixtures() []simulateResponseInvalidFixture {
	return []simulateResponseInvalidFixture{
		{
			name: "reject both transaction output variants",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"tx_blob":"1200"
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject neither transaction output variant",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject JSON output with meta_blob",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"meta_blob":"1200"
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject binary output with meta",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"1200","meta":{}
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject malformed binary transaction",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"XYZ"
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject malformed binary metadata",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"1200","meta_blob":"XYZ"
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject missing applied flag",
			json: `{
			"engine_result":"tesSUCCESS","engine_result_code":0,"engine_result_message":"ok",
			"ledger_index":1,"tx_blob":"1200"
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject applied true",
			json: `{
			"applied":true,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"1200"
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject null meta_blob in JSON response",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"meta_blob":null
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject empty meta_blob in JSON response",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"meta_blob":""
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject null meta_blob in binary response",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_blob":"1200","meta_blob":null
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
		{
			name: "reject null metadata",
			json: `{
			"applied":false,"engine_result":"tesSUCCESS","engine_result_code":0,
			"engine_result_message":"ok","ledger_index":1,"tx_json":{"TransactionType":"Payment"},"meta":null
		}`,
			wantErr: transactions.ErrInvalidSimulateResponse,
		},
	}
}

func TestSimulateResponseInvalidJSONDecode(t *testing.T) {
	for _, tt := range simulateResponseInvalidFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var got transactions.SimulateResponse
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			err := decoder.Decode(&got)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSimulateResponseInvalidClientDecode(t *testing.T) {
	for _, tt := range simulateResponseInvalidFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&data))
			var got transactions.SimulateResponse
			err := clientinternal.DecodeResultInto(data, &got)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSimulateResponseValidateForRequest(t *testing.T) {
	jsonResponse := transactions.SimulateResponse{
		EngineResult:        "tesSUCCESS",
		EngineResultMessage: "ok",
		LedgerIndex:         1,
		TxJSON:              transaction.FlatTransaction{"TransactionType": "Payment"},
	}
	binaryResponse := transactions.SimulateResponse{
		EngineResult:        "tesSUCCESS",
		EngineResultMessage: "ok",
		LedgerIndex:         1,
		TxBlob:              "1200",
	}
	jsonRequest := &transactions.SimulateRequest{TxJSON: validSimulateTxJSON()}
	binaryRequest := &transactions.SimulateRequest{TxBlob: simulateTxBlob, Binary: true}

	tests := []struct {
		name     string
		response transactions.SimulateResponse
		request  *transactions.SimulateRequest
		wantErr  error
	}{
		{name: "JSON request with JSON response", response: jsonResponse, request: jsonRequest},
		{name: "binary request with binary response", response: binaryResponse, request: binaryRequest},
		{name: "JSON request with binary response", response: binaryResponse, request: jsonRequest, wantErr: transactions.ErrInvalidSimulateResponse},
		{name: "binary request with JSON response", response: jsonResponse, request: binaryRequest, wantErr: transactions.ErrInvalidSimulateResponse},
		{name: "nil request", response: jsonResponse, wantErr: transactions.ErrInvalidSimulateRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.response.ValidateForRequest(tt.request)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
