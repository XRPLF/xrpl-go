package rpc

import (
	"testing"

	ledgerentry "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/clio"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/stretchr/testify/require"
)

const sponsoringAccount = "rf1BiGeXwwQoi8Z2ueFYTEXSwuJYfV2Jpn"

func TestGetAccountSponsoring(t *testing.T) {
	cl, request := setupQueryTestClient(t, map[string]any{
		"account": sponsoringAccount,
		"sponsored_objects": []any{
			map[string]any{"LedgerEntryType": "Sponsorship"},
		},
	})

	result, err := cl.GetAccountSponsoring(&clio.AccountSponsoringRequest{Account: sponsoringAccount})
	require.NoError(t, err)
	require.Equal(t, &clio.AccountSponsoringResponse{
		Account:          sponsoringAccount,
		SponsoredObjects: []ledgerentry.FlatLedgerObject{{"LedgerEntryType": "Sponsorship"}},
	}, result)
	got := request()
	require.Equal(t, "account_sponsoring", got["command"])
	require.Equal(t, sponsoringAccount, got["account"])
}

func TestGetAccountObjectsSponsoredFalse(t *testing.T) {
	cl, request := setupQueryTestClient(t, map[string]any{"account": sponsoringAccount, "account_objects": []any{}})
	no := false
	_, err := cl.GetAccountObjects(&account.ObjectsRequest{
		Account: sponsoringAccount, Type: account.SponsorshipObject, Sponsored: &no,
	})
	require.NoError(t, err)
	got := request()
	require.Equal(t, "account_objects", got["command"])
	require.Contains(t, got, "sponsored")
	require.Equal(t, false, got["sponsored"])
	require.Equal(t, "sponsorship", got["type"])
}

func TestGetLedgerEntrySponsorship(t *testing.T) {
	const index = "7DB0788C020F02780A673DC74757F23823FA3014C1866E72CC4CD8B226CD6EF4"
	for _, tt := range []struct {
		name     string
		selector ledgerquery.SponsorshipSelector
		want     any
	}{
		{
			name:     "index",
			selector: ledgerquery.SponsorshipSelector{Index: index},
			want:     index,
		},
		{
			name: "pair",
			selector: ledgerquery.SponsorshipSelector{Object: &ledgerquery.SponsorshipSelectorFields{
				Sponsor: sponsoringAccount, Sponsee: sponsoringAccount,
			}},
			want: map[string]any{"sponsor": sponsoringAccount, "sponsee": sponsoringAccount},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cl, request := setupQueryTestClient(t, map[string]any{
				"node": map[string]any{"LedgerEntryType": "Sponsorship", "Owner": sponsoringAccount},
			})
			result, err := cl.GetLedgerEntry(&ledgerquery.EntryRequest{Sponsorship: tt.selector})
			require.NoError(t, err)
			require.Equal(t, ledgerentry.SponsorshipEntry, result.Node.EntryType())
			got := request()
			require.Equal(t, "ledger_entry", got["command"])
			require.Equal(t, tt.want, got["sponsorship"])
			require.NotContains(t, got, "amm")
		})
	}
}
