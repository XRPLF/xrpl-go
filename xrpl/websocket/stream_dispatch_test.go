package websocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	streamtypes "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/testutil"
	gorillaws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestClient_HandleMessageDispatchesExportedStreams(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		register func(*Client, chan<- bool)
	}{
		{
			name:    "ledger closed",
			message: `{"type":"ledgerClosed","ledger_index":11}`,
			register: func(c *Client, received chan<- bool) {
				c.OnLedgerClosed(func(event *streamtypes.LedgerStream) {
					received <- event.Type == streamtypes.LedgerStreamType && event.LedgerIndex == 11
				})
			},
		},
		{
			name:    "validation received",
			message: `{"type":"validationReceived","ledger_index":12}`,
			register: func(c *Client, received chan<- bool) {
				c.OnValidationReceived(func(event *streamtypes.ValidationStream) {
					received <- event.Type == streamtypes.ValidationStreamType && event.LedgerIndex == 12
				})
			},
		},
		{
			name:    "transaction",
			message: `{"type":"transaction","engine_result":"tesSUCCESS","tx_json":{"TransactionType":"OfferCreate"}}`,
			register: func(c *Client, received chan<- bool) {
				c.OnTransactions(func(event *streamtypes.TransactionStream) {
					received <- event.Type == streamtypes.TransactionStreamType && event.EngineResult == "tesSUCCESS"
				})
			},
		},
		{
			name:    "order book transaction",
			message: `{"type":"transaction","engine_result":"tesSUCCESS","tx_json":{"TransactionType":"OfferCreate","owner_funds":"100"}}`,
			register: func(c *Client, received chan<- bool) {
				c.OnOrderBook(func(event *streamtypes.OrderBookStream) {
					ownerFunds, ok := event.Transaction["owner_funds"]
					received <- event.Type == streamtypes.OrderBookStreamType &&
						event.EngineResult == "tesSUCCESS" && ok && ownerFunds == "100"
				})
			},
		},
		{
			name:    "peer status change",
			message: `{"type":"peerStatusChange","action":"ACCEPTED_LEDGER","ledger_index":13}`,
			register: func(c *Client, received chan<- bool) {
				c.OnPeerStatusChange(func(event *streamtypes.PeerStatusStream) {
					received <- event.Type == streamtypes.PeerStatusStreamType && event.Action == streamtypes.PeerStatusAcceptedLedger && event.LedgerIndex == 13
				})
			},
		},
		{
			name:    "book changes",
			message: `{"type":"bookChanges","ledger_index":14,"changes":[{"currency_a":"XRP_drops","currency_b":"issuer/USD","volume_a":"1","volume_b":"2","high":"3","low":"4","open":"5","close":"6"}]}`,
			register: func(c *Client, received chan<- bool) {
				c.OnBookChanges(func(event *streamtypes.BookChangesStream) {
					received <- event.Type == streamtypes.BookChangesStreamType &&
						event.LedgerIndex == 14 && len(event.Changes) == 1 &&
						event.Changes[0].CurrencyA == "XRP_drops" && event.Changes[0].VolumeA == "1"
				})
			},
		},
		{
			name:    "consensus phase",
			message: `{"type":"consensusPhase","consensus":"accepted"}`,
			register: func(c *Client, received chan<- bool) {
				c.OnConsensusPhase(func(event *streamtypes.ConsensusStream) {
					received <- event.Type == streamtypes.ConsensusStreamType && event.Consensus == "accepted"
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(*NewClientConfig())
			ctx := client.resetLifecycle()
			defer client.cancelLifecycle()

			received := make(chan bool, 1)
			tt.register(client, received)
			client.handleMessage(ctx, []byte(tt.message))

			select {
			case decoded := <-received:
				require.True(t, decoded, "handler received incorrectly decoded event")
			case <-time.After(time.Second):
				t.Fatal("timed out waiting for stream handler")
			}
		})
	}
}

func TestClient_StreamHandlerOrderingAndBackpressure(t *testing.T) {
	client := NewClient(*NewClientConfig())
	ctx := client.resetLifecycle()
	defer client.cancelLifecycle()

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	defer func() {
		select {
		case <-releaseFirst:
		default:
			close(releaseFirst)
		}
	}()
	handled := make(chan uint64, 2)
	client.OnLedgerClosed(func(event *streamtypes.LedgerStream) {
		if event.LedgerIndex == 1 {
			close(firstStarted)
			<-releaseFirst
		}
		handled <- uint64(event.LedgerIndex)
	})

	firstDispatchDone := make(chan struct{})
	go func() {
		client.handleMessage(ctx, []byte(`{"type":"ledgerClosed","ledger_index":1}`))
		close(firstDispatchDone)
	}()

	select {
	case <-firstStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first handler call")
	}
	select {
	case <-firstDispatchDone:
	case <-time.After(time.Second):
		t.Fatal("reader-side dispatch waited for the handler to finish")
	}

	secondDispatchDone := make(chan struct{})
	go func() {
		client.handleMessage(ctx, []byte(`{"type":"ledgerClosed","ledger_index":2}`))
		close(secondDispatchDone)
	}()

	select {
	case <-secondDispatchDone:
		t.Fatal("second same-stream dispatch bypassed handler backpressure")
	case <-time.After(50 * time.Millisecond):
	}

	close(releaseFirst)
	select {
	case <-secondDispatchDone:
	case <-time.After(time.Second):
		t.Fatal("second dispatch remained blocked after handler resumed")
	}

	handledInOrder := make([]uint64, 0, 2)
	for range 2 {
		select {
		case ledgerIndex := <-handled:
			handledInOrder = append(handledInOrder, ledgerIndex)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for ordered handler calls")
		}
	}
	require.Equal(t, []uint64{1, 2}, handledInOrder)
}

func TestClient_StreamHandlersRunConcurrentlyAcrossStreams(t *testing.T) {
	client := NewClient(*NewClientConfig())
	ctx := client.resetLifecycle()
	defer client.cancelLifecycle()

	ledgerStarted := make(chan struct{})
	releaseLedger := make(chan struct{})
	defer close(releaseLedger)
	client.OnLedgerClosed(func(*streamtypes.LedgerStream) {
		close(ledgerStarted)
		<-releaseLedger
	})

	bookChangesReceived := make(chan struct{})
	client.OnBookChanges(func(*streamtypes.BookChangesStream) {
		close(bookChangesReceived)
	})

	client.handleMessage(ctx, []byte(`{"type":"ledgerClosed","ledger_index":1}`))
	select {
	case <-ledgerStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ledger handler")
	}

	client.handleMessage(ctx, []byte(`{"type":"bookChanges","ledger_index":2}`))
	select {
	case <-bookChangesReceived:
	case <-time.After(time.Second):
		t.Fatal("different stream handler was blocked by ledger handler")
	}
}

func TestClient_StreamHandlerSingleDeliveryAcrossRepeatedReconnects(t *testing.T) {
	const reconnectCount = 2
	var connectionCount atomic.Int32
	upgrader := gorillaws.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionNumber := connectionCount.Add(1)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		message := fmt.Appendf(nil, `{"type":"ledgerClosed","ledger_index":%d}`, connectionNumber)
		if err := conn.WriteMessage(gorillaws.TextMessage, message); err != nil {
			return
		}
		if connectionNumber <= reconnectCount {
			_ = conn.WriteMessage(
				gorillaws.CloseMessage,
				gorillaws.FormatCloseMessage(gorillaws.CloseNormalClosure, ""),
			)
			return
		}

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer server.Close()

	url, err := testutil.ConvertHTTPToWS(server.URL)
	require.NoError(t, err)

	client := NewClient(withReconnectDelays(
		NewClientConfig().
			WithHost(url).
			WithTimeout(time.Second).
			WithMaxReconnects(reconnectCount).
			WithNetworkIdentity(0, "2.0.0"),
		time.Millisecond,
		time.Millisecond,
	))

	received := make(chan uint64, reconnectCount+2)
	client.OnLedgerClosed(func(event *streamtypes.LedgerStream) {
		received <- uint64(event.LedgerIndex)
	})

	require.NoError(t, client.Connect())
	defer client.Disconnect()

	for expected := uint64(1); expected <= reconnectCount+1; expected++ {
		select {
		case actual := <-received:
			require.Equal(t, expected, actual)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for reconnect stream event")
		}
	}

	select {
	case <-received:
		t.Fatal("stream event was delivered more than once")
	case <-time.After(50 * time.Millisecond):
	}

	// An explicit reconnect restarts the reader and handler runners. The
	// registration persists, but exactly one runner must receive the event.
	require.NoError(t, client.Disconnect())
	require.NoError(t, client.Connect())
	select {
	case actual := <-received:
		require.Equal(t, uint64(reconnectCount+2), actual)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for reader-restart stream event")
	}
	select {
	case <-received:
		t.Fatal("reader-restart stream event was delivered more than once")
	case <-time.After(50 * time.Millisecond):
	}
	require.Equal(t, int32(reconnectCount+2), connectionCount.Load())
}
