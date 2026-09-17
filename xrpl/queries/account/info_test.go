package account_test

import (
	"encoding/json"
	"strings"
	"testing"

	clientinternal "github.com/Peersyst/xrpl-go/xrpl/internal/client"
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

type accountInfoSponsorshipFieldsFixture struct {
	name string
	want account.InfoResponse
	json string
}

func accountInfoSponsorshipFieldsFixtures() []accountInfoSponsorshipFieldsFixture {
	zero, one, two, maximum := uint32(0), uint32(1), uint32(2), uint32(4294967295)
	fixtures := []struct {
		name                            string
		fields                          string
		sponsored, sponsoring, accounts *uint32
	}{
		{name: "absent"},
		{name: "sponsored zero", fields: `,"SponsoredOwnerCount":0`, sponsored: &zero},
		{name: "sponsoring zero", fields: `,"SponsoringOwnerCount":0`, sponsoring: &zero},
		{name: "accounts zero", fields: `,"SponsoringAccountCount":0`, accounts: &zero},
		{
			name:       "all nonzero counts",
			fields:     `,"SponsoredOwnerCount":1,"SponsoringOwnerCount":2,"SponsoringAccountCount":4294967295`,
			sponsored:  &one,
			sponsoring: &two,
			accounts:   &maximum,
		},
		{
			name:       "all counts",
			fields:     `,"SponsoredOwnerCount":1,"SponsoringOwnerCount":4294967295,"SponsoringAccountCount":0`,
			sponsored:  &one,
			sponsoring: &maximum,
			accounts:   &zero,
		},
	}
	result := make([]accountInfoSponsorshipFieldsFixture, 0, len(fixtures))
	for _, tt := range fixtures {
		want := account.InfoResponse{
			AccountData: ledger.AccountRoot{
				Account:                "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				Sponsor:                "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				Balance:                types.XRPCurrencyAmount(1000000),
				LedgerEntryType:        ledger.AccountRootEntry,
				SponsoredOwnerCount:    tt.sponsored,
				SponsoringOwnerCount:   tt.sponsoring,
				SponsoringAccountCount: tt.accounts,
			},
			Validated: true,
		}
		payload := `{
	"account_data": {
		"Account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		"Sponsor": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
		"Balance": "1000000",
		"LedgerEntryType": "AccountRoot",
		"Flags": 0,
		"OwnerCount": 0,
		"PreviousTxnID": "",
		"PreviousTxnLgrSeq": 0,
		"Sequence": 0` + tt.fields + `
	},
	"validated": true
}`
		result = append(result, accountInfoSponsorshipFieldsFixture{name: tt.name, want: want, json: payload})
	}
	return result
}

func TestAccountInfoSponsorshipFieldsSerialize(t *testing.T) {
	for _, tt := range accountInfoSponsorshipFieldsFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.want)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(encoded))
		})
	}
}

