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
			require.Equal(t, tt.want, isTransactionNotFoundError(tt.err))
		})
	}
}

func TestClientWaitForTransactionFinalityMatrix(t *testing.T) {
	lastLedger := uint32(20)

	tests := []struct {
		name               string
		lastLedgerSequence *uint32
		maxAttempts        int
		requestTimeout     time.Duration
		contextTimeout     time.Duration
		steps              []wsFinalityStep
		wantResult         string
		wantError          error
		wantCause          error
	}{
		{
			name:               "validation exactly at LastLedgerSequence",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        1,
			steps: []wsFinalityStep{
				{method: "tx", errorCode: txnNotFound},
				{method: "ledger", result: wsLedgerResult(20)},
				{method: "tx", result: wsValidatedTxResult(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:               "validation racing expiry is rechecked",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        1,
			steps: []wsFinalityStep{
				{method: "tx", errorCode: txnNotFound},
				{method: "ledger", result: wsLedgerResult(21)},
				{method: "tx", result: wsValidatedTxResult(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:               "expiry only after LastLedgerSequence",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        1,
			steps: []wsFinalityStep{
				{method: "tx", errorCode: txnNotFound},
				{method: "ledger", result: wsLedgerResult(20)},
				{method: "tx", errorCode: txnNotFound},
				{method: "ledger", result: wsLedgerResult(21)},
				{method: "tx", errorCode: txnNotFound},
			},
			wantError: ErrTransactionExpired,
		},
		{
			name:               "validated tec returns response and failure",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        1,
			steps: []wsFinalityStep{
				{method: "tx", result: wsValidatedTxResult(19, "tecPATH_DRY")},
			},
			wantResult: "tecPATH_DRY",
			wantError:  ErrValidatedTransaction,
		},
		{
			name:        "missing LastLedgerSequence has bounded unknown outcome",
			maxAttempts: 2,
			steps: []wsFinalityStep{
				{method: "tx", errorCode: txnNotFound},
				{method: "tx", result: wsUnvalidatedTxResult(19)},
			},
			wantError: ErrFinalityNotDetermined,
		},
		{
			name:               "transient transport timeout does not become transaction failure",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        2,
			requestTimeout:     10 * time.Millisecond,
			steps: []wsFinalityStep{
				{method: "tx", noResponse: true},
				{method: "tx", result: wsValidatedTxResult(20, "tesSUCCESS")},
			},
			wantResult: "tesSUCCESS",
		},
		{
			name:               "repeated transport timeout remains observable",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        2,
			requestTimeout:     10 * time.Millisecond,
			steps: []wsFinalityStep{
				{method: "tx", noResponse: true},
				{method: "tx", noResponse: true},
			},
			wantError: ErrFinalityTransport,
			wantCause: ErrRequestTimedOut,
		},
		{
			name:               "caller deadline during request is not expiry",
			lastLedgerSequence: &lastLedger,
			maxAttempts:        2,
			requestTimeout:     time.Second,
			contextTimeout:     10 * time.Millisecond,
			steps: []wsFinalityStep{
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
			require.Equal(t, int32(len(tt.steps)), requestCount.Load())
			requireNoWSFinalityServerError(t, serverErrors)
		})
	}
}

func TestClientSubmitTxBlobAndWaitWithoutLastLedgerSequence(t *testing.T) {
	blob := signedWSFinalityBlob(t, nil)
	steps := []wsFinalityStep{
		{method: "submit", result: map[string]any{"engine_result": "tesSUCCESS"}},
		{method: "tx", errorCode: txnNotFound},
		{method: "tx", errorCode: txnNotFound},
	}
	client, requestCount, serverErrors, cleanup := setupWSFinalityClient(
		t,
		steps,
		2,
		time.Second,
	)
	defer cleanup()

	response, err := client.SubmitTxBlobAndWait(blob, false)
	require.Nil(t, response)
	require.ErrorIs(t, err, ErrFinalityNotDetermined)
	require.NotErrorIs(t, err, ErrMissingLastLedgerSequenceInTransaction)
	require.Equal(t, int32(len(steps)), requestCount.Load())
	requireNoWSFinalityServerError(t, serverErrors)
}

func TestClientSubmitTxBlobAndWaitPreliminaryResultFamilies(t *testing.T) {
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
		{name: "tef fails fast", preliminaryResult: "tefPAST_SEQ"},
		{name: "tem fails fast", preliminaryResult: "temBAD_AMOUNT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := []wsFinalityStep{{
				method: "submit",
				result: map[string]any{"engine_result": tt.preliminaryResult},
			}}
			if tt.wantMonitored {
				steps = append(steps, wsFinalityStep{
					method: "tx",
					result: wsValidatedTxResult(20, "tesSUCCESS"),
				})
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

func wsUnvalidatedTxResult(index uint32) map[string]any {
	return map[string]any{
		"ledger_index": index,
		"meta":         map[string]any{"TransactionResult": "terQUEUED"},
		"validated":    false,
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
