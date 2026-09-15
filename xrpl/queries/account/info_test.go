package account

import (
	"encoding/json"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestAccountInfoRequest(t *testing.T) {
	s := InfoRequest{
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

func TestAccountInfoResponsePseudoAccountLinks(t *testing.T) {
	const id = "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
	tests := []struct {
		name                         string
		ammID, vaultID, loanBrokerID types.Hash256
		linkJSON                     string
	}{
		{name: "ordinary account"},
		{name: "AMM account", ammID: id, linkJSON: `,"AMMID":"` + id + `"`},
		{name: "vault account", vaultID: id, linkJSON: `,"VaultID":"` + id + `"`},
		{name: "loan broker account", loanBrokerID: id, linkJSON: `,"LoanBrokerID":"` + id + `"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := InfoResponse{
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
			encoded, err := json.Marshal(s)
			require.NoError(t, err)
			require.JSONEq(t, j, string(encoded))
			var decoded InfoResponse
			require.NoError(t, json.Unmarshal([]byte(j), &decoded))
			require.Equal(t, s, decoded)
		})
	}
}

func TestAccountInfoResponse(t *testing.T) {
	s := InfoResponse{
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
	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}