func TestAccountInfoSponsorshipFieldsJSONDecode(t *testing.T) {
	for _, tt := range accountInfoSponsorshipFieldsFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var got account.InfoResponse
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAccountInfoSponsorshipFieldsClientDecode(t *testing.T) {
	for _, tt := range accountInfoSponsorshipFieldsFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&data))
			var got account.InfoResponse
			require.NoError(t, clientinternal.DecodeResultInto(data, &got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAccountInfoRequest(t *testing.T) {
	s := account.InfoRequest{
		Account:     "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
		LedgerIndex: common.Closed,
		Queue:       true,
		SignerLists: false,
		Strict:      true,
	}

	// SignerLists assigned to default, omitted due to omitempty
	j := `{
	"account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
	"ledger_index": "closed",
	"queue": true,
	"strict": true
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

type accountInfoResponsePseudoAccountLinksFixture struct {
	name string
	want account.InfoResponse
	json string
}

func accountInfoResponsePseudoAccountLinksFixtures() []accountInfoResponsePseudoAccountLinksFixture {
	const id = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
	fixtures := []struct {
		name                         string
		ammID, vaultID, loanBrokerID types.Hash256
		linkJSON                     string
	}{
		{name: "ordinary account"},
		{name: "AMM account", ammID: id, linkJSON: `,"AMMID":"` + id + `"`},
		{name: "vault account", vaultID: id, linkJSON: `,"VaultID":"` + id + `"`},
		{name: "loan broker account", loanBrokerID: id, linkJSON: `,"LoanBrokerID":"` + id + `"`},
	}
	result := make([]accountInfoResponsePseudoAccountLinksFixture, 0, len(fixtures))
	for _, tt := range fixtures {
		s := account.InfoResponse{
			AccountData: ledger.AccountRoot{
				Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				LedgerEntryType: ledger.AccountRootEntry,
				AMMID:           tt.ammID,
				VaultID:         tt.vaultID,
				LoanBrokerID:    tt.loanBrokerID,
			},
			Validated: true,
		}
		j := `{"account_data":{"Account":"rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD","LedgerEntryType":"AccountRoot","Flags":0,"OwnerCount":0,"PreviousTxnID":"","PreviousTxnLgrSeq":0,"Sequence":0` + tt.linkJSON + `},"validated":true}`

		result = append(result, accountInfoResponsePseudoAccountLinksFixture{name: tt.name, want: s, json: j})
	}
	return result
}

func TestAccountInfoResponsePseudoAccountLinksSerialize(t *testing.T) {
	for _, tt := range accountInfoResponsePseudoAccountLinksFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.want)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(encoded))
		})
	}
}

func TestAccountInfoResponsePseudoAccountLinksJSONDecode(t *testing.T) {
	for _, tt := range accountInfoResponsePseudoAccountLinksFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var got account.InfoResponse
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&got))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAccountInfoResponsePseudoAccountLinksClientDecode(t *testing.T) {
	for _, tt := range accountInfoResponsePseudoAccountLinksFixtures() {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]any
			decoder := json.NewDecoder(strings.NewReader(tt.json))
			decoder.UseNumber()
			require.NoError(t, decoder.Decode(&data))
			var got account.InfoResponse
			require.NoError(t, clientinternal.DecodeResultInto(data, &got))
			require.Equal(t, tt.want, got)
		})
	}
}

func accountInfoResponseFixture() (account.InfoResponse, string) {
	s := account.InfoResponse{
		AccountData: ledger.AccountRoot{
			Account:           "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
			Balance:           types.XRPCurrencyAmount(999999999960),
			Flags:             8388608,
			LedgerEntryType:   ledger.AccountRootEntry,
			OwnerCount:        0,
			PreviousTxnID:     "4294BEBE5B569A18C0A2702387C9B1E7146DC3A5850C1E87204951C6FDAA4C42",
			PreviousTxnLgrSeq: 3,
			Sequence:          6,
		},
		LedgerCurrentIndex: 4,
		QueueData: accounttypes.QueueData{
			TxnCount:           5,
			AuthChangeQueued:   true,
			LowestSequence:     6,
			HighestSequence:    10,
			MaxSpendDropsTotal: types.XRPCurrencyAmount(500),
			Transactions: []accounttypes.QueueTransaction{
				{
					AuthChange:    false,
					Fee:           types.XRPCurrencyAmount(100),
					FeeLevel:      types.XRPCurrencyAmount(2560),
					MaxSpendDrops: types.XRPCurrencyAmount(100),
					Seq:           6,
				},
				{
					AuthChange:    true,
					Fee:           types.XRPCurrencyAmount(100),
					FeeLevel:      types.XRPCurrencyAmount(2560),
					MaxSpendDrops: types.XRPCurrencyAmount(100),
					Seq:           10,
				},
			},
		},
		Validated: false,
	}
	j := `{
	"account_data": {
		"Flags": 8388608,
		"LedgerEntryType": "AccountRoot",
		"Account": "rG1QQv2nh2gr7RCZ1P8YYcBUKCCN633jCn",
		"Balance": "999999999960",
		"OwnerCount": 0,
		"PreviousTxnID": "4294BEBE5B569A18C0A2702387C9B1E7146DC3A5850C1E87204951C6FDAA4C42",
		"PreviousTxnLgrSeq": 3,
		"Sequence": 6
	},
	"ledger_current_index": 4,
	"queue_data": {
		"txn_count": 5,
		"auth_change_queued": true,
		"lowest_sequence": 6,
		"highest_sequence": 10,
		"max_spend_drops_total": "500",
		"transactions": [
			{
				"auth_change": false,
				"fee": "100",
				"fee_level": "2560",
				"max_spend_drops": "100",
				"seq": 6
			},
			{
				"auth_change": true,
				"fee": "100",
				"fee_level": "2560",
				"max_spend_drops": "100",
				"seq": 10
			}
		]
	},
	"validated": false
}`
	return s, j
}

func TestAccountInfoResponseSerialize(t *testing.T) {
	value, payload := accountInfoResponseFixture()
	require.NoError(t, testutil.Serialize(t, value, payload))
}

func TestAccountInfoResponseJSONDecode(t *testing.T) {
	want, payload := accountInfoResponseFixture()
	var got account.InfoResponse
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&got))
	require.Equal(t, want, got)
}

func TestAccountInfoResponseClientDecode(t *testing.T) {
	want, payload := accountInfoResponseFixture()
	var data map[string]any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&data))
	var got account.InfoResponse
	require.NoError(t, clientinternal.DecodeResultInto(data, &got))
	require.Equal(t, want, got)
}
