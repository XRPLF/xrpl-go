package websocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

const (
	deleteAccount = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	deleteSponsor = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
)

func TestAccountDeleteUsesValidatedInfoAfterSequenceAutofill(t *testing.T) {
	tagged, err := addresscodec.ClassicAddressToXAddress(deleteSponsor, 17, true, false)
	require.NoError(t, err)
	for _, multisigned := range []bool{false, true} {
		t.Run(fmt.Sprintf("multisigned=%t", multisigned), func(t *testing.T) {
			cl, requests := newAccountDeleteTestClient(t, []map[string]any{
				{"account_data": map[string]any{"Sequence": 20, "Sponsor": deleteSponsor}},
				{"account_objects": []any{}},
				{"account_data": map[string]any{"Sequence": 19, "Sponsor": deleteAccount}},
			})
			tx := transaction.FlatTransaction{
				"TransactionType": transaction.AccountDeleteTx,
				"Account":         deleteAccount, "Destination": tagged,
				"Fee": "10", "LastLedgerSequence": uint32(100),
			}
			before := clientinternal.CloneTransaction(tx)
			var err error
			if multisigned {
				err = cl.AutofillMultisigned(&tx, 2)
			} else {
				err = cl.Autofill(&tx)
			}
			require.ErrorIs(t, err, ErrAccountDeleteSponsorMismatch)
			require.Equal(t, transaction.FlatTransaction(before), tx)
			got := requests()
			require.Len(t, got, 3)
			require.Equal(t, "account_info", got[0]["command"])
			require.Equal(t, "current", got[0]["ledger_index"])
			require.Equal(t, "account_objects", got[1]["command"])
			require.Equal(t, true, got[1]["deletion_blockers_only"])
			require.Equal(t, "validated", got[1]["ledger_index"])
			require.Equal(t, "account_info", got[2]["command"])
			require.Equal(t, "validated", got[2]["ledger_index"])
			for _, req := range got {
				require.Equal(t, deleteAccount, req["account"])
			}
		})
	}
}

func newAccountDeleteTestClient(t *testing.T, results []map[string]any) (*Client, func() []map[string]any) {
	t.Helper()
	var mu sync.Mutex
	var requests []map[string]any
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = conn.Close() }()
		for {
			var req map[string]any
			if err := conn.ReadJSON(&req); err != nil {
				return
			}
			mu.Lock()
			index := len(requests)
			requests = append(requests, req)
			mu.Unlock()
			if index >= len(results) {
				t.Errorf("unexpected request: %v", req)
				return
			}
			response := map[string]any{"id": req["id"], "status": "success", "type": "response", "result": results[index]}
			if err := conn.WriteJSON(response); err != nil {
				t.Error(err)
				return
			}
		}
	}))
	t.Cleanup(s.Close)
	cl := NewClient(NewClientConfig().WithHost("ws" + strings.TrimPrefix(s.URL, "http")))
	setTrustedTestNetworkIdentity(cl, 0)
	require.NoError(t, cl.Connect())
	t.Cleanup(func() { _ = cl.Disconnect() })
	return cl, func() []map[string]any {
		mu.Lock()
		defer mu.Unlock()
		return append([]map[string]any(nil), requests...)
	}
}
