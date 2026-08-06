package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type wsFinalityStep struct {
	method     string
	result     map[string]any
	errorCode  string
	noResponse bool
}

func TestIsTransactionNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "exact typed error",
			err:  &ErrorWebsocketClientXrplResponse{Type: txnNotFound},
			want: true,
		},
		{
			name: "substring is not enough",
			err:  &ErrorWebsocketClientXrplResponse{Type: "transport mentioned txnNotFound"},
		},
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

	tests := []struct {
		name           string
		maxAttempts    int
		requestTimeout time.Duration
		contextTimeout time.Duration
		steps          []wsFinalityStep
		wantResult     string
		wantError      error
		wantCause      error
		wantExpiryAt   uint32
	}{
		{
			name:        "validation exactly at LastLedgerSequence",
			maxAttempts: 1,
			steps: []wsFinalityStep{
				{method: "ledger", result: wsLedgerResult(20)},
				{method: "tx", result: wsValidatedTxResult(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:        "passed LastLedgerSequence expires before transaction lookup",
			maxAttempts: 1,
			steps: []wsFinalityStep{
				{method: "ledger", result: wsLedgerResult(21)},
			},
			wantError:    ErrTransactionExpired,
			wantExpiryAt: 21,
		},
		{
			name:        "expiry only after LastLedgerSequence",
			maxAttempts: 1,
			steps: []wsFinalityStep{
				{method: "ledger", result: wsLedgerResult(20)},
				{method: "tx", errorCode: txnNotFound},
				{method: "ledger", result: wsLedgerResult(21)},
			},
			wantError:    ErrTransactionExpired,
			wantExpiryAt: 21,
		},
		{
			name:        "validated tec returns response without error",
			maxAttempts: 1,
			steps: []wsFinalityStep{
				{method: "ledger", result: wsLedgerResult(20)},
				{method: "tx", result: wsValidatedTxResult(20, "tecPATH_DRY")},
			},
			wantResult: "tecPATH_DRY",
		},
		{
			name:           "transient transport timeout does not become transaction failure",
			maxAttempts:    2,
			requestTimeout: 10 * time.Millisecond,
			steps: []wsFinalityStep{
				{method: "ledger", result: wsLedgerResult(19)},
				{method: "tx", noResponse: true},
				{method: "ledger", result: wsLedgerResult(20)},
				{method: "tx", result: wsValidatedTxResult(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:           "repeated transport timeout remains observable",
			maxAttempts:    2,
			requestTimeout: 10 * time.Millisecond,
			steps: []wsFinalityStep{
				{method: "ledger", result: wsLedgerResult(19)},
				{method: "tx", noResponse: true},
				{method: "ledger", result: wsLedgerResult(19)},
				{method: "tx", noResponse: true},
			},
			wantError: ErrFinalityTransport,
			wantCause: ErrRequestTimedOut,
		},
		{
			name:           "caller deadline during request is not expiry",
			maxAttempts:    2,
			requestTimeout: time.Second,
			contextTimeout: 10 * time.Millisecond,
			steps: []wsFinalityStep{
				{method: "ledger", result: wsLedgerResult(19)},
				{method: "tx", noResponse: true},
			},
			wantError: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestTimeout := tt.requestTimeout
			if requestTimeout == 0 {
				requestTimeout = time.Second
			}
			client, requestCount, serverErrors, cleanup := setupWSFinalityClient(
				t,
				tt.steps,
				tt.maxAttempts,
				requestTimeout,
			)
			defer cleanup()

			ctx := context.Background()
			cancel := func() {}
			if tt.contextTimeout > 0 {
				ctx, cancel = context.WithTimeout(ctx, tt.contextTimeout)
			}
			defer cancel()

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
			require.Equal(t, int32(len(tt.steps)), requestCount.Load())
			requireNoWSFinalityServerError(t, serverErrors)
		})
	}
}

func TestClientSubmitTxBlobAndWaitRequiresLastLedgerSequence(t *testing.T) {
	blob := signedWSFinalityBlob(t, nil)
	client, requestCount, serverErrors, cleanup := setupWSFinalityClient(t, nil, 2, time.Second)
	defer cleanup()

	response, err := client.SubmitTxBlobAndWait(blob, false)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrMissingLastLedgerSequenceInTransaction)
	require.Zero(t, requestCount.Load())
	requireNoWSFinalityServerError(t, serverErrors)
}

func TestClientSubmitTxBlobAndWaitRejectsNegativePollInterval(t *testing.T) {
	lastLedger := uint32(20)
	blob := signedWSFinalityBlob(t, &lastLedger)
	client, requestCount, serverErrors, cleanup := setupWSFinalityClient(t, nil, 2, time.Second)
	defer cleanup()
	client.cfg.retryDelay = -time.Nanosecond

	response, err := client.SubmitTxBlobAndWait(blob, false)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrInvalidPollInterval)
	require.Zero(t, requestCount.Load())
	requireNoWSFinalityServerError(t, serverErrors)
}

func TestClientSubmitTxBlobAndWaitExpiryRetainsPreliminaryResult(t *testing.T) {
	const preliminaryResult = "terQUEUED"
	lastLedger := uint32(20)
	blob := signedWSFinalityBlob(t, &lastLedger)
	steps := []wsFinalityStep{
		{
			method: "submit",
			result: map[string]any{
				"engine_result":          preliminaryResult,
				"engine_result_message":  "queued",
				"validated_ledger_index": uint32(17),
			},
		},
		{method: "ledger", result: wsLedgerResult(21)},
	}
	client, requestCount, serverErrors, cleanup := setupWSFinalityClient(t, steps, 1, time.Second)
	defer cleanup()

	response, err := client.SubmitTxBlobAndWait(blob, false)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrTransactionExpired)
	var expiryErr *TransactionExpiredError
	require.ErrorAs(t, err, &expiryErr)
	require.Equal(t, lastLedger, expiryErr.LastLedgerSequence)
	require.Equal(t, uint32(21), expiryErr.ValidatedLedger)
	require.Equal(t, preliminaryResult, expiryErr.PreliminaryResult)
	require.Equal(t, int32(len(steps)), requestCount.Load())
	requireNoWSFinalityServerError(t, serverErrors)
}

func TestClientSubmitTxBlobAndWaitPreliminaryResultFamilies(t *testing.T) {
	const resultMessage = "preliminary result message"
	lastLedger := uint32(20)
	blob := signedWSFinalityBlob(t, &lastLedger)

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
			steps := []wsFinalityStep{{
				method: "submit",
				result: map[string]any{
					"engine_result":         tt.preliminaryResult,
					"engine_result_message": resultMessage,
				},
			}}
			if tt.wantMonitored {
				steps = append(
					steps,
					wsFinalityStep{method: "ledger", result: wsLedgerResult(20)},
					wsFinalityStep{method: "tx", result: wsValidatedTxResult(20, "tesSUCCESS")},
				)
			}

			client, requestCount, serverErrors, cleanup := setupWSFinalityClient(
				t,
				steps,
				1,
				time.Second,
			)
			defer cleanup()

			response, err := client.SubmitTxBlobAndWait(blob, false)
			if tt.wantMonitored {
				require.NoError(t, err)
				require.NotNil(t, response)
			} else {
				require.Nil(t, response)
				require.ErrorIs(t, err, ErrPreliminaryResult)
				var preliminaryErr *PreliminaryResultError
				require.ErrorAs(t, err, &preliminaryErr)
				require.Equal(t, tt.preliminaryResult, preliminaryErr.EngineResult)
				require.Equal(t, resultMessage, preliminaryErr.EngineResultMessage)
				require.ErrorContains(t, err, resultMessage)
			}
			require.Equal(t, int32(len(steps)), requestCount.Load())
			requireNoWSFinalityServerError(t, serverErrors)
		})
	}
}

