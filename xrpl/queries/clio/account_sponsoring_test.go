package clio

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/stretchr/testify/require"
)

func TestAccountSponsoringRequestMissingAccount(t *testing.T) {
	request := AccountSponsoringRequest{}
	require.ErrorIs(t, request.Validate(), account.ErrNoAccountID)
}

func TestAccountSponsoringRequest(t *testing.T) {
	for _, tt := range []struct {
		name    string
		request AccountSponsoringRequest
		want    string
	}{
		{"minimal", AccountSponsoringRequest{Account: "rAccount"}, `{"account":"rAccount"}`},
		{"complete", AccountSponsoringRequest{Account: "rAccount", Type: account.SponsorshipObject, DeletionBlockersOnly: true, LedgerHash: "ABC", LedgerIndex: common.LedgerIndex(123), Limit: 20, Marker: map[string]any{"cursor": "next"}}, `{"account":"rAccount","type":"sponsorship","deletion_blockers_only":true,"ledger_hash":"ABC","ledger_index":123,"limit":20,"marker":{"cursor":"next"}}`},
		{"named ledger", AccountSponsoringRequest{Account: "rAccount", LedgerIndex: common.LedgerTitle("validated"), Marker: "next"}, `{"account":"rAccount","ledger_index":"validated","marker":"next"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, "account_sponsoring", tt.request.Method())
			require.Equal(t, 2, tt.request.APIVersion())
			require.NoError(t, tt.request.Validate())
			data, err := json.Marshal(tt.request)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(data))
		})
	}
}

type accountSponsoringResponseFixture struct {
	name string
	want AccountSponsoringResponse
	json string
}

func accountSponsoringResponseFixtures() []accountSponsoringResponseFixture {
	return []accountSponsoringResponseFixture{
		{
			name: "minimal response without a marker",
			want: AccountSponsoringResponse{
				Account:          "rAccount",
				SponsoredObjects: []ledger.FlatLedgerObject{},
			},
			json: `{"account":"rAccount","sponsored_objects":[]}`,
		},
		{
			name: "complete response with an object marker",
			want: AccountSponsoringResponse{
				Account: "rAccount",
				SponsoredObjects: []ledger.FlatLedgerObject{
					{"LedgerEntryType": "Sponsorship", "Owner": "rAccount", "Sponsee": "rSponsee", "FeeAmount": "0"},
					{"LedgerEntryType": "FutureEntry", "FutureField": "retained"},
				},
				LedgerHash:         "ABC",
				LedgerIndex:        123,
				LedgerCurrentIndex: 124,
				Limit:              20,
				Marker:             map[string]any{"cursor": "next"},
				Validated:          true,
			},
			json: `{
				"account": "rAccount",
				"sponsored_objects": [
					{"LedgerEntryType":"Sponsorship","Owner":"rAccount","Sponsee":"rSponsee","FeeAmount":"0"},
					{"LedgerEntryType":"FutureEntry","FutureField":"retained"}
				],
				"ledger_hash": "ABC",
				"ledger_index": 123,
				"ledger_current_index": 124,
				"limit": 20,
				"marker": {"cursor":"next"},
				"validated": true
			}`,
		},
		{
			name: "string marker",
			want: AccountSponsoringResponse{
				Account:          "rAccount",
				SponsoredObjects: []ledger.FlatLedgerObject{},
				Marker:           "next",
			},
			json: `{"account":"rAccount","sponsored_objects":[],"marker":"next"}`,
		},
	}
}

func TestAccountSponsoringResponseSerialize(t *testing.T) {
	for _, tt := range accountSponsoringResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.want)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(encoded))
		})
	}
}

func TestAccountSponsoringResponseJSONDecode(t *testing.T) {
	for _, tt := range accountSponsoringResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var got AccountSponsoringResponse
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAccountSponsoringResponseClientDecode(t *testing.T) {
	for _, tt := range accountSponsoringResponseFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&data))
			var got AccountSponsoringResponse
			require.NoError(t, clientinternal.DecodeResultInto(data, &got))
			require.Equal(t, tt.want, got)
		})
	}
}
