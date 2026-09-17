package client

import (
	"context"
	"encoding/json"
	"testing"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
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
		{name: "owner count zero", root: map[string]any{"SponsoringOwnerCount": 0}, wantErr: ErrAccountHasSponsorshipObligations},
		{name: "owner count nonzero", root: map[string]any{"SponsoringOwnerCount": 2}, wantErr: ErrAccountHasSponsorshipObligations},
		{name: "account count zero", root: map[string]any{"SponsoringAccountCount": 0}, wantErr: ErrAccountHasSponsorshipObligations},
		{name: "account count nonzero", root: map[string]any{"SponsoringAccountCount": 2}, wantErr: ErrAccountHasSponsorshipObligations},
		{name: "sponsored owner count zero", root: map[string]any{"SponsoredOwnerCount": 0}},
		{name: "sponsored owner count nonzero", root: map[string]any{"SponsoredOwnerCount": 2}},
		{name: "matching sponsor", root: map[string]any{"Sponsor": deleteSponsor}, destination: deleteSponsor},
		{name: "equivalent X-address", root: map[string]any{"Sponsor": deleteSponsor}, destination: x},
		{name: "equivalent tagged testnet X-address", root: map[string]any{"Sponsor": deleteSponsor}, destination: tagged},
		{name: "sponsor mismatch", root: map[string]any{"Sponsor": deleteSponsor}, destination: deleteAccount, wantErr: ErrAccountDeleteSponsorMismatch},
		{name: "missing destination", root: map[string]any{"Sponsor": deleteSponsor}},
		{name: "missing destination still checks counters", root: map[string]any{"Sponsor": deleteSponsor, "SponsoringOwnerCount": 0}, wantErr: ErrAccountHasSponsorshipObligations},
		{name: "malformed sponsor", root: map[string]any{"Sponsor": "invalid"}, destination: deleteSponsor, wantErr: addresscodec.ErrInvalidAddressFormat, wantContext: "decode account Sponsor:"},
		{name: "malformed destination", root: map[string]any{"Sponsor": deleteSponsor}, destination: "invalid", wantErr: addresscodec.ErrInvalidAddressFormat, wantContext: "decode AccountDelete Destination:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, methods := accountDeleteRequests(t, []map[string]any{
				{"account_objects": []any{}}, {"account_data": tt.root},
			})
			err := CheckAccountDeleteBlockers(context.Background(), request, types.Address(deleteAccount), tt.destination)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				if tt.wantContext != "" {
					require.ErrorContains(t, err, tt.wantContext)
				}
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, []string{"account_objects", "account_info"}, methods())
		})
	}
}

func TestAccountDeleteObjectsStillBlock(t *testing.T) {
	for _, kind := range []string{"Escrow", "Sponsorship"} {
		t.Run(kind, func(t *testing.T) {
			request, methods := accountDeleteRequests(t, []map[string]any{
				{"account_objects": []any{map[string]any{"LedgerEntryType": kind, "Sponsor": deleteSponsor, "Sponsee": deleteAccount}}},
			})
			err := CheckAccountDeleteBlockers(context.Background(), request, types.Address(deleteAccount), deleteSponsor)
			require.ErrorIs(t, err, ErrAccountCannotBeDeleted)
			require.Equal(t, []string{"account_objects"}, methods())
		})
	}
}

func TestAccountDeleteRequestErrors(t *testing.T) {
	for _, method := range []string{"account_objects", "account_info"} {
		for _, wantErr := range []error{errRequest, context.Canceled} {
			t.Run(method+"/"+wantErr.Error(), func(t *testing.T) {
				ctx := t.Context()
				var methods []string
				request := func(gotCtx context.Context, req Request, _ any) error {
					require.Equal(t, ctx, gotCtx)
					methods = append(methods, req.Method())
					if req.Method() == method {
						return wantErr
					}
					return nil
				}
				err := CheckAccountDeleteBlockers(ctx, request, types.Address(deleteAccount), deleteSponsor)
				require.ErrorIs(t, err, wantErr)
				expected := []string{"account_objects"}
				if method == "account_info" {
					expected = append(expected, "account_info")
				}
				require.Equal(t, expected, methods)
			})
		}
	}
}

func accountDeleteRequests(t *testing.T, results []map[string]any) (RequestResultFunc, func() []string) {
	t.Helper()
	var methods []string
	request := func(_ context.Context, req Request, result any) error {
		switch req := req.(type) {
		case *account.ObjectsRequest:
			require.Equal(t, &account.ObjectsRequest{
				Account:              types.Address(deleteAccount),
				LedgerIndex:          common.LedgerTitle("validated"),
				DeletionBlockersOnly: true,
			}, req)
		case *account.InfoRequest:
			require.Equal(t, &account.InfoRequest{
				Account:     types.Address(deleteAccount),
				LedgerIndex: common.LedgerTitle("validated"),
			}, req)
		default:
			t.Fatalf("unexpected request %T", req)
		}
		index := len(methods)
		methods = append(methods, req.Method())
		require.Less(t, index, len(results), "unexpected request: %s", req.Method())
		body, err := json.Marshal(results[index])
		require.NoError(t, err)
		return json.Unmarshal(body, result)
	}
	return request, func() []string { return methods }
}
