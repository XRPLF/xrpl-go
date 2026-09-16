package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	deleteAccount = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	deleteSponsor = "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59"
)

func TestAccountDeleteSponsorshipBlockers(t *testing.T) {
	x, err := addresscodec.ClassicAddressToXAddress(deleteSponsor, 0, false, false)
	require.NoError(t, err)
	tagged, err := addresscodec.ClassicAddressToXAddress(deleteSponsor, 17, true, true)
	require.NoError(t, err)
	tests := []struct {
		name        string
		root        map[string]any
		destination string
		wantErr     error
		wantContext string
	}{
		{name: "ordinary account", root: map[string]any{}},
		{name: "owner count zero", root: map[string]any{"SponsoringOwnerCount": 0}, wantErr: errAccountHasSponsorshipObligations},
		{name: "owner count nonzero", root: map[string]any{"SponsoringOwnerCount": 2}, wantErr: errAccountHasSponsorshipObligations},
		{name: "account count zero", root: map[string]any{"SponsoringAccountCount": 0}, wantErr: errAccountHasSponsorshipObligations},
		{name: "account count nonzero", root: map[string]any{"SponsoringAccountCount": 2}, wantErr: errAccountHasSponsorshipObligations},
		{name: "sponsored owner count zero", root: map[string]any{"SponsoredOwnerCount": 0}},
		{name: "sponsored owner count nonzero", root: map[string]any{"SponsoredOwnerCount": 2}},
		{name: "matching sponsor", root: map[string]any{"Sponsor": deleteSponsor}, destination: deleteSponsor},
		{name: "equivalent X-address", root: map[string]any{"Sponsor": deleteSponsor}, destination: x},
		{name: "equivalent tagged testnet X-address", root: map[string]any{"Sponsor": deleteSponsor}, destination: tagged},
		{name: "sponsor mismatch", root: map[string]any{"Sponsor": deleteSponsor}, destination: deleteAccount, wantErr: errAccountDeleteSponsorMismatch},
		{name: "missing destination", root: map[string]any{"Sponsor": deleteSponsor}},
		{name: "missing destination still checks counters", root: map[string]any{"Sponsor": deleteSponsor, "SponsoringOwnerCount": 0}, wantErr: errAccountHasSponsorshipObligations},
		{name: "malformed sponsor", root: map[string]any{"Sponsor": "invalid"}, destination: deleteSponsor, wantErr: addresscodec.ErrInvalidAddressFormat, wantContext: "decode account Sponsor:"},
		{name: "malformed destination", root: map[string]any{"Sponsor": deleteSponsor}, destination: "invalid", wantErr: addresscodec.ErrInvalidAddressFormat, wantContext: "decode AccountDelete Destination:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, requests := newAccountDeleteTestClient(t, []map[string]any{
				{"account_objects": []any{}}, {"account_data": tt.root},
			})
			err := cl.checkAccountDeleteBlockers(context.Background(), types.Address(deleteAccount), tt.destination)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				if tt.wantContext != "" {
					require.ErrorContains(t, err, tt.wantContext)
				}
			} else {
				require.NoError(t, err)
			}
			got := requests()
			require.Len(t, got, 2)
			require.Equal(t, "account_objects", got[0]["command"])
			require.Equal(t, true, got[0]["deletion_blockers_only"])
			require.Equal(t, "account_info", got[1]["command"])
			for _, req := range got {
				require.Equal(t, deleteAccount, req["account"])
				require.Equal(t, "validated", req["ledger_index"])
			}
		})
	}
}

func TestAccountDeleteObjectsStillBlock(t *testing.T) {
	for _, kind := range []string{"Escrow", "Sponsorship"} {
		t.Run(kind, func(t *testing.T) {
			cl, requests := newAccountDeleteTestClient(t, []map[string]any{
				{"account_objects": []any{map[string]any{"LedgerEntryType": kind, "Sponsor": deleteSponsor, "Sponsee": deleteAccount}}},
			})
			err := cl.checkAccountDeleteBlockers(context.Background(), types.Address(deleteAccount), deleteSponsor)
			require.ErrorIs(t, err, ErrAccountCannotBeDeleted)
			require.Len(t, requests(), 1)
		})
	}
}