func setupWSFinalityClient(
	t *testing.T,
	steps []wsFinalityStep,
	maxAttempts int,
	requestTimeout time.Duration,
) (*Client, *atomic.Int32, <-chan error, func()) {
	t.Helper()

	var requestCount atomic.Int32
	serverErrors := make(chan error, 1)
	mockServer := testutil.MockWebSocketServer{}
	server := mockServer.TestWebSocketServer(func(conn *ws.Conn) {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var request struct {
				ID      uint64 `json:"id"`
				Command string `json:"command"`
			}
			if err := json.Unmarshal(message, &request); err != nil {
				select {
				case serverErrors <- err:
				default:
				}
				return
			}

			stepIndex := int(requestCount.Add(1)) - 1
			if stepIndex >= len(steps) {
				select {
				case serverErrors <- fmt.Errorf("unexpected WebSocket request %s", request.Command):
				default:
				}
				return
			}
			step := steps[stepIndex]
			if request.Command != step.method {
				select {
				case serverErrors <- fmt.Errorf("request %d: got method %s, want %s", stepIndex, request.Command, step.method):
				default:
				}
			}
			if step.noResponse {
				continue
			}

			response := map[string]any{
				"id":     request.ID,
				"status": "success",
				"type":   "response",
				"result": step.result,
			}
			if step.errorCode != "" {
				response["status"] = "error"
				response["error"] = step.errorCode
			}
			if err := conn.WriteJSON(response); err != nil {
				return
			}
		}
	})

	host, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)
	config := NewClientConfig().
		WithHost(host).
		WithNetworkIdentity(0, "1.12.0").
		WithMaxRetries(maxAttempts).
		WithRetryDelay(0).
		WithTimeout(requestTimeout)
	client := NewClient(config)
	require.NoError(t, client.Connect())

	cleanup := func() {
		if client.IsConnected() {
			require.NoError(t, client.Disconnect())
		}
		server.Close()
	}
	return client, &requestCount, serverErrors, cleanup
}

func requireNoWSFinalityServerError(t *testing.T, serverErrors <-chan error) {
	t.Helper()
	select {
	case err := <-serverErrors:
		require.NoError(t, err)
	default:
	}
}

func wsLedgerResult(index uint32) map[string]any {
	return map[string]any{"ledger_index": index, "validated": true}
}

func wsValidatedTxResult(index uint32, result string) map[string]any {
	return map[string]any{
		"ledger_index": index,
		"meta":         map[string]any{"TransactionResult": result},
		"validated":    true,
	}
}

func signedWSFinalityBlob(t *testing.T, lastLedgerSequence *uint32) string {
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
