package builder

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	xrplhash "github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/stretchr/testify/require"
)

// Both shipped clients must keep satisfying LedgerQuerier, because the entryNotFound
// classification below is pinned against the error each one actually returns.
var (
	_ LedgerQuerier = (*rpc.Client)(nil)
	_ LedgerQuerier = (*websocket.Client)(nil)
)

// queuedRPCTransport replays canned JSON-RPC responses in order, so a test can drive a
// real rpc.Client without a node.
type queuedRPCTransport struct {
	responses []string
}

func (t *queuedRPCTransport) Do(*http.Request) (*http.Response, error) {
	if len(t.responses) == 0 {
		return nil, fmt.Errorf("unexpected RPC request")
	}
	response := t.responses[0]
	t.responses = t.responses[1:]
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(response)),
		Header:     make(http.Header),
	}, nil
}

// TestGetMPTokenStateClassifiesRPCEntryNotFound pins the classification against a real
// rpc.Client rather than a hand-built error, because ErrReceiverNotOptedIn depends on the
// node's "entryNotFound" reaching the builder as exactly that string.
func TestGetMPTokenStateClassifiesRPCEntryNotFound(t *testing.T) {
	transport := &queuedRPCTransport{responses: []string{
		`{"result":{"error":"entryNotFound"}}`,
	}}
	config, err := rpc.NewClientConfig("http://testnode/", rpc.WithHTTPClient(transport))
	require.NoError(t, err)
	client := rpc.NewClient(config)

	_, err = getMPTokenState(snapshotFor(client), issuanceState{}, testIssuanceID, testAccount)
	require.ErrorIs(t, err, ErrMPTokenNotFound)
	require.NotErrorIs(t, err, ErrLedgerQuery)
	require.Empty(t, transport.responses)
}

func TestGetMPTokenStateRejectsRPCNullBalanceVersion(t *testing.T) {
	index, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)
	transport := &queuedRPCTransport{responses: []string{
		fmt.Sprintf(`{"result":{"index":"%s","ledger_hash":"%s","ledger_index":%d,"node":{"LedgerEntryType":"MPToken","HolderEncryptionKey":"key","ConfidentialBalanceSpending":"ciphertext","IssuerEncryptedBalance":"mirror","ConfidentialBalanceVersion":null},"validated":true}}`, index, mockLedgerHash, mockLedgerIndex),
	}}
	config, err := rpc.NewClientConfig("http://testnode/", rpc.WithHTTPClient(transport))
	require.NoError(t, err)
	client := rpc.NewClient(config)

	_, err = getMPTokenState(snapshotFor(client), issuanceState{}, testIssuanceID, testAccount)
	require.ErrorIs(t, err, ErrInvalidLedgerState)
	require.Empty(t, transport.responses)
}

// TestGetMPTokenStateClassifiesWebsocketEntryNotFound pins the same classification against
// the error the WebSocket client returns. It carries the node's string in a struct field
// instead of a wrapped message, so it reaches the builder by a different path than the
// rpc.Client case above.
func TestGetMPTokenStateClassifiesWebsocketEntryNotFound(t *testing.T) {
	index, err := xrplhash.MPToken(testIssuanceID, testAccount)
	require.NoError(t, err)

	tests := []struct {
		name     string
		queryErr error
		wantErr  error
		notErr   error
	}{
		{
			name:     "entryNotFound",
			queryErr: &websocket.ErrorWebsocketClientXrplResponse{Type: ledgerEntryNotFound},
			wantErr:  ErrMPTokenNotFound,
			notErr:   ErrLedgerQuery,
		},
		{
			name:     "wrapped entryNotFound",
			queryErr: fmt.Errorf("request failed: %w", &websocket.ErrorWebsocketClientXrplResponse{Type: ledgerEntryNotFound}),
			wantErr:  ErrMPTokenNotFound,
			notErr:   ErrLedgerQuery,
		},
		{
			name:     "other xrpl error",
			queryErr: &websocket.ErrorWebsocketClientXrplResponse{Type: "invalidParams"},
			wantErr:  ErrLedgerQuery,
			notErr:   ErrMPTokenNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &mockQuerier{entryErrs: map[string]error{index: tt.queryErr}}

			_, err := getMPTokenState(snapshotFor(q), issuanceState{}, testIssuanceID, testAccount)
			require.ErrorIs(t, err, tt.wantErr)
			require.NotErrorIs(t, err, tt.notErr)
		})
	}
}
