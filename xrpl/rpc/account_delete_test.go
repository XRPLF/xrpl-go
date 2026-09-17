package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
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
	var requests []map[string]any
	mc := &testutil.JSONRPCMockClient{}
	mc.DoFunc = func(req *http.Request) (*http.Response, error) {
		var wire struct {
			Method string           `json:"method"`
			Params []map[string]any `json:"params"`
		}
		if err := json.NewDecoder(req.Body).Decode(&wire); err != nil {
			return nil, err
		}
		if len(wire.Params) != 1 {
			return nil, fmt.Errorf("unexpected params: %v", wire.Params)
		}
		params := wire.Params[0]
		params["command"] = wire.Method
		index := len(requests)
		requests = append(requests, params)
		if index >= len(results) {
			return nil, fmt.Errorf("unexpected request: %v", params)
		}
		body, err := json.Marshal(map[string]any{"result": results[index]})
		if err != nil {
			return nil, err
		}
		return testutil.MockResponse(string(body), 200, mc)(req)
	}
	cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc), WithNetworkIdentity(0, "1.12.0"))
	require.NoError(t, err)
	return NewClient(cfg), func() []map[string]any { return requests }
}