func deletionTransaction() transaction.FlatTransaction {
	return transaction.FlatTransaction{
		"TransactionType": transaction.AccountDeleteTx,
		"Account":         deleteAccount, "Destination": deleteSponsor,
		"Fee": "10", "Sequence": uint32(1), "LastLedgerSequence": uint32(100),
	}
}

func TestAccountDeleteAutofillDestinations(t *testing.T) {
	x, err := addresscodec.ClassicAddressToXAddress(deleteSponsor, 0, false, false)
	require.NoError(t, err)
	tagged, err := addresscodec.ClassicAddressToXAddress(deleteSponsor, 17, true, false)
	require.NoError(t, err)
	zeroTagged, err := addresscodec.ClassicAddressToXAddress(deleteSponsor, 0, true, true)
	require.NoError(t, err)
	tests := []struct {
		name        string
		destination any
		omit        bool
		explicitTag *uint32
		wantTag     *uint32
		wantErr     error
	}{
		{name: "classic", destination: deleteSponsor},
		{name: "named classic", destination: types.Address(deleteSponsor)},
		{name: "X-address", destination: x},
		{name: "named X-address", destination: types.Address(x)},
		{name: "tagged X-address", destination: tagged, wantTag: uint32Ptr(17)},
		{name: "zero tagged X-address", destination: zeroTagged, wantTag: uint32Ptr(0)},
		{name: "matching explicit tag", destination: tagged, explicitTag: uint32Ptr(17), wantTag: uint32Ptr(17)},
		{name: "conflicting tag", destination: tagged, explicitTag: uint32Ptr(18), wantErr: clientinternal.ErrMismatchedTag},
		{name: "missing", omit: true},
		{name: "nil"},
		{name: "empty", destination: "", wantErr: clientinternal.ErrInvalidAddress},
		{name: "invalid", destination: "invalid", wantErr: clientinternal.ErrInvalidAddress},
		{name: "nonstring", destination: 123, wantErr: clientinternal.ErrAddressFieldIsNotAString},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, requests := newAccountDeleteTestClient(t, []map[string]any{
				{"account_objects": []any{}}, {"account_data": map[string]any{"Sponsor": deleteSponsor}},
			})
			tx := deletionTransaction()
			if tt.omit {
				delete(tx, "Destination")
			} else {
				tx["Destination"] = tt.destination
			}
			if tt.explicitTag != nil {
				tx["DestinationTag"] = *tt.explicitTag
			}
			before := clientinternal.CloneTransaction(tx)
			err := cl.Autofill(&tx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, transaction.FlatTransaction(before), tx)
				require.Empty(t, requests())
				return
			}
			require.NoError(t, err)
			require.Len(t, requests(), 2)
			if tt.omit {
				require.NotContains(t, tx, "Destination")
			} else if tt.destination == nil {
				require.Nil(t, tx["Destination"])
			} else {
				require.Equal(t, deleteSponsor, tx["Destination"])
			}
			if tt.wantTag != nil {
				require.Equal(t, *tt.wantTag, tx["DestinationTag"])
			} else {
				require.NotContains(t, tx, "DestinationTag")
			}
		})
	}
}

