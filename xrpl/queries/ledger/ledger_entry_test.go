package ledger

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	types "github.com/Peersyst/xrpl-go/xrpl/queries/ledger/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestLedgerEntryRequest(t *testing.T) {
	s := EntryRequest{
		MPTIssuance:    true,
		Binary:         true,
		Index:          "some_index",
		AccountRoot:    "some_account",
		Check:          "some_check",
		Did:            "some_did",
		PaymentChannel: "some_channel",
		NFTPage:        "some_nft_page",
		BridgeAccount:  "some_bridge_account",
		Bridge: types.EntryXChainBridge{
			LockingChainDoor:  "door",
			LockingChainIssue: "issue",
			IssuingChainDoor:  "door2",
			IssuingChainIssue: "issue2",
		},
		IncludeDeleted: true,
		AMM: types.EntryAssetPair{
			Asset: ledger.Asset{
				Currency: "XRP",
			},
			Asset2: ledger.Asset{
				Currency: "USD",
			},
		},
		RippleState: types.EntryRippleState{
			Accounts: []string{"acc1", "acc2"},
			Currency: "XRP",
		},
	}
	j := `{
	"mpt_issuance": true,
	"amm": {
		"asset": {
			"currency": "XRP"
		},
		"asset2": {
			"currency": "USD"
		}
	},
	"include_deleted": true,
	"binary": true,
	"index": "some_index",
	"account_root": "some_account",
	"check": "some_check",
	"did": "some_did",
	"payment_channel": "some_channel",
	"ripple_state": {
		"accounts": [
			"acc1",
			"acc2"
		],
		"currency": "XRP"
	},
	"nft_page": "some_nft_page",
	"bridge_account": "some_bridge_account",
	"bridge": {
		"locking_chain_door": "door",
		"locking_chain_issue": "issue",
		"issuing_chain_door": "door2",
		"issuing_chain_issue": "issue2"
	}
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
