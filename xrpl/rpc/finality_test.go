package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/stretchr/testify/require"
)

type rpcFinalityStep struct {
	method string
	body   string
	err    error
}

func TestIsTransactionNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "exact typed error", err: &ClientError{ErrorString: txnNotFound}, want: true},
		{name: "substring is not enough", err: &ClientError{ErrorString: "transport mentioned txnNotFound"}},
		{name: "different error type", err: context.DeadlineExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isTransactionNotFoundError(tt.err))
		})
	}
}

func TestClientWaitForTransactionFinalityMatrix(t *testing.T) {
	lastLedger := uint32(20)
	transportTimeout := context.DeadlineExceeded

	tests := []struct {
		name               string
		lastLedgerSequence *uint32
		maxAttempts        int
		steps              []rpcFinalityStep
		cancelBefore       bool
		wantResult         string
		wantError          error
		wantCause          error
	}{
		{
			name:               "validation exactly at LastLedgerSequence",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        1,
			steps: []rpcFinalityStep{
				{method: "tx", body: rpcTxnNotFoundResponse},
				{method: "ledger", body: rpcLedgerResponse(20)},
				{method: "tx", body: rpcValidatedTxResponse(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:               "validation racing expiry is rechecked",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        1,
			steps: []rpcFinalityStep{
				{method: "tx", body: rpcTxnNotFoundResponse},
				{method: "ledger", body: rpcLedgerResponse(21)},
				{method: "tx", body: rpcValidatedTxResponse(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:               "expiry only after LastLedgerSequence",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        1,
			steps: []rpcFinalityStep{
				{method: "tx", body: rpcTxnNotFoundResponse},
				{method: "ledger", body: rpcLedgerResponse(20)},
				{method: "tx", body: rpcTxnNotFoundResponse},
				{method: "ledger", body: rpcLedgerResponse(21)},
				{method: "tx", body: rpcTxnNotFoundResponse},
			},
			wantError: ErrTransactionExpired,
		},
		{
			name:               "validated tec returns response and failure",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        1,
			steps: []rpcFinalityStep{
				{method: "tx", body: rpcValidatedTxResponse(19, "tecPATH_DRY")},
			},
			wantResult: "tecPATH_DRY",
			wantError:  ErrValidatedTransaction,
		},
		{
			name:        "missing LastLedgerSequence has bounded unknown outcome",
			maxAttempts: 2,
			steps: []rpcFinalityStep{
				{method: "tx", body: rpcTxnNotFoundResponse},
				{method: "tx", body: rpcUnvalidatedTxResponse(19)},
			},
			wantError: ErrFinalityNotDetermined,
		},
		{
			name:               "transient transport error does not become transaction failure",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        2,
			steps: []rpcFinalityStep{
				{method: "tx", err: transportTimeout},
				{method: "tx", body: rpcValidatedTxResponse(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:               "repeated transport timeout remains observable",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        2,
			steps: []rpcFinalityStep{
				{method: "tx", err: transportTimeout},
				{method: "tx", err: transportTimeout},
			},
			wantError: ErrFinalityTransport,
			wantCause: transportTimeout,
		},
		{
			name:               "caller cancellation is not expiry",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        2,
			cancelBefore:       true,
			wantError:          context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stepIndex := 0
			mockClient := &testutil.JSONRPCMockClient{}
			mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
				require.Less(t, stepIndex, len(tt.steps), "unexpected RPC request")
				var request struct {
					Method string `json:"method"`
				}
				require.NoError(t, json.NewDecoder(req.Body).Decode(&request))
				step := tt.steps[stepIndex]
				stepIndex++
				require.Equal(t, step.method, request.Method)
				if step.err != nil {
					return nil, step.err
				}
				return testutil.MockResponse(step.body, http.StatusOK, mockClient)(req)
			}
			cfg, err := NewClientConfig(
				"http://testnode/",
				WithHTTPClient(mockClient),
				WithRetryDelay(0),
				WithMaxRetries(tt.maxAttempts),
			)
			require.NoError(t, err)
			client := NewClient(cfg)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancelBefore {
				cancel()
			}
			response, err := client.waitForTransaction(ctx, "ABC", tt.lastLedgerSequence)

			if tt.wantResult == "" {
				require.Nil(t, response)
			} else {
				require.NotNil(t, response)
				require.Equal(t, tt.wantResult, response.Meta.TransactionResult)
			}
			if tt.wantError == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantError)
			}
			if tt.wantCause != nil {
				require.ErrorIs(t, err, tt.wantCause)
			}
			require.Equal(t, len(tt.steps), stepIndex)
		})
	}
}

func TestClientSubmitTxBlobAndWaitWithoutLastLedgerSequence(t *testing.T) {
	blob := signedRPCFinalityBlob(t, nil)
	methods := make([]string, 0, 3)
	mockClient := &testutil.JSONRPCMockClient{}
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		var request struct {
			Method string `json:"method"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&request))
		methods = append(methods, request.Method)
		if request.Method == "submit" {
			return testutil.MockResponse(rpcSubmitResponse("tesSUCCESS"), http.StatusOK, mockClient)(req)
		}
		return testutil.MockResponse(rpcTxnNotFoundResponse, http.StatusOK, mockClient)(req)
	}
	cfg, err := NewClientConfig(
		"http://testnode/",
		WithHTTPClient(mockClient),
		WithRetryDelay(0),
		WithMaxRetries(2),
	)
	require.NoError(t, err)

	response, err := NewClient(cfg).SubmitTxBlobAndWait(blob, false)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrFinalityNotDetermined)
	require.NotErrorIs(t, err, ErrMissingLastLedgerSequenceInTransaction)
	require.Equal(t, []string{"submit", "tx", "tx"}, methods)
}

func TestClientSubmitTxBlobAndWaitPreliminaryResultFamilies(t *testing.T) {
	lastLedger := uint32(20)
	blob := signedRPCFinalityBlob(t, &lastLedger)

	tests := []struct {
		name              string
		preliminaryResult string
		wantMonitored     bool
	}{
		{name: "tes is monitored", preliminaryResult: "tesSUCCESS", wantMonitored: true},
		{name: "ter is monitored", preliminaryResult: "terQUEUED", wantMonitored: true},
		{name: "tec is monitored", preliminaryResult: "tecPATH_DRY", wantMonitored: true},
		{name: "tef fails fast", preliminaryResult: "tefPAST_SEQ"},
		{name: "tem fails fast", preliminaryResult: "temBAD_AMOUNT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methods := make([]string, 0, 2)
			mockClient := &testutil.JSONRPCMockClient{}
			mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
				var request struct {
					Method string `json:"method"`
				}
				require.NoError(t, json.NewDecoder(req.Body).Decode(&request))
				methods = append(methods, request.Method)
				switch request.Method {
				case "submit":
					return testutil.MockResponse(rpcSubmitResponse(tt.preliminaryResult), http.StatusOK, mockClient)(req)
				case "tx":
					return testutil.MockResponse(rpcValidatedTxResponse(20, "tesSUCCESS"), http.StatusOK, mockClient)(req)
				default:
					t.Fatalf("unexpected RPC method %s", request.Method)
					return nil, nil
				}
			}
			cfg, err := NewClientConfig(
				"http://testnode/",
				WithHTTPClient(mockClient),
				WithRetryDelay(0),
				WithMaxRetries(1),
			)
			require.NoError(t, err)

			response, err := NewClient(cfg).SubmitTxBlobAndWait(blob, false)
			if tt.wantMonitored {
				require.NoError(t, err)
				require.NotNil(t, response)
				require.Equal(t, []string{"submit", "tx"}, methods)
				return
			}

			require.Nil(t, response)
			require.ErrorIs(t, err, ErrPreliminaryResult)
			var preliminaryErr *PreliminaryResultError
			require.ErrorAs(t, err, &preliminaryErr)
			require.Equal(t, tt.preliminaryResult, preliminaryErr.EngineResult)
			require.Equal(t, []string{"submit"}, methods)
		})
	}
}

func rpcLedgerResponse(index uint32) string {
	return `{"result":{"ledger_index":` + strconv.FormatUint(uint64(index), 10) + `,"validated":true}}`
}

func rpcValidatedTxResponse(index uint32, result string) string {
	return `{"result":{"ledger_index":` + strconv.FormatUint(uint64(index), 10) + `,"meta":{"TransactionResult":"` + result + `"},"validated":true}}`
}

func rpcUnvalidatedTxResponse(index uint32) string {
	return `{"result":{"ledger_index":` + strconv.FormatUint(uint64(index), 10) + `,"meta":{"TransactionResult":"terQUEUED"},"validated":false}}`
}

func rpcSubmitResponse(result string) string {
	return `{"result":{"engine_result":"` + result + `"}}`
}

func signedRPCFinalityBlob(t *testing.T, lastLedgerSequence *uint32) string {
	t.Helper()

	signer, err := wallet.FromSeed("sEdSuqBPSQaood2DmNYVkwWTn1oQTj2", "")
	require.NoError(t, err)
	tx := transaction.FlatTransaction{
		"TransactionType": "Payment",
		"Account":         signer.ClassicAddress.String(),
		"Destination":     "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		"Amount":          "1",
		"Fee":             "10",
		"Sequence":        uint32(1),
	}
	if lastLedgerSequence != nil {
		tx["LastLedgerSequence"] = *lastLedgerSequence
	}

	blob, _, err := signer.Sign(tx)
	require.NoError(t, err)
	return blob
}

const rpcTxnNotFoundResponse = `{"result":{"error":"txnNotFound"}}`
