package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

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
			got := isTransactionNotFoundError(tt.err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestClientWaitForTransactionFinalityMatrix(t *testing.T) {
	const (
		lastLedger        = uint32(20)
		preliminaryResult = "tesSUCCESS"
	)
	transportTimeout := context.DeadlineExceeded

	tests := []struct {
		name         string
		maxAttempts  int
		steps        []rpcFinalityStep
		cancelBefore bool
		wantResult   string
		wantError    error
		wantCause    error
		wantExpiryAt uint32
	}{
		{
			name:        "validation exactly at LastLedgerSequence",
			maxAttempts: 1,
			steps: []rpcFinalityStep{
				{method: "ledger", body: rpcLedgerResponse(20)},
				{method: "tx", body: rpcValidatedTxResponse(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:        "passed LastLedgerSequence expires before transaction lookup",
			maxAttempts: 1,
			steps: []rpcFinalityStep{
				{method: "ledger", body: rpcLedgerResponse(21)},
			},
			wantError:    ErrTransactionExpired,
			wantExpiryAt: 21,
		},
		{
			name:        "expiry only after LastLedgerSequence",
			maxAttempts: 1,
			steps: []rpcFinalityStep{
				{method: "ledger", body: rpcLedgerResponse(20)},
				{method: "tx", body: rpcTxnNotFoundResponse},
				{method: "ledger", body: rpcLedgerResponse(21)},
			},
			wantError:    ErrTransactionExpired,
			wantExpiryAt: 21,
		},
		{
			name:        "validated tec returns response without error",
			maxAttempts: 1,
			steps: []rpcFinalityStep{
				{method: "ledger", body: rpcLedgerResponse(20)},
				{method: "tx", body: rpcValidatedTxResponse(20, "tecPATH_DRY")},
			},
			wantResult: "tecPATH_DRY",
		},
		{
			name:        "transient transport error does not become transaction failure",
			maxAttempts: 2,
			steps: []rpcFinalityStep{
				{method: "ledger", body: rpcLedgerResponse(19)},
				{method: "tx", err: transportTimeout},
				{method: "ledger", body: rpcLedgerResponse(20)},
				{method: "tx", body: rpcValidatedTxResponse(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:        "repeated transport timeout remains observable",
			maxAttempts: 2,
			steps: []rpcFinalityStep{
				{method: "ledger", body: rpcLedgerResponse(19)},
				{method: "tx", err: transportTimeout},
				{method: "ledger", body: rpcLedgerResponse(19)},
				{method: "tx", err: transportTimeout},
			},
			wantError: ErrFinalityTransport,
			wantCause: transportTimeout,
		},
		{
			name:         "caller cancellation is not expiry",
			maxAttempts:  2,
			cancelBefore: true,
			wantError:    context.Canceled,
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
			response, err := client.waitForTransaction(ctx, "ABC", lastLedger, preliminaryResult)

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
			if tt.wantExpiryAt != 0 {
				var expiryErr *TransactionExpiredError
				require.ErrorAs(t, err, &expiryErr)
				require.Equal(t, lastLedger, expiryErr.LastLedgerSequence)
				require.Equal(t, tt.wantExpiryAt, expiryErr.ValidatedLedger)
				require.Equal(t, preliminaryResult, expiryErr.PreliminaryResult)
			}
			require.Equal(t, len(tt.steps), stepIndex)
		})
	}
}

func TestClientSubmitTxBlobAndWaitRequiresLastLedgerSequence(t *testing.T) {
	blob := signedRPCFinalityBlob(t, nil)
	requestCount := 0
	mockClient := &testutil.JSONRPCMockClient{}
	mockClient.DoFunc = func(*http.Request) (*http.Response, error) {
		requestCount++
		return nil, nil
	}
	cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mockClient))
	require.NoError(t, err)

	response, err := NewClient(cfg).SubmitTxBlobAndWait(blob, false)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrMissingLastLedgerSequenceInTransaction)
	require.Zero(t, requestCount)
}

func TestClientSubmitTxBlobAndWaitRejectsNegativePollInterval(t *testing.T) {
	lastLedger := uint32(20)
	blob := signedRPCFinalityBlob(t, &lastLedger)
	requestCount := 0
	mockClient := &testutil.JSONRPCMockClient{}
	mockClient.DoFunc = func(*http.Request) (*http.Response, error) {
		requestCount++
		return nil, nil
	}
	cfg, err := NewClientConfig(
		"http://testnode/",
		WithHTTPClient(mockClient),
		WithRetryDelay(-time.Nanosecond),
	)
	require.NoError(t, err)

	response, err := NewClient(cfg).SubmitTxBlobAndWait(blob, false)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrInvalidPollInterval)
	require.Zero(t, requestCount)
}

func TestClientSubmitTxBlobAndWaitExpiryRetainsPreliminaryResult(t *testing.T) {
	const preliminaryResult = "terQUEUED"
	lastLedger := uint32(20)
	blob := signedRPCFinalityBlob(t, &lastLedger)
	steps := []rpcFinalityStep{
		{method: "submit", body: rpcSubmitResponse(preliminaryResult, "queued")},
		{method: "ledger", body: rpcLedgerResponse(21)},
	}
	stepIndex := 0
	mockClient := &testutil.JSONRPCMockClient{}
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		require.Less(t, stepIndex, len(steps), "unexpected RPC request")
		var request struct {
			Method string `json:"method"`
		}
		require.NoError(t, json.NewDecoder(req.Body).Decode(&request))
		step := steps[stepIndex]
		stepIndex++
		require.Equal(t, step.method, request.Method)
		return testutil.MockResponse(step.body, http.StatusOK, mockClient)(req)
	}
	cfg, err := NewClientConfig(
		"http://testnode/",
		WithHTTPClient(mockClient),
		WithRetryDelay(0),
		WithMaxRetries(1),
	)
	require.NoError(t, err)

	response, err := NewClient(cfg).SubmitTxBlobAndWait(blob, false)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrTransactionExpired)
	var expiryErr *TransactionExpiredError
	require.ErrorAs(t, err, &expiryErr)
	require.Equal(t, lastLedger, expiryErr.LastLedgerSequence)
	require.Equal(t, uint32(21), expiryErr.ValidatedLedger)
	require.Equal(t, preliminaryResult, expiryErr.PreliminaryResult)
	require.Equal(t, len(steps), stepIndex)
}