func TestAccountDeleteAutofillFailuresAreAtomic(t *testing.T) {
	x, err := addresscodec.ClassicAddressToXAddress(deleteSponsor, 17, true, false)
	require.NoError(t, err)
	tests := []struct {
		name        string
		results     []map[string]any
		wantErr     error
		wantMessage string
	}{
		{name: "obligations", results: []map[string]any{{"account_objects": []any{}}, {"account_data": map[string]any{"SponsoringAccountCount": 0}}}, wantErr: errAccountHasSponsorshipObligations},
		{name: "mismatch", results: []map[string]any{{"account_objects": []any{}}, {"account_data": map[string]any{"Sponsor": deleteAccount}}}, wantErr: errAccountDeleteSponsorMismatch},
		{name: "object query failure", results: []map[string]any{{"error": "actNotFound"}}, wantMessage: "actNotFound"},
		{name: "info query failure", results: []map[string]any{{"account_objects": []any{}}, {"error": "actNotFound"}}, wantMessage: "actNotFound"},
		{name: "info decode failure", results: []map[string]any{{"account_objects": []any{}}, {"account_data": "not an object"}}, wantMessage: "account_data"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, requests := newAccountDeleteTestClient(t, tt.results)
			tx := deletionTransaction()
			tx["Destination"] = x
			before := clientinternal.CloneTransaction(tx)
			err := cl.Autofill(&tx)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.ErrorContains(t, err, tt.wantMessage)
			}
			require.Equal(t, transaction.FlatTransaction(before), tx)
			require.Len(t, requests(), len(tt.results))
		})
	}
}

func TestAccountDeleteMissingDestinationStillChecksObligations(t *testing.T) {
	for _, omit := range []bool{true, false} {
		t.Run(fmt.Sprint("omit=", omit), func(t *testing.T) {
			cl, requests := newAccountDeleteTestClient(t, []map[string]any{
				{"account_objects": []any{}}, {"account_data": map[string]any{"SponsoringOwnerCount": 0}},
			})
			tx := deletionTransaction()
			if omit {
				delete(tx, "Destination")
			} else {
				tx["Destination"] = nil
			}
			before := clientinternal.CloneTransaction(tx)
			require.ErrorIs(t, cl.Autofill(&tx), errAccountHasSponsorshipObligations)
			require.Equal(t, transaction.FlatTransaction(before), tx)
			require.Len(t, requests(), 2)
		})
	}
}

func TestAccountDeleteUsesValidatedInfoAfterSequenceAutofill(t *testing.T) {
	cl, requests := newAccountDeleteTestClient(t, []map[string]any{
		{"account_data": map[string]any{"Sequence": 20}},
		{"account_objects": []any{}},
		{"account_data": map[string]any{"Sequence": 19, "SponsoringOwnerCount": 0}},
	})
	tx := deletionTransaction()
	delete(tx, "Sequence")
	before := clientinternal.CloneTransaction(tx)
	require.ErrorIs(t, cl.Autofill(&tx), errAccountHasSponsorshipObligations)
	require.Equal(t, transaction.FlatTransaction(before), tx)
	got := requests()
	require.Len(t, got, 3)
	require.Equal(t, "account_info", got[0]["command"])
	require.Equal(t, "current", got[0]["ledger_index"])
	require.Equal(t, "account_info", got[2]["command"])
	require.Equal(t, "validated", got[2]["ledger_index"])
}

func TestAccountDeleteMultisignedChecksSponsorship(t *testing.T) {
	cl, requests := newAccountDeleteTestClient(t, []map[string]any{
		{"account_objects": []any{}}, {"account_data": map[string]any{"Sponsor": deleteAccount}},
	})
	tx := deletionTransaction()
	before := clientinternal.CloneTransaction(tx)
	require.ErrorIs(t, cl.AutofillMultisigned(&tx, 2), errAccountDeleteSponsorMismatch)
	require.Equal(t, transaction.FlatTransaction(before), tx)
	require.Len(t, requests(), 2)
}

func uint32Ptr(v uint32) *uint32 { return &v }

// Each configured result corresponds to one request. Unexpected requests fail
// instead of silently returning an empty AccountRoot.
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

func TestAccountDeleteInfoCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	mc := &testutil.JSONRPCMockClient{}
	mc.DoFunc = func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return testutil.MockResponse(`{"result":{"account_objects":[]}}`, 200, mc)(req)
		}
		cancel()
		require.ErrorIs(t, req.Context().Err(), context.Canceled)
		return nil, req.Context().Err()
	}
	cfg, err := NewClientConfig("http://testnode/", WithHTTPClient(mc), WithNetworkIdentity(0, "1.12.0"))
	require.NoError(t, err)
	err = NewClient(cfg).checkAccountDeleteBlockers(ctx, types.Address(deleteAccount), deleteSponsor)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 2, calls)
}