func TestClientSubmitTxBlobAndWaitPreliminaryResultFamilies(t *testing.T) {
	const resultMessage = "preliminary result message"
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
		{name: "tef is monitored", preliminaryResult: "tefPAST_SEQ", wantMonitored: true},
		{name: "tel is monitored", preliminaryResult: "telINSUF_FEE_P", wantMonitored: true},
		{name: "unknown is monitored", preliminaryResult: "customResult", wantMonitored: true},
		{name: "tem fails fast", preliminaryResult: "temBAD_AMOUNT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methods := make([]string, 0, 3)
			mockClient := &testutil.JSONRPCMockClient{}
			mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
				var request struct {
					Method string `json:"method"`
				}
				require.NoError(t, json.NewDecoder(req.Body).Decode(&request))
				methods = append(methods, request.Method)
				switch request.Method {
				case "submit":
					return testutil.MockResponse(
						rpcSubmitResponse(tt.preliminaryResult, resultMessage),
						http.StatusOK,
						mockClient,
					)(req)
				case "ledger":
					return testutil.MockResponse(rpcLedgerResponse(20), http.StatusOK, mockClient)(req)
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
				require.Equal(t, []string{"submit", "ledger", "tx"}, methods)
				return
			}

			require.Nil(t, response)
			require.ErrorIs(t, err, ErrPreliminaryResult)
			var preliminaryErr *PreliminaryResultError
			require.ErrorAs(t, err, &preliminaryErr)
			require.Equal(t, tt.preliminaryResult, preliminaryErr.EngineResult)
			require.Equal(t, resultMessage, preliminaryErr.EngineResultMessage)
			require.ErrorContains(t, err, resultMessage)
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

func rpcSubmitResponse(result, message string) string {
	return `{"result":{"engine_result":"` + result + `","engine_result_message":"` + message + `","validated_ledger_index":17}}`
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
